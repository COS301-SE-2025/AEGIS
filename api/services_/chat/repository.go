package chat

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	//	"aegis-api/services_/chat/ipfs_uploader"
)

const (
	GroupsCollection   = "chat_groups"
	MessagesCollection = "chat_messages"
)

type MongoRepository struct {
	db *mongo.Database
}

// NewChatRepository creates a new MongoDB chat repository
func NewChatRepository(db *mongo.Database) ChatRepository {
	repo := &MongoRepository{db: db}

	// Create indexes
	repo.createIndexes()

	return repo
}

// createIndexes creates necessary MongoDB indexes
func (r *MongoRepository) createIndexes() {
	ctx := context.Background()

	// Groups collection indexes
	groupsCollection := r.db.Collection(GroupsCollection)
	_, _ = groupsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "members.user_email", Value: 1}},
	})
	_, _ = groupsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "created_by", Value: 1}},
	})
	_, _ = groupsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "is_active", Value: 1}},
	})

	// Messages collection indexes
	messagesCollection := r.db.Collection(MessagesCollection)
	_, _ = messagesCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "group_id", Value: 1}, {Key: "created_at", Value: -1}},
	})
	_, _ = messagesCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "group_id", Value: 1}, {Key: "is_deleted", Value: 1}},
	})
	_, _ = messagesCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "sender_email", Value: 1}},
	})
	_, _ = messagesCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "status.read_by.user_email", Value: 1}},
	})
	// Text index for message search
	_, _ = messagesCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "content", Value: "text"}},
	})
}

// Group operations
func (r *MongoRepository) CreateGroup(ctx context.Context, group *ChatGroup) error {
	collection := r.db.Collection(GroupsCollection)

	group.CreatedAt = time.Now()
	group.UpdatedAt = time.Now()
	group.IsActive = true

	result, err := collection.InsertOne(ctx, group)
	if err != nil {
		return fmt.Errorf("failed to create group: %w", err)
	}

	group.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *MongoRepository) GetGroupByID(ctx context.Context, groupID primitive.ObjectID) (*ChatGroup, error) {
	collection := r.db.Collection(GroupsCollection)

	var group ChatGroup
	err := collection.FindOne(ctx, bson.M{
		"_id":       groupID,
		"is_active": true,
	}).Decode(&group)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("group not found")
		}
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	return &group, nil
}

func (r *MongoRepository) GetUserGroups(ctx context.Context, userEmail string) ([]*ChatGroup, error) {
	collection := r.db.Collection(GroupsCollection)

	filter := bson.M{
		"members.user_email": userEmail,
		"members.is_active":  true,
		"is_active":          true,
	}

	opts := options.Find().SetSort(bson.D{{Key: "updated_at", Value: -1}})
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get user groups: %w", err)
	}
	defer cursor.Close(ctx)

	var groups []*ChatGroup
	for cursor.Next(ctx) {
		var group ChatGroup
		if err := cursor.Decode(&group); err != nil {
			continue
		}
		groups = append(groups, &group)
	}

	return groups, nil
}

func (r *MongoRepository) UpdateGroup(ctx context.Context, group *ChatGroup) error {
	collection := r.db.Collection(GroupsCollection)

	group.UpdatedAt = time.Now()

	filter := bson.M{"_id": group.ID}
	update := bson.M{"$set": group}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update group: %w", err)
	}

	return nil
}

