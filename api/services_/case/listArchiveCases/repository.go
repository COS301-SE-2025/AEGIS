package listArchiveCases

import (
	"gorm.io/gorm"
)

type ArchiveCaseRepository struct {
	db *gorm.DB
}

func NewArchiveCaseRepository(db *gorm.DB) *ArchiveCaseRepository {
	return &ArchiveCaseRepository{db: db}
}

func (r *ArchiveCaseRepository) ListArchivedCases(userID, tenantID, teamID string) ([]ArchivedCase, error) {
	var cases []ArchivedCase
	// Query from the main 'cases' table, using 'created_by' instead of 'user_id'
	err := r.db.Table("cases").Where("status = ? AND created_by = ? AND tenant_id = ? AND team_id = ?", "archived", userID, tenantID, teamID).Find(&cases).Error
	return cases, err
}
