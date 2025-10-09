// websocket_manager.go
package chatModels

import (
	"aegis-api/pkg/sharedws"
	"context"
	"os"

	"encoding/json"

	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MessageType represents different types of WebSocket messages
type MessageType string

const (
	MessageTypeChat          MessageType = "chat"
	MessageTypeTyping        MessageType = "typing"
	MessageTypeStopTyping    MessageType = "stop_typing"
	MessageTypeUserJoined    MessageType = "user_joined"
	MessageTypeUserLeft      MessageType = "user_left"
	MessageTypeMessageRead   MessageType = "message_read"
	MessageTypeMessageUpdate MessageType = "message_update"
	MessageTypeMessageDelete MessageType = "message_delete"
	MessageTypeError         MessageType = "error"
	MessageTypePing          MessageType = "ping"
	MessageTypePong          MessageType = "pong"
	MessageTypeDelivered     MessageType = "delivered"
	NewMessageType           MessageType = "NEW_MESSAGE"
)

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Type      MessageType `json:"type"`
	GroupID   string      `json:"group_id,omitempty"`
	UserEmail string      `json:"user_email,omitempty"`
	Payload   interface{} `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
}

// TypingIndicator represents typing status
type TypingIndicator struct {
	UserEmail string `json:"user_email"`
	IsTyping  bool   `json:"is_typing"`
}

// webSocketManager implements the WebSocketManager interface
type webSocketManager struct {
	connections      map[string]*websocket.Conn      // userEmail -> connection
	groupUsers       map[string][]string             // groupID -> []userEmail
	userGroups       map[string][]string             // userEmail -> []groupID
	typingUsers      map[string]map[string]time.Time // groupID -> userEmail -> lastTypingTime
	groupConnections map[string]map[string]*websocket.Conn
	caseGroups       map[string][]string // caseID → groupIDs
	// caseID → groupIDs
	clients    map[string]*sharedws.Client
	connection map[string]map[string]*websocket.Conn // caseID → userEmail → Conn

	mutex        sync.RWMutex
	upgrader     websocket.Upgrader
	userService  UserService
	pingInterval time.Duration
	pongTimeout  time.Duration

	repo ChatRepository
}

// NewWebSocketManager creates a new WebSocket manager
func NewWebSocketManager(userService UserService, repo ChatRepository) WebSocketManager {

	manager := &webSocketManager{
		connections:      make(map[string]*websocket.Conn),
		groupUsers:       make(map[string][]string),
		userGroups:       make(map[string][]string),
		typingUsers:      make(map[string]map[string]time.Time),
		groupConnections: make(map[string]map[string]*websocket.Conn),
		caseGroups:       make(map[string][]string),                   // ✅ ADD THIS
		clients:          make(map[string]*sharedws.Client),           // ✅ ADD THIS
		connection:       make(map[string]map[string]*websocket.Conn), // ✅ ADD THIS
		userService:      userService,

		repo: repo,

		pingInterval: 30 * time.Second,
		pongTimeout:  60 * time.Second,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Configure CORS properly for production
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}

	stopChan := make(chan struct{})

	// Start cleanup routine for typing indicators
	go manager.cleanupTypingIndicators(stopChan)

	return manager
}
func (w *webSocketManager) AddConnection(userEmail, caseID string, conn *websocket.Conn) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// Create map for caseID if not present
	if w.connections == nil {
		w.connection = make(map[string]map[string]*websocket.Conn)
	}

	if _, exists := w.connections[caseID]; !exists {
		w.connection[caseID] = make(map[string]*websocket.Conn)
	}

	// Close any previous connection for this user in this case
	if oldConn, exists := w.connection[caseID][userEmail]; exists {
		oldConn.Close()
	}

	// Save new connection
	w.connection[caseID][userEmail] = conn
}

func (w *webSocketManager) BroadcastToCase(caseID string, message WebSocketMessage) error {
	w.mutex.RLock()
	groupIDs, exists := w.caseGroups[caseID]
	w.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("no groups found for case %s", caseID)
	}

	for _, groupID := range groupIDs {
		if err := w.BroadcastToGroup(groupID, message); err != nil {
			log.Printf("❌ Failed to broadcast to group %s: %v", groupID, err)
		}
	}
	return nil
}

// HandleConnection upgrades HTTP connection to WebSocket and manages it
var jwtSecret = []byte(os.Getenv("JWT_SECRET_KEY"))

func (w *webSocketManager) HandleConnection(wr http.ResponseWriter, r *http.Request, c *gin.Context) error {
	// ✅ USE THE GIN CONTEXT instead of re-validating the token
	userID := c.GetString("userID")
	userEmail := c.GetString("email")
	//tenantID := c.GetString("tenantID")

	// Check if authentication was already done by middleware
	if userID == "" || userEmail == "" {
		http.Error(wr, "Authentication required", http.StatusUnauthorized)
		return fmt.Errorf("user not authenticated in context")
	}

	log.Printf("✅ WebSocket connection for user %s (%s)", userEmail, userID)

	// ✅ Get groupID from query param
	groupID := r.URL.Query().Get("groupId")
	if groupID == "" {
		http.Error(wr, "Missing groupId", http.StatusBadRequest)
		return fmt.Errorf("groupId query param is required")
	}

	// ✅ Get caseID from URL path (since it's /cases/:caseId)
	caseID := c.Param("caseId")
	if caseID == "" {
		http.Error(wr, "Missing caseId", http.StatusBadRequest)
		return fmt.Errorf("caseId path parameter is required")
	}

	//tenantID = "" // Not used in this example, but could be validated

	// ✅ Upgrade to WebSocket
	conn, err := w.upgrader.Upgrade(wr, r, nil)
	if err != nil {
		return fmt.Errorf("failed to upgrade connection: %w", err)
	}

	// ✅ Register connection
	w.mutex.Lock()
	if existingConn, exists := w.connections[userEmail]; exists {
		existingConn.Close()
	}
	w.connections[userEmail] = conn
	w.mutex.Unlock()

	// ✅ Add user to group
	if err := w.AddUserToGroup(userEmail, groupID, caseID, conn); err != nil {
		log.Printf("Failed to add user %s to group %s: %v", userEmail, groupID, err)
	}

	log.Printf("✅ User %s connected via WebSocket to group %s, case %s", userEmail, groupID, caseID)

	// ✅ Start listener & ping routines
	go w.handleConnectionMessages(userEmail, groupID, conn)
	go w.pingConnection(userEmail, conn)

	// ✅ Optionally deliver queued messages
	go w.deliverQueuedMessages(userEmail)

	return nil
}

// handleConnectionMessages handles incoming messages from a WebSocket connection
func (w *webSocketManager) handleConnectionMessages(userEmail, groupID string, conn *websocket.Conn) {
	defer func() {
		w.removeConnection(userEmail)
		conn.Close()
		log.Printf("⚠️ WebSocket closed for user %s (group %s)", userEmail, groupID)
	}()

	// Set pong handler
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(w.pongTimeout))
		return nil
	})

	// Set read deadline
	conn.SetReadDeadline(time.Now().Add(w.pongTimeout))

	for {
		var message WebSocketMessage
		err := conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error for user %s: %v", userEmail, err)
			}
			break
		}

		// Reset read deadline
		conn.SetReadDeadline(time.Now().Add(w.pongTimeout))

		// Handle different message types
		switch message.Type {
		case MessageTypeTyping:
			w.handleTypingIndicator(userEmail, message.GroupID, true)
		case MessageTypeStopTyping:
			w.handleTypingIndicator(userEmail, message.GroupID, false)

		case MessageTypeDelivered:
			var ack struct {
				MessageID string `json:"message_id"`
			}
			if err := json.Unmarshal([]byte(fmt.Sprint(message.Payload)), &ack); err == nil {
				objID, err := primitive.ObjectIDFromHex(ack.MessageID)
				if err != nil {
					log.Println("Invalid ObjectID:", ack.MessageID)
					break
				}

				// Retrieve the message to get its GroupID
				msg, err := w.repo.GetMessageByID(context.TODO(), objID)
				if err != nil {
					log.Println("Could not fetch message for delivery ack:", err)
					break
				}

				// Now mark it as delivered
				err = w.repo.MarkMessagesAsDelivered(
					context.TODO(),
					msg.GroupID,
					[]primitive.ObjectID{objID},
					userEmail,
				)
				if err != nil {
					log.Printf("Failed to mark message %s as delivered: %v", msg.ID, err)
				}
			}

		case MessageTypePong:
			// Pong received, connection is alive
			continue
		default:
			log.Printf("Unknown message type: %s from user: %s", message.Type, userEmail)
		}
	}

}

// pingConnection sends periodic ping messages to keep connection alive
func (w *webSocketManager) pingConnection(userEmail string, conn *websocket.Conn) {
	ticker := time.NewTicker(w.pingInterval)
	defer ticker.Stop()

	for range ticker.C {
		w.mutex.RLock()
		_, exists := w.connections[userEmail]
		w.mutex.RUnlock()

		if !exists {
			return
		}

		if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
			log.Printf("Failed to send ping to user %s: %v", userEmail, err)
			return
		}
	}
}

// removeConnection removes a user's connection and cleans up related data
func (w *webSocketManager) removeConnection(userEmail string) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// Remove connection
	delete(w.connections, userEmail)

	// Remove user from all groups
	if groups, exists := w.userGroups[userEmail]; exists {
		for _, groupID := range groups {
			w.removeUserFromGroupUnsafe(userEmail, groupID)
		}
		delete(w.userGroups, userEmail)
	}

	// Remove typing indicators
	for groupID := range w.typingUsers {
		delete(w.typingUsers[groupID], userEmail)
	}
}

// BroadcastToGroup sends a message to all users in a group
func (w *webSocketManager) BroadcastToGroup(groupID string, message WebSocketMessage) error {
	w.mutex.RLock()
	users, exists := w.groupUsers[groupID]
	w.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("group %s not found", groupID)
	}

	encoded, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to encode message: %w", err)
	}

	var failedUsers []string
	for _, userEmail := range users {
		if err := w.sendMessageToUser(userEmail, encoded); err != nil {
			failedUsers = append(failedUsers, userEmail)
			log.Printf("❌ Failed to send message to user %s: %v", userEmail, err)
		}
	}

	if len(failedUsers) > 0 {
		w.mutex.Lock()
		for _, userEmail := range failedUsers {
			w.removeUserFromGroupUnsafe(userEmail, groupID)
		}
		w.mutex.Unlock()
	}

	return nil
}

// SendToUser sends a message to a specific user
func (w *webSocketManager) SendToUser(userEmail string, message interface{}) error {
	wsMessage := WebSocketMessage{
		Type:      MessageTypeChat,
		UserEmail: userEmail,
		Payload:   message,
		Timestamp: time.Now(),
	}

	// 🔄 Convert to []byte
	data, err := json.Marshal(wsMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal WebSocketMessage: %w", err)
	}

	return w.sendMessageToUser(userEmail, data)
}

// sendMessageToUser sends a WebSocket message to a specific user
func (w *webSocketManager) sendMessageToUser(userEmail string, data []byte) error {
	client, ok := w.clients[userEmail]
	if !ok {
		return fmt.Errorf("no client for %s", userEmail)
	}

	select {
	case client.Send <- data:
		return nil
	default:
		return fmt.Errorf("client send buffer full for %s", userEmail)
	}
}

// AddUserToGroup adds a user to a group
// func (w *webSocketManager) AddUserToGroup(userEmail, groupID string) error {
// 	w.mutex.Lock()
// 	defer w.mutex.Unlock()

// 	// Add user to group
// 	if users, exists := w.groupUsers[groupID]; exists {
// 		// Check if user is already in group
// 		for _, user := range users {
// 			if user == userEmail {
// 				return nil // User already in group
// 			}
// 		}
// 		w.groupUsers[groupID] = append(users, userEmail)
// 	} else {
// 		w.groupUsers[groupID] = []string{userEmail}
// 	}

// 	// Add group to user's groups
// 	if groups, exists := w.userGroups[userEmail]; exists {
// 		// Check if group is already in user's groups
// 		for _, group := range groups {
// 			if group == groupID {
// 				return nil // Group already in user's groups
// 			}
// 		}
// 		w.userGroups[userEmail] = append(groups, groupID)
// 	} else {
// 		w.userGroups[userEmail] = []string{groupID}
// 	}

// 	// Notify other users in the group
// 	go w.notifyUserJoined(groupID, userEmail)

//		return nil
//	}
func (w *webSocketManager) AddUserToGroup(userEmail, groupID, caseID string, conn *websocket.Conn) error {
	log.Printf("[DEBUG] AddUserToGroup START user=%s group=%s case=%s", userEmail, groupID, caseID)
	w.mutex.Lock()
	defer func() {
		w.mutex.Unlock()
		log.Printf("[DEBUG] AddUserToGroup END user=%s group=%s case=%s", userEmail, groupID, caseID)
	}()

	// Check and add user to groupUsers
	userAlreadyInGroup := false
	users := w.groupUsers[groupID]
	for _, u := range users {
		if u == userEmail {
			userAlreadyInGroup = true
			break
		}
	}
	if !userAlreadyInGroup {
		w.groupUsers[groupID] = append(users, userEmail)
	}

	// Check and add group to userGroups
	groupAlreadyInUser := false
	groups := w.userGroups[userEmail]
	for _, g := range groups {
		if g == groupID {
			groupAlreadyInUser = true
			break
		}
	}
	if !groupAlreadyInUser {
		w.userGroups[userEmail] = append(groups, groupID)
	}

	// ✅ Save connection
	if w.groupConnections[groupID] == nil {
		w.groupConnections[groupID] = make(map[string]*websocket.Conn)
	}
	w.groupConnections[groupID][userEmail] = conn

	// Notify others
	go w.notifyUserJoined(groupID, userEmail)

	// Track group under caseID
	w.caseGroups[caseID] = appendUnique(w.caseGroups[caseID], groupID)
	return nil
}
func appendUnique(slice []string, value string) []string {
	for _, v := range slice {
		if v == value {
			return slice
		}
	}
	return append(slice, value)
}

// RemoveUserFromGroup removes a user from a group
func (w *webSocketManager) RemoveUserFromGroup(userEmail, groupID string) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.removeUserFromGroupUnsafe(userEmail, groupID)

	// Notify other users in the group
	go w.notifyUserLeft(groupID, userEmail)

	return nil
}

// removeUserFromGroupUnsafe removes a user from a group without locking (internal use)
func (w *webSocketManager) removeUserFromGroupUnsafe(userEmail, groupID string) {
	// Remove user from group
	if users, exists := w.groupUsers[groupID]; exists {
		for i, user := range users {
			if user == userEmail {
				w.groupUsers[groupID] = append(users[:i], users[i+1:]...)
				break
			}
		}
		// Remove group if empty
		if len(w.groupUsers[groupID]) == 0 {
			delete(w.groupUsers, groupID)
		}
	}

	// Remove group from user's groups
	if groups, exists := w.userGroups[userEmail]; exists {
		for i, group := range groups {
			if group == groupID {
				w.userGroups[userEmail] = append(groups[:i], groups[i+1:]...)
				break
			}
		}
		// Remove user if no groups
		if len(w.userGroups[userEmail]) == 0 {
			delete(w.userGroups, userEmail)
		}
	}

	// Remove typing indicator
	if typingUsers, exists := w.typingUsers[groupID]; exists {
		delete(typingUsers, userEmail)
		if len(typingUsers) == 0 {
			delete(w.typingUsers, groupID)
		}
	}
}

// GetActiveUsers returns the list of active users in a group
func (w *webSocketManager) GetActiveUsers(groupID string) []string {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	if users, exists := w.groupUsers[groupID]; exists {
		// Return copy to avoid race conditions
		result := make([]string, len(users))
		copy(result, users)
		return result
	}

	return []string{}
}

// handleTypingIndicator handles typing indicators
func (w *webSocketManager) handleTypingIndicator(userEmail, groupID string, isTyping bool) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if isTyping {
		if _, exists := w.typingUsers[groupID]; !exists {
			w.typingUsers[groupID] = make(map[string]time.Time)
		}
		w.typingUsers[groupID][userEmail] = time.Now()
	} else {
		if typingUsers, exists := w.typingUsers[groupID]; exists {
			delete(typingUsers, userEmail)
			if len(typingUsers) == 0 {
				delete(w.typingUsers, groupID)
			}
		}
	}

	// Broadcast typing indicator to other users in the group
	go w.broadcastTypingIndicator(groupID, userEmail, isTyping)
}

// broadcastTypingIndicator broadcasts typing indicator to group members
func (w *webSocketManager) broadcastTypingIndicator(groupID, userEmail string, isTyping bool) {
	w.mutex.RLock()
	users, exists := w.groupUsers[groupID]
	w.mutex.RUnlock()

	if !exists {
		return
	}

	message := WebSocketMessage{
		Type:    MessageTypeTyping,
		GroupID: groupID,
		Payload: TypingIndicator{
			UserEmail: userEmail,
			IsTyping:  isTyping,
		},
		Timestamp: time.Now(),
	}

	if !isTyping {
		message.Type = MessageTypeStopTyping
	}

	for _, user := range users {
		if user != userEmail { // Don't send to the typing user
			data, err := json.Marshal(message)
			if err != nil {
				log.Printf("❌ Failed to marshal typing indicator: %v", err)
				continue
			}
			w.sendMessageToUser(user, data)

		}
	}
}

func (w *webSocketManager) cleanupTypingIndicators(stopChan <-chan struct{}) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.mutex.Lock()
			now := time.Now()
			for groupID, typingUsers := range w.typingUsers {
				for userEmail, lastTyping := range typingUsers {
					if now.Sub(lastTyping) > 10*time.Second {
						delete(typingUsers, userEmail)
						go w.broadcastTypingIndicator(groupID, userEmail, false)
					}
				}
				if len(typingUsers) == 0 {
					delete(w.typingUsers, groupID)
				}
			}
			w.mutex.Unlock()

		case <-stopChan:
			return
		}
	}
}

// notifyUserJoined notifies group members that a user joined
func (w *webSocketManager) notifyUserJoined(groupID, userEmail string) {
	message := WebSocketMessage{
		Type:      MessageTypeUserJoined,
		GroupID:   groupID,
		UserEmail: userEmail,
		Timestamp: time.Now(),
	}

	w.mutex.RLock()
	users, exists := w.groupUsers[groupID]
	w.mutex.RUnlock()

	if !exists {
		return
	}

	for _, user := range users {
		if user != userEmail {
			data, err := json.Marshal(message)
			if err != nil {
				log.Printf("❌ Failed to marshal typing indicator: %v", err)
				continue
			}
			w.sendMessageToUser(user, data)

		}
	}
}

// notifyUserLeft notifies group members that a user left
func (w *webSocketManager) notifyUserLeft(groupID, userEmail string) {
	message := WebSocketMessage{
		Type:      MessageTypeUserLeft,
		GroupID:   groupID,
		UserEmail: userEmail,
		Timestamp: time.Now(),
	}

	w.mutex.RLock()
	users, exists := w.groupUsers[groupID]
	w.mutex.RUnlock()

	if !exists {
		return
	}

	for _, user := range users {
		if user != userEmail {
			data, err := json.Marshal(message)
			if err != nil {
				log.Printf("❌ Failed to marshal typing indicator: %v", err)
				continue
			}
			w.sendMessageToUser(user, data)

		}
	}
}

// to do thati
func (w *webSocketManager) deliverQueuedMessages(userEmail string) {
	ctx := context.TODO()

	messages, err := w.repo.GetUndeliveredMessages(ctx, userEmail, 100, nil)
	if err != nil {
		log.Println("Failed to fetch undelivered messages:", err)
		return
	}

	groupMsgMap := make(map[primitive.ObjectID][]primitive.ObjectID)

	for _, msg := range messages {
		event := WebSocketMessage{
			Type:      MessageType(EventNewMessage),
			GroupID:   msg.GroupID.Hex(),
			UserEmail: msg.SenderEmail,
			Payload:   msg,
			Timestamp: time.Now(),
		}

		data, err := json.Marshal(event)
		if err != nil {
			log.Printf("❌ Failed to marshal message for user %s: %v", userEmail, err)
			continue
		}

		if err := w.sendMessageToUser(userEmail, data); err == nil {
			if objID, err := primitive.ObjectIDFromHex(msg.ID); err == nil {
				groupMsgMap[msg.GroupID] = append(groupMsgMap[msg.GroupID], objID)
			} else {
				log.Printf("Invalid ObjectID for message: %v", msg.ID)
			}
		}
	}

	for groupID, messageIDs := range groupMsgMap {
		if err := w.repo.MarkMessagesAsDelivered(ctx, groupID, messageIDs, userEmail); err != nil {
			log.Printf("Failed to mark messages delivered for group %s: %v", groupID.Hex(), err)
		}
	}
}
