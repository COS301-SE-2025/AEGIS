package update_case_Investigation_stage

import (
	"aegis-api/db"
	"aegis-api/services_/case/case_creation"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// interface for updating the case stage, just defines the method signature
type UpdateCaseStageRepository interface {
	UpdateStage(caseID string, newStage InvestigationStage) error
	CaseExists(caseID uuid.UUID) (bool, error)
}

type caseRepo struct {
	DB *gorm.DB
}

func NewCaseRepo() UpdateCaseStageRepository {
	return &caseRepo{}
}

func (r *caseRepo) UpdateStage(caseID string, newStage InvestigationStage) error {
	// Also update progress based on newStage
	// Import GetProgressForStage from case_creation
	progress := case_creation.GetProgressForStage(string(newStage))
	return db.DB.Model(&Case{}).
		Where("id = ?", caseID).
		Updates(map[string]interface{}{
			"InvestigationStage": newStage,
			"Progress":           progress,
		}).Error
}

func (r *caseRepo) CaseExists(caseID uuid.UUID) (bool, error) {
	var count int64
	err := db.DB.Model(&Case{}).
		Where("id = ?", caseID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
