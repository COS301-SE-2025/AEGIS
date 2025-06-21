package profile

import (
	"fmt"
	"strings"
)

// ProfileService is the business logic layer for profile-related operations.
type ProfileService struct {
	repo     ProfileRepository
	uploader IPFSUploader
}

// NewProfileService initializes a new ProfileService using a repository and IPFS uploader.
func NewProfileService(repo ProfileRepository, uploader IPFSUploader) *ProfileService {
    return &ProfileService{
        repo:     repo,
        uploader: uploader,
    }
}

// GetProfile fetches profile information for a user by ID.
func (s *ProfileService) GetProfile(userID string) (*UserProfile, error) {
	return s.repo.GetProfileByID(userID)
}

// UpdateProfile updates a user's profile using validated input.
func (s *ProfileService) UpdateProfile(data *UpdateProfileRequest) error {
	return s.repo.UpdateProfile(data)
}

// UpdateProfileWithImage updates a user's profile and handles image upload to IPFS.
func (s *ProfileService) UpdateProfileWithImage(data *UpdateProfileRequest, imageData []byte, filename string) error {
	// If image data is provided, upload to IPFS first
	if len(imageData) > 0 && filename != "" {
		imageURL, err := s.uploader.UploadProfilePicture(filename, imageData, data.ID)
		if err != nil {
			return fmt.Errorf("failed to upload profile picture: %w", err)
		}
		
		// Update the image URL in the request data
		data.ImageURL = imageURL
	}

	// Update the profile with the new data
	return s.repo.UpdateProfile(data)
}

// DeleteProfilePicture removes the profile picture from IPFS and updates the user's profile
func (s *ProfileService) DeleteProfilePicture(userID string) error {
	// First, get the current profile to find the image URL
	profile, err := s.repo.GetProfileByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user profile: %w", err)
	}

	// Extract IPFS hash from the image URL if it exists
	if profile.ImageURL != "" && strings.Contains(profile.ImageURL, "ipfs") {
		hash := s.extractIPFSHash(profile.ImageURL)
		if hash != "" {
			// Attempt to delete from IPFS (this only unpins from your node)
			err := s.uploader.DeleteFile(hash)
			if err != nil {
				// Log the error but don't fail the operation
				// since the file might still be accessible from other nodes
				fmt.Printf("Warning: failed to unpin file from IPFS: %v\n", err)
			}
		}
	}

	// Update the profile to remove the image URL
	updateData := &UpdateProfileRequest{
		ID:       userID,
		Name:     profile.Name,
		Email:    profile.Email,
		ImageURL: "", // Clear the image URL
	}

	return s.repo.UpdateProfile(updateData)
}

// extractIPFSHash extracts the IPFS hash from a URL
func (s *ProfileService) extractIPFSHash(url string) string {
	// Handle different IPFS URL formats
	// https://ipfs.io/ipfs/QmXXXXXX
	// https://gateway.ipfs.io/ipfs/QmXXXXXX
	// ipfs://QmXXXXXX
	
	if strings.Contains(url, "/ipfs/") {
		parts := strings.Split(url, "/ipfs/")
		if len(parts) > 1 {
			return strings.Split(parts[1], "/")[0] // Get hash before any additional path
		}
	}
	
	if strings.HasPrefix(url, "ipfs://") {
		return strings.TrimPrefix(url, "ipfs://")
	}
	
	return ""
}

// ValidateProfileUpdate performs business logic validation on profile update data
func (s *ProfileService) ValidateProfileUpdate(data *UpdateProfileRequest) error {
	if data.ID == "" {
		return fmt.Errorf("user ID is required")
	}
	
	if data.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	
	if data.Email == "" {
		return fmt.Errorf("email cannot be empty")
	}
	
	// Basic email validation
	if !strings.Contains(data.Email, "@") {
		return fmt.Errorf("invalid email format")
	}
	
	return nil
}