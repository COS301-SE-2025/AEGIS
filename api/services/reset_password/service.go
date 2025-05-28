package reset_password

import (
	//"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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

type PasswordResetService struct {
	repo       ResetTokenRepository
	users      UserRepository
	emailer    EmailSender
}

func NewPasswordResetService(repo ResetTokenRepository, users UserRepository, emailer EmailSender) *PasswordResetService {
	return &PasswordResetService{repo: repo, users: users, emailer: emailer}
}

func (s *PasswordResetService) RequestPasswordReset(userID uuid.UUID, email string) error {
	token := uuid.New().String()
	expires := time.Now().Add(30 * time.Minute)

	err := s.repo.CreateToken(userID, token, expires)
	if err != nil {
		return err
	}

	return s.emailer.SendPasswordResetEmail(email, token)
}

func (s *PasswordResetService) ResetPassword(token string, newPassword string) error {
	userID, expiresAt, err := s.repo.GetUserIDByToken(token)
	if err != nil {
		return errors.New("invalid or expired token")
	}

	if time.Now().After(expiresAt) {
		return errors.New("token has expired")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	err = s.users.UpdatePassword(userID, string(hashed))
	if err != nil {
		return err
	}

	return s.repo.MarkTokenUsed(token)
}



