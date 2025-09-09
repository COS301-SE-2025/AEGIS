package case_assign

import (
	"aegis-api/pkg/websocket"
	"aegis-api/services_/notification"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func NewCaseAssignmentService(
	repo CaseAssignmentRepoInterface,
	adminChecker AdminChecker,
	userRepo UserRepo,
	notificationService notification.NotificationServiceInterface,
	hub *websocket.Hub,
) *CaseAssignmentService {
	return &CaseAssignmentService{
		repo:                repo,
		adminChecker:        adminChecker,
		userRepo:            userRepo,
		notificationService: notificationService,
		hub:                 hub,
	}
}

// AFTER
func (s *CaseAssignmentService) AssignUserToCase(
	assignerRole string,
	assigneeID uuid.UUID,
	caseID uuid.UUID,
	assignerID uuid.UUID,
	role string,
	tenantID uuid.UUID,
	teamID uuid.UUID,
) error {
	if assignerRole != "DFIR Admin" {
		return errors.New("forbidden: admin privileges required")
	}
	if s.notificationService == nil {
		return errors.New("notification service not initialized")
	}
	if s.hub == nil {
		return errors.New("websocket hub not initialized")
	}

	assignee, err := s.userRepo.GetUserByID(assigneeID)
	if err != nil {
		return fmt.Errorf("get assignee: %w", err)
	}

	// Resolve final IDs
	tID := tenantID
	if tID == uuid.Nil && assignee.TenantID != uuid.Nil {
		tID = assignee.TenantID
	}
	tmID := teamID
	if tmID == uuid.Nil && assignee.TeamID != uuid.Nil {
		tmID = assignee.TeamID
	}

	// Persist notification
	n := &notification.Notification{
		UserID:   assigneeID.String(),
		TenantID: tID.String(),
		TeamID:   tmID.String(),
		Title:    "Assigned to Case",
		Message:  "You have been assigned to a case: " + caseID.String(),
	}
	if err := s.notificationService.SaveNotification(n); err != nil {
		return fmt.Errorf("save notification: %w", err)
	}

	// Best-effort WS push
	if err := websocket.NotifyUser(
		s.hub, s.notificationService,
		assigneeID.String(), tID.String(), tmID.String(),
		n.Title, n.Message,
	); err != nil {
		fmt.Printf("websocket.NotifyUser failed (notification_id=%s): %v\n", n.ID, err)
	}

	// Now pass teamID as well (5 args)
	if err := s.repo.AssignRole(assigneeID, caseID, role, tID, tmID); err != nil {
		return fmt.Errorf("assign role: %w", err)
	}
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
