package profile

import(
	"context"
	"aegis-api/pkg/encryption"
	gorm "gorm.io/gorm"
)

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



func (u *UpdateProfileRequest) BeforeSave(tx *gorm.DB) error {
	ctx := context.Background()
	service := encryption.GetService()

	if u.Email != "" && !encryption.IsEncryptedFormat(u.Email) {
		encrypted, err := encryption.EncryptString(ctx, service, u.Email)
		if err != nil {
			return err
		}
		u.Email = encrypted
	}

	if u.ImageURL != "" && !encryption.IsEncryptedFormat(u.ImageURL) {
		encrypted, err := encryption.EncryptString(ctx, service, u.ImageURL)
		if err != nil {
			return err
		}
		u.ImageURL = encrypted
	}

	return nil
}

func (u *UpdateProfileRequest) AfterFind(tx *gorm.DB) error {
	ctx := context.Background()
	service := encryption.GetService()

	if u.Email != "" {
		decrypted, err := encryption.DecryptString(ctx, service, u.Email)
		if err != nil {
			return err
		}
		u.Email = decrypted
	}

	if u.ImageURL != "" {
		decrypted, err := encryption.DecryptString(ctx, service, u.ImageURL)
		if err != nil {
			return err
		}
		u.ImageURL = decrypted
	}

	return nil
}

func (u *UpdateProfileRequest) AfterCreate(tx *gorm.DB) error {
	return u.AfterFind(tx)
}

func (u *UserProfile) AfterFind(tx *gorm.DB) error {
	ctx := context.Background()
	service := encryption.GetService()

	if u.Email != "" {
		decrypted, err := encryption.DecryptString(ctx, service, u.Email)
		if err != nil {
			return err
		}
		u.Email = decrypted
	}

	if u.ImageURL != "" {
		decrypted, err := encryption.DecryptString(ctx, service, u.ImageURL)
		if err != nil {
			return err
		}
		u.ImageURL = decrypted
	}

	return nil
}