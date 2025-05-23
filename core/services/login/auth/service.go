package auth

import (
	"aegis-api/services/registration"
	database "aegis-api/db"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func Login(email, password string) (*LoginResponse, error) {
	var user registration.User

	// Query the user by email
	if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Compare hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Return a clean response
	return &LoginResponse{
		ID:       user.ID,
        Token:    user.VerificationToken, // Assuming this is a JWT or similar token
		Email:    user.Email,
	}, nil
}
