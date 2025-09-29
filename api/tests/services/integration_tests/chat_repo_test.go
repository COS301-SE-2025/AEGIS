package integration_test

import (
	"context"
	"testing"
	"time"

	"aegis-api/services_/chat"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func cleanChatCollections(t *testing.T) {
	t.Helper()
	ctx, cancel := context.WithTimeout(tcCtx, 10*time.Second)
	defer cancel()

	_, _ = mongoDB.Collection(chat.GroupsCollection).DeleteMany(ctx, bson.D{})
	_, _ = mongoDB.Collection(chat.MessagesCollection).DeleteMany(ctx, bson.D{})
}

func TestChat_GetUserGroups(t *testing.T) {
	cleanChatCollections(t)
	ctx, cancel := context.WithTimeout(tcCtx, 15*time.Second)
	defer cancel()

	repo := chat.NewChatRepository(mongoDB, nil, nil, nil)

	user := "user@example.com"
	other := "other@example.com"

	groups := mongoDB.Collection(chat.GroupsCollection)

	// Insert two groups where the user is involved (creator in one, member in another)
	g1 := bson.M{
		"_id":        primitive.NewObjectID(),
		"name":       "Alpha",
		"created_by": user,
		"is_active":  true,
		"members":    []bson.M{{"user_email": user, "is_active": true}},
		"updated_at": time.Now(),
	}
	g2 := bson.M{
		"_id":        primitive.NewObjectID(),
		"name":       "Beta",
		"created_by": other,
		"is_active":  true,
		"members":    []bson.M{{"user_email": user, "is_active": true}},
		"updated_at": time.Now(),
	}
	_, err := groups.InsertMany(ctx, []interface{}{g1, g2})
	require.NoError(t, err)

	out, err := repo.GetUserGroups(ctx, user)
	require.NoError(t, err)
	require.Len(t, out, 2)
	require.ElementsMatch(t, []string{"Alpha", "Beta"}, []string{out[0].Name, out[1].Name})
}

func TestChat_CreateMessage_And_GetMessages(t *testing.T) {
	cleanChatCollections(t)
	ctx, cancel := context.WithTimeout(tcCtx, 15*time.Second)
	defer cancel()

	repo := chat.NewChatRepository(mongoDB, nil, nil, nil)

	// Create a group first (raw insert so we don't depend on struct layout)
	groupID := primitive.NewObjectID()
	_, err := mongoDB.Collection(chat.GroupsCollection).InsertOne(ctx, bson.M{
		"_id":        groupID,
		"name":       "Gamma",
		"created_by": "creator@example.com",
		"is_active":  true,
		"members": []bson.M{
			{"user_email": "creator@example.com", "is_active": true},
			{"user_email": "alice@example.com", "is_active": true},
		},
		"updated_at": time.Now(),
	})
	require.NoError(t, err)

	// Create a message via repository
	msg := &chat.Message{
		GroupID:     groupID,
		SenderEmail: "alice@example.com",
		Content:     "hello world",
	}
	err = repo.CreateMessage(ctx, msg)
	require.NoError(t, err)
	require.NotEmpty(t, msg.ID)

	// Fetch messages back
	got, err := repo.GetMessages(ctx, groupID, 50, nil)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, "hello world", got[0].Content)
	require.Equal(t, "alice@example.com", got[0].SenderEmail)
}

func TestChat_MarkAsRead_And_UnreadCount(t *testing.T) {
	cleanChatCollections(t)
	ctx, cancel := context.WithTimeout(tcCtx, 15*time.Second)
	defer cancel()

	repo := chat.NewChatRepository(mongoDB, nil, nil, nil)

	// Group
	groupID := primitive.NewObjectID()
	_, err := mongoDB.Collection(chat.GroupsCollection).InsertOne(ctx, bson.M{
		"_id":        groupID,
		"name":       "Delta",
		"created_by": "owner@example.com",
		"is_active":  true,
		"members": []bson.M{
			{"user_email": "owner@example.com", "is_active": true},
			{"user_email": "bob@example.com", "is_active": true},
		},
		"updated_at": time.Now(),
	})
	require.NoError(t, err)

	// One message from owner
	m := &chat.Message{
		GroupID:     groupID,
		SenderEmail: "owner@example.com",
		Content:     "secret",
	}
	require.NoError(t, repo.CreateMessage(ctx, m))
	require.NotEmpty(t, m.ID)

	// Initially unread for bob
	unread, err := repo.GetUnreadCount(ctx, groupID, "bob@example.com")
	require.NoError(t, err)
	require.Equal(t, 1, unread)

	// Mark as read by bob
	oid, err := primitive.ObjectIDFromHex(m.ID)
	require.NoError(t, err)

	require.NoError(t, repo.MarkMessagesAsRead(ctx, groupID, []primitive.ObjectID{oid}, "bob@example.com"))

	// Now zero unread
	unread2, err := repo.GetUnreadCount(ctx, groupID, "bob@example.com")
	require.NoError(t, err)
	require.Equal(t, 0, unread2)
}

func TestChat_UpdateGroupImage_And_DeleteMessage(t *testing.T) {
	cleanChatCollections(t)
	ctx, cancel := context.WithTimeout(tcCtx, 15*time.Second)
	defer cancel()

	repo := chat.NewChatRepository(mongoDB, nil, nil, nil)

	// Group
	groupID := primitive.NewObjectID()
	_, err := mongoDB.Collection(chat.GroupsCollection).InsertOne(ctx, bson.M{
		"_id":        groupID,
		"name":       "Omega",
		"created_by": "owner@example.com",
		"is_active":  true,
		"members":    []bson.M{{"user_email": "owner@example.com", "is_active": true}},
		"updated_at": time.Now(),
	})
	require.NoError(t, err)

	// Update image
	err = repo.UpdateGroupImage(ctx, groupID, "https://cdn.test/new.png")
	require.NoError(t, err)

	var groupDoc bson.M
	require.NoError(t, mongoDB.Collection(chat.GroupsCollection).
		FindOne(ctx, bson.M{"_id": groupID}).Decode(&groupDoc))
	require.Equal(t, "https://cdn.test/new.png", groupDoc["group_url"])

	// Insert a message and delete it through repository
	msg := &chat.Message{
		GroupID:     groupID,
		SenderEmail: "owner@example.com",
		Content:     "to-delete",
	}
	require.NoError(t, repo.CreateMessage(ctx, msg))

	// Delete
	oid, err := primitive.ObjectIDFromHex(msg.ID)
	require.NoError(t, err)
	require.NoError(t, repo.DeleteMessage(ctx, oid))

	// Ensure is_deleted=true
	var msgDoc bson.M
	require.NoError(t, mongoDB.Collection(chat.MessagesCollection).
		FindOne(ctx, bson.M{"_id": oid}).Decode(&msgDoc))
	require.Equal(t, true, msgDoc["is_deleted"])
}
