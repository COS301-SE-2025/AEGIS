package case_creation

import (
	"context"

	"github.com/google/uuid"
)

type CaseRepository interface {
	CreateCase(c *Case) error
	GetCaseByID(ctx context.Context, id uuid.UUID) (*Case, error)
}
