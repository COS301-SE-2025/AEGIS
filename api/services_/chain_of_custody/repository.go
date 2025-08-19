package chain_of_custody

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ChainOfCustodyRepository interface {
	Create(ctx context.Context, custody *ChainOfCustody) error
	Update(ctx context.Context, custody *ChainOfCustody) error
	GetByEvidenceID(ctx context.Context, evidenceID uuid.UUID) ([]ChainOfCustody, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ChainOfCustody, error)
}

type chainOfCustodyRepo struct {
	db *gorm.DB
}

func NewChainOfCustodyRepository(db *gorm.DB) ChainOfCustodyRepository {
	return &chainOfCustodyRepo{db: db}
}

func (r *chainOfCustodyRepo) Create(ctx context.Context, custody *ChainOfCustody) error {
	return r.db.WithContext(ctx).Create(custody).Error

}

func (r *chainOfCustodyRepo) Update(ctx context.Context, custody *ChainOfCustody) error {
	return r.db.WithContext(ctx).Save(custody).Error
}

func (r *chainOfCustodyRepo) GetByEvidenceID(ctx context.Context, evidenceID uuid.UUID) ([]ChainOfCustody, error) {
	var entries []ChainOfCustody
	err := r.db.WithContext(ctx).Where("evidence_id = ?", evidenceID).Order("acquisition_date").Find(&entries).Error
	return entries, err
}

func (r *chainOfCustodyRepo) GetByID(ctx context.Context, id uuid.UUID) (*ChainOfCustody, error) {
	var entry ChainOfCustody
	err := r.db.WithContext(ctx).First(&entry, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &entry, nil
}
