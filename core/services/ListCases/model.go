// Folder: services/ListCases/

// File: model.go
package ListCases

import (
	"github.com/google/uuid"
	"time"
)

type Case struct {
	ID                  uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Title               string
	Description         string
	Status              string
	InvestigationStage  string
	Priority            string
	CreatedBy           uuid.UUID
	CreatedAt           time.Time `gorm:"autoCreateTime"`
}
type CaseFilter struct {
	Status     string
	Priority   string
	CreatedBy  string
	TitleTerm  string
	SortBy     string
	SortOrder  string
}

func (Case) TableName() string {
	return "cases"
}
