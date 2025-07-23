package case_assign

import (
	"time"

	"github.com/google/uuid"
)

type CaseAssignmentService struct {
	repo         CaseAssignmentRepoInterface
	adminChecker AdminChecker
}
type CaseUserRole struct {
	UserID     uuid.UUID
	CaseID     uuid.UUID
	Role       string // One of your user_role ENUM values
	AssignedAt time.Time
}
