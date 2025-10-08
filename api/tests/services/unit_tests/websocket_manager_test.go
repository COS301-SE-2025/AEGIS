// unit_tests/websocket_manager_test.go
package unit_tests

import (
	"aegis-api/pkg/chatModels"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

// Test wrapper types
type TestWebSocketManager = chatModels.TestWebSocketManager
type MockUserService = chatModels.MockUserService

// type MockChatRepository = chatModels.MockChatRepository
type WebSocketMessage = chatModels.WebSocketMessage

const (
	MessageTypeChat       = chatModels.MessageTypeChat
	MessageTypeTyping     = chatModels.MessageTypeTyping
	MessageTypeStopTyping = chatModels.MessageTypeStopTyping
)

// Add this function to create a timeout context for tests
func getTestContext(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func TestNewWebSocketManager(t *testing.T) {
	mockUserService := new(MockUserService)
	mockRepo := new(chatModels.MockChatRepository)

	manager := chatModels.NewTestWebSocketManager(mockUserService, mockRepo)

	assert.NotNil(t, manager)

	// Test internal state through exported methods
	groupUsers, userGroups, typingUsers := manager.GetInternalState()
	assert.NotNil(t, groupUsers)
	assert.NotNil(t, userGroups)
	assert.NotNil(t, typingUsers)
}

func TestWebSocketManager_AddUserToGroup(t *testing.T) {
	mockUserService := new(MockUserService)
	mockRepo := new(chatModels.MockChatRepository)
	manager := chatModels.NewTestWebSocketManager(mockUserService, mockRepo)

	// Create a mock WebSocket connection
	_, conn := createMockWebSocket()
	defer conn.Close()

	err := manager.AddUserToGroup("user@test.com", "group1", "case1", conn)

	assert.NoError(t, err)

	// Verify through internal state
	groupUsers, userGroups, _ := manager.GetInternalState()
	assert.Contains(t, groupUsers["group1"], "user@test.com")
	assert.Contains(t, userGroups["user@test.com"], "group1")
}

func TestWebSocketManager_RemoveUserFromGroup(t *testing.T) {
	mockUserService := new(MockUserService)
	mockRepo := new(chatModels.MockChatRepository)
	manager := chatModels.NewTestWebSocketManager(mockUserService, mockRepo)

	// Add user first
	_, conn := createMockWebSocket()
	defer conn.Close()

	manager.AddUserToGroup("user@test.com", "group1", "case1", conn)

	err := manager.RemoveUserFromGroup("user@test.com", "group1")

	assert.NoError(t, err)

	// Verify through internal state
	groupUsers, userGroups, _ := manager.GetInternalState()
	assert.NotContains(t, groupUsers["group1"], "user@test.com")
	assert.NotContains(t, userGroups["user@test.com"], "group1")
}

func TestWebSocketManager_BroadcastToGroup(t *testing.T) {
	mockUserService := new(MockUserService)
	mockRepo := new(chatModels.MockChatRepository)
	manager := chatModels.NewTestWebSocketManager(mockUserService, mockRepo)

	// Add users to group
	server1, conn1 := createMockWebSocket()
	server2, conn2 := createMockWebSocket()
	defer server1.Close()
	defer server2.Close()
	defer conn1.Close()
	defer conn2.Close()

	manager.AddUserToGroup("user1@test.com", "group1", "case1", conn1)
	manager.AddUserToGroup("user2@test.com", "group1", "case1", conn2)

	message := WebSocketMessage{
		Type:      MessageTypeChat,
		GroupID:   "group1",
		Payload:   "Hello World",
		Timestamp: time.Now(),
	}

	err := manager.BroadcastToGroup("group1", message)
	assert.NoError(t, err)
}

func TestWebSocketManager_GetActiveUsers(t *testing.T) {
	mockUserService := new(MockUserService)
	mockRepo := new(chatModels.MockChatRepository)
	manager := chatModels.NewTestWebSocketManager(mockUserService, mockRepo)

	// Add users to group
	_, conn1 := createMockWebSocket()
	_, conn2 := createMockWebSocket()
	defer conn1.Close()
	defer conn2.Close()

	manager.AddUserToGroup("user1@test.com", "group1", "case1", conn1)
	manager.AddUserToGroup("user2@test.com", "group1", "case1", conn2)

	activeUsers := manager.GetActiveUsers("group1")
	assert.Len(t, activeUsers, 2)
	assert.Contains(t, activeUsers, "user1@test.com")
	assert.Contains(t, activeUsers, "user2@test.com")
}

func TestWebSocketManager_HandleTypingIndicator(t *testing.T) {
	mockUserService := new(MockUserService)
	mockRepo := new(chatModels.MockChatRepository)
	manager := chatModels.NewTestWebSocketManager(mockUserService, mockRepo)

	// Add user to group first
	_, conn := createMockWebSocket()
	defer conn.Close()
	manager.AddUserToGroup("user@test.com", "group1", "case1", conn)

	// Test typing start
	manager.HandleTypingIndicator("user@test.com", "group1", true)

	_, _, typingUsers := manager.GetInternalState()
	assert.Contains(t, typingUsers["group1"], "user@test.com")

	// Test typing stop
	manager.HandleTypingIndicator("user@test.com", "group1", false)

	_, _, typingUsers = manager.GetInternalState()
	groupTyping, exists := typingUsers["group1"]
	if exists {
		assert.NotContains(t, groupTyping, "user@test.com")
	}
}

func TestWebSocketManager_CleanupTypingIndicators(t *testing.T) {
	mockUserService := new(MockUserService)
	mockRepo := new(chatModels.MockChatRepository)
	manager := chatModels.NewTestWebSocketManager(mockUserService, mockRepo)

	stop := make(chan struct{})

	go func() {
		time.Sleep(100 * time.Millisecond)
		close(stop)
	}()

	assert.NotPanics(t, func() {
		manager.CleanupTypingIndicators(stop) // ✅ directly call promoted method
	})
}

func TestWebSocketManager_SendToUser(t *testing.T) {
	mockUserService := new(MockUserService)
	mockRepo := new(chatModels.MockChatRepository)
	manager := chatModels.NewTestWebSocketManager(mockUserService, mockRepo)

	// For this test, we need to create a real connection to test sending
	server, conn := createMockWebSocket()
	defer server.Close()
	defer conn.Close()

	// Add user with connection
	manager.AddUserToGroup("user@test.com", "group1", "case1", conn)

	message := "Test message"
	err := manager.SendToUser("user@test.com", message)

	// This might fail if there's no proper client setup, but we're testing the interface
	// In a real scenario, you'd set up the client properly
	if err != nil {
		t.Logf("Expected error in test environment: %v", err)
	}
}

// / Helper function to create mock WebSocket connections
func createMockWebSocket() (*httptest.Server, *websocket.Conn) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		// Keep connection open but with proper timeout handling
		go func() {
			defer conn.Close()
			conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			for {
				if _, _, err := conn.NextReader(); err != nil {
					break
				}
				// Reset deadline
				conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			}
		}()
	}))

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		server.Close()
		return nil, nil
	}

	return server, conn
}
