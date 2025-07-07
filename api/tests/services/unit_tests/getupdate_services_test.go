package unit_tests

import (
	"testing"

	"aegis-api/services_/user/GetUpdate_UserInfo"


	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetProfileUser_Success(t *testing.T) {
	mockRepo := new(GetUpdate_UserInfo.MockUserRepo)
	service := GetUpdate_UserInfo.NewUserService(mockRepo)

	id := uuid.MustParse("00000000-0000-0000-0000-000000000123")
	user := &GetUpdate_UserInfo.User{ID: id, FullName: "John Doe"}

	mockRepo.On("GetUserByID", id).Return(user, nil)

	result, err := service.GetProfile(id)

	assert.NoError(t, err)
	assert.Equal(t, user, result)
	mockRepo.AssertExpectations(t)
}

func TestGetProfile_NotFound(t *testing.T) {
	mockRepo := new(GetUpdate_UserInfo.MockUserRepo)
	service := GetUpdate_UserInfo.NewUserService(mockRepo)

	id := uuid.MustParse("00000000-0000-0000-0000-000000000123")

	mockRepo.On("GetUserByID", id).Return(nil, nil)

	result, err := service.GetProfile(id)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "user not found", err.Error())
}

func TestUpdateProfile(t *testing.T) {
	mockRepo := new(GetUpdate_UserInfo.MockUserRepo)
	service := GetUpdate_UserInfo.NewUserService(mockRepo)

	id := uuid.MustParse("00000000-0000-0000-0000-000000000123")
	updates := map[string]interface{}{
		"full_name": "Updated Name",
	}
	mockRepo.On("UpdateUser", id, updates).Return(nil)

	err := service.UpdateProfile(id, updates)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAuthenticate_Success(t *testing.T) {
	mockRepo := new(GetUpdate_UserInfo.MockUserRepo)
	service := GetUpdate_UserInfo.NewUserService(mockRepo)

	user := &GetUpdate_UserInfo.User{Email: "test@example.com"}
	mockRepo.On("GetUserByEmail", "test@example.com").Return(user, nil)

	result, err := service.Authenticate("test@example.com", "password")

	assert.NoError(t, err)
	assert.Equal(t, user, result)
	mockRepo.AssertExpectations(t)
}

func TestAuthenticate_InvalidCredentials(t *testing.T) {
	mockRepo := new(GetUpdate_UserInfo.MockUserRepo)
	service := GetUpdate_UserInfo.NewUserService(mockRepo)

	mockRepo.On("GetUserByEmail", "invalid@example.com").Return(nil, nil)

	result, err := service.Authenticate("invalid@example.com", "wrongpass")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "invalid credentials", err.Error())
}

func TestGetUserRoles_Success(t *testing.T) {
	mockRepo := new(GetUpdate_UserInfo.MockUserRepo)
	service := GetUpdate_UserInfo.NewUserService(mockRepo)

	id := uuid.MustParse("00000000-0000-0000-0000-000000000123")

	mockRepo.On("GetUserRoles", id).Return([]string{"admin", "user"}, nil)

	roles, err := service.GetUserRoles(id)

	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"admin", "user"}, roles)
	mockRepo.AssertExpectations(t)
}
