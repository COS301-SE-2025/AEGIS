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
	// Step 1: Check if the assigner has admin privileges
	if assignerRole != "DFIR Admin" {
		return errors.New("forbidden: admin privileges required")
	}

	// Step 2: Ensure the notification and websocket services are initialized
	if s.notificationService == nil {
		return errors.New("notification service not initialized")
	}
	if s.hub == nil {
		return errors.New("websocket hub not initialized")
	}

	// Step 3: Fetch the case details using the caseID
	var caseDetails Case
	err := s.repo.GetCaseByID(caseID, &caseDetails)
	if err != nil {
		return fmt.Errorf("get case details: %w", err)
	}

	// Step 4: Check if the case is active by checking its Status
	if caseDetails.Status != "open" && caseDetails.Status != "ongoing" {
		return fmt.Errorf("case with ID %s is not active", caseID)
	}

	// Step 5: Create and persist the notification with the case title
	n := &notification.Notification{
		UserID:   assigneeID.String(),
		TenantID: tenantID.String(),
		TeamID:   teamID.String(),
		Title:    "Assigned to Case",
		Message:  fmt.Sprintf("You have been assigned to case: %s", caseDetails.Title), // Use case title here
	}
	if err := s.notificationService.SaveNotification(n); err != nil {
		return fmt.Errorf("save notification: %w", err)
	}

	// Step 6: Send a best-effort websocket push notification
	if err := websocket.NotifyUser(
		s.hub, s.notificationService,
		assigneeID.String(), tenantID.String(), teamID.String(),
		n.Title, n.Message,
	); err != nil {
		fmt.Printf("websocket.NotifyUser failed (notification_id=%s): %v\n", n.ID, err)
	}

	// Step 7: Assign the role to the user for the case
	if err := s.repo.AssignRole(assigneeID, caseID, role, tenantID, teamID); err != nil {
		return fmt.Errorf("assign role: %w", err)
	}

	// Step 8: Return nil indicating success
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
