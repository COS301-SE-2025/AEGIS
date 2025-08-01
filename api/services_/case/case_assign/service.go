package case_assign

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func NewCaseAssignmentService(
	repo CaseAssignmentRepoInterface,
	adminChecker AdminChecker,
	userRepo UserRepo,
) *CaseAssignmentService {
	return &CaseAssignmentService{repo: repo, adminChecker: adminChecker, userRepo: userRepo}
}

// This method now takes the assigner's role directly
func (s *CaseAssignmentService) AssignUserToCase(
	assignerRole string,
	assigneeID uuid.UUID,
	caseID uuid.UUID,
	assignerID uuid.UUID,
	role string,
	tenantID uuid.UUID, // Pass the tenant ID of the assigner
) error {
	if assignerRole != "DFIR Admin" {
		return errors.New("forbidden: admin privileges required")
	}

	// // Fetch both users
	// assigner, err := s.userRepo.GetUserByID(assignerID)
	// if err != nil {
	// 	return err
	// }
	// assignee, err := s.userRepo.GetUserByID(assigneeID)
	// if err != nil {
	// 	return err
	// }

	// // Ensure both belong to the same tenant
	// if assigner.TenantID != assignee.TenantID {
	// 	return errors.New("cannot assign users from a different tenant")
	// }

	return s.repo.AssignRole(assigneeID, caseID, role, tenantID)
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
