package delete_user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey"`
	FullName          string
	Email             string    `gorm:"unique"`
	PasswordHash      string
	Role              string    `gorm:"type:user_role"`
	IsVerified        bool
	VerificationToken string
	CreatedAt         time.Time
}
