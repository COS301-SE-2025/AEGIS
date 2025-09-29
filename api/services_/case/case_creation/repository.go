package case_creation

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormCaseRepository struct {
	db *gorm.DB
}

func NewGormCaseRepository(db *gorm.DB) *GormCaseRepository {
	return &GormCaseRepository{db: db}
}

func (r *GormCaseRepository) CreateCase(c *Case) error {
	return r.db.Create(c).Error
}

// GetCaseByID fetches a case by its ID
func (r *GormCaseRepository) GetCaseByID(ctx context.Context, id uuid.UUID) (*Case, error) {
	var c Case
	err := r.db.WithContext(ctx).First(&c, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}
