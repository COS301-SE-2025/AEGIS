package case_status_update

import (
	"github.com/google/uuid"
	"time"
)

type Case struct {
	ID                 uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primary_key" json:"id"`
	Title              string    `json:"title"`
	Description        string    `json:"description"`
	Status             string    `gorm:"type:case_status;default:'open'" json:"status"`
	Priority           string    `gorm:"type:case_priority;default:'medium'" json:"priority"`
	InvestigationStage string    `gorm:"type:investigation_stage;default:'analysis'" json:"investigation_stage"`
	CreatedBy          uuid.UUID `gorm:"type:uuid" json:"created_by"`
	CreatedAt          time.Time `gorm:"autoCreateTime" json:"created_at"`
}
