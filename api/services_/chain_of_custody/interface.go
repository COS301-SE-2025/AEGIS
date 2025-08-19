package chain_of_custody

import (
	"context"

	"github.com/google/uuid"
)

type ChainOfCustodyService interface {
	AddEntry(ctx context.Context, custody *ChainOfCustody) error
	UpdateEntry(ctx context.Context, custody *ChainOfCustody) error
	GetEntries(ctx context.Context, evidenceID uuid.UUID) ([]ChainOfCustody, error)
	GetEntry(ctx context.Context, id uuid.UUID) (*ChainOfCustody, error)
}
