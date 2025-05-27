// user_service_test.go
package tests

import (
	"errors"
	"testing"
	"time"

	"aegis-api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetUserByID(userID string) (*models.UserDTO, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserDTO), args.Error(1)
}

func (m *MockUserRepository) GetUserByEmail(email string) (*models.UserDTO, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserDTO), args.Error(1)
}

func (m *MockUserRepository) UpdateUser(userID string, updates map[string]interface{}) error {
	args := m.Called(userID, updates)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserRoles(userID string) ([]string, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// Test UserService
func TestNewUserService(t *testing.T) {
	mockRepo := &MockUserRepository{}
	service := NewUserService(mockRepo)
	
	assert.NotNil(t, service)
	assert.Equal(t, mockRepo, service.repo)
}

func TestUserService_GetProfile_Success(t *testing.T) {
	mockRepo := &MockUserRepository{}
	service := NewUserService(mockRepo)
	
	expectedUser := &models.UserDTO{
		ID:       "123",
		FullName: "John Doe",
		Email:    "john@example.com",
		Role:     "user",
	}
	
	mockRepo.On("GetUserByID", "123").Return(expectedUser, nil)
	
	result, err := service.GetProfile("123")
	
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, result)
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetProfile_UserNotFound(t *testing.T) {
	mockRepo := &MockUserRepository{}
	service := NewUserService(mockRepo)
	
	mockRepo.On("GetUserByID", "123").Return(nil, nil)
	
	result, err := service.GetProfile("123")
	
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "user not found", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetProfile_RepositoryError(t *testing.T) {
	mockRepo := &MockUserRepository{}
	service := NewUserService(mockRepo)
	
	expectedError := errors.New("database error")
	mockRepo.On("GetUserByID", "123").Return(nil, expectedError)
	
	result, err := service.GetProfile("123")
	
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
}

func TestUserService_UpdateProfile_Success(t *testing.T) {
	mockRepo := &MockUserRepository{}
	service := NewUserService(mockRepo)
	
	updates := map[string]interface{}{
		"full_name": "Jane Doe",
		"email":     "jane@example.com",
	}
	
	mockRepo.On("UpdateUser", "123", updates).Return(nil)
	
	err := service.UpdateProfile("123", updates)
	
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUserService_UpdateProfile_Error(t *testing.T) {
	mockRepo := &MockUserRepository{}
	service := NewUserService(mockRepo)
	
	updates := map[string]interface{}{
		"full_name": "Jane Doe",
	}
	expectedError := errors.New("update failed")
	
	mockRepo.On("UpdateUser", "123", updates).Return(expectedError)
	
	err := service.UpdateProfile("123", updates)
	
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Authenticate_Success(t *testing.T) {
	mockRepo := &MockUserRepository{}
	service := NewUserService(mockRepo)
	
	expectedUser := &models.UserDTO{
		ID:           "123",
		FullName:     "John Doe",
		Email:        "john@example.com",
		PasswordHash: "hashedpassword",
		Role:         "user",
	}
	
	mockRepo.On("GetUserByEmail", "john@example.com").Return(expectedUser, nil)
	
	result, err := service.Authenticate("john@example.com", "password")
	
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, result)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Authenticate_UserNotFound(t *testing.T) {
	mockRepo := &MockUserRepository{}
	service := NewUserService(mockRepo)
	
	mockRepo.On("GetUserByEmail", "john@example.com").Return(nil, nil)
	
	result, err := service.Authenticate("john@example.com", "password")
	
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "invalid credentials", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestUserService_Authenticate_RepositoryError(t *testing.T) {
	mockRepo := &MockUserRepository{}
	service := NewUserService(mockRepo)
	
	expectedError := errors.New("database error")
	mockRepo.On("GetUserByEmail", "john@example.com").Return(nil, expectedError)
	
	result, err := service.Authenticate("john@example.com", "password")
	
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetUserRoles_Success(t *testing.T) {
	mockRepo := &MockUserRepository{}
	service := NewUserService(mockRepo)
	
	expectedRoles := []string{"admin", "user"}
	mockRepo.On("GetUserRoles", "123").Return(expectedRoles, nil)
	
	result, err := service.GetUserRoles("123")
	
	assert.NoError(t, err)
	assert.Equal(t, expectedRoles, result)
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetUserRoles_Error(t *testing.T) {
	mockRepo := &MockUserRepository{}
	service := NewUserService(mockRepo)
	
	expectedError := errors.New("roles query failed")
	mockRepo.On("GetUserRoles", "123").Return(nil, expectedError)
	
	result, err := service.GetUserRoles("123")
	
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
}




/