func (r *MongoRepository) DeleteGroup(ctx context.Context, groupID primitive.ObjectID) error {
	collection := r.db.Collection(GroupsCollection)

	filter := bson.M{"_id": groupID}
	update := bson.M{
		"$set": bson.M{
			"is_active":  false,
			"updated_at": time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}

	return nil
}

func (r *MongoRepository) AddMemberToGroup(ctx context.Context, groupID primitive.ObjectID, member *Member) error {
	collection := r.db.Collection(GroupsCollection)

	member.JoinedAt = time.Now()
	member.IsActive = true

	filter := bson.M{"_id": groupID}
	update := bson.M{
		"$push": bson.M{"members": member},
		"$set":  bson.M{"updated_at": time.Now()},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to add member to group: %w", err)
	}

	return nil
}

func (r *MongoRepository) RemoveMemberFromGroup(ctx context.Context, groupID primitive.ObjectID, userEmail string) error {
	collection := r.db.Collection(GroupsCollection)

	filter := bson.M{
		"_id":                groupID,
		"members.user_email": userEmail,
	}
	update := bson.M{
		"$set": bson.M{
			"members.$.is_active": false,
			"updated_at":          time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to remove member from group: %w", err)
	}

	return nil
}

func (r *MongoRepository) IsUserInGroup(ctx context.Context, groupID primitive.ObjectID, userEmail string) (bool, error) {
	collection := r.db.Collection(GroupsCollection)

	filter := bson.M{
		"_id":                groupID,
		"members.user_email": userEmail,
		"members.is_active":  true,
		"is_active":          true,
	}

	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("failed to check user membership: %w", err)
	}

	return count > 0, nil
}

func (r *MongoRepository) UpdateLastMessage(ctx context.Context, groupID primitive.ObjectID, lastMessage *LastMessage) error {
	collection := r.db.Collection(GroupsCollection)

	filter := bson.M{"_id": groupID}
	update := bson.M{
		"$set": bson.M{
			"last_message": lastMessage,
			"updated_at":   time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update last message: %w", err)
	}

	return nil
}

// Message operations
func (r *MongoRepository) CreateMessage(ctx context.Context, message *Message) error {
	collection := r.db.Collection(MessagesCollection)

	message.CreatedAt = time.Now()
	message.UpdatedAt = time.Now()
	message.IsDeleted = false
	message.Status.Sent = time.Now()

	result, err := collection.InsertOne(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	message.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *MongoRepository) GetMessageByID(ctx context.Context, messageID primitive.ObjectID) (*Message, error) {
	collection := r.db.Collection(MessagesCollection)

	var message Message
	err := collection.FindOne(ctx, bson.M{
		"_id":        messageID,
		"is_deleted": false,
	}).Decode(&message)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("message not found")
		}
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	return &message, nil
}

func (r *MongoRepository) GetMessages(ctx context.Context, groupID primitive.ObjectID, limit int, before *primitive.ObjectID) ([]*Message, error) {
	collection := r.db.Collection(MessagesCollection)

	filter := bson.M{
		"group_id":   groupID,
		"is_deleted": false,
	}

	if before != nil {
		// Get messages before the specified message ID (for pagination)
		beforeMessage, err := r.GetMessageByID(ctx, *before)
		if err == nil {
			filter["created_at"] = bson.M{"$lt": beforeMessage.CreatedAt}
		}
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	defer cursor.Close(ctx)

	var messages []*Message
	for cursor.Next(ctx) {
		var message Message
		if err := cursor.Decode(&message); err != nil {
			continue
		}
		messages = append(messages, &message)
	}

	// Reverse to get chronological order (oldest first)
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

func (r *MongoRepository) SearchMessages(ctx context.Context, groupID primitive.ObjectID, query string, limit int, skip int) ([]*Message, error) {
	collection := r.db.Collection(MessagesCollection)

	filter := bson.M{
		"group_id":   groupID,
		"is_deleted": false,
		"$text":      bson.M{"$search": query},
	}

	opts := options.Find().
		SetSort(bson.D{
			{Key: "score", Value: bson.M{"$meta": "textScore"}},
			{Key: "created_at", Value: -1},
		}).
		SetLimit(int64(limit)).
		SetSkip(int64(skip))

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to search messages: %w", err)
	}
	defer cursor.Close(ctx)

	var messages []*Message
	for cursor.Next(ctx) {
		var message Message
		if err := cursor.Decode(&message); err != nil {
			continue
		}
		messages = append(messages, &message)
	}

	return messages, nil
}

func (r *MongoRepository) UpdateMessage(ctx context.Context, message *Message) error {
	collection := r.db.Collection(MessagesCollection)

	message.UpdatedAt = time.Now()
	editedTime := time.Now()
	message.Status.Edited = &editedTime

	filter := bson.M{
		"_id":        message.ID,
		"is_deleted": false,
	}
	update := bson.M{"$set": message}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update message: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("message not found or already deleted")
	}

	return nil
}

func (r *MongoRepository) DeleteMessage(ctx context.Context, messageID primitive.ObjectID) error {
	collection := r.db.Collection(MessagesCollection)

	filter := bson.M{
		"_id":        messageID,
		"is_deleted": false,
	}
	update := bson.M{
		"$set": bson.M{
			"is_deleted": true,
			"updated_at": time.Now(),
		},
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("message not found or already deleted")
	}

	return nil
}

func (r *MongoRepository) MarkMessagesAsRead(ctx context.Context, groupID primitive.ObjectID, messageIDs []primitive.ObjectID, userEmail string) error {
	collection := r.db.Collection(MessagesCollection)

	if len(messageIDs) == 0 {
		return nil // Nothing to update
	}

	filter := bson.M{
		"_id":                       bson.M{"$in": messageIDs},
		"group_id":                  groupID,
		"is_deleted":                false,
		"sender_email":              bson.M{"$ne": userEmail}, // Don't mark own messages as read
		"status.read_by.user_email": bson.M{"$ne": userEmail}, // Only update if not already read
	}
	update := bson.M{
		"$push": bson.M{
			"status.read_by": bson.M{
				"user_email": userEmail,
				"read_at":    time.Now(),
			},
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	_, err := collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to mark messages as read: %w", err)
	}

	return nil
}

func (r *MongoRepository) GetUnreadCount(ctx context.Context, groupID primitive.ObjectID, userEmail string) (int, error) {
	collection := r.db.Collection(MessagesCollection)

	filter := bson.M{
		"group_id":                  groupID,
		"is_deleted":                false,
		"sender_email":              bson.M{"$ne": userEmail}, // Exclude own messages
		"status.read_by.user_email": bson.M{"$ne": userEmail}, // Not read by user
	}

	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}

	return int(count), nil
}

// Utility operations
func (r *MongoRepository) GetGroupMembers(ctx context.Context, groupID primitive.ObjectID) ([]*Member, error) {
	collection := r.db.Collection(GroupsCollection)

	var group ChatGroup
	err := collection.FindOne(ctx, bson.M{
		"_id":       groupID,
		"is_active": true,
	}).Decode(&group)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("group not found")
		}
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	// Filter active members only
	var activeMembers []*Member
	for _, member := range group.Members {
		if member.IsActive {
			activeMembers = append(activeMembers, member)
		}
	}

	return activeMembers, nil
}

func (r *MongoRepository) IsGroupAdmin(ctx context.Context, groupID primitive.ObjectID, userEmail string) (bool, error) {
	collection := r.db.Collection(GroupsCollection)

	filter := bson.M{
		"_id":                groupID,
		"is_active":          true,
		"members.user_email": userEmail,
		"members.is_active":  true,
		"$or": []bson.M{
			{"created_by": userEmail}, // Group creator is always admin
			{
				"members": bson.M{
					"$elemMatch": bson.M{
						"user_email": userEmail,
						"role":       "admin",
						"is_active":  true,
					},
				},
			},
		},
	}

	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("failed to check admin status: %w", err)
	}

	return count > 0, nil
}

