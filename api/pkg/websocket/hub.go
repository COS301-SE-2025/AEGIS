package websocket

import (
	"aegis-api/pkg/sharedws"
	"aegis-api/services_/chat"
	"aegis-api/services_/notification"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

type Hub struct {
	clients             map[string]map[*Client]bool // caseID -> clients
	broadcast           chan MessageEnvelope
	register            chan *Client
	unregister          chan *Client
	mu                  sync.Mutex
	connections         map[string]map[string]*websocket.Conn
	NotificationService *notification.NotificationService
}
type Claims struct {
	Email        string `json:"email"`
	UserID       string `json:"user_id"`
	Role         string `json:"role"`
	FullName     string `json:"fullName"`
	TokenVersion int    `json:"token_version"`
	jwt.RegisteredClaims
}

// Ensure Hub implements the chat.WebSocketManager interface
var _ chat.WebSocketManager = (*Hub)(nil)

func NewHub(notificationService *notification.NotificationService) *Hub {
	return &Hub{
		clients:             make(map[string]map[*Client]bool),
		broadcast:           make(chan MessageEnvelope),
		register:            make(chan *Client),
		unregister:          make(chan *Client),
		NotificationService: notificationService,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			if h.clients[client.CaseID] == nil {
				h.clients[client.CaseID] = make(map[*Client]bool)
			}
			h.clients[client.CaseID][client] = true

		case client := <-h.unregister:
			if clients, ok := h.clients[client.CaseID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.Send)
				}
			}

		case message := <-h.broadcast:
			if clients, ok := h.clients[message.CaseID]; ok {
				for client := range clients {
					select {
					case client.Send <- message.Data:
					default:
						close(client.Send)
						delete(clients, client)
					}
				}
			}
		}
	}

}
func (h *Hub) AddConnection(userEmail, caseID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.connections == nil {
		h.connections = make(map[string]map[string]*websocket.Conn)
	}

	if _, ok := h.connections[caseID]; !ok {
		h.connections[caseID] = make(map[string]*websocket.Conn)
	}

	// Optionally close old connection
	if oldConn, exists := h.connections[caseID][userEmail]; exists {
		oldConn.Close()
	}

	h.connections[caseID][userEmail] = conn
}

func (h *Hub) CountClients(caseID string) int {
	if clients, ok := h.clients[caseID]; ok {
		return len(clients)
	}
	return 0
}

func (h *Hub) ListConnectedUsers(caseID string) []string {
	users := []string{}
	if clients, ok := h.clients[caseID]; ok {
		for client := range clients {
			users = append(users, client.UserID)
		}
	}
	return users
}
func (h *Hub) GetActiveUsers(caseID string) []string {
	users := []string{}
	if clients, ok := h.clients[caseID]; ok {
		for client := range clients {
			users = append(users, client.UserID)
		}
	}
	return users
}

// for integration tests
func (h *Hub) WaitForClient(caseID string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if clients, ok := h.clients[caseID]; ok && len(clients) > 0 {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}
func (h *Hub) AddUserToGroup(userID, userEmail, caseID string, conn *websocket.Conn) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.connections == nil {
		h.connections = make(map[string]map[string]*websocket.Conn)
	}

	if _, ok := h.connections[caseID]; !ok {
		h.connections[caseID] = make(map[string]*websocket.Conn)
	}

	// Close existing connection if any
	if oldConn, exists := h.connections[caseID][userEmail]; exists && oldConn != conn {
		_ = oldConn.Close()
	}

	h.connections[caseID][userEmail] = conn

	log.Printf("✅ Added user %s (ID: %s) to group %s\n", userEmail, userID, caseID)
	return nil
}

func (h *Hub) BroadcastToGroup(groupID string, message chat.WebSocketMessage) error {
	// Marshal the message to JSON
	encoded, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to encode WebSocketMessage: %w", err)
	}

	envelope := MessageEnvelope{
		CaseID: groupID,
		Data:   encoded,
	}

	// Non-blocking broadcast
	select {
	case h.broadcast <- envelope:
		return nil
	default:
		return fmt.Errorf("broadcast channel full or not ready")
	}
}
func toWebSocketMessage(event chat.WebSocketEvent) chat.WebSocketMessage {
	return chat.WebSocketMessage{
		Type:      chat.MessageType(event.Type),
		Payload:   event.Payload,
		GroupID:   event.GroupID,
		UserEmail: event.UserEmail,
		Timestamp: event.Timestamp,
	}
}

