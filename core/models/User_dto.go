package models

import "time"

type UserDTO struct {
	ID                string    `json:"id"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	PasswordHash      string    `json:"-"` // Exclude from JSON responses for security
	Role              string    `json:"role"`
	IsVerified        bool      `json:"is_verified"`
	VerificationToken string    `json:"verification_token,omitempty"` // Omit if empty
	CreatedAt         time.Time `json:"created_at"`
}