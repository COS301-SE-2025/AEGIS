package case_assign

import (
	//"aegis-api/db"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormCaseAssignmentRepo struct {
	db *gorm.DB
}

func NewGormCaseAssignmentRepo(database *gorm.DB) *GormCaseAssignmentRepo {
	return &GormCaseAssignmentRepo{db: database}
}

func (r *GormCaseAssignmentRepo) AssignRole(userID, caseID uuid.UUID, role string, tenantID uuid.UUID) error {
	entry := CaseUserRole{
		UserID:     userID,
		CaseID:     caseID,
		Role:       role,
		AssignedAt: time.Now(),
		TenantID:   tenantID, // Ensure you pass the tenant ID here
	}
	return r.db.Create(&entry).Error
}

func (r *GormCaseAssignmentRepo) UnassignRole(userID, caseID uuid.UUID) error {
	return r.db.Where("user_id = ? AND case_id = ?", userID, caseID).Delete(&CaseUserRole{}).Error
}
