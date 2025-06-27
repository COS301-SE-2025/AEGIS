package ListActiveCases

import (
	"context"
)

type ActiveCaseQueryRepository interface {
	// QueryActiveCases retrieves active cases based on the provided filter.
	GetActiveCasesByUserID(ctx context.Context, userID string) ([]ActiveCase, error)
}
