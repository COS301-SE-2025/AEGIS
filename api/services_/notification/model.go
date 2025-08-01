package notification

import "time"

type Notification struct {
	ID        string    `gorm:"primaryKey;type:uuid" json:"id"`
	UserID    string    `gorm:"not null" json:"user_id"`
	TenantID  string    `gorm:"not null" json:"tenant_id"`
	TeamID    string    `gorm:"not null" json:"team_id"`
	Title     string    `gorm:"not null" json:"title"`
	Message   string    `gorm:"not null" json:"message"`
	Timestamp time.Time `gorm:"autoCreateTime" json:"timestamp"`
	Read      bool      `gorm:"default:false" json:"read"`
	Archived  bool      `gorm:"default:false" json:"archived"`
}
