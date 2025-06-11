package reset_password

import (
	"time"

	"github.com/google/uuid"
)

type PasswordResetService struct {
	repo    ResetTokenRepository
	users   UserRepository
	emailer EmailSender
}

type PasswordResetToken struct {
	Token     string    `gorm:"primaryKey"`
	UserID    uuid.UUID `gorm:"index"`
	ExpiresAt time.Time
	Used      bool
}
