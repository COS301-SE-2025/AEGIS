package GetUpdate_UserInfo

import (
	"time"

	"github.com/google/uuid"
	"aegis-api/pkg/encryption"
	gorm "gorm.io/gorm"
	"context"
)

type User struct {
	ID                   uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	FullName             string     `gorm:"not null" json:"full_name"`
	Email                string     `gorm:"unique;not null" json:"email"`
	PasswordHash         string     `gorm:"not null;column:password_hash" json:"-"` // hidden from JSON
	Role                 string     `gorm:"type:user_role" json:"role"`
	IsVerified           bool       `gorm:"default:false" json:"is_verified"`

	ProfilePictureURL    string     `json:"profile_picture_url,omitempty"`
	TokenVersion         int        `gorm:"default:1" json:"token_version,omitempty"`
	ExternalTokenExpiry  *time.Time `json:"external_token_expiry,omitempty"`
	ExternalTokenStatus  string     `gorm:"type:token_status;default:'active'" json:"external_token_status,omitempty"`

	UpdatedAt            time.Time  `gorm:"autoUpdateTime" json:"updated_at,omitempty"`
	CreatedAt            time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

type UserRole struct {
	UserID uuid.UUID `gorm:"type:uuid;primaryKey"`
	Role   string    `gorm:"type:user_role;primaryKey"`
}



func (u *User) BeforeSave(tx *gorm.DB) error {
	ctx := context.Background()
	service := encryption.GetService()

	// Encrypt Email
	if u.Email != "" && !encryption.IsEncryptedFormat(u.Email) {
		encrypted, err := encryption.EncryptString(ctx, service, u.Email)
		if err != nil {
			return err
		}
		u.Email = encrypted
	}

	// Encrypt PasswordHash
	if u.PasswordHash != "" && !encryption.IsEncryptedFormat(u.PasswordHash) {
		encrypted, err := encryption.EncryptString(ctx, service, u.PasswordHash)
		if err != nil {
			return err
		}
		u.PasswordHash = encrypted
	}

	// Encrypt ProfilePictureURL
	if u.ProfilePictureURL != "" && !encryption.IsEncryptedFormat(u.ProfilePictureURL) {
		encrypted, err := encryption.EncryptString(ctx, service, u.ProfilePictureURL)
		if err != nil {
			return err
		}
		u.ProfilePictureURL = encrypted
	}

	return nil
}

func (u *User) AfterFind(tx *gorm.DB) error {
	ctx := context.Background()
	service := encryption.GetService()

	// Decrypt Email
	if u.Email != "" {
		decrypted, err := encryption.DecryptString(ctx, service, u.Email)
		if err != nil {
			return err
		}
		u.Email = decrypted
	}

	// Decrypt PasswordHash
	if u.PasswordHash != "" {
		decrypted, err := encryption.DecryptString(ctx, service, u.PasswordHash)
		if err != nil {
			return err
		}
		u.PasswordHash = decrypted
	}

	// Decrypt ProfilePictureURL
	if u.ProfilePictureURL != "" {
		decrypted, err := encryption.DecryptString(ctx, service, u.ProfilePictureURL)
		if err != nil {
			return err
		}
		u.ProfilePictureURL = decrypted
	}

	return nil
}

func (u *User) AfterCreate(tx *gorm.DB) error {
	return u.AfterFind(tx)
}