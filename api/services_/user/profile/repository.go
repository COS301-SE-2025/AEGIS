package profile

import (
	"aegis-api/db"
	"gorm.io/gorm"
)

// GormProfileRepository provides database operations using GORM.
type GormProfileRepository struct{
		db *gorm.DB

}

// NewGormProfileRepository creates and returns a new repository instance.
func NewGormProfileRepository() *GormProfileRepository {
	return &GormProfileRepository{}
}

// GetProfileByID retrieves the user's profile by their UUID.
// It maps DB fields (full_name, profile_picture_url) to struct fields (Name, ImageURL).
func (r *GormProfileRepository) GetProfileByID(userID string) (*UserProfile, error) {
	var user UserProfile
	err := db.DB.Raw(`
		SELECT id, full_name AS name, email, role::text AS role, profile_picture_url AS image_url
		FROM users
		WHERE id = ?
	`, userID).Scan(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateProfile updates the full_name, email, and profile_picture_url of the user.
// The ID field is used to identify which user to update.
func (r *GormProfileRepository) UpdateProfile(data *UpdateProfileRequest) error {
	return db.DB.Exec(`
		UPDATE users
		SET full_name = ?, email = ?, profile_picture_url = ?
		WHERE id = ?
	`, data.Name, data.Email, data.ImageURL, data.ID).Error
}
