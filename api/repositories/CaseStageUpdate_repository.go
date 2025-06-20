package repositories

import (
	"aegis-api/models"
	"aegis-api/db"
	"gorm.io/gorm"
)

type UpdateCaseStageRepository interface {
    UpdateStage(caseID string, newStage models.InvestigationStage) error
}

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
