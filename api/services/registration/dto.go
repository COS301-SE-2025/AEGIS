package registration

//Web Layer
// This layer is responsible for handling HTTP requests and responses.
// It decodes incoming requests, calls the service layer, and encodes the responses.
// It should not contain any business logic or data access code.
// It should only handle HTTP-specific concerns like request/response encoding/decoding.

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type RegistrationRequest struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
	/*
		Password is required to be hashed.
		From client side, password is sent in plain text.
		On the server side, it is hashed using bcrypt before storage.
	*/

}
type ResendVerificationRequest struct {
	Email string `json:"email"`
}

type UserResponse struct {
	ID       string `json:"id"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
}

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"` // use "verify" for email token
	jwt.RegisteredClaims
}

type User struct {
	ID                  string     `gorm:"primaryKey" json:"id"`
	FullName            string     `json:"full_name"`
	Email               string     `json:"email"`
	PasswordHash        string     `json:"-"` // Do not expose in JSON responses
	Role                string     `json:"role"`
	TokenVersion        int        `json:"token_version"`
	IsVerified          bool       `json:"is_verified"`
	ExternalTokenStatus string     `json:"external_token_status"` // "active", "revoked", etc.
	ExternalTokenExpiry *time.Time `json:"external_token_expiry"` // nullable
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}
