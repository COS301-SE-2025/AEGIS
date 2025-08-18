package chain_of_custody

import (
	"context"

	"github.com/google/uuid"
)

type chainOfCustodyService struct {
	repo ChainOfCustodyRepository
}

func NewChainOfCustodyService(repo ChainOfCustodyRepository) ChainOfCustodyService {
	return &chainOfCustodyService{repo: repo}
}

func (s *chainOfCustodyService) AddEntry(ctx context.Context, custody *ChainOfCustody) error {
	custody.ID = uuid.New()
	return s.repo.Create(ctx, custody)
}

func (s *chainOfCustodyService) UpdateEntry(ctx context.Context, custody *ChainOfCustody) error {
	return s.repo.Update(ctx, custody)
}

func (s *chainOfCustodyService) GetEntries(ctx context.Context, evidenceID uuid.UUID) ([]ChainOfCustody, error) {
	return s.repo.GetByEvidenceID(ctx, evidenceID)
}

func (s *chainOfCustodyService) GetEntry(ctx context.Context, id uuid.UUID) (*ChainOfCustody, error) {
	return s.repo.GetByID(ctx, id)
}
