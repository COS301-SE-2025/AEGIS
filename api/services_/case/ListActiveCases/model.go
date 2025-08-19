package ListActiveCases

import (
	"time"

	"github.com/google/uuid"
)

// ActiveCase represents a case that is currently active
type ActiveCase struct {
	ID                 uuid.UUID `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Title              string    `gorm:"column:title;type:varchar(255);not null" json:"title"`
	Description        string    `gorm:"column:description;type:text" json:"description"`
	Status             string    `gorm:"column:status;type:varchar(50);not null" json:"status"`
	InvestigationStage string    `gorm:"column:investigation_stage;type:varchar(50);not null" json:"investigation_stage"`
	Priority           string    `gorm:"column:priority;type:varchar(50);not null" json:"priority"`
	TeamName           string    `gorm:"column:team_name;type:text;not null" json:"team_name"`
	TeamID             uuid.UUID `gorm:"column:team_id;type:uuid;not null" json:"team_id"`     // ✅ Added for multi-tenancy
	TenantID           uuid.UUID `gorm:"column:tenant_id;type:uuid;not null" json:"tenant_id"` // ✅ Added for multi-tenancy
	CreatedBy          uuid.UUID `gorm:"column:created_by;type:uuid;not null" json:"created_by"`
	CreatedAt          time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

type RequestDTO struct {
	// UserID is the ID of the user for whom to list active cases.
	UserID string `json:"user_id" validate:"required,uuid"`
}

type ResponseDTO struct {
	// Cases is a list of active cases.
	Cases []ActiveCase `json:"cases"`
}
