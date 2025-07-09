package chat

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const UsersCollection = "users"

// userService implements the UserService interface
type userService struct {
	db *mongo.Database
}
type ChatService struct {
	repo         ChatRepository
	ipfsUploader IPFSUploader
	wsManager    WebSocketManager
}

func NewChatService(repo ChatRepository, uploader IPFSUploader, ws WebSocketManager) *ChatService {
	return &ChatService{
		repo:         repo,
		ipfsUploader: uploader,
		wsManager:    ws,
	}
}
func (s *ChatService) SendMessageWithAttachment(
	ctx context.Context,
	senderEmail, senderName string,
	groupID primitive.ObjectID,
	content string,
	file multipart.File,
	fileHeader *multipart.FileHeader,
) error {
	if fileHeader.Size == 0 {
		return errors.New("empty file")
	}
	if file == nil {
		return errors.New("no file provided")
	}
	if fileHeader == nil || fileHeader.Size == 0 {
		return errors.New("empty or missing file")
	}

	ipfsResult, err := s.ipfsUploader.UploadFile(ctx, file, fileHeader.Filename)

	if err != nil {
		return fmt.Errorf("IPFS upload failed: %w", err)
	}

	attachment := &Attachment{
		ID:       primitive.NewObjectID().Hex(),
		FileName: ipfsResult.FileName,
		FileType: fileHeader.Header.Get("Content-Type"),
		FileSize: ipfsResult.Size,
		URL:      ipfsResult.URL,
		Hash:     ipfsResult.Hash,
	}

	message := &Message{
		GroupID:     groupID,
		SenderEmail: senderEmail,
		SenderName:  senderName,
		Content:     content,
		MessageType: "file",
		Attachments: []*Attachment{attachment},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		IsDeleted:   false,
		Status: MessageStatus{
			Sent: time.Now(),
		},
	}

	if err := s.repo.CreateMessage(ctx, message); err != nil {
		return fmt.Errorf("failed to store message: %w", err)
	}

	_ = s.wsManager.BroadcastToGroup(groupID.Hex(), message)

	return nil
}

// NewUserService creates a new user service instance
func NewUserService(db *mongo.Database) UserService {
	service := &userService{db: db}

	// Create indexes for better performance
	service.createIndexes()

	return service
}

// createIndexes creates necessary indexes for the users collection
func (s *userService) createIndexes() {
	ctx := context.Background()
	collection := s.db.Collection(UsersCollection)

	// Email index (unique)
	_, _ = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	// Status index for filtering online users
	_, _ = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "status", Value: 1}},
	})

	// Last seen index for sorting
	_, _ = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "last_seen", Value: -1}},
	})
}

// GetUserByEmail retrieves a user by email address
func (s *userService) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	collection := s.db.Collection(UsersCollection)

	var user User
	err := collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user with email %s not found", email)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetUsersByEmails retrieves multiple users by their email addresses
func (s *userService) GetUsersByEmails(ctx context.Context, emails []string) ([]*User, error) {
	if len(emails) == 0 {
		return []*User{}, nil
	}

	collection := s.db.Collection(UsersCollection)

	filter := bson.M{"email": bson.M{"$in": emails}}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	defer cursor.Close(ctx)

	var users []*User
	for cursor.Next(ctx) {
		var user User
		if err := cursor.Decode(&user); err != nil {
			continue // Skip invalid documents
		}
		users = append(users, &user)
	}

	return users, nil
}

// ValidateUserExists checks if a user exists by email
func (s *userService) ValidateUserExists(ctx context.Context, email string) (bool, error) {
	collection := s.db.Collection(UsersCollection)

	count, err := collection.CountDocuments(ctx, bson.M{"email": email})
	if err != nil {
		return false, fmt.Errorf("failed to validate user existence: %w", err)
	}

	return count > 0, nil
}

// Additional helper methods for user management

// CreateUser creates a new user
func (s *userService) CreateUser(ctx context.Context, user *User) error {
	collection := s.db.Collection(UsersCollection)

	user.Status = "offline"
	now := time.Now()
	user.LastSeen = &now

	result, err := collection.InsertOne(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		user.ID = oid.Hex()
	}

	return nil
}

// UpdateUserStatus updates a user's online/offline status
func (s *userService) UpdateUserStatus(ctx context.Context, email string, status string) error {
	collection := s.db.Collection(UsersCollection)

	update := bson.M{
		"$set": bson.M{
			"status": status,
		},
	}

	if status == "offline" {
		update["$set"].(bson.M)["last_seen"] = time.Now()
	}

	filter := bson.M{"email": email}
	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}

	return nil
}

// GetOnlineUsers retrieves all currently online users
func (s *userService) GetOnlineUsers(ctx context.Context) ([]*User, error) {
	collection := s.db.Collection(UsersCollection)

	filter := bson.M{"status": "online"}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get online users: %w", err)
	}
	defer cursor.Close(ctx)

	var users []*User
	for cursor.Next(ctx) {
		var user User
		if err := cursor.Decode(&user); err != nil {
			continue
		}
		users = append(users, &user)
	}

	return users, nil
}

// SearchUsers searches for users by name or email
func (s *userService) SearchUsers(ctx context.Context, query string, limit int) ([]*User, error) {
	collection := s.db.Collection(UsersCollection)

	// Create a case-insensitive regex search
	filter := bson.M{
		"$or": []bson.M{
			{"full_name": bson.M{"$regex": query, "$options": "i"}},
			{"email": bson.M{"$regex": query, "$options": "i"}},
		},
	}

	opts := options.Find().SetLimit(int64(limit))
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}
	defer cursor.Close(ctx)

	var users []*User
	for cursor.Next(ctx) {
		var user User
		if err := cursor.Decode(&user); err != nil {
			continue
		}
		users = append(users, &user)
	}

	return users, nil
}

// UpdateUserProfile updates user profile information
func (s *userService) UpdateUserProfile(ctx context.Context, email string, updates map[string]interface{}) error {
	collection := s.db.Collection(UsersCollection)

	// Remove email from updates to prevent changing it
	delete(updates, "email")
	delete(updates, "_id")

	if len(updates) == 0 {
		return fmt.Errorf("no valid fields to update")
	}

	filter := bson.M{"email": email}
	update := bson.M{"$set": updates}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update user profile: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// GetUserStats returns user statistics
func (s *userService) GetUserStats(ctx context.Context) (map[string]interface{}, error) {
	collection := s.db.Collection(UsersCollection)

	// Count total users
	totalUsers, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to count total users: %w", err)
	}

	// Count online users
	onlineUsers, err := collection.CountDocuments(ctx, bson.M{"status": "online"})
	if err != nil {
		return nil, fmt.Errorf("failed to count online users: %w", err)
	}

	// Count users active in last 24 hours
	last24h := time.Now().Add(-24 * time.Hour)
	activeUsers, err := collection.CountDocuments(ctx, bson.M{
		"last_seen": bson.M{"$gte": last24h},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to count active users: %w", err)
	}

	return map[string]interface{}{
		"total_users":   totalUsers,
		"online_users":  onlineUsers,
		"active_24h":    activeUsers,
		"offline_users": totalUsers - onlineUsers,
	}, nil
}

func (s *ChatService) Repo() ChatRepository {
	return s.repo
}

func (s *ChatService) IPFSUploader() IPFSUploader {
	return s.ipfsUploader
}

func (s *ChatService) WsManager() WebSocketManager {
	return s.wsManager
}
