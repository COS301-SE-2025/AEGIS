package isolation

import (
	
	"aegis-api/services_/user/GetUpdate_UserInfo"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewUserService(t *testing.T) {
	mockRepo := &GetUpdate_UserInfo.MockUserRepo{}
	service := GetUpdate_UserInfo.NewUserService(mockRepo)

	assert.NotNil(t, service)
	assert.Equal(t, mockRepo, service.GetRepo())
}

func TestUserService_GetProfile_Success(t *testing.T) {
	mockRepo := &GetUpdate_UserInfo.MockUserRepo{}
	service := GetUpdate_UserInfo.NewUserService(mockRepo)

	id := uuid.MustParse("00000000-0000-0000-0000-000000000123")
	expectedUser := &GetUpdate_UserInfo.User{ID: id, FullName: "John Doe"}

	mockRepo.On("GetUserByID", id).Return(expectedUser, nil)

	result, err := service.GetProfile(id)

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, result)
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetProfile_UserNotFound(t *testing.T) {
	mockRepo := &GetUpdate_UserInfo.MockUserRepo{}
	service := GetUpdate_UserInfo.NewUserService(mockRepo)

	id := uuid.MustParse("00000000-0000-0000-0000-000000000123")

	mockRepo.On("GetUserByID", id).Return(nil, nil)

	result, err := service.GetProfile(id)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "user not found", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestUserService_UpdateProfile_Success(t *testing.T) {
	mockRepo := &GetUpdate_UserInfo.MockUserRepo{}
	service := GetUpdate_UserInfo.NewUserService(mockRepo)

	id := uuid.New()
	updates := map[string]interface{}{"full_name": "Jane Doe"}

	mockRepo.On("UpdateUser", id, updates).Return(nil)

	err := service.UpdateProfile(id, updates)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Authenticate_Success(t *testing.T) {
	mockRepo := &GetUpdate_UserInfo.MockUserRepo{}
	service := GetUpdate_UserInfo.NewUserService(mockRepo)

	expectedUser := &GetUpdate_UserInfo.User{
		ID:    uuid.New(),
		Email: "john@example.com",
	}

	mockRepo.On("GetUserByEmail", "john@example.com").Return(expectedUser, nil)

	result, err := service.Authenticate("john@example.com", "password")

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, result)
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetUserRoles_Success(t *testing.T) {
	mockRepo := &GetUpdate_UserInfo.MockUserRepo{}
	service := GetUpdate_UserInfo.NewUserService(mockRepo)

	id := uuid.New()
	expectedRoles := []string{"admin", "user"}

	mockRepo.On("GetUserRoles", id).Return(expectedRoles, nil)

	result, err := service.GetUserRoles(id)

	assert.NoError(t, err)
	assert.Equal(t, expectedRoles, result)
	mockRepo.AssertExpectations(t)
}
