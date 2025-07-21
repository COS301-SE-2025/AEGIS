package profile

// UpdateProfileRequest represents the data that a user can update in their profile.
type UpdateProfileRequest struct {
	ID          string `json:"id"`          // UUID of the user
	Name        string `json:"name"`        // Full name to update
	Email       string `json:"email"`       // Email to update
	ImageBase64 string `json:"imageBase64"` // New profile picture URL (optional)
	ImageURL    string `json:"-"`           // internal use only,
}

// UserProfile represents the full profile information that can be retrieved for a user.
type UserProfile struct {
	ID       string `json:"id"`        // UUID of the user
	Name     string `json:"name"`      // Full name of the user
	Email    string `json:"email"`     // Email address of the user
	Role     string `json:"role"`      // User's role (e.g., admin, responder)
	ImageURL string `json:"image_url"` // URL to the user's profile picture
}
