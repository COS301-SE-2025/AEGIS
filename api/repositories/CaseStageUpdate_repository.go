package repositories

import (
	"aegis-api/models"
	"aegis-api/db"
	"gorm.io/gorm"
	 "github.com/google/uuid"

)

//interface for updating the case stage, just defines the method signature
type UpdateCaseStageRepository interface {
    UpdateStage(caseID string, newStage models.InvestigationStage) error
	CaseExists(caseID uuid.UUID) (bool, error)
}

//
type caseRepo struct {
    DB *gorm.DB
}

func NewCaseRepo() UpdateCaseStageRepository{
    return &caseRepo{}
}

func (r *caseRepo) UpdateStage(caseID string, newStage models.InvestigationStage) error {
   return db.DB.Model(&models.Case{}).
	Where("id = ?", caseID).
	Update("InvestigationStage", newStage).Error
}


func (r *caseRepo) CaseExists(caseID uuid.UUID) (bool, error) {
	var count int64
	err := db.DB.Model(&models.Case{}).
		Where("id = ?", caseID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
