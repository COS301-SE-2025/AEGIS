// user_service_test.go
package tests

import (
	mocks "aegis-api/mock"
	"aegis-api/models"
	"aegis-api/services/GetUpdate_Users"
	"errors"
	"testing"
	

	"github.com/stretchr/testify/assert"
	
)

// Test UserService
func TestNewUserService(t *testing.T) {
	mockRepo := &mocks.MockUserRepo{}
	service := GetUpdate_Users.NewUserService(mockRepo)
	
	assert.NotNil(t, service)
	assert.Equal(t, mockRepo, service.GetRepo())
}

func TestUserService_GetProfile_Success(t *testing.T) {
	mockRepo := &mocks.MockUserRepo{}
	service := GetUpdate_Users.NewUserService(mockRepo)
	
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
	mockRepo := &mocks.MockUserRepo{}
	service := GetUpdate_Users.NewUserService(mockRepo)
	
	mockRepo.On("GetUserByID", "123").Return(nil, nil)
	
	result, err := service.GetProfile("123")
	
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "user not found", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetProfile_RepositoryError(t *testing.T) {
	mockRepo := &mocks.MockUserRepo{}
	service :=GetUpdate_Users.NewUserService(mockRepo)
	
	expectedError := errors.New("database error")
	mockRepo.On("GetUserByID", "123").Return(nil, expectedError)
	
	result, err := service.GetProfile("123")
	
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
}

func TestUserService_UpdateProfile_Success(t *testing.T) {
	mockRepo := &mocks.MockUserRepo{}
	service := GetUpdate_Users.NewUserService(mockRepo)
	
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
	mockRepo := &mocks.MockUserRepo{}
	service := GetUpdate_Users.NewUserService(mockRepo)
	
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
	mockRepo := &mocks.MockUserRepo{}
	service :=GetUpdate_Users.NewUserService(mockRepo)
	
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
	mockRepo := &mocks.MockUserRepo{}
	service := GetUpdate_Users.NewUserService(mockRepo)
	
	mockRepo.On("GetUserByEmail", "john@example.com").Return(nil, nil)
	
	result, err := service.Authenticate("john@example.com", "password")
	
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "invalid credentials", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestUserService_Authenticate_RepositoryError(t *testing.T) {
	mockRepo := &mocks.MockUserRepo{}
	service := GetUpdate_Users.NewUserService(mockRepo)
	
	expectedError := errors.New("database error")
	mockRepo.On("GetUserByEmail", "john@example.com").Return(nil, expectedError)
	
	result, err := service.Authenticate("john@example.com", "password")
	
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetUserRoles_Success(t *testing.T) {
	mockRepo := &mocks.MockUserRepo{}
	service := GetUpdate_Users.NewUserService(mockRepo)
	
	expectedRoles := []string{"admin", "user"}
	mockRepo.On("GetUserRoles", "123").Return(expectedRoles, nil)
	
	result, err := service.GetUserRoles("123")
	
	assert.NoError(t, err)
	assert.Equal(t, expectedRoles, result)
	mockRepo.AssertExpectations(t)
}





func TestUserService_GetUserRoles_Error(t *testing.T) {
	mockRepo := new(mocks.MockUserRepo)
	service := GetUpdate_Users.NewUserService(mockRepo)

	expectedError := errors.New("roles query failed")

	// 👇 Ensure nil is typed correctly to match the expected return type
	mockRepo.On("GetUserRoles", "123").Return(([]string)(nil), expectedError)

	result, err := service.GetUserRoles("123")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedError, err)

	mockRepo.AssertExpectations(t)
}
