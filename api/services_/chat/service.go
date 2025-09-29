package chat

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"mime"
	"mime/multipart"
	"path/filepath"
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
	isEncrypted bool,
	msgEnvelope *CryptoEnvelopeV1,
	fileBase64 string,
	fileHeader *multipart.FileHeader,
	attEnvelope *CryptoEnvelopeV1,
) (*Message, error) {
	if fileBase64 == "" || fileHeader == nil {
		return nil, errors.New("fileBase64 and fileHeader required")
	}
	if fileHeader.Size == 0 {
		return nil, errors.New("empty file")
	}

	var data []byte
	var err error
	if isEncrypted {
		if attEnvelope == nil || attEnvelope.CT == "" {
			return nil, errors.New("missing attEnvelope.CT for encrypted attachment")
		}
		log.Printf("attEnvelope.CT length: %d, sample (first 100 chars): %s", len(attEnvelope.CT), attEnvelope.CT[:min(len(attEnvelope.CT), 100)])
		data, err = base64.URLEncoding.DecodeString(attEnvelope.CT)
		if err != nil {
			log.Printf("Failed to decode attEnvelope.CT: %v", err)
			data, err = base64.StdEncoding.DecodeString(fileBase64)
			if err != nil {
				return nil, fmt.Errorf("decode fileBase64 fallback: %w", err)
			}
			log.Printf("Using fallback fileBase64")
		}
	} else {
		data, err = base64.StdEncoding.DecodeString(fileBase64)
		if err != nil {
			return nil, fmt.Errorf("decode fileBase64: %w", err)
		}
	}

	log.Printf("File data (first 16 bytes): %x", data[:min(len(data), 16)])

	declared := fileHeader.Header.Get("Content-Type")
	if declared == "" || declared == "application/octet-stream" {
		if ext := filepath.Ext(fileHeader.Filename); ext != "" {
			if mt := mime.TypeByExtension(ext); mt != "" {
				declared = mt
			}
		}
		if declared == "" {
			declared = "application/octet-stream"
		}
	}

	ipfsResult, err := s.ipfsUploader.UploadBytes(ctx, data, fileHeader.Filename)
	if err != nil {
		return nil, fmt.Errorf("IPFS upload failed: %w", err)
	}
	log.Printf("IPFS CID: %s, URL: %s", ipfsResult.Hash, ipfsResult.URL)

	att := &Attachment{
		ID:          primitive.NewObjectID().Hex(),
		FileName:    ipfsResult.FileName,
		FileType:    declared,
		FileSize:    fileHeader.Size,
		URL:         ipfsResult.URL,
		Hash:        ipfsResult.Hash,
		IsEncrypted: isEncrypted,
		Envelope:    attEnvelope,
	}
	log.Printf("Attachment before normalize: %+v", att)

	now := time.Now().UTC()
	msg := &Message{
		GroupID:     groupID,
		SenderEmail: senderEmail,
		SenderName:  senderName,
		MessageType: "file",
		Attachments: []*Attachment{att},
		IsEncrypted: isEncrypted,
		Envelope:    msgEnvelope,
		Content:     content, // Preserve content as caption, even for encrypted messages
		CreatedAt:   now,
		UpdatedAt:   now,
		IsDeleted:   false,
		Status:      MessageStatus{Sent: now},
	}

	log.Printf("Message before normalize: %+v", msg)
	NormalizeMessageEncryption(msg)
	log.Printf("Message after normalize: %+v", msg)

	if err := s.repo.CreateMessage(ctx, msg); err != nil {
		return nil, fmt.Errorf("store message: %w", err)
	}

	err = s.wsManager.BroadcastToGroup(groupID.Hex(), WebSocketMessage{
		Type:      MessageType(EventNewMessage),
		GroupID:   groupID.Hex(),
		Payload:   msg,
		Timestamp: now,
		UserEmail: senderEmail,
	})
	if err != nil {
		log.Printf("Failed to broadcast message: %v", err)
	}

	return msg, nil
}

// Helper function to avoid index out of range
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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

var MessageCollection *mongo.Collection // Inject this from main.go or init

func SaveMessageToDB(payload NewMessagePayload) error {
	if MessageCollection == nil {
		return fmt.Errorf("MessageCollection not initialized")
	}

	createdAt, err := time.Parse(time.RFC3339, payload.Timestamp)
	if err != nil {
		return err
	}

	gid, err := primitive.ObjectIDFromHex(payload.GroupID)
	if err != nil {
		return fmt.Errorf("invalid group ID: %w", err)
	}

	msg := Message{
		GroupID:     gid,
		SenderEmail: payload.SenderEmail,
		SenderName:  payload.SenderName,
		MessageType: "text",
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
		IsDeleted:   false,
		Status:      MessageStatus{Sent: createdAt},
		IsEncrypted: payload.IsEncrypted,
		Envelope:    payload.Envelope,
	}

	if len(payload.Attachments) > 0 {
		msg.Attachments = payload.Attachments
		msg.MessageType = "file"
	}

	if !payload.IsEncrypted {
		msg.Content = payload.Text
	} else {
		msg.Content = "" // ensure no plaintext
	}

	_, err = MessageCollection.InsertOne(context.Background(), msg)
	return err
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

// In services_/chat/chat_service.go
func (s *ChatService) HandleMessage(
	ctx context.Context,
	senderEmail, senderName string,
	content string,
	fileName string,
	contentType string, // <-- from client; don't infer from ciphertext
	fileBytes []byte, // ciphertext if E2EE
	groupID primitive.ObjectID,
	isEncrypted bool,
	msgEnvelope *CryptoEnvelopeV1,
	attEnvelope *CryptoEnvelopeV1,
) error {
	var att *Attachment
	if len(fileBytes) > 0 && fileName != "" {
		declared := contentType
		if declared == "" {
			declared = "application/octet-stream"
		}
		ipfsResult, err := s.ipfsUploader.UploadBytes(ctx, fileBytes, fileName)
		if err != nil {
			return fmt.Errorf("IPFS upload failed: %w", err)
		}
		att = &Attachment{
			ID:          primitive.NewObjectID().Hex(),
			FileName:    fileName,
			FileType:    declared,
			FileSize:    int64(len(fileBytes)),
			URL:         ipfsResult.URL,
			Hash:        ipfsResult.Hash,
			IsEncrypted: isEncrypted || msgEnvelope != nil || attEnvelope != nil,
			Envelope:    attEnvelope,
		}
	}

	now := time.Now().UTC()
	msg := &Message{
		GroupID:     groupID,
		SenderEmail: senderEmail,
		SenderName:  senderName,
		MessageType: ifThen(att != nil, "file", "text"),
		IsEncrypted: isEncrypted,
		Envelope:    msgEnvelope,
		CreatedAt:   now,
		UpdatedAt:   now,
		IsDeleted:   false,
		Status:      MessageStatus{Sent: now},
	}
	if att != nil {
		msg.Attachments = []*Attachment{att}
	}
	if isEncrypted {
		msg.Content = ""
	} else {
		msg.Content = content
	}

	// ✅ enforce flags consistently
	NormalizeMessageEncryption(msg)

	if err := s.repo.CreateMessage(ctx, msg); err != nil {
		return fmt.Errorf("store message: %w", err)
	}

	_ = s.wsManager.BroadcastToGroup(groupID.Hex(), WebSocketMessage{
		Type:      MessageType(EventNewMessage),
		GroupID:   groupID.Hex(),
		Payload:   msg,
		Timestamp: now,
		UserEmail: senderEmail,
	})
	return nil
}

func ifThen[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
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
