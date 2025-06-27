package unit_tests

import (
	"aegis-api/services_/user/profile"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProfileRepository is a mocked implementation of ProfileRepository.
type MockProfileRepository struct {
	mock.Mock
}

func (m *MockProfileRepository) GetProfileByID(userID string) (*profile.UserProfile, error) {
	args := m.Called(userID)
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	return result.(*profile.UserProfile), args.Error(1)
}

func (m *MockProfileRepository) UpdateProfile(data *profile.UpdateProfileRequest) error {
	args := m.Called(data)
	return args.Error(0)
}

func TestGetProfile_Success(t *testing.T) {
	mockRepo := new(MockProfileRepository)
	service := profile.NewProfileService(mockRepo)

	expectedProfile := &profile.UserProfile{
		ID:       "123",
		Name:     "Alice",
		Email:    "alice@example.com",
		Role:     "responder",
		ImageURL: "https://cdn.com/image.jpg",
	}

	mockRepo.On("GetProfileByID", "123").Return(expectedProfile, nil)

	result, err := service.GetProfile("123")

	assert.NoError(t, err)
	assert.Equal(t, expectedProfile, result)
	mockRepo.AssertExpectations(t)
}

func TestUpdateProfile_ValidInput(t *testing.T) {
	mockRepo := new(MockProfileRepository)
	service := profile.NewProfileService(mockRepo)

	updateReq := &profile.UpdateProfileRequest{
		ID:       "123",
		Name:     "Bob",
		Email:    "bob@example.com",
		ImageURL: "https://cdn.com/bob.jpg",
	}

	mockRepo.On("UpdateProfile", updateReq).Return(nil)

	err := service.UpdateProfile(updateReq)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUpdateProfile_InvalidEmail(t *testing.T) {
	mockRepo := new(MockProfileRepository)
	service := profile.NewProfileService(mockRepo)

	updateReq := &profile.UpdateProfileRequest{
		ID:    "123",
		Name:  "Bob",
		Email: "invalid-email",
	}

	err := service.UpdateProfile(updateReq)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid email address")
}
