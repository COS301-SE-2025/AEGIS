package ListClosedCases

import (
	"context"
)

type ListClosedCasesRepository interface {
	GetClosedCasesByUserID(ctx context.Context, userID, tenantID, teamID string) ([]ClosedCase, error)
}
