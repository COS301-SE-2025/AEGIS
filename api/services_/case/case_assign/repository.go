package case_assign

import (
	//"aegis-api/db"
	"fmt"
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

func (r *GormCaseAssignmentRepo) AssignRole(userID, caseID uuid.UUID, role string, tenantID, teamID uuid.UUID) error {
	entry := CaseUserRole{
		UserID:     userID,
		CaseID:     caseID,
		Role:       role,
		AssignedAt: time.Now(),
		TenantID:   tenantID, // Ensure you pass the tenant ID here
		TeamID:     teamID,   // Ensure you pass the team ID here
	}
	return r.db.Create(&entry).Error
}

func (r *GormCaseAssignmentRepo) UnassignRole(userID, caseID uuid.UUID) error {
	return r.db.Where("user_id = ? AND case_id = ?", userID, caseID).Delete(&CaseUserRole{}).Error
}

func (r *GormCaseAssignmentRepo) GetCaseByID(caseID uuid.UUID, caseDetails *Case) error {
	err := r.db.Where("id = ?", caseID).First(caseDetails).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("case with ID %s not found", caseID)
		}
		return fmt.Errorf("failed to retrieve case: %w", err)
	}
	return nil
}
