package ListUsers

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID                  uuid.UUID `gorm:"type:uuid;primaryKey"`
	FullName            string    `gorm:"not null"` // This is a derived field, not stored in the DB
	Email               string    `gorm:"uniqueIndex"`
	PasswordHash        string
	Role                string `gorm:"type:user_role"`
	IsVerified          bool
	EmailVerifiedAt     *time.Time
	AcceptedTermsAt     *time.Time
	TokenVersion        int    `gorm:"default:1"`
	ExternalTokenStatus string `gorm:"default:''"`
	ExternalTokenExpiry *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
	TenantID            uuid.UUID `gorm:"type:uuid;index"` // Foreign key to Tenant
}

type ListUsersResponse struct {
	Users []User `json:"users"`
}

type listUserService struct {
	repo ListUserRepository
}
