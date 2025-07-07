package case_assign

import (
	"errors"

	"github.com/google/uuid"
)

func NewCaseAssignmentService(
	repo CaseAssignmentRepoInterface,
) *CaseAssignmentService {
	return &CaseAssignmentService{repo: repo}
}

// This method now takes the assigner's role directly
func (s *CaseAssignmentService) AssignUserToCase(assignerRole string, assigneeID, caseID uuid.UUID, role string) error {
	if assignerRole != "Admin" {
		return errors.New("forbidden: admin privileges required")
	}
	return s.repo.AssignRole(assigneeID, caseID, role)
}
