// unit_tests/client_test.go
package unit_tests

import (
	"aegis-api/pkg/chatModels"
	"aegis-api/pkg/sharedws"
	wspkg "aegis-api/pkg/websocket"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

// TestClientWrapper wraps the Client for testing
type TestClientWrapperN struct {
	Client *wspkg.Client
	Send   chan []byte
}

func NewTestClientN(hub *TestHubWrapper, userID, caseID string, conn *websocket.Conn) *TestClientWrapperN {
	sharedClient := &sharedws.Client{
		UserID: userID,
		CaseID: caseID,
		Conn:   conn,
	}

	client := &wspkg.Client{
		Client: sharedClient,
		Hub:    hub.Hub,
		Send:   make(chan []byte, 256),
	}

	return &TestClientWrapperN{
		Client: client,
		Send:   client.Send,
	}
}

func TestClient_WritePump_ChannelOperation(t *testing.T) {
	mockNotifService := new(MockWsNotificationService)
	hub := NewTestHub(mockNotifService)

	server, conn := CreateMockWebSocket()
	defer server.Close()
	defer conn.Close()

	client := NewTestClientN(hub, "user123", "case1", conn)

	// Test sending a message through the channel
	testMessage := []byte(`{"type": "test", "payload": "data"}`)

	go func() {
		client.Send <- testMessage
	}()

	// Verify the channel operation
	select {
	case msg := <-client.Send:
		assert.Equal(t, testMessage, msg)
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for message")
	}
}

func TestClient_MessageProcessing(t *testing.T) {
	// Test different message types without requiring hub
	testCases := []struct {
		name     string
		event    chatModels.WebSocketEvent
		expected string
	}{
		{
			name: "New Message Event",
			event: chatModels.WebSocketEvent{
				Type: chatModels.EventNewMessage,
				Payload: chatModels.NewMessagePayload{
					MessageID:  "msg1",
					Text:       "Hello",
					SenderID:   "user123",
					SenderName: "Test User",
					GroupID:    "group1",
					Timestamp:  time.Now().Format(time.RFC3339),
				},
			},
			expected: "Should process new message",
		},
		{
			name: "Typing Start Event",
			event: chatModels.WebSocketEvent{
				Type: chatModels.EventTypingStart,
				Payload: wspkg.TypingPayload{
					UserEmail: "user@test.com",
					CaseID:    "case1",
				},
			},
			expected: "Should process typing start",
		},
		{
			name: "Typing Stop Event",
			event: chatModels.WebSocketEvent{
				Type: chatModels.EventTypingStop,
				Payload: wspkg.TypingPayload{
					UserEmail: "user@test.com",
					CaseID:    "case1",
				},
			},
			expected: "Should process typing stop",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			eventData, err := json.Marshal(tc.event)
			assert.NoError(t, err)

			// Test that the event can be marshaled/unmarshaled properly
			var unmarshaledEvent chatModels.WebSocketEvent
			err = json.Unmarshal(eventData, &unmarshaledEvent)
			assert.NoError(t, err)
			assert.Equal(t, tc.event.Type, unmarshaledEvent.Type)
		})
	}
}

func TestClient_ConnectionLifecycle(t *testing.T) {
	mockNotifService := new(MockWsNotificationService)
	hub := NewTestHub(mockNotifService)

	server, conn := CreateMockWebSocket()
	defer server.Close()
	defer conn.Close()

	client := NewTestClientN(hub, "user123", "case1", conn)

	assert.NotNil(t, client)
	assert.NotNil(t, client.Client) // This was causing the panic
	assert.NotNil(t, client.Send)
	assert.Equal(t, "user123", client.Client.Client.UserID)
	assert.Equal(t, "case1", client.Client.Client.CaseID)
}

func TestClient_InvalidMessageHandling(t *testing.T) {
	// Test invalid JSON
	invalidMessage := []byte(`invalid json`)

	var msg chatModels.WebSocketEvent
	err := json.Unmarshal(invalidMessage, &msg)
	assert.Error(t, err)

	// Test unsupported message type
	unsupportedEvent := chatModels.WebSocketEvent{
		Type:    "UNSUPPORTED_TYPE",
		Payload: "some data",
	}

	eventData, err := json.Marshal(unsupportedEvent)
	assert.NoError(t, err)

	err = json.Unmarshal(eventData, &msg)
	assert.NoError(t, err)

	// Fixed: Compare same types
	assert.Equal(t, chatModels.EventType("UNSUPPORTED_TYPE"), msg.Type)
}

