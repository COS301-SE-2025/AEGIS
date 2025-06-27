package accept_terms

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey"`
	FullName          string    `gorm:"not null"` // This is a derived field, not stored in the DB
	Email             string    `gorm:"uniqueIndex"`
	PasswordHash      string
	Role              string    `gorm:"type:user_role"`
	CreatedAt         time.Time `gorm:"autoCreateTime"`
	IsVerified        bool
	VerificationToken string //We send the token to the user’s email as a verification link, e.g.:
	EmailVerifiedAt   *time.Time
	AcceptedTermsAt   *time.Time
}
