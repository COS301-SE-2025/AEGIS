package websocket

import (
	"aegis-api/pkg/chatModels"
	"aegis-api/services_/notification"
	"log"
	"time"

	"github.com/google/uuid"
)

func (h *Hub) BroadcastNotificationToUser(tenantID, teamID, userID string, notif notification.Notification) error {
	notif.TenantID = tenantID
	notif.TeamID = teamID

	msg := chatModels.WebSocketEvent{
		Type:      chatModels.EventNotification,
		Payload:   notif,
		UserEmail: userID,
		Timestamp: notif.Timestamp,
	}

	return h.SendToUser(userID, msg)
}

func NotifyUser(
	hub *Hub,
	service notification.NotificationServiceInterface,
	userID, tenantID, teamID, title, message string,
) error {
	notif := notification.Notification{
		ID:        uuid.New().String(),
		UserID:    userID,
		TenantID:  tenantID,
		TeamID:    teamID,
		Title:     title,
		Message:   message,
		Timestamp: time.Now(),
		Read:      false,
		Archived:  false,
	}

	if err := service.SaveNotification(&notif); err != nil {
		return err
	}

	event := chatModels.WebSocketEvent{
		Type:      chatModels.EventNotification,
		Payload:   notif,
		UserEmail: userID,
		Timestamp: notif.Timestamp,
	}

	if err := hub.SendToUser(userID, event); err != nil {
		// 👇 Do not treat “offline” as an error; we’ve persisted already
		if err == ErrNoActiveConnection {
			return nil
		}
		return err
	}
	return nil
}

func (h *Hub) syncNotificationsOnConnect(userID, tenantID, teamID string) {
	if h.NotificationService == nil {
		return
	}

	notifs, err := h.NotificationService.GetNotificationsForUser(tenantID, teamID, userID)
	if err != nil {
		log.Printf("❌ Failed fetching unread notifications for %s: %v", userID, err)
		return
	}

	// Option A: batch event
	evt := chatModels.WebSocketEvent{
		Type:      chatModels.EventNotificationSync, // define this in your chatModels package
		Payload:   notifs,                           // []notification.Notification
		UserEmail: userID,
		Timestamp: time.Now(),
	}
	if err := h.SendToUser(userID, evt); err != nil && err != ErrNoActiveConnection {
		log.Printf("❌ Failed to send notification sync to %s: %v", userID, err)
	}

	// Option B (alternative): send count only
	// unread := 0
	// for _, n := range notifs { if !n.Read && !n.Archived { unread++ } }
	// _ = h.SendToUser(userID, chatModels.WebSocketEvent{ Type: chatModels.EventUnreadCount, Payload: unread, ... })
}
