package GetUpdate_UserInfo

import (
	"testing"


	"aegis-api/pkg/encryption"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupUserTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.Exec(`
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			full_name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			role TEXT,
			is_verified BOOLEAN DEFAULT 0,
			profile_picture_url TEXT,
			token_version INTEGER DEFAULT 1,
			external_token_expiry DATETIME,
			external_token_status TEXT DEFAULT 'active',
			updated_at DATETIME,
			created_at DATETIME
		)
	`).Error
	assert.NoError(t, err)

	return db
}

func TestUserEncryption(t *testing.T) {
	err := encryption.InitializeService()
	assert.NoError(t, err)

	db := setupUserTestDB(t)

	t.Run("encrypts sensitive fields on create", func(t *testing.T) {
		user := &User{
			ID:                uuid.New(),
			FullName:          "John Doe",
			Email:             "john@example.com",
			PasswordHash:      "$2a$10$hashedpassword",
			Role:              "admin",
			IsVerified:        true,
			ProfilePictureURL: "https://example.com/pic.jpg",
			TokenVersion:      1,
		}

		err := db.Create(user).Error
		assert.NoError(t, err)

		var retrieved User
		err = db.First(&retrieved, "id = ?", user.ID).Error
		assert.NoError(t, err)

		assert.Equal(t, "john@example.com", retrieved.Email)
		assert.Equal(t, "$2a$10$hashedpassword", retrieved.PasswordHash)
		assert.Equal(t, "https://example.com/pic.jpg", retrieved.ProfilePictureURL)
		assert.Equal(t, "John Doe", retrieved.FullName)
	})

	t.Run("prevents double encryption on update", func(t *testing.T) {
		user := &User{
			ID:                uuid.New(),
			FullName:          "Jane Smith",
			Email:             "jane@example.com",
			PasswordHash:      "$2a$10$anotherhashedpw",
			Role:              "responder",
			ProfilePictureURL: "https://example.com/jane.jpg",
		}

		err := db.Create(user).Error
		assert.NoError(t, err)

		// Update
		user.FullName = "Jane Doe"
		err = db.Save(user).Error
		assert.NoError(t, err)

		var updated User
		err = db.First(&updated, "id = ?", user.ID).Error
		assert.NoError(t, err)

		assert.Equal(t, "jane@example.com", updated.Email)
		assert.Equal(t, "Jane Doe", updated.FullName)
	})

	t.Run("handles empty optional fields", func(t *testing.T) {
		user := &User{
			ID:           uuid.New(),
			FullName:     "Bob Test",
			Email:        "bob@example.com",
			PasswordHash: "$2a$10$bobhash",
			Role:         "viewer",
		}

		err := db.Create(user).Error
		assert.NoError(t, err)

		var retrieved User
		err = db.First(&retrieved, "id = ?", user.ID).Error
		assert.NoError(t, err)

		assert.Equal(t, "", retrieved.ProfilePictureURL)
		assert.Equal(t, "bob@example.com", retrieved.Email)
	})

	t.Run("bulk query decrypts all records", func(t *testing.T) {
		users := []*User{
			{
				ID:           uuid.New(),
				FullName:     "User One",
				Email:        "user1@example.com",
				PasswordHash: "$2a$10$hash1",
				Role:         "admin",
			},
			{
				ID:           uuid.New(),
				FullName:     "User Two",
				Email:        "user2@example.com",
				PasswordHash: "$2a$10$hash2",
				Role:         "admin",
			},
		}

		for _, u := range users {
			err := db.Create(u).Error
			assert.NoError(t, err)
		}

		var results []User
		err := db.Where("role = ?", "admin").Find(&results).Error
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 2)

		// Verify decryption worked
		for _, u := range results {
			assert.NotContains(t, u.Email, "mock::")
			assert.NotContains(t, u.PasswordHash, "mock::")
		}
	})
}