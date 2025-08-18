package sharecase

import (
	"time"

	"github.com/google/uuid"
)

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
