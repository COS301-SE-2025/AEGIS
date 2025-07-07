package ListActiveCases

import (
	"time"

	"github.com/google/uuid"
)

// ActiveCase represents a case that is currently active
type ActiveCase struct {
	ID                 uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Title              string    `gorm:"type:varchar(255);not null" json:"title"`
	Description        string    `gorm:"type:text" json:"description"`
	Status             string    `gorm:"type:varchar(50);not null" json:"status"`
	InvestigationStage string    `gorm:"type:varchar(50);not null" json:"investigation_stage"`
	Priority           string    `gorm:"type:varchar(50);not null" json:"priority"`
	TeamName           string    `gorm:"type:text;not null"`
	CreatedBy          uuid.UUID `gorm:"type:uuid;not null" json:"created_by"`
	CreatedAt          time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type RequestDTO struct {
	// UserID is the ID of the user for whom to list active cases.
	UserID string `json:"user_id" validate:"required,uuid"`
}

type ResponseDTO struct {
	// Cases is a list of active cases.
	Cases []ActiveCase `json:"cases"`
}
