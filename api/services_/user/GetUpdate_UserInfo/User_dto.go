package GetUpdate_UserInfo

import (
	"time"

	"github.com/google/uuid"
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
