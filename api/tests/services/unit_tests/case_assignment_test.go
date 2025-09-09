package unit_tests

import (
	"errors"
	"testing"

	"aegis-api/pkg/websocket"
	"aegis-api/services_/case/case_assign"
	"aegis-api/services_/notification"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ---- Mocks ----

type MockCaseAssignmentRepo struct{ mock.Mock }

type MockUserRepository struct{ mock.Mock }

// Minimal mock for NotificationServiceInterface
type MockNotificationService struct{}

func (m *MockNotificationService) SaveNotification(n *notification.Notification) error { return nil }

// case_assign.UserRepo requirement
func (m *MockUserRepository) GetUserByID(userID uuid.UUID) (*case_assign.User, error) {
	args := m.Called(userID)
	user, _ := args.Get(0).(*case_assign.User)
	return user, args.Error(1)
}

// UPDATED to match new interface: AssignRole(userID, caseID, role, tenantID, teamID)
func (m *MockCaseAssignmentRepo) AssignRole(
	userID, caseID uuid.UUID, role string, tenantID, teamID uuid.UUID,
) error {
	args := m.Called(userID, caseID, role, tenantID, teamID)
	return args.Error(0)
}

func (m *MockCaseAssignmentRepo) UnassignRole(userID, caseID uuid.UUID) error {
	args := m.Called(userID, caseID)
	return args.Error(0)
}

type MockAdminChecker struct{ mock.Mock }

func (m *MockAdminChecker) IsAdminFromContext(ctx *gin.Context) (bool, error) {
	args := m.Called(ctx)
	return args.Bool(0), args.Error(1)
}

// ---- Tests ----

// Assign flow: verifies team/tenant are passed through and repo gets called with both.
func TestAssignUserToCase_Admin_PassesTenantAndTeam(t *testing.T) {
	repo := new(MockCaseAssignmentRepo)
	admin := new(MockAdminChecker) // not used in Assign path
	userRepo := new(MockUserRepository)
	notif := new(MockNotificationService)
	hub := &websocket.Hub{}

	svc := case_assign.NewCaseAssignmentService(repo, admin, userRepo, notif, hub)

	assigneeID := uuid.New()
	caseID := uuid.New()
	assignerID := uuid.New()
	tID := uuid.New()
	teamID := uuid.New()

	// The service fetches assignee (for fallback/logging). Provide sane Tenant/Team.
	userRepo.
		On("GetUserByID", assigneeID).
		Return(&case_assign.User{
			ID:       assigneeID,
			TenantID: tID,
			TeamID:   teamID,
		}, nil)

	// Expect AssignRole to be called with tenantID and teamID
	repo.
		On("AssignRole", assigneeID, caseID, "SOC Analyst", tID, teamID).
		Return(nil)

	err := svc.AssignUserToCase(
		"DFIR Admin",
		assigneeID,
		caseID,
		assignerID,
		"SOC Analyst",
		tID,
		teamID,
	)
	assert.NoError(t, err)

	userRepo.AssertExpectations(t)
	repo.AssertExpectations(t)
}

// Unassign: authorized admin → calls repo.UnassignRole and returns nil.
func TestUnassignUserFromCase_Authorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := &gin.Context{}
	ctx.Set("userRole", "Admin")

	repo := new(MockCaseAssignmentRepo)
	admin := new(MockAdminChecker)
	userRepo := new(MockUserRepository)
	notif := new(MockNotificationService)
	hub := &websocket.Hub{}
	svc := case_assign.NewCaseAssignmentService(repo, admin, userRepo, notif, hub)

	assigneeID := uuid.New()
	caseID := uuid.New()

	admin.On("IsAdminFromContext", ctx).Return(true, nil)
	repo.On("UnassignRole", assigneeID, caseID).Return(nil)

	// Provide tenant/team for notification path
	tenantID := uuid.New()
	teamID := uuid.New()
	userRepo.On("GetUserByID", assigneeID).
		Return(&case_assign.User{
			ID:       assigneeID,
			TenantID: tenantID,
			TeamID:   teamID,
		}, nil)

	err := svc.UnassignUserFromCase(ctx, assigneeID, caseID)
	assert.NoError(t, err)

	admin.AssertExpectations(t)
	repo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

// Unassign: forbidden when not admin
func TestUnassignUserFromCase_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := &gin.Context{}
	ctx.Set("userRole", "User")

	repo := new(MockCaseAssignmentRepo)
	admin := new(MockAdminChecker)
	userRepo := new(MockUserRepository)
	notif := new(MockNotificationService)
	hub := &websocket.Hub{}
	svc := case_assign.NewCaseAssignmentService(repo, admin, userRepo, notif, hub)

	assigneeID := uuid.New()
	caseID := uuid.New()

	admin.On("IsAdminFromContext", ctx).Return(false, nil)

	err := svc.UnassignUserFromCase(ctx, assigneeID, caseID)
	assert.EqualError(t, err, "forbidden: admin privileges required")

	admin.AssertExpectations(t)
	repo.AssertNotCalled(t, "UnassignRole", mock.Anything, mock.Anything)
}

// Unassign: admin check fails → returns the error
func TestUnassignUserFromCase_AdminCheckFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := &gin.Context{}
	ctx.Set("userRole", "Admin")

	repo := new(MockCaseAssignmentRepo)
	admin := new(MockAdminChecker)
	userRepo := new(MockUserRepository)
	notif := new(MockNotificationService)
	hub := &websocket.Hub{}
	svc := case_assign.NewCaseAssignmentService(repo, admin, userRepo, notif, hub)

	assigneeID := uuid.New()
	caseID := uuid.New()

	admin.On("IsAdminFromContext", ctx).Return(false, errors.New("db error"))

	err := svc.UnassignUserFromCase(ctx, assigneeID, caseID)
	assert.EqualError(t, err, "db error")

	admin.AssertExpectations(t)
	repo.AssertNotCalled(t, "UnassignRole", mock.Anything, mock.Anything)
}