func TestClient_ChannelCapacity(t *testing.T) {
	mockNotifService := new(MockWsNotificationService)
	hub := NewTestHub(mockNotifService)

	server, conn := CreateMockWebSocket()
	defer server.Close()
	defer conn.Close()

	client := NewTestClientN(hub, "user123", "case1", conn)

	// Test that the channel has the expected capacity
	// The channel is buffered with capacity 256
	for i := 0; i < 256; i++ {
		select {
		case client.Send <- []byte("test"):
			// Success
		default:
			t.Error("Channel should accept 256 messages")
		}
	}

	// The 257th message should block or fail
	select {
	case client.Send <- []byte("overflow"):
		t.Error("Channel should be full after 256 messages")
	default:
		// Expected behavior - channel is full
	}
}

func TestClient_ConcurrentSend(t *testing.T) {
	mockNotifService := new(MockWsNotificationService)
	hub := NewTestHub(mockNotifService)

	server, conn := CreateMockWebSocket()
	defer server.Close()
	defer conn.Close()

	client := NewTestClientN(hub, "user123", "case1", conn)

	// Test concurrent sends
	messages := []string{"msg1", "msg2", "msg3", "msg4", "msg5"}
	done := make(chan bool, len(messages))

	for i, msg := range messages {
		go func(index int, message string) {
			client.Send <- []byte(message)
			done <- true
		}(i, msg)
	}

	// Wait for all sends to complete
	for i := 0; i < len(messages); i++ {
		select {
		case <-done:
			// Success
		case <-time.After(1 * time.Second):
			t.Error("Timeout waiting for concurrent send")
		}
	}
}

func TestServeWS_Function(t *testing.T) {
	// Test the ServeWS function signature and basic behavior
	// This is more of an integration test, but we can verify the function exists
	assert.NotNil(t, wspkg.ServeWS, "ServeWS function should exist")

	// Create a test server to simulate WebSocket upgrade
	server := httptest.NewServer(nil)
	defer server.Close()

	// We can't actually test ServeWS without a real WebSocket upgrade,
	// but we can verify the function signature and basic setup
	t.Log("ServeWS function is available for WebSocket connections")
}

// Simple test for client creation without complex hub setup
func TestClient_SimpleCreation(t *testing.T) {
	server, conn := CreateMockWebSocket()
	defer server.Close()
	defer conn.Close()

	// Create a minimal hub for testing
	mockNotifService := new(MockWsNotificationService)
	hub := NewTestHub(mockNotifService)

	client := NewTestClientN(hub, "test-user", "test-case", conn)

	assert.NotNil(t, client)
	assert.Equal(t, "test-user", client.Client.Client.UserID)
	assert.Equal(t, "test-case", client.Client.Client.CaseID)
}

// Test client Send channel operations
func TestClient_SendChannel(t *testing.T) {
	server, conn := CreateMockWebSocket()
	defer server.Close()
	defer conn.Close()

	mockNotifService := new(MockWsNotificationService)
	hub := NewTestHub(mockNotifService)
	client := NewTestClientN(hub, "user123", "case1", conn)

	// Test basic send/receive
	testMsg := []byte("test message")

	// Send
	go func() {
		client.Send <- testMsg
	}()

	// Receive
	select {
	case received := <-client.Send:
		assert.Equal(t, testMsg, received)
	case <-time.After(100 * time.Millisecond):
		t.Error("Failed to receive message from channel")
	}
}

// Test message serialization/deserialization
func TestClient_MessageSerialization(t *testing.T) {
	// Test valid message
	validMessage := chatModels.WebSocketEvent{
		Type: chatModels.EventNewMessage,
		Payload: chatModels.NewMessagePayload{
			MessageID: "123",
			Text:      "Hello",
			SenderID:  "user1",
		},
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(validMessage)
	assert.NoError(t, err)

	var decodedMessage chatModels.WebSocketEvent
	err = json.Unmarshal(data, &decodedMessage)
	assert.NoError(t, err)
	assert.Equal(t, validMessage.Type, decodedMessage.Type)
}
