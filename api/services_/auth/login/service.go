package login

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func NewAuthService(repo UserRepository) *AuthService {
	return &AuthService{repo: repo}
}

func (s *AuthService) Login(email, password string) (*LoginResponse, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil || user == nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	return &LoginResponse{
		ID:    user.ID,
		Email: user.Email,
		Token: user.VerificationToken, // Later replace with JWT
	}, nil
}
