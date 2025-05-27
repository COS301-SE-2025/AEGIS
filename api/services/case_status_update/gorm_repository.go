package case_status_update

import (
	"aegis-api/db"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormCaseStatusRepository struct {
	db *gorm.DB
}

func NewGormCaseStatusRepository() *GormCaseStatusRepository {
	return &GormCaseStatusRepository{db: db.DB}
}

func (r *GormCaseStatusRepository) UpdateStatus(caseID string, newStatus string) error {
	id, err := uuid.Parse(caseID)
	if err != nil {
		return err
	}

	return r.db.Model(&Case{}).Where("id = ?", id).Update("status", newStatus).Error
}
