package integration

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	gws "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"

	"aegis-api/pkg/websocket" // Adjust this path to match your module
	"github.com/google/uuid"

)

func TestWebSocket_ReceivesThreadCreatedEvent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create and run hub
	hub := websocket.NewHub()
	go hub.Run()

	// Set up test Gin route
	r := gin.New()
	r.GET("/ws/cases/:case_id", func(c *gin.Context) {
		caseID := c.Param("case_id")
		userID := "test-user"
		websocket.ServeWS(hub, c.Writer, c.Request, userID, caseID)
	})

	// Start test server
	srv := httptest.NewServer(r)
	defer srv.Close()

	// ✅ Use valid UUID for caseID
	caseID := uuid.New().String()
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws/cases/" + caseID

	// Connect WebSocket client
	conn, _, err := gws.DefaultDialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	defer conn.Close()

	time.Sleep(100 * time.Millisecond) // Give hub time to register connection

	// Send THREAD_CREATED event to correct caseID
	payload := websocket.ThreadCreatedPayload{
		ThreadID:  "thread-123",
		Title:     "New Thread",
		CaseID:    caseID, // ✅ matches above
		FileID:    "file-abc",
		CreatedBy: "user-xyz",
		CreatedAt: "2025-07-22T12:00:00Z",
		Priority:  "High",
	}

	assert.True(t, hub.WaitForClient(caseID, 1*time.Second), "WebSocket client was not registered in time")
	err = websocket.SendThreadCreated(hub, payload)
	assert.NoError(t, err)

	// Read WebSocket message
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err := conn.ReadMessage()
	assert.NoError(t, err)

	var wsMsg websocket.WebSocketMessage
	err = json.Unmarshal(msg, &wsMsg)
	assert.NoError(t, err)
	assert.Equal(t, websocket.ThreadCreatedType, wsMsg.Type)

	// Decode payload properly
	payloadBytes, err := json.Marshal(wsMsg.Payload)
	assert.NoError(t, err)

	var payloadMap map[string]interface{}
	err = json.Unmarshal(payloadBytes, &payloadMap)
	assert.NoError(t, err)

	assert.Equal(t, "thread-123", payloadMap["thread_id"])
	assert.Equal(t, "New Thread", payloadMap["title"])
	assert.Equal(t, caseID, payloadMap["case_id"]) // ✅ matches
}
