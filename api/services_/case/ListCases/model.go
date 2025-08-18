// Folder: services/ListCases/

// File: model.go
package ListCases

import (
	"time"

	"github.com/google/uuid"
)

type Case struct {
	ID                 uuid.UUID `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Title              string    `gorm:"column:title;not null" json:"title"`
	Description        string    `gorm:"column:description" json:"description"`
	Status             string    `gorm:"column:status;type:case_status;default:'open'" json:"status"`
	Priority           string    `gorm:"column:priority;type:case_priority;default:'medium'" json:"priority"`
	InvestigationStage string    `gorm:"column:investigation_stage;type:investigation_stage;default:'analysis'" json:"investigation_stage"`
	CreatedBy          uuid.UUID `gorm:"column:created_by;type:uuid;not null" json:"created_by"`
	TeamName           string    `gorm:"column:team_name;type:text;not null" json:"team_name"`
	CreatedAt          time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	TenantID           uuid.UUID `gorm:"column:tenant_id;type:uuid;not null" json:"tenant_id"`
}

type CaseFilter struct {
	Status    string
	Priority  string
	CreatedBy string
	TeamName  string // ← new
	TitleTerm string
	SortBy    string
	SortOrder string
	TenantID  uuid.UUID // ← new
}

// Service provides operations for listing and filtering cases.
type Service struct {
	repo CaseQueryRepository
}

func (Case) TableName() string {
	return "cases"
}
