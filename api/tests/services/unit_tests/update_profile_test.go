package unit_tests

import (
	"errors"
	"strings"
	"testing"
	"aegis-api/services_/user/profile"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)
var _ profile.IPFSUploader = (*MockUploader)(nil)


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

// MockUploader is a mocked IPFSUploader.
type MockUploader struct {
	mock.Mock
}

func (m *MockUploader) UploadProfilePicture(filename string, fileData []byte, userID string) (string, error) {
	args := m.Called(filename, fileData, userID)
	return args.String(0), args.Error(1)
}

func (m *MockUploader) DeleteFile(hash string) error {
	args := m.Called(hash)
	return args.Error(0)
}

func TestGetProfile_Success(t *testing.T) {
	repo := new(MockProfileRepository)
	uploader := new(MockUploader)
	service := profile.NewProfileService(repo, uploader)

	expected := &profile.UserProfile{
		ID:       "123",
		Name:     "John Doe",
		Email:    "john@example.com",
		Role:     "admin",
		ImageURL: "https://ipfs.io/ipfs/QmXYZ",
	}

	repo.On("GetProfileByID", "123").Return(expected, nil)

	profileData, err := service.GetProfile("123")
	assert.NoError(t, err)
	assert.Equal(t, expected, profileData)
}

func TestGetProfile_UserNotFound(t *testing.T) {
	repo := new(MockProfileRepository)
	uploader := new(MockUploader)
	service := profile.NewProfileService(repo, uploader)

	repo.On("GetProfileByID", "not-found").Return(nil, errors.New("user not found"))

	profileData, err := service.GetProfile("not-found")
	assert.Nil(t, profileData)
	assert.Error(t, err)
	assert.EqualError(t, err, "user not found")
}

func TestUpdateProfile_Success(t *testing.T) {
	repo := new(MockProfileRepository)
	uploader := new(MockUploader)
	service := profile.NewProfileService(repo, uploader)

	req := &profile.UpdateProfileRequest{
		ID:       "123",
		Name:     "Jane Doe",
		Email:    "jane@example.com",
		ImageURL: "https://ipfs.io/ipfs/QmABC",
	}

	repo.On("UpdateProfile", req).Return(nil)

	err := service.UpdateProfile(req)
	assert.NoError(t, err)
}
func TestUpdateProfile_DBError(t *testing.T) {
	repo := new(MockProfileRepository)
	uploader := new(MockUploader)
	service := profile.NewProfileService(repo, uploader)

	req := &profile.UpdateProfileRequest{
		ID:    "123",
		Name:  "Jane Doe",
		Email: "jane@example.com",
	}

	repo.On("UpdateProfile", req).Return(errors.New("db update failed"))

	err := service.UpdateProfile(req)
	assert.Error(t, err)
	assert.EqualError(t, err, "db update failed")
}

func TestValidateProfileUpdate_InvalidEmail(t *testing.T) {
	service := profile.NewProfileService(nil, nil)

	req := &profile.UpdateProfileRequest{
		ID:    "123",
		Name:  "Alice",
		Email: "invalid-email", // no '@'
	}

	err := service.ValidateProfileUpdate(req)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "invalid email"))
}

func TestDeleteProfilePicture_UnpinFailsButStillUpdates(t *testing.T) {
	repo := new(MockProfileRepository)
	uploader := new(MockUploader)
	service := profile.NewProfileService(repo, uploader)

	// Simulated existing profile
	existing := &profile.UserProfile{
		ID:       "user-1",
		Name:     "Alice",
		Email:    "alice@example.com",
		ImageURL: "https://ipfs.io/ipfs/QmFakeHash",
	}

	repo.On("GetProfileByID", "user-1").Return(existing, nil)
	uploader.On("DeleteFile", "QmFakeHash").Return(errors.New("node offline")) // Simulated IPFS failure
	repo.On("UpdateProfile", mock.AnythingOfType("*profile.UpdateProfileRequest")).Return(nil)

	err := service.DeleteProfilePicture("user-1")
	assert.NoError(t, err) // should not fail even if unpin fails
}
func TestDeleteProfilePicture_Success(t *testing.T) {
	repo := new(MockProfileRepository)
	uploader := new(MockUploader)
	service := profile.NewProfileService(repo, uploader)

	existing := &profile.UserProfile{
		ID:       "user-2",
		Name:     "Bob",
		Email:    "bob@example.com",
		ImageURL: "https://ipfs.io/ipfs/QmHashToDelete",
	}

	repo.On("GetProfileByID", "user-2").Return(existing, nil)
	uploader.On("DeleteFile", "QmHashToDelete").Return(nil)

	repo.On("UpdateProfile", mock.MatchedBy(func(data *profile.UpdateProfileRequest) bool {
		return data.ID == "user-2" && data.ImageURL == ""
	})).Return(nil)

	err := service.DeleteProfilePicture("user-2")
	assert.NoError(t, err)
}