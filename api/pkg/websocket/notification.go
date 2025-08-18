package websocket

import (
	"aegis-api/services_/chat"
	"aegis-api/services_/notification"
	"time"

	"github.com/google/uuid"
)

func (h *Hub) BroadcastNotificationToUser(tenantID, teamID, userID string, notif notification.Notification) error {
	notif.TenantID = tenantID
	notif.TeamID = teamID

	msg := chat.WebSocketEvent{
		Type:      chat.EventNotification,
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

	event := chat.WebSocketEvent{
		Type:      chat.EventNotification,
		Payload:   notif,
		UserEmail: userID,
		Timestamp: notif.Timestamp,
	}

	return hub.SendToUser(userID, event)
}
