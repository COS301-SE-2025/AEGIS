package registration

import (
	"time"

	"github.com/google/uuid"
)

// UserModel represents the domain model used in the business logic layer.
// It excludes database-specific concerns like ID and timestamps.
type UserModel struct {
	FullName     string
	Email        string
	PasswordHash string
	Role         string // ENUM: Incident Responder, Forensic Analyst, etc.
}

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

type UserResponse struct {
	ID       string `json:"id"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
}

type User struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey"`
	FullName          string    `gorm:"not null"` // This is a derived field, not stored in the DB
	Email             string    `gorm:"uniqueIndex"`
	PasswordHash      string
	Role              string    `gorm:"type:user_role"`
	CreatedAt         time.Time `gorm:"autoCreateTime"`
	IsVerified        bool
	VerificationToken string //We send the token to the user’s email as a verification link, e.g.:
	EmailVerifiedAt   *time.Time
	AcceptedTermsAt   *time.Time
}

type Token struct {
	ID        uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null"`
	CaseID    *uuid.UUID `gorm:"type:uuid"` // 👈 only used for CASE_SHARE
	Token     string     `gorm:"uniqueIndex;not null"`
	Type      string     `gorm:"type:VARCHAR(30);not null"`
	ExpiresAt *time.Time
	Used      bool `gorm:"default:false"`
	Uses      int  `gorm:"default:0"`
	MaxUses   *int
	CreatedAt time.Time
}
