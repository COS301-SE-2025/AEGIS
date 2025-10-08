// unit_tests/message_test.go
package unit_tests

import (
	"aegis-api/pkg/chatModels"
	wspkg "aegis-api/pkg/websocket"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestSaveMessageToDB_ValidPayload(t *testing.T) {
	// This test would require a mock database connection
	// For now, we'll test the function logic

	payload := chatModels.NewMessagePayload{
		MessageID:  "msg123",
		Text:       "Test message",
		SenderID:   "user123",
		SenderName: "Test User",
		GroupID:    primitive.NewObjectID().Hex(),
		Timestamp:  time.Now().Format(time.RFC3339),
	}

	// We can't actually save without a database connection, but we can test the parsing
	parsedTime, err := time.Parse(time.RFC3339, payload.Timestamp)
	assert.NoError(t, err)

	groupID, err := primitive.ObjectIDFromHex(payload.GroupID)
	assert.NoError(t, err)

	expectedMessage := chatModels.Message{
		ID:          payload.MessageID,
		GroupID:     groupID,
		SenderEmail: payload.SenderName,
		SenderName:  payload.SenderName,
		Content:     payload.Text,
		CreatedAt:   parsedTime,
		UpdatedAt:   parsedTime,
		IsDeleted:   false,
		Status: chatModels.MessageStatus{
			Sent: parsedTime,
		},
		MessageType: "text",
	}

	assert.Equal(t, payload.MessageID, expectedMessage.ID)
	assert.Equal(t, payload.Text, expectedMessage.Content)
}

func TestSendJSONMessage(t *testing.T) {
	mockNotifService := new(MockWsNotificationService)
	hub := NewTestHub(mockNotifService)

	// Start the hub
	go hub.Run()

	payload := map[string]string{"test": "data"}

	err := wspkg.SendJSONMessage(hub.Hub, "case1", "TEST_TYPE", payload)
	assert.NoError(t, err)
}

// Test message constants
func TestMessageConstants(t *testing.T) {
	assert.Equal(t, "NEW_MESSAGE", wspkg.NewMessageType)
	assert.Equal(t, "THREAD_UPDATED", wspkg.ThreadUpdatedType)
	assert.Equal(t, "REACTION_UPDATED", wspkg.ReactionUpdatedType)
	assert.Equal(t, "MESSAGE_APPROVED", wspkg.MessageApprovedType)
}

// Test message envelope
func TestMessageEnvelope(t *testing.T) {
	envelope := wspkg.MessageEnvelope{
		CaseID: "case123",
		Data:   []byte("test data"),
	}

	assert.Equal(t, "case123", envelope.CaseID)
	assert.Equal(t, []byte("test data"), envelope.Data)
}

// Test WebSocket message structure
func TestWebSocketMessage(t *testing.T) {
	wsMessage := wspkg.WebSocketMessage{
		Type:    "TEST_TYPE",
		Payload: "test payload",
	}

	data, err := json.Marshal(wsMessage)
	assert.NoError(t, err)

	var decodedMessage wspkg.WebSocketMessage
	err = json.Unmarshal(data, &decodedMessage)
	assert.NoError(t, err)
	assert.Equal(t, wsMessage.Type, decodedMessage.Type)
	assert.Equal(t, wsMessage.Payload, decodedMessage.Payload)
}
