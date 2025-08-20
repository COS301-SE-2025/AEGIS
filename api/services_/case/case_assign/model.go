package case_assign

import (
	"aegis-api/pkg/websocket"
	"aegis-api/services_/notification"
	"time"

	"github.com/google/uuid"
)

type CaseAssignmentService struct {
	repo                CaseAssignmentRepoInterface
	adminChecker        AdminChecker
	userRepo            UserRepo
	notificationService notification.NotificationServiceInterface
	hub                 *websocket.Hub
}
type CaseUserRole struct {
	UserID     uuid.UUID
	CaseID     uuid.UUID
	Role       string // One of your user_role ENUM values
	AssignedAt time.Time
	TenantID   uuid.UUID
}
type User struct {
	ID        uuid.UUID
	FullName  string
	Email     string
	Role      string    // One of your user_role ENUM values
	TenantID  uuid.UUID `gorm:"type:uuid;not null" json:"tenant_id"`
	CreatedAt time.Time
	UpdatedAt time.Time
	TeamID    uuid.UUID // Optional, if users can belong to teams
}