// Send typing start notification
// Send typing start notification
func (h *Hub) BroadcastTypingStart(groupID string, userEmail string) error {
	typingPayload := TypingPayload{
		UserEmail: userEmail,
		CaseID:    groupID,
	}

	typingMessage := chat.WebSocketEvent{
		Type:      chat.EventTypingStart, // 👈 this is of type EventType, as expected
		Payload:   typingPayload,
		GroupID:   groupID,
		UserEmail: userEmail,
		Timestamp: time.Now(),
	}

	log.Printf("📤 Broadcasting typing_start for user %s in group %s", userEmail, groupID)

	encoded, err := json.Marshal(typingMessage)
	if err != nil {
		log.Printf("❌ Failed to marshal typing_start message: %v", err)
		return err
	}

	h.broadcast <- MessageEnvelope{
		CaseID: groupID,
		Data:   encoded,
	}
	return nil
}

// Send typing stop notification
// Send typing stop notification
func (h *Hub) BroadcastTypingStop(groupID string, userEmail string) error {
	typingPayload := TypingPayload{
		UserEmail: userEmail,
		CaseID:    groupID,
	}

	typingMessage := chat.WebSocketEvent{
		Type:      chat.EventTypingStop, // ✅ correct EventType
		Payload:   typingPayload,
		GroupID:   groupID,
		UserEmail: userEmail,
		Timestamp: time.Now(),
	}

	log.Printf("📤 Broadcasting typing_stop for user %s in group %s", userEmail, groupID)

	encoded, err := json.Marshal(typingMessage)
	if err != nil {
		log.Printf("❌ Failed to marshal typing_stop message: %v", err)
		return err
	}

	h.broadcast <- MessageEnvelope{
		CaseID: groupID,
		Data:   encoded,
	}
	return nil
}

func (h *Hub) BroadcastToCase(caseID string, message chat.WebSocketMessage) error {
	// Marshal message to JSON bytes
	encoded, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to encode WebSocketMessage: %w", err)
	}

	envelope := MessageEnvelope{
		CaseID: caseID,
		Data:   encoded,
	}

	// Non-blocking send to broadcast channel
	select {
	case h.broadcast <- envelope:
		return nil
	default:
		return fmt.Errorf("broadcast channel is full or unavailable")
	}
}
func extractCaseIDFromPath(path string) string {
	// Example: /ws/cases/<caseId>
	parts := strings.Split(path, "/")
	if len(parts) >= 4 && parts[2] == "cases" {
		return parts[3]
	}
	return ""
}
func GetJWTSecret() []byte {
	return []byte(os.Getenv("JWT_SECRET_KEY"))
}
func (h *Hub) HandleConnection(w http.ResponseWriter, r *http.Request) error {
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return fmt.Errorf("missing token in query params")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return GetJWTSecret(), nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return fmt.Errorf("invalid or expired token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || claims.UserID == "" {
		http.Error(w, "Invalid claims", http.StatusUnauthorized)
		return fmt.Errorf("invalid token claims")
	}

	userID := claims.UserID
	caseID := extractCaseIDFromPath(r.URL.Path) // or use groupID if you prefer

	if userID == "" || caseID == "" {
		return fmt.Errorf("missing userID or caseID in query params")
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return fmt.Errorf("websocket upgrade failed: %w", err)
	}

	client := &Client{
		Client: &sharedws.Client{
			UserID: userID,
			CaseID: caseID,
			Conn:   conn,
		},
		Hub:  h,
		Send: make(chan []byte, 256),
	}

	h.register <- client
	go client.readPump()
	go client.writePump()

	log.Printf("✅ WebSocket upgraded for user %s in case %s\n", userID, caseID)
	h.AddUserToGroup(userID, claims.Email, caseID, conn) // or groupID if applicable

	return nil
}

