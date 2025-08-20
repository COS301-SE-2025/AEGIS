package unit_tests

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"aegis-api/services_/admin/delete_user"
)

// Mock repository
type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) DeleteUserByID(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}

func TestDeleteUser_Success(t *testing.T) {
	mockRepo := new(MockUserRepo)
	service := delete_user.NewUserDeleteService(mockRepo)

	userID := uuid.New()
	req := delete_user.DeleteUserRequest{
		UserID: userID.String(),
	}

	mockRepo.On("DeleteUserByID", userID).Return(nil)

	err := service.DeleteUser(req, "Admin")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteUser_Unauthorized(t *testing.T) {
	mockRepo := new(MockUserRepo)
	service := delete_user.NewUserDeleteService(mockRepo)

	req := delete_user.DeleteUserRequest{
		UserID: uuid.New().String(),
	}

	err := service.DeleteUser(req, "User") // Not Admin

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized")
}

func TestDeleteUser_InvalidUUID(t *testing.T) {
	mockRepo := new(MockUserRepo)
	service := delete_user.NewUserDeleteService(mockRepo)

	req := delete_user.DeleteUserRequest{
		UserID: "invalid-uuid",
	}

	err := service.DeleteUser(req, "Admin")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid user UUID")
}

func TestDeleteUser_RepoFailure(t *testing.T) {
	mockRepo := new(MockUserRepo)
	service := delete_user.NewUserDeleteService(mockRepo)

	userID := uuid.New()
	req := delete_user.DeleteUserRequest{
		UserID: userID.String(),
	}

	mockRepo.On("DeleteUserByID", userID).Return(errors.New("db error"))

	err := service.DeleteUser(req, "Admin")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete user")
	mockRepo.AssertExpectations(t)
}
