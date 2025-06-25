package ListClosedCases

import (
	"context"
)

type ListClosedCasesRepository interface {
	GetClosedCasesByUserID(ctx context.Context, userID string) ([]ClosedCase, error)
}
