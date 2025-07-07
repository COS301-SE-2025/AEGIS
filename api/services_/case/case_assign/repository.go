package case_assign

import (
	//"aegis-api/db"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormCaseAssignmentRepo struct {
	db *gorm.DB
}

func NewGormCaseAssignmentRepo(database *gorm.DB) *GormCaseAssignmentRepo {
	return &GormCaseAssignmentRepo{db: database}
}

func (r *GormCaseAssignmentRepo) AssignRole(userID, caseID uuid.UUID, role string) error {
	entry := CaseUserRole{
		UserID: userID,
		CaseID: caseID,
		Role:   role,
	}
	return r.db.Create(&entry).Error
}
