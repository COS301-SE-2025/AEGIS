package update_case

import (
	"context"
)

// UpdateCaseRepository defines how we update cases in DB
type UpdateCaseRepository interface {
	UpdateCase(ctx context.Context, req *UpdateCaseRequest) error
}
