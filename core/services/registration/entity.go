package registration

import (
	"time"
)
// This struct is saved in or loaded from the database. 
// It should contain the same fields as the DB table.

type User struct {
	ID           string    `gorm:"primaryKey"`
	FullName     string    `gorm:"not null"` // This is a derived field, not stored in the DB
	Email        string    `gorm:"uniqueIndex"`
	PasswordHash string
	Role         string    `gorm:"type:user_role`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	IsVerified        bool   
	VerificationToken string //We send the token to the user’s email as a verification link, e.g.:
}