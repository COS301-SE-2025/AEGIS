package login

import (
	"aegis-api/services_/auth/registration"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	ID         string `json:"id"`
	Email      string `json:"email"`
	Token      string `json:"token"`
	Role       string `json:"role"`
	IsVerified bool   `json:"isVerified"`
}
type RegenerateTokenRequest struct {
	ExpiresInDays int `json:"expires_in_days"` // how many days until it expires
}

type AuthService struct {
	repo registration.UserRepository
}
