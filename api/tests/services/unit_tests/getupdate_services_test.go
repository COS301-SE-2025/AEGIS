package tests

import (
	
	"testing"

	"aegis-api/models"
	"aegis-api/services/GetUpdate_Users"
	"aegis-api/mock"
	"github.com/stretchr/testify/assert"
)




func TestGetProfile_Success(t *testing.T) {
	mockRepo := new(mocks.MockUserRepo)
	service := GetUpdate_Users.NewUserService(mockRepo)

	user := &models.UserDTO{ID: "123", FullName: "John Doe"}
	mockRepo.On("GetUserByID", "123").Return(user, nil)

	result, err := service.GetProfile("123")

	assert.NoError(t, err)
	assert.Equal(t, user, result)
	mockRepo.AssertExpectations(t)
}

func TestGetProfile_NotFound(t *testing.T) {
	mockRepo := new(mocks.MockUserRepo)
	service := GetUpdate_Users.NewUserService(mockRepo)

	mockRepo.On("GetUserByID", "not_exist").Return(nil, nil)

	result, err := service.GetProfile("not_exist")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "user not found", err.Error())
}

func TestUpdateProfile(t *testing.T) {
	mockRepo := new(mocks.MockUserRepo)
	service := GetUpdate_Users.NewUserService(mockRepo)

	updates := map[string]interface{}{
		"full_name": "Updated Name",
	}
	mockRepo.On("UpdateUser", "123", updates).Return(nil)

	err := service.UpdateProfile("123", updates)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAuthenticate_Success(t *testing.T) {
	mockRepo := new(mocks.MockUserRepo)
	service := GetUpdate_Users.NewUserService(mockRepo)

	user := &models.UserDTO{Email: "test@example.com"}
	mockRepo.On("GetUserByEmail", "test@example.com").Return(user, nil)

	result, err := service.Authenticate("test@example.com", "password")

	assert.NoError(t, err)
	assert.Equal(t, user, result)
	mockRepo.AssertExpectations(t)
}

func TestAuthenticate_InvalidCredentials(t *testing.T) {
	mockRepo := new(mocks.MockUserRepo)
	service := GetUpdate_Users.NewUserService(mockRepo)

	mockRepo.On("GetUserByEmail", "invalid@example.com").Return(nil, nil)

	result, err := service.Authenticate("invalid@example.com", "wrongpass")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "invalid credentials", err.Error())
}

func TestGetUserRoles_Success(t *testing.T) {
	mockRepo := new(mocks.MockUserRepo)
	service := GetUpdate_Users.NewUserService(mockRepo)

	mockRepo.On("GetUserRoles", "123").Return([]string{"admin", "user"}, nil)

	roles, err := service.GetUserRoles("123")

	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"admin", "user"}, roles)
	mockRepo.AssertExpectations(t)
}
