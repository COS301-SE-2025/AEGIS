package reset_password

import (
	"time"

	"github.com/google/uuid"
)

// ResetTokenRepository defines operations for managing password reset tokens.
type ResetTokenRepository interface {
	// CreateToken stores a new reset token for a user with an expiry time.
	CreateToken(userID uuid.UUID, token string, expiresAt time.Time) error

	// GetUserIDByToken returns the associated user ID and expiry for a given valid token.
	GetUserIDByToken(token string) (uuid.UUID, time.Time, error)

	// MarkTokenUsed marks a token as used to prevent reuse.
	MarkTokenUsed(token string) error
}

// UserRepository defines user account operations.
type UserRepository interface {
	// UpdatePassword hashes and updates a user's password in the database.
	UpdatePassword(userID uuid.UUID, hashedPassword string) error
}

// EmailSender defines the behavior for sending password reset emails.
type EmailSender interface {
	// SendPasswordResetEmail sends a reset email with the reset token to the user.
	SendPasswordResetEmail(email string, token string) error
}
