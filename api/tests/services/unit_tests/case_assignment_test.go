package unit_tests

import (
	"errors"
	"testing"

	"aegis-api/services_/case/case_assign"

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

// Implement GetUserByID to satisfy case_assign.UserRepo interface
func (m *MockUserRepository) GetUserByID(userID uuid.UUID) (interface{}, error) {
	args := m.Called(userID)
	return args.Get(0), args.Error(1)
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
	svc := case_assign.NewCaseAssignmentService(repo, admin, userRepo)

	assigneeID := uuid.New()
	caseID := uuid.New()

	admin.On("IsAdminFromContext", ctx).Return(true, nil)
	repo.On("UnassignRole", assigneeID, caseID).Return(nil)

	err := svc.UnassignUserFromCase(ctx, assigneeID, caseID)
	assert.NoError(t, err)

	admin.AssertExpectations(t)
	repo.AssertExpectations(t)
}

// TestUnassignUserFromCase_Forbidden tests unassigning a user from a case when the user is not an admin.
func TestUnassignUserFromCase_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := &gin.Context{}
	ctx.Set("userRole", "User")

	repo := new(MockCaseAssignmentRepo)
	admin := new(MockAdminChecker)
	userRepo := new(MockUserRepository)
	svc := case_assign.NewCaseAssignmentService(repo, admin, userRepo)

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
	svc := case_assign.NewCaseAssignmentService(repo, admin)

	assigneeID := uuid.New()
	caseID := uuid.New()

	admin.On("IsAdminFromContext", ctx).Return(false, errors.New("db error"))

	err := svc.UnassignUserFromCase(ctx, assigneeID, caseID)
	assert.EqualError(t, err, "db error")

	admin.AssertExpectations(t)
	repo.AssertNotCalled(t, "UnassignRole")
}
