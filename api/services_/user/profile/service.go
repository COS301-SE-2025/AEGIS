package profile

import (
	"fmt"
	"strings"
)

type ProfileService struct {
	repo ProfileRepository
}

func NewProfileService(repo ProfileRepository) *ProfileService {
	return &ProfileService{repo: repo}
}

func (s *ProfileService) GetProfile(userID string) (*UserProfile, error) {
	return s.repo.GetProfileByID(userID)
}

func (s *ProfileService) UpdateProfile(data *UpdateProfileRequest) error {
	// Validate the input
	if err := s.ValidateProfileUpdate(data); err != nil {
		return err
	}
	return s.repo.UpdateProfile(data)
}

func (s *ProfileService) ValidateProfileUpdate(data *UpdateProfileRequest) error {
	if data.ID == "" {
		return fmt.Errorf("user ID is required")
	}
	if strings.TrimSpace(data.Name) == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if strings.TrimSpace(data.Email) == "" || !strings.Contains(data.Email, "@") {
		return fmt.Errorf("invalid email address")
	}
	return nil
}
