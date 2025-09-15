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
	TeamID     uuid.UUID
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

type Case struct {
	ID                 uuid.UUID `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Title              string    `gorm:"column:title;not null" json:"title"`
	Description        string    `gorm:"column:description" json:"description"`
	Status             string    `gorm:"column:status;type:case_status;default:'open'" json:"status"`
	Priority           string    `gorm:"column:priority;type:case_priority;default:'medium'" json:"priority"`
	InvestigationStage string    `gorm:"column:investigation_stage;type:investigation_stage;default:'analysis'" json:"investigation_stage"`
	CreatedBy          uuid.UUID `gorm:"column:created_by;type:uuid;not null" json:"created_by"`
	TeamName           string    `gorm:"column:team_name;type:text;not null" json:"team_name"`
	TenantID           uuid.UUID `gorm:"column:tenant_id;type:uuid;not null" json:"tenant_id"`
	TeamID             uuid.UUID `gorm:"column:team_id;type:uuid" json:"team_id"`
	CreatedAt          time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}
