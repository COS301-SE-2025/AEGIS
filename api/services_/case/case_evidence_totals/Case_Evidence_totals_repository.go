package case_evidence_totals

import (
    "gorm.io/gorm"
    
)

type CountCasesEvidenceRepo interface {
    CountCases() (int64, error)
	CountEvidence() (int64, error)
}

type caseEviRepository struct {
    db *gorm.DB
}

func NewCaseEviRepository(db *gorm.DB) CountCasesEvidenceRepo {
    return &caseEviRepository{db: db}
}

func (r *caseEviRepository) CountCases() (int64, error) {
    var count int64
    err := r.db.Model(&Case{}).Count(&count).Error
    return count, err
}

func (r *caseEviRepository) CountEvidence() (int64, error) {
    var count int64
    err := r.db.Model(&EvidenceDTO{}).Count(&count).Error
    return count, err
}