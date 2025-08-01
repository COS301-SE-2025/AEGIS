package case_assign

import (
	"aegis-api/pkg/websocket"
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
	assignerID uuid.UUID,
	assigneeID uuid.UUID,
	caseID uuid.UUID,
	role string,
	tenantID uuid.UUID,
) error {
	if assignerRole != "DFIR Admin" {
		return errors.New("forbidden: admin privileges required")
	}

	// Fetch users
	assigner, err := s.userRepo.GetUserByID(assignerID)
	if err != nil {
		return err
	}
	assignee, err := s.userRepo.GetUserByID(assigneeID)
	if err != nil {
		return err
	}

	// Ensure same tenant
	if assigner.TenantID != assignee.TenantID {
		return errors.New("cannot assign users from a different tenant")
	}

	// Assign the role
	if err := s.repo.AssignRole(assigneeID, caseID, role); err != nil {
		return err
	}

	// ✅ Trigger WebSocket + DB notification
	go websocket.NotifyUser(
		s.hub,
		s.notificationService,
		assigneeID.String(),
		tenantID.String(),
		assignee.TeamID.String(),
		"Assigned to Case",
		"You have been assigned to a new case.",
	)

	return nil
}

func (s *CaseAssignmentService) UnassignUserFromCase(ctx *gin.Context, assigneeID, caseID uuid.UUID) error {
	isAdmin, err := s.adminChecker.IsAdminFromContext(ctx)
	if err != nil {
		return err
	}
	if !isAdmin {
		return errors.New("forbidden: admin privileges required")
	}

	// Perform unassignment
	if err := s.repo.UnassignRole(assigneeID, caseID); err != nil {
		return err
	}

	// Fetch the user for tenant/team info
	user, err := s.userRepo.GetUserByID(assigneeID)
	if err == nil {
		// ✅ Trigger WebSocket + DB notification
		go websocket.NotifyUser(
			s.hub,
			s.notificationService,
			assigneeID.String(),
			user.TenantID.String(),
			user.TeamID.String(),
			"Unassigned from Case",
			"You have been unassigned from a case.",
		)
	}

	return nil
}
