package models

import (
	"time"

	"github.com/google/uuid"
)

type InvestigationStage string

const (
	StageAnalysis     InvestigationStage = "analysis"
	StageResearch     InvestigationStage = "research"
	StageEvaluation   InvestigationStage = "evaluation"
	StageFinalization InvestigationStage = "finalization"
)

type Case struct {
	ID                 uuid.UUID          `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Title              string             `gorm:"not null"`
	Description        *string
	Status             string             `gorm:"default:'open'"`
	InvestigationStage InvestigationStage `gorm:"default:'analysis'"`
	Priority           string             `gorm:"default:'medium'"`
	TeamName           string             `gorm:"not null"`
	CreatedBy          uuid.UUID          `gorm:"type:uuid"`
	CreatedAt          time.Time          `gorm:"autoCreateTime"`
		Tags []*Tag `gorm:"many2many:case_tags;constraint:OnDelete:CASCADE;"`
}

func (s InvestigationStage) IsValid() bool {
	switch s {
	case StageAnalysis, StageResearch, StageEvaluation, StageFinalization:
		return true
	default:
		return false
	}
}

type CaseTag struct {
	CaseID uuid.UUID `gorm:"type:uuid;primaryKey"`
	TagID  int       `gorm:"primaryKey"`
}