func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(75 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		messageType, rawMessage, err := c.Conn.ReadMessage()
		if err != nil {
			log.Printf("❌ Read error (%v): %v", messageType, err)
			break
		}

		if err != nil {
			log.Printf("❌ Read error from %s: %v", c.UserID, err)
			break
		}

		var msg chat.WebSocketEvent
		if err := json.Unmarshal(rawMessage, &msg); err != nil {
			log.Printf("❌ Invalid WebSocket message format: %v", err)
			continue
		}

		switch msg.Type {
		case chat.EventNewMessage:
			// Marshal and re-unmarshal payload into NewMessagePayload
			payloadBytes, err := json.Marshal(msg.Payload)
			if err != nil {
				log.Printf("❌ Failed to re-marshal payload: %v", err)
				continue
			}

			var payload chat.NewMessagePayload
			if err := json.Unmarshal(payloadBytes, &payload); err != nil {
				log.Printf("❌ Failed to decode NEW_MESSAGE payload: %v", err)
				continue
			}

			// Save message to DB
			log.Printf("📥 Persisting message with ID: %s", payload.MessageID)
			if err := chat.SaveMessageToDB(payload); err != nil {
				log.Printf("❌ Failed to persist message: %v", err)
			} else {
				log.Printf("✅ Message persisted successfully: %s", payload.MessageID)
			}

			// Re-encode full event for broadcasting
			broadcastMsg := chat.WebSocketEvent{
				Type:      chat.EventNewMessage,
				GroupID:   payload.GroupID,
				Payload:   payload,
				Timestamp: time.Now(),
				UserEmail: payload.SenderID, // or SenderEmail if available
			}

			encoded, err := json.Marshal(broadcastMsg)
			if err != nil {
				log.Printf("❌ Failed to encode message for broadcast: %v", err)
				continue
			}

			c.Hub.broadcast <- MessageEnvelope{
				CaseID: c.CaseID,
				Data:   encoded,
			}

		case chat.EventTypingStart:
			payloadBytes, _ := json.Marshal(msg.Payload)
			var payload TypingPayload
			if err := json.Unmarshal(payloadBytes, &payload); err != nil {
				log.Printf("❌ Failed to decode typing_start payload: %v", err)
				continue
			}
			log.Printf("✍️ Typing START received from %s in case %s", payload.UserEmail, c.CaseID)
			c.Hub.BroadcastTypingStart(c.CaseID, payload.UserEmail)

		case chat.EventTypingStop:
			payloadBytes, _ := json.Marshal(msg.Payload)
			var payload TypingPayload
			if err := json.Unmarshal(payloadBytes, &payload); err != nil {
				log.Printf("❌ Failed to decode typing_stop payload: %v", err)
				continue
			}
			log.Printf("🛑 Typing STOP received from %s in case %s", payload.UserEmail, c.CaseID)
			c.Hub.BroadcastTypingStop(c.CaseID, payload.UserEmail)

		case chat.EventMarkNotificationRead:
			payloadBytes, _ := json.Marshal(msg.Payload)
			var payload chat.MarkReadPayload
			if err := json.Unmarshal(payloadBytes, &payload); err != nil {
				log.Printf("❌ Failed to decode MARK_READ payload: %v", err)
				continue
			}

			err := c.Hub.NotificationService.MarkAsRead(payload.NotificationIDs)

			if err != nil {
				log.Printf("❌ Failed to mark notifications as read: %v", err)
				continue
			}

		default:
			log.Printf("⚠️ Unsupported WebSocket message type: %s", msg.Type)
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			w.Close()

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *Hub) RemoveUserFromGroup(userID string, groupID string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Check if the group exists
	group, ok := h.connections[groupID]
	if !ok {
		return fmt.Errorf("group %s not found", groupID)
	}

	// Check if the user exists in the group
	conn, ok := group[userID]
	if !ok {
		return fmt.Errorf("user %s not found in group %s", userID, groupID)
	}

	// Close the connection and remove the user
	conn.Close()
	delete(group, userID)

	// Clean up the group map if it's empty
	if len(group) == 0 {
		delete(h.connections, groupID)
	}

	log.Printf("👋 Removed user %s from group %s\n", userID, groupID)
	return nil
}

func (h *Hub) SendToUser(userID string, message interface{}) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Search for the user in all groups (optional: you can optimize if needed)
	for groupID, group := range h.connections {
		if conn, ok := group[userID]; ok && conn != nil {
			// Marshal the interface message
			encoded, err := json.Marshal(message)
			if err != nil {
				return fmt.Errorf("failed to encode message: %w", err)
			}

			// Write outside lock to avoid blocking other goroutines
			h.mu.Unlock()
			err = conn.WriteMessage(websocket.TextMessage, encoded)
			h.mu.Lock()
			if err != nil {
				conn.Close()
				delete(h.connections[groupID], userID)
				return fmt.Errorf("failed to send message to user %s: %w", userID, err)
			}

			return nil
		}
	}

	return fmt.Errorf("no active connection found for user %s", userID)
}
