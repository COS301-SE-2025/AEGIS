package reset_password

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// NewPasswordResetService creates a new PasswordResetService with the required dependencies.
func NewPasswordResetService(repo ResetTokenRepository, users UserRepository, emailer EmailSender) *PasswordResetService {
	return &PasswordResetService{repo: repo, users: users, emailer: emailer}
}

// RequestPasswordReset generates a password reset token, stores it, and sends an email to the user.
func (s *PasswordResetService) RequestPasswordReset(userID uuid.UUID, email string) error {
	token := uuid.New().String()
	expires := time.Now().Add(30 * time.Minute)

	// Save the token to the repository
	err := s.repo.CreateToken(userID, token, expires)
	if err != nil {
		return err
	}

	// Send the token via email
	return s.emailer.SendPasswordResetEmail(email, token)
}

// ResetPassword validates the token, checks expiry, updates the user's password, and marks the token as used.
func (s *PasswordResetService) ResetPassword(token string, newPassword string) error {
	userID, expiresAt, err := s.repo.GetUserIDByToken(token)
	if err != nil {
		return errors.New("invalid or expired token")
	}

	if time.Now().After(expiresAt) {
		return errors.New("token has expired")
	}

	// Hash the new password
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Save the new password
	err = s.users.UpdatePassword(userID, string(hashed))
	if err != nil {
		return err
	}

	// Mark the token as used
	return s.repo.MarkTokenUsed(token)
}
