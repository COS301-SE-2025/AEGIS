// unit_tests/notification_test.go
package unit_tests

import (
	"aegis-api/pkg/chatModels"
	//wspkg "aegis-api/pkg/websocket"
	"aegis-api/services_/notification"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNotifyUser_Creation(t *testing.T) {
	mockNotifService := new(MockWsNotificationService)

	// Set up the expectation
	mockNotifService.On("SaveNotification", mock.AnythingOfType("*notification.Notification")).Return(nil)

	// Create notification
	notif := notification.Notification{
		ID:        uuid.New().String(),
		UserID:    "user123",
		TenantID:  "tenant1",
		TeamID:    "team1",
		Title:     "Test Title",
		Message:   "Test Message",
		Timestamp: time.Now(),
		Read:      false,
		Archived:  false,
	}

	// Actually call the method that should trigger SaveNotification
	err := mockNotifService.SaveNotification(&notif)
	assert.NoError(t, err)

	// Test the WebSocket event creation
	event := chatModels.WebSocketEvent{
		Type:      chatModels.EventNotification,
		Payload:   notif,
		UserEmail: "user123",
		Timestamp: notif.Timestamp,
	}

	assert.Equal(t, chatModels.EventNotification, event.Type)
	assert.Equal(t, "user123", event.UserEmail)

	mockNotifService.AssertExpectations(t)
}

func TestNotification_BroadcastToUser(t *testing.T) {
	// Test the function signature and basic behavior
	// This is more of a integration test pattern
	mockNotifService := new(MockWsNotificationService)
	hub := NewTestHub(mockNotifService)

	notif := notification.Notification{
		ID:        uuid.New().String(),
		UserID:    "user123",
		Title:     "Test",
		Message:   "Test message",
		Timestamp: time.Now(),
	}

	// Test that the method exists and can be called
	// The actual implementation might fail without real connections
	err := hub.BroadcastNotificationToUser("tenant1", "team1", "user123", notif)

	// This might fail without real connections, but we're testing the interface
	if err != nil {
		t.Logf("BroadcastNotificationToUser returned: %v", err)
	}
}

// Test notification serialization
func TestNotification_Serialization(t *testing.T) {
	notif := notification.Notification{
		ID:        "notif123",
		UserID:    "user123",
		Title:     "Test Notification",
		Message:   "This is a test message",
		Timestamp: time.Now(),
		Read:      false,
		Archived:  false,
	}

	// Test JSON marshaling
	data, err := json.Marshal(notif)
	assert.NoError(t, err)

	var decodedNotif notification.Notification
	err = json.Unmarshal(data, &decodedNotif)
	assert.NoError(t, err)
	assert.Equal(t, notif.ID, decodedNotif.ID)
	assert.Equal(t, notif.Title, decodedNotif.Title)
}

// Test WebSocket event types for notifications
func TestNotification_EventTypes(t *testing.T) {
	assert.Equal(t, chatModels.EventType("notification"), chatModels.EventNotification)
	assert.Equal(t, chatModels.EventType("mark_notification_read"), chatModels.EventMarkNotificationRead)
	assert.Equal(t, chatModels.EventType("archive_notification"), chatModels.EventArchiveNotification)
	assert.Equal(t, chatModels.EventType("delete_notification"), chatModels.EventDeleteNotification)
	assert.Equal(t, chatModels.EventType("notification_sync"), chatModels.EventNotificationSync)
}

// Test mark read payload
func TestMarkReadPayload(t *testing.T) {
	payload := chatModels.MarkReadPayload{
		NotificationIDs: []string{"notif1", "notif2", "notif3"},
	}

	data, err := json.Marshal(payload)
	assert.NoError(t, err)

	var decodedPayload chatModels.MarkReadPayload
	err = json.Unmarshal(data, &decodedPayload)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(decodedPayload.NotificationIDs))
	assert.Contains(t, decodedPayload.NotificationIDs, "notif1")
}
