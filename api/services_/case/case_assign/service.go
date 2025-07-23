package case_assign

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func NewCaseAssignmentService(repo CaseAssignmentRepoInterface, adminChecker AdminChecker) *CaseAssignmentService {
	return &CaseAssignmentService{repo: repo, adminChecker: adminChecker}
}

// This method now takes the assigner's role directly
func (s *CaseAssignmentService) AssignUserToCase(assignerRole string, assigneeID, caseID uuid.UUID, role string) error {
	if assignerRole != "Admin" {
		return errors.New("forbidden: admin privileges required")
	}
	return s.repo.AssignRole(assigneeID, caseID, role)
}

func (s *CaseAssignmentService) UnassignUserFromCase(ctx *gin.Context, assigneeID, caseID uuid.UUID) error {
	isAdmin, err := s.adminChecker.IsAdminFromContext(ctx)
	if err != nil {
		return err
	}
	if !isAdmin {
		return errors.New("forbidden: admin privileges required")
	}
	return s.repo.UnassignRole(assigneeID, caseID)
}
