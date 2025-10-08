// unit_tests/thread_test.go
package unit_tests

import (
	wspkg "aegis-api/pkg/websocket"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSendThreadEvent(t *testing.T) {
	mockNotifService := new(MockWsNotificationService)
	hub := NewTestHub(mockNotifService)

	// Start the hub
	go hub.Run()
	defer func() {
		// Use the exported method to close channels if available
		// Or just let the test finish naturally
	}()

	payload := wspkg.ThreadEventPayload{
		ThreadID: "thread123",
		Title:    "Test Thread",
		CaseID:   "case1",
	}

	err := wspkg.SendThreadEvent(hub.Hub, payload, wspkg.ThreadCreatedType)
	assert.NoError(t, err)
}

func TestSendThreadParticipantAdded(t *testing.T) {
	mockNotifService := new(MockWsNotificationService)
	hub := NewTestHub(mockNotifService)

	// Start the hub
	go hub.Run()

	payload := wspkg.ThreadParticipantPayload{
		ThreadID: "thread123",
		UserID:   "user123",
		UserName: "Test User",
		CaseID:   "case1",
		JoinedAt: time.Now().Format(time.RFC3339),
	}

	err := wspkg.SendThreadParticipantAdded(hub.Hub, payload)
	assert.NoError(t, err)
}

func TestSendThreadCreated(t *testing.T) {
	mockNotifService := new(MockWsNotificationService)
	hub := NewTestHub(mockNotifService)

	// Start the hub
	go hub.Run()

	payload := wspkg.ThreadCreatedPayload{
		ThreadID:  "thread123",
		Title:     "Test Thread",
		CaseID:    "case1",
		CreatedBy: "user123",
		CreatedAt: time.Now().Format(time.RFC3339),
		Priority:  "high",
	}

	err := wspkg.SendThreadCreated(hub.Hub, payload)
	assert.NoError(t, err)
}

// Test thread payload serialization
func TestThreadPayload_Serialization(t *testing.T) {
	payload := wspkg.ThreadEventPayload{
		ThreadID:  "thread123",
		Title:     "Test Thread",
		CaseID:    "case1",
		UpdatedBy: "user123",
		NewStatus: "open",
	}

	// Test JSON marshaling
	data, err := json.Marshal(payload)
	assert.NoError(t, err)

	var decodedPayload wspkg.ThreadEventPayload
	err = json.Unmarshal(data, &decodedPayload)
	assert.NoError(t, err)
	assert.Equal(t, payload.ThreadID, decodedPayload.ThreadID)
	assert.Equal(t, payload.Title, decodedPayload.Title)
}

// Test thread constants
func TestThreadConstants(t *testing.T) {
	assert.Equal(t, "THREAD_CREATED", wspkg.ThreadCreatedType)
	assert.Equal(t, "THREAD_RESOLVED", wspkg.ThreadResolvedType)
	assert.Equal(t, "THREAD_PARTICIPANT_ADDED", wspkg.ThreadParticipantAddedType)
}
