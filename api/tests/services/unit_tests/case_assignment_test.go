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

type MockCaseAssignmentRepo struct{ mock.Mock }
type MockUserRepository struct{ mock.Mock }
type MockNotificationService struct{ mock.Mock }
type MockAdminChecker struct{ mock.Mock }

func (m *MockNotificationService) SaveNotification(n *notification.Notification) error {
	args := m.Called(n)
	return args.Error(0)
}

func (m *MockCaseAssignmentRepo) AssignRole(
	userID, caseID uuid.UUID, role string, tenantID, teamID uuid.UUID,
) error {
	args := m.Called(userID, caseID, role, tenantID, teamID)
	return args.Error(0)
}

func (m *MockCaseAssignmentRepo) GetCaseByID(caseID uuid.UUID, caseDetails *case_assign.Case) error {
	args := m.Called(caseID, caseDetails)
	return args.Error(0)
}

func (m *MockAdminChecker) IsAdminFromContext(ctx *gin.Context) (bool, error) {
	args := m.Called(ctx)
	return args.Bool(0), args.Error(1)
}

// Implement GetUserByID method to satisfy the UserRepo interface
func (m *MockUserRepository) GetUserByID(userID uuid.UUID) (*case_assign.User, error) {
	args := m.Called(userID)
	user, _ := args.Get(0).(*case_assign.User) // Extract the return value as *User
	return user, args.Error(1)                 // Return the error (if any)
}

// Add UnassignRole method to the MockCaseAssignmentRepo struct
func (m *MockCaseAssignmentRepo) UnassignRole(userID, caseID uuid.UUID) error {
	args := m.Called(userID, caseID)
	return args.Error(0)
}

// Test AssignUserToCase: verifies that the method calls necessary functions to assign a user to a case
func TestAssignUserToCase(t *testing.T) {
	repo := new(MockCaseAssignmentRepo)
	adminChecker := new(MockAdminChecker)
	userRepo := new(MockUserRepository)
	notifService := new(MockNotificationService)
	hub := &websocket.Hub{}
	svc := case_assign.NewCaseAssignmentService(repo, adminChecker, userRepo, notifService, hub)

	// Initialize test data
	assigneeID := uuid.New()
	caseID := uuid.New()
	assignerID := uuid.New()
	tID := uuid.New()
	teamID := uuid.New()

	// Mock GetCaseByID to return an active case
	repo.On("GetCaseByID", caseID, mock.AnythingOfType("*case_assign.Case")).Return(nil).Run(func(args mock.Arguments) {
		caseDetails := args.Get(1).(*case_assign.Case)
		caseDetails.Status = "open" // Case is active
		caseDetails.Title = "Test Case"
	})

	// Mock SaveNotification to confirm it gets called
	notifService.On("SaveNotification", mock.AnythingOfType("*notification.Notification")).Return(nil)

	// Mock AssignRole to verify that the role is being assigned
	repo.On("AssignRole", assigneeID, caseID, "SOC Analyst", tID, teamID).Return(nil)

	// Call the method to test
	err := svc.AssignUserToCase(
		"DFIR Admin",  // assignerRole
		assigneeID,    // assigneeID
		caseID,        // caseID
		assignerID,    // assignerID
		"SOC Analyst", // role
		tID,           // tenantID
		teamID,        // teamID
	)

	// Assertions
	assert.NoError(t, err)
	repo.AssertExpectations(t)
	notifService.AssertExpectations(t)
}

// // Test AssignUserToCase: verifies the error when the case is inactive
func TestAssignUserToCase_CaseNotActive(t *testing.T) {
	repo := new(MockCaseAssignmentRepo)
	adminChecker := new(MockAdminChecker)
	userRepo := new(MockUserRepository)
	notifService := new(MockNotificationService)
	hub := &websocket.Hub{}
	svc := case_assign.NewCaseAssignmentService(repo, adminChecker, userRepo, notifService, hub)

	assigneeID := uuid.New()
	caseID := uuid.New()
	assignerID := uuid.New()
	tID := uuid.New()
	teamID := uuid.New()

	// Mock GetCaseByID to return a closed case
	repo.On("GetCaseByID", caseID, mock.AnythingOfType("*case_assign.Case")).Return(nil).Run(func(args mock.Arguments) {
		caseDetails := args.Get(1).(*case_assign.Case)
		caseDetails.Status = "closed" // Case is inactive
		caseDetails.Title = "Closed Test Case"
	})

	// Call the method to test
	err := svc.AssignUserToCase(
		"DFIR Admin",  // assignerRole
		assigneeID,    // assigneeID
		caseID,        // caseID
		assignerID,    // assignerID
		"SOC Analyst", // role
		tID,           // tenantID
		teamID,        // teamID
	)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is not active")

	repo.AssertExpectations(t)
	notifService.AssertExpectations(t)
}

