package notification

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NotificationService struct {
	DB *gorm.DB
}

// NewNotificationService creates a new NotificationService with the given DB
func NewNotificationService(db *gorm.DB) *NotificationService {
	return &NotificationService{
		DB: db,
	}
}

// SaveNotification stores a new notification in the database
func (s *NotificationService) SaveNotification(n *Notification) error {
	n.ID = uuid.New().String()
	n.Timestamp = time.Now()
	return s.DB.Create(n).Error
}

// GetNotificationsForUser fetches notifications by tenant, team, and user
func (s *NotificationService) GetNotificationsForUser(tenantID, teamID, userID string) ([]Notification, error) {
	var notifications []Notification
	err := s.DB.
		Where("tenant_id = ? AND team_id = ? AND user_id = ?", tenantID, teamID, userID).
		Order("timestamp DESC").
		Find(&notifications).Error
	return notifications, err
}

// MarkAsRead marks specific notifications as read
func (s *NotificationService) MarkAsRead(notificationIDs []string) error {
	return s.DB.
		Model(&Notification{}).
		Where("id IN ?", notificationIDs).
		Update("read", true).Error
}

// ArchiveNotifications marks specific notifications as archived
func (s *NotificationService) ArchiveNotifications(notificationIDs []string) error {
	return s.DB.
		Model(&Notification{}).
		Where("id IN ?", notificationIDs).
		Update("archived", true).Error
}

// DeleteNotifications deletes selected notifications
func (s *NotificationService) DeleteNotifications(notificationIDs []string) error {
	return s.DB.
		Where("id IN ?", notificationIDs).
		Delete(&Notification{}).Error
}
