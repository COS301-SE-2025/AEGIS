package unit_tests

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"aegis-api/services_/case/remove_user_from_case"
)

// MockRepository implements the Repository interface
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) IsAdmin(userID uuid.UUID) (bool, error) {
	args := m.Called(userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockRepository) RemoveUserFromCase(userID, caseID uuid.UUID) error {
	args := m.Called(userID, caseID)
	return args.Error(0)
}

func TestRemoveUser_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := remove_user_from_case.NewService(mockRepo)

	adminID := uuid.New()
	userID := uuid.New()
	caseID := uuid.New()

	req := remove_user_from_case.RemoveUserRequest{
		AdminID: adminID,
		UserID:  userID,
		CaseID:  caseID,
	}

	mockRepo.On("IsAdmin", adminID).Return(true, nil)
	mockRepo.On("RemoveUserFromCase", userID, caseID).Return(nil)

	err := service.RemoveUser(req)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestRemoveUser_NotAdmin(t *testing.T) {
	mockRepo := new(MockRepository)
	service := remove_user_from_case.NewService(mockRepo)

	adminID := uuid.New()
	userID := uuid.New()
	caseID := uuid.New()

	req := remove_user_from_case.RemoveUserRequest{
		AdminID: adminID,
		UserID:  userID,
		CaseID:  caseID,
	}

	mockRepo.On("IsAdmin", adminID).Return(false, nil)

	err := service.RemoveUser(req)
	assert.EqualError(t, err, "unauthorized: only admins can remove users from a case")
	mockRepo.AssertExpectations(t)
}

func TestRemoveUser_IsAdminError(t *testing.T) {
	mockRepo := new(MockRepository)
	service := remove_user_from_case.NewService(mockRepo)

	adminID := uuid.New()
	userID := uuid.New()
	caseID := uuid.New()

	req := remove_user_from_case.RemoveUserRequest{
		AdminID: adminID,
		UserID:  userID,
		CaseID:  caseID,
	}

	mockRepo.On("IsAdmin", adminID).Return(false, errors.New("db error"))

	err := service.RemoveUser(req)
	assert.EqualError(t, err, "db error")
	mockRepo.AssertExpectations(t)
}

func TestRemoveUser_DeleteError(t *testing.T) {
	mockRepo := new(MockRepository)
	service := remove_user_from_case.NewService(mockRepo)

	adminID := uuid.New()
	userID := uuid.New()
	caseID := uuid.New()

	req := remove_user_from_case.RemoveUserRequest{
		AdminID: adminID,
		UserID:  userID,
		CaseID:  caseID,
	}

	mockRepo.On("IsAdmin", adminID).Return(true, nil)
	mockRepo.On("RemoveUserFromCase", userID, caseID).Return(errors.New("delete error"))

	err := service.RemoveUser(req)
	assert.EqualError(t, err, "delete error")
	mockRepo.AssertExpectations(t)
}