// Test UnassignUserFromCase: verifies that an admin can unassign a user from a case
// func TestUnassignUserFromCase_Authorized(t *testing.T) {
// 	ctx := &gin.Context{}
// 	ctx.Set("userRole", "DFIR Admin") // Simulating an admin role

// 	repo := new(MockCaseAssignmentRepo)
// 	adminChecker := new(MockAdminChecker)
// 	userRepo := new(MockUserRepository)
// 	notifService := new(MockNotificationService)
// 	hub := &websocket.Hub{}
// 	svc := case_assign.NewCaseAssignmentService(repo, adminChecker, userRepo, notifService, hub)

// 	assigneeID := uuid.New()
// 	caseID := uuid.New()

// 	// Mock IsAdminFromContext to return true (admin role)
// 	adminChecker.On("IsAdminFromContext", ctx).Return(true, nil)

// 	// Mock UnassignRole to verify that the user is unassigned
// 	repo.On("UnassignRole", assigneeID, caseID).Return(nil)

// 	// Mock GetUserByID to return user details for WebSocket notification
// 	userRepo.On("GetUserByID", assigneeID).Return(&case_assign.User{
// 		ID:       assigneeID,
// 		TenantID: uuid.New(),
// 		TeamID:   uuid.New(),
// 	}, nil)

// 	// Call the method to test
// 	err := svc.UnassignUserFromCase(ctx, assigneeID, caseID)

// 	// Assertions
// 	assert.NoError(t, err)
// 	repo.AssertExpectations(t)
// 	notifService.AssertExpectations(t)
// 	userRepo.AssertExpectations(t)
// }

// // Test UnassignUserFromCase: verifies that non-admins cannot unassign users
func TestUnassignUserFromCase_Forbidden(t *testing.T) {
	ctx := &gin.Context{}
	ctx.Set("userRole", "User") // Simulating a non-admin role

	repo := new(MockCaseAssignmentRepo)
	adminChecker := new(MockAdminChecker)
	userRepo := new(MockUserRepository)
	notifService := new(MockNotificationService)
	hub := &websocket.Hub{}
	svc := case_assign.NewCaseAssignmentService(repo, adminChecker, userRepo, notifService, hub)

	assigneeID := uuid.New()
	caseID := uuid.New()

	// Mock IsAdminFromContext to return false (non-admin role)
	adminChecker.On("IsAdminFromContext", ctx).Return(false, nil)

	// Call the method to test
	err := svc.UnassignUserFromCase(ctx, assigneeID, caseID)

	// Assertions
	assert.EqualError(t, err, "forbidden: admin privileges required")

	repo.AssertNotCalled(t, "UnassignRole", assigneeID, caseID)
	userRepo.AssertNotCalled(t, "GetUserByID", assigneeID)
}

// Test UnassignUserFromCase: verifies the failure when admin check fails
func TestUnassignUserFromCase_AdminCheckFails(t *testing.T) {
	ctx := &gin.Context{}
	repo := new(MockCaseAssignmentRepo)
	adminChecker := new(MockAdminChecker)
	userRepo := new(MockUserRepository)
	notifService := new(MockNotificationService)
	hub := &websocket.Hub{}
	svc := case_assign.NewCaseAssignmentService(repo, adminChecker, userRepo, notifService, hub)

	assigneeID := uuid.New()
	caseID := uuid.New()

	// Mock IsAdminFromContext to return an error
	adminChecker.On("IsAdminFromContext", ctx).Return(false, errors.New("db error"))

	// Call the method to test
	err := svc.UnassignUserFromCase(ctx, assigneeID, caseID)

	// Assertions
	assert.EqualError(t, err, "db error")

	repo.AssertNotCalled(t, "UnassignRole", assigneeID, caseID)
	userRepo.AssertNotCalled(t, "GetUserByID", assigneeID)
}
