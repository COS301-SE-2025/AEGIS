package case_creation

import (
    "github.com/google/uuid"
    "time"
)

// Case represents a case record in the database.
type Case struct {
    ID                 uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
    Title              string    `gorm:"not null" json:"title"`
    Description        string    `json:"description"`
    Status             string    `gorm:"type:case_status;default:'open'" json:"status"`
    Priority           string    `gorm:"type:case_priority;default:'medium'" json:"priority"`
    InvestigationStage string    `gorm:"type:investigation_stage;default:'analysis'" json:"investigation_stage"`
    CreatedBy          uuid.UUID `gorm:"type:uuid;not null" json:"created_by"`
    TeamName           string    `gorm:"type:text;not null" json:"team_name"`
    CreatedAt          time.Time `gorm:"autoCreateTime" json:"created_at"`
}
