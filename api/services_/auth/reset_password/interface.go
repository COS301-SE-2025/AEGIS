package reset_password

import (
	"time"

	"github.com/google/uuid"
)

type ResetTokenRepository interface {
	CreateToken(userID uuid.UUID, token string, expiresAt time.Time) error
	GetUserIDByToken(token string) (uuid.UUID, time.Time, error)
	MarkTokenUsed(token string) error
}

type UserRepository interface {
	UpdatePassword(userID uuid.UUID, hashedPassword string) error
}

type EmailSender interface {
	SendPasswordResetEmail(email string, token string) error
}
