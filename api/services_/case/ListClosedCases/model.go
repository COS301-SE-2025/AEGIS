package ListClosedCases

import (
	"time"

	"github.com/google/uuid"
)

type ClosedCase struct {
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

type ListClosedCasesRequest struct {
	UserID string `json:"user_id"`
}

type ListClosedCasesResponse struct {
	ClosedCases []ClosedCase `json:"closed_cases"`
}
