package case_assign

import (
	"errors"
	"github.com/google/uuid"
)

type AdminChecker interface {
	IsAdmin(userID uuid.UUID) (bool, error)
}

type CaseAssignmentRepoInterface interface {
	AssignRole(userID, caseID uuid.UUID, role string) error
	UnassignRole(userID, caseID uuid.UUID) error
	IsAdmin(userID uuid.UUID) (bool, error) // integrated for simplicity
}

type CaseAssignmentService struct {
	repo CaseAssignmentRepoInterface
}

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