package ListUsers

import (
	"github.com/google/uuid"
	"time"
)
// User represents a user in the system
type User struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Full_name string    `gorm:"type:varchar(255);not null" json:"full_name"`
	Email     string    `gorm:"type:varchar(255);unique;not null" json:"email"`
	password_hash string    `gorm:"type:varchar(255);not null" json:"-"`
	Role      string    `gorm:"type:varchar(50);default:'user'" json:"role"`
	IsVerified bool      `gorm:"default:false" json:"is_verified"`
	verification_token string    `gorm:"type:varchar(255);unique" json:"-"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}



