package profile

// ProfileRepository defines the behavior for working with user profile data.
type ProfileRepository interface {
	// GetProfileByID retrieves the profile information of a user by their ID.
	GetProfileByID(userID string) (*UserProfile, error)

	// UpdateProfile updates the user's name, email, and optionally profile picture.
	UpdateProfile(data *UpdateProfileRequest) error
}
