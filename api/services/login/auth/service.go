package auth

import (
	//"aegis-api/services/registration"
	//database "aegis-api/db"
	"aegis-api/services/registration"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo registration.UserRepository
}

func NewAuthService(repo registration.UserRepository) *AuthService {
	return &AuthService{repo: repo}
}

func (s *AuthService) Login(email, password string) (*LoginResponse, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if user.Role == "External Collaborator" {
		if user.ExternalTokenStatus == "revoked" {
			return nil, fmt.Errorf("access revoked by administrator")
		}
		if user.ExternalTokenExpiry != nil && user.ExternalTokenExpiry.Before(time.Now()) {
			return nil, fmt.Errorf("access expired. contact administrator")
		}
	}

	token, err := GenerateJWT(user.ID, user.Email, user.Role, user.TokenVersion, user.ExternalTokenExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token")
	}

	return &LoginResponse{
		ID:         user.ID,
		Email:      user.Email,
		Token:      token,
		IsVerified: user.IsVerified,
		Role:       user.Role,
	}, nil
}

func (s *AuthService) RegenerateExternalToken(adminID, targetUserID string, req RegenerateTokenRequest) (*LoginResponse, error) {
	// Verify the admin is actually an admin (optional safety check)
	admin, err := s.repo.GetUserByID(adminID)
	if err != nil || admin.Role != "Admin" {
		return nil, fmt.Errorf("unauthorized: only admins can regenerate tokens")
	}

	// Get the target user (the external user)
	user, err := s.repo.GetUserByID(targetUserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	if user.Role != "External Collaborator" {
		return nil, fmt.Errorf("cannot regenerate token for non-external user")
	}

	// Set new expiry
	expiry := time.Now().Add(time.Duration(req.ExpiresInDays) * 24 * time.Hour)

	// Increment token version and activate token
	user.TokenVersion += 1
	user.ExternalTokenExpiry = &expiry
	user.ExternalTokenStatus = "active"

	err = s.repo.UpdateUserTokenInfo(user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user token info")
	}

	// Generate new JWT
	token, err := GenerateJWT(user.ID, user.Email, user.Role, user.TokenVersion, user.ExternalTokenExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token")
	}

	return &LoginResponse{
		ID:         user.ID,
		Email:      user.Email,
		Token:      token,
		IsVerified: user.IsVerified,
		Role:       user.Role,
	}, nil
}
