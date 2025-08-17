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

type MockCaseAssignmentRepo struct {
	mock.Mock
}
type MockUserRepository struct {
	mock.Mock
}

// Minimal mock for NotificationService to prevent nil pointer panics in tests
type MockNotificationService struct {
	// No embedded type needed, just implement the interface
}

func (m *MockNotificationService) SaveNotification(n *notification.Notification) error { return nil }

// Implement GetUserByID to satisfy case_assign.UserRepo interface
func (m *MockUserRepository) GetUserByID(userID uuid.UUID) (*case_assign.User, error) {
	args := m.Called(userID)
	user, _ := args.Get(0).(*case_assign.User)
	return user, args.Error(1)
}

func (m *MockCaseAssignmentRepo) AssignRole(userID, caseID uuid.UUID, role string, assignedBy uuid.UUID) error {
	args := m.Called(userID, caseID, role, assignedBy)
	return args.Error(0)
}

func (m *MockCaseAssignmentRepo) UnassignRole(userID, caseID uuid.UUID) error {
	args := m.Called(userID, caseID)
	return args.Error(0)
}

type MockAdminChecker struct {
	mock.Mock
}

func (m *MockAdminChecker) IsAdminFromContext(ctx *gin.Context) (bool, error) {
	args := m.Called(ctx)
	return args.Bool(0), args.Error(1)
}

// ---- Tests ----

// TestAssignUserToCase_Authorized tests assigning a user to a case when the user is an admin.
func TestUnassignUserFromCase_Authorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := &gin.Context{}
	ctx.Set("userRole", "Admin")

	repo := new(MockCaseAssignmentRepo)
	admin := new(MockAdminChecker)
	userRepo := new(MockUserRepository)
	mockNotification := new(MockNotificationService)
	mockHub := &websocket.Hub{}
	svc := case_assign.NewCaseAssignmentService(repo, admin, userRepo, mockNotification, mockHub)

	assigneeID := uuid.New()
	caseID := uuid.New()

	admin.On("IsAdminFromContext", ctx).Return(true, nil)
	repo.On("UnassignRole", assigneeID, caseID).Return(nil)
	userRepo.On("GetUserByID", mock.Anything).Return(&case_assign.User{ID: assigneeID}, nil)

	err := svc.UnassignUserFromCase(ctx, assigneeID, caseID)
	assert.NoError(t, err)

	admin.AssertExpectations(t)
	repo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

// TestUnassignUserFromCase_Forbidden tests unassigning a user from a case when the user is not an admin.
func TestUnassignUserFromCase_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := &gin.Context{}
	ctx.Set("userRole", "User")

	repo := new(MockCaseAssignmentRepo)
	admin := new(MockAdminChecker)
	userRepo := new(MockUserRepository)
	mockNotification := new(MockNotificationService)
	mockHub := &websocket.Hub{}
	svc := case_assign.NewCaseAssignmentService(repo, admin, userRepo, mockNotification, mockHub)

	assigneeID := uuid.New()
	caseID := uuid.New()

	admin.On("IsAdminFromContext", ctx).Return(false, nil)

	err := svc.UnassignUserFromCase(ctx, assigneeID, caseID)
	assert.EqualError(t, err, "forbidden: admin privileges required")

	admin.AssertExpectations(t)
	repo.AssertNotCalled(t, "UnassignRole")
}

// TestUnassignUserFromCase_AdminCheckFails tests the case where the admin check fails.
func TestUnassignUserFromCase_AdminCheckFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := &gin.Context{}
	ctx.Set("userRole", "Admin")

	repo := new(MockCaseAssignmentRepo)
	admin := new(MockAdminChecker)
	userRepo := new(MockUserRepository)
	mockNotification := new(MockNotificationService)
	mockHub := &websocket.Hub{}
	svc := case_assign.NewCaseAssignmentService(repo, admin, userRepo, mockNotification, mockHub)

	assigneeID := uuid.New()
	caseID := uuid.New()

	admin.On("IsAdminFromContext", ctx).Return(false, errors.New("db error"))

	err := svc.UnassignUserFromCase(ctx, assigneeID, caseID)
	assert.EqualError(t, err, "db error")

	admin.AssertExpectations(t)
	repo.AssertNotCalled(t, "UnassignRole")
}
