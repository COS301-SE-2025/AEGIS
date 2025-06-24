// websocket_manager.go
package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

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
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// TypingIndicator represents typing status
type TypingIndicator struct {
	UserEmail string `json:"user_email"`
	IsTyping  bool   `json:"is_typing"`
}

// webSocketManager implements the WebSocketManager interface
type webSocketManager struct {
	connections  map[string]*websocket.Conn      // userEmail -> connection
	groupUsers   map[string][]string             // groupID -> []userEmail
	userGroups   map[string][]string             // userEmail -> []groupID
	typingUsers  map[string]map[string]time.Time // groupID -> userEmail -> lastTypingTime
	mutex        sync.RWMutex
	upgrader     websocket.Upgrader
	userService  UserService
	pingInterval time.Duration
	pongTimeout  time.Duration
	repo         ChatRepository
}

// NewWebSocketManager creates a new WebSocket manager
func NewWebSocketManager(userService UserService, repo ChatRepository) WebSocketManager {
	manager := &webSocketManager{
		connections:  make(map[string]*websocket.Conn),
		groupUsers:   make(map[string][]string),
		userGroups:   make(map[string][]string),
		typingUsers:  make(map[string]map[string]time.Time),
		userService:  userService,
		repo:         repo,
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

	// Start cleanup routine for typing indicators
	go manager.cleanupTypingIndicators()

	return manager
}

// HandleConnection upgrades HTTP connection to WebSocket and manages it
func (w *webSocketManager) HandleConnection(userEmail string, wr http.ResponseWriter, r *http.Request) error {
	// Validate user exists
	exists, err := w.userService.ValidateUserExists(context.Background(), userEmail)
	if err != nil {
		return fmt.Errorf("failed to validate user: %w", err)
	}
	if !exists {
		return fmt.Errorf("user not found: %s", userEmail)
	}

	// Upgrade connection
	conn, err := w.upgrader.Upgrade(wr, r, nil)
	if err != nil {
		return fmt.Errorf("failed to upgrade connection: %w", err)
	}

	// Store connection
	w.mutex.Lock()
	// Close existing connection if any
	if existingConn, exists := w.connections[userEmail]; exists {
		existingConn.Close()
	}
	w.connections[userEmail] = conn
	w.mutex.Unlock()

	log.Printf("User %s connected", userEmail)

	// Start connection management
	go w.handleConnectionMessages(userEmail, conn)
	go w.pingConnection(userEmail, conn)

	return nil
}

// handleConnectionMessages handles incoming messages from a WebSocket connection
func (w *webSocketManager) handleConnectionMessages(userEmail string, conn *websocket.Conn) {
	defer func() {
		w.removeConnection(userEmail)
		conn.Close()
		log.Printf("User %s disconnected", userEmail)
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
	if err := json.Unmarshal([]byte(fmt.Sprint(message.Data)), &ack); err == nil {
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
			[]primitive.ObjectID{msg.ID},
			userEmail,
		)
		if err != nil {
			log.Printf("Failed to mark message %s as delivered: %v", msg.ID.Hex(), err)
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
func (w *webSocketManager) BroadcastToGroup(groupID string, message interface{}) error {
	w.mutex.RLock()
	users, exists := w.groupUsers[groupID]
	w.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("group %s not found", groupID)
	}

	wsMessage := WebSocketMessage{
		Type:      MessageTypeChat,
		GroupID:   groupID,
		Data:      message,
		Timestamp: time.Now(),
	}

	var failedUsers []string
	for _, userEmail := range users {
		if err := w.sendMessageToUser(userEmail, wsMessage); err != nil {
			failedUsers = append(failedUsers, userEmail)
			log.Printf("Failed to send message to user %s: %v", userEmail, err)
		}
	}

	// Remove failed users from group
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
		Data:      message,
		Timestamp: time.Now(),
	}

	return w.sendMessageToUser(userEmail, wsMessage)
}

// sendMessageToUser sends a WebSocket message to a specific user
func (w *webSocketManager) sendMessageToUser(userEmail string, message WebSocketMessage) error {
	w.mutex.RLock()
	conn, exists := w.connections[userEmail]
	w.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("user %s not connected", userEmail)
	}

	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return conn.WriteJSON(message)
}

// AddUserToGroup adds a user to a group
func (w *webSocketManager) AddUserToGroup(userEmail, groupID string) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// Add user to group
	if users, exists := w.groupUsers[groupID]; exists {
		// Check if user is already in group
		for _, user := range users {
			if user == userEmail {
				return nil // User already in group
			}
		}
		w.groupUsers[groupID] = append(users, userEmail)
	} else {
		w.groupUsers[groupID] = []string{userEmail}
	}

	// Add group to user's groups
	if groups, exists := w.userGroups[userEmail]; exists {
		// Check if group is already in user's groups
		for _, group := range groups {
			if group == groupID {
				return nil // Group already in user's groups
			}
		}
		w.userGroups[userEmail] = append(groups, groupID)
	} else {
		w.userGroups[userEmail] = []string{groupID}
	}

	// Notify other users in the group
	go w.notifyUserJoined(groupID, userEmail)

	return nil
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
		Data: TypingIndicator{
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
			w.sendMessageToUser(user, message)
		}
	}
}

// cleanupTypingIndicators removes stale typing indicators
func (w *webSocketManager) cleanupTypingIndicators() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		w.mutex.Lock()
		now := time.Now()
		for groupID, typingUsers := range w.typingUsers {
			for userEmail, lastTyping := range typingUsers {
				if now.Sub(lastTyping) > 10*time.Second {
					delete(typingUsers, userEmail)
					// Notify others that user stopped typing
					go w.broadcastTypingIndicator(groupID, userEmail, false)
				}
			}
			if len(typingUsers) == 0 {
				delete(w.typingUsers, groupID)
			}
		}
		w.mutex.Unlock()
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
			w.sendMessageToUser(user, message)
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
			w.sendMessageToUser(user, message)
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
			Type:      NewMessageType,
			GroupID:   msg.GroupID.Hex(),
			UserEmail: msg.SenderEmail,
			Data:      msg,
			Timestamp: time.Now(),
		}

		if err := w.sendMessageToUser(userEmail, event); err == nil {
			groupMsgMap[msg.GroupID] = append(groupMsgMap[msg.GroupID], msg.ID)
		}
	}

	for groupID, messageIDs := range groupMsgMap {
		if err := w.repo.MarkMessagesAsDelivered(ctx, groupID, messageIDs, userEmail); err != nil {
			log.Printf("Failed to mark messages delivered for group %s: %v", groupID.Hex(), err)
		}
	}
}
