package unit_tests

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"aegis-api/services_/case/case_assign"
)

type MockCaseAssignmentRepo struct {
	mock.Mock
}

// Implement IsAdmin to satisfy CaseAssignmentRepoInterface
func (m *MockCaseAssignmentRepo) IsAdmin(userID uuid.UUID) (bool, error) {
	args := m.Called(userID)
	return args.Bool(0), args.Error(1)
}

type MockAdminChecker struct {
	mock.Mock
}

func (m *MockAdminChecker) IsAdmin(userID uuid.UUID) (bool, error) {
	args := m.Called(userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockCaseAssignmentRepo) AssignRole(userID, caseID uuid.UUID, role string) error {
	args := m.Called(userID, caseID, role)
	return args.Error(0)
}

func (m *MockCaseAssignmentRepo) UnassignRole(userID, caseID uuid.UUID) error {
	args := m.Called(userID, caseID)
	return args.Error(0)
}

type MockPermissionChecker struct {
	mock.Mock
}

func (m *MockPermissionChecker) HasPermission(userID uuid.UUID, permissionName string) (bool, error) {
	args := m.Called(userID, permissionName)
	return args.Bool(0), args.Error(1)
}

func TestUnassignUserFromCase_Authorized(t *testing.T) {
	repo := new(MockCaseAssignmentRepo)
	perm := new(MockAdminChecker)

	svc := case_assign.NewCaseAssignmentService(repo)

	assignerID := uuid.New()
	assigneeID := uuid.New()
	caseID := uuid.New()

	perm.On("IsAdmin", assignerID).Return(true, nil)

	repo.On("UnassignRole", assigneeID, caseID).Return(nil)

	err := svc.UnassignUserFromCase(assignerID, assigneeID, caseID)
	assert.NoError(t, err)

	perm.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestUnassignUserFromCase_Forbidden(t *testing.T) {
	repo := new(MockCaseAssignmentRepo)
	perm := new(MockAdminChecker)

	svc := case_assign.NewCaseAssignmentService(repo)

	assignerID := uuid.New()
	assigneeID := uuid.New()
	caseID := uuid.New()

	perm.On("IsAdmin", assignerID).Return(false, errors.New("db error"))

	err := svc.UnassignUserFromCase(assignerID, assigneeID, caseID)
	assert.Error(t, err)
	assert.Equal(t, "forbidden: missing assign_case permission", err.Error())

	perm.AssertExpectations(t)
	repo.On("IsAdmin", assignerID).Return(true, nil)

}

func TestUnassignUserFromCase_PermissionError(t *testing.T) {
	repo := new(MockCaseAssignmentRepo)
	perm := new(MockAdminChecker)

	svc := case_assign.NewCaseAssignmentService(repo)

	assignerID := uuid.New()
	assigneeID := uuid.New()
	caseID := uuid.New()

	perm.On("HasPermission", assignerID, "assign_case").Return(false, errors.New("db error"))

	err := svc.UnassignUserFromCase(assignerID, assigneeID, caseID)
	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())

	perm.AssertExpectations(t)
	repo.On("IsAdmin", assignerID).Return(true, nil)

}
