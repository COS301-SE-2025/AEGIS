package metadata

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormRepository is a concrete implementation of the Repository interface using GORM.
type GormRepository struct {
	db *gorm.DB
}

// NewGormRepository creates a new instance of the repository using a GORM DB instance.
func NewGormRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

// SaveEvidence inserts a new evidence record into the database, including metadata stored as JSONB.
func (r *GormRepository) SaveEvidence(e *Evidence) error {
	return r.db.Create(e).Error
}

// FindEvidenceByID fetches a single evidence record by ID.
func (r *GormRepository) FindEvidenceByID(id uuid.UUID) (*Evidence, error) {
	var evidence Evidence
	err := r.db.First(&evidence, "id = ?", id).Error
	return &evidence, err
}

// FindEvidenceByCaseID retrieves all evidence records associated with a specific case ID.
func (r *GormRepository) FindEvidenceByCaseID(caseID uuid.UUID) ([]Evidence, error) {
	var evidences []Evidence
	err := r.db.Where("case_id = ?", caseID).Find(&evidences).Error

	return evidences, err
}

// AppendEvidenceLog inserts a new evidence log record (append-only)
func (r *GormRepository) AppendEvidenceLog(log *EvidenceLog) error {
	return r.db.Create(log).Error
}

func (r *GormRepository) GetLastEvidenceLog(evidenceID uuid.UUID) (*EvidenceLog, error) {
	var lastLog EvidenceLog
	err := r.db.Where("evidence_id = ?", evidenceID).Order("created_at DESC").First(&lastLog).Error
	if err != nil {
		return nil, err
	}
	return &lastLog, nil
}
