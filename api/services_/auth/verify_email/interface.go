package verifyemail

import (
	"github.com/google/uuid"
)

type UserRepository interface {
	GetValidToken(token string) (*Token, error)
	incrementTokenUsage(token *Token) error
	CreateEmailVerificationToken(userID uuid.UUID) error
}

type VerifyEmailService interface {
	VerifyEmailService(rawToken string) error
	sendVerificationEmail(email string, token string) error
}
