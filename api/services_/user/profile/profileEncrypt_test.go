package profile

import (
	"testing"

	"aegis-api/pkg/encryption"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"context"
)

func setupProfileTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.Exec(`
		CREATE TABLE update_profile_requests (
			id TEXT PRIMARY KEY,
			name TEXT,
			email TEXT,
			image_base64 TEXT,
			image_url TEXT
		)
	`).Error
	assert.NoError(t, err)

	err = db.Exec(`
		CREATE TABLE user_profiles (
			id TEXT PRIMARY KEY,
			name TEXT,
			email TEXT,
			role TEXT,
			image_url TEXT
		)
	`).Error
	assert.NoError(t, err)

	return db
}

func TestUpdateProfileRequestEncryption(t *testing.T) {
	err := encryption.InitializeService()
	assert.NoError(t, err)

	db := setupProfileTestDB(t)

	t.Run("encrypts profile update request", func(t *testing.T) {
		req := &UpdateProfileRequest{
			ID:       uuid.New().String(),
			Name:     "Alice Johnson",
			Email:    "alice@example.com",
			ImageURL: "https://cdn.example.com/alice.png",
		}

		err := db.Table("update_profile_requests").Create(req).Error
		assert.NoError(t, err)

		var retrieved UpdateProfileRequest
		err = db.Table("update_profile_requests").First(&retrieved, "id = ?", req.ID).Error
		assert.NoError(t, err)

		assert.Equal(t, "alice@example.com", retrieved.Email)
		assert.Equal(t, "https://cdn.example.com/alice.png", retrieved.ImageURL)
	})
}

func TestUserProfileEncryption(t *testing.T) {
	err := encryption.InitializeService()
	assert.NoError(t, err)

	db := setupProfileTestDB(t)

	t.Run("decrypts user profile on retrieval", func(t *testing.T) {
		// Manually insert encrypted data
		ctx := context.Background()
		service := encryption.GetService()
		
		encryptedEmail, _ := encryption.EncryptString(ctx, service, "user@example.com")
		encryptedURL, _ := encryption.EncryptString(ctx, service, "https://example.com/user.jpg")

		id := uuid.New().String()
		err := db.Exec(`
			INSERT INTO user_profiles (id, name, email, role, image_url)
			VALUES (?, ?, ?, ?, ?)
		`, id, "Test User", encryptedEmail, "admin", encryptedURL).Error
		assert.NoError(t, err)

		// Retrieve and verify decryption
		var profile UserProfile
		err = db.Table("user_profiles").First(&profile, "id = ?", id).Error
		assert.NoError(t, err)

		assert.Equal(t, "user@example.com", profile.Email)
		assert.Equal(t, "https://example.com/user.jpg", profile.ImageURL)
		assert.Equal(t, "Test User", profile.Name)
	})
}