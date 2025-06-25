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

func (s *CaseAssignmentService) AssignUserToCase(assignerID, assigneeID, caseID uuid.UUID, role string) error {
	ok, err := s.repo.IsAdmin(assignerID)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("forbidden: admin privileges required")
	}
	return s.repo.AssignRole(assigneeID, caseID, role)
}

func (s *CaseAssignmentService) UnassignUserFromCase(assignerID, userID, caseID uuid.UUID) error {
	ok, err := s.repo.IsAdmin(assignerID)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("forbidden: admin privileges required")
	}
	return s.repo.UnassignRole(userID, caseID)
}
