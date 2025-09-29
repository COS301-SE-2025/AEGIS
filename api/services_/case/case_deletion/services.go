package case_deletion

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"aegis-api/services_/case/case_creation"
)

// CaseRepository defines persistence operations for case deletion
type CaseRepository interface {
	ArchiveCase(ctx context.Context, id uuid.UUID) error
}

type GormCaseRepository struct {
	db *gorm.DB
}

func NewGormCaseRepository(db *gorm.DB) *GormCaseRepository {
	return &GormCaseRepository{db: db}
}

// ArchiveCase sets status to 'archived' and updates UpdatedAt
func (r *GormCaseRepository) ArchiveCase(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&case_creation.Case{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     "archived",
		"updated_at": time.Now(),
	}).Error
}

// Service handles business logic for case deletion
type Service struct {
	repo CaseRepository
}

func NewCaseDeletionService(repo CaseRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ArchiveCase(ctx context.Context, caseID string) error {
	id, err := uuid.Parse(caseID)
	if err != nil {
		return err
	}
	return s.repo.ArchiveCase(ctx, id)
}
