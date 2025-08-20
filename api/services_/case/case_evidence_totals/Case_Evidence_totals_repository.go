package case_evidence_totals

import (
    "gorm.io/gorm"
    
)

type CountCasesEvidenceRepo interface {
    CountCases(string,[]string) (int64, error)
	CountEvidence(string) (int64, error)
}

type caseEviRepository struct {
    db *gorm.DB
}

func NewCaseEviRepository(db *gorm.DB) CountCasesEvidenceRepo {
    return &caseEviRepository{db: db}
}



func (r *caseEviRepository) CountCases(userID string, statuses []string) (int64, error) {
	var count int64
	err := r.db.Model(&Case{}).
		Where("status IN ?", statuses).
		Where("created_by = ?", userID). // Or use a join if team-based access
		Count(&count).Error
	return count, err
}

func (r *caseEviRepository) CountEvidence(userID string) (int64, error) {
	var count int64
	err := r.db.Model(&EvidenceDTO{}).
		Where("uploaded_by = ?", userID).
		Count(&count).Error
	return count, err
}
