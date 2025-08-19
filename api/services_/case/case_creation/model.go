package case_creation

import (
	"time"

	"github.com/google/uuid"
)

// Case represents a case record in the database.
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

type CreateCaseRequest struct {
	Title              string    `json:"title" binding:"required"`
	Description        string    `json:"description"`
	Status             string    `json:"status"`
	Priority           string    `json:"priority"`
	InvestigationStage string    `json:"investigation_stage"`
	CreatedBy          uuid.UUID `json:"created_by" binding:"required"`
	TeamName           string    `json:"team_name" binding:"required"`
	TenantID           uuid.UUID `json:"tenant_id" `
	TeamID             uuid.UUID `json:"team_id"`
}