//mark messages as delivered
// This function marks messages as delivered for a specific group and user.
func (r *MongoRepository) MarkMessagesAsDelivered(ctx context.Context, groupID primitive.ObjectID, messageIDs []primitive.ObjectID, userEmail string) error {
	collection := r.db.Collection(MessagesCollection)

	if len(messageIDs) == 0 {
		return nil
	}

	now := time.Now()

	filter := bson.M{
		"_id":              bson.M{"$in": messageIDs},
		"group_id":         groupID,
		"is_deleted":       false,
		"sender_email":     bson.M{"$ne": userEmail},
		"status.delivered": nil,
	}

	update := bson.M{
		"$set": bson.M{
			"status.delivered": now,
			"updated_at":       now,
		},
	}

	_, err := collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to mark messages as delivered: %w", err)
	}

	return nil
}

func (r *MongoRepository) GetUndeliveredMessages(ctx context.Context, userEmail string, limit int, before *primitive.ObjectID) ([]*Message, error) {
	collection := r.db.Collection(MessagesCollection)

	filter := bson.M{
		"is_deleted":       false,
		"sender_email":     bson.M{"$ne": userEmail},
		"status.delivered": nil,
	}

	if before != nil {
		beforeMessage, err := r.GetMessageByID(ctx, *before)
		if err == nil {
			filter["created_at"] = bson.M{"$lt": beforeMessage.CreatedAt}
		}
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get undelivered messages: %w", err)
	}
	defer cursor.Close(ctx)

	var messages []*Message
	for cursor.Next(ctx) {
		var message Message
		if err := cursor.Decode(&message); err != nil {
			continue
		}
		messages = append(messages, &message)
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}
