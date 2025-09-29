package messages

import (
	"time"

	"aegis-api/pkg/encryption"

	"github.com/google/uuid"
	gorm "gorm.io/gorm"
)

// ThreadMessage represents a message in an annotation thread.
type ThreadMessage struct {
	ID              uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	ThreadID        uuid.UUID  `gorm:"not null"`
	ParentMessageID *uuid.UUID `gorm:"default:null"`
	UserID          uuid.UUID  `gorm:"not null"`
	Message         string     `gorm:"type:text;not null"`
	IsApproved      *bool
	ApprovedBy      *uuid.UUID
	ApprovedAt      *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time

	Mentions  []MessageMention  `gorm:"foreignKey:MessageID"`
	Reactions []MessageReaction `gorm:"foreignKey:MessageID"`
}

// ✅ Add GORM lifecycle hooks for encryption/decryption
func (m *ThreadMessage) BeforeSave(tx *gorm.DB) (err error) {
	if m.Message != "" {
		encrypted, err := encryption.Encrypt(m.Message)
		if err != nil {
			return err
		}
		m.Message = encrypted
	}
	return nil
}

func (m *ThreadMessage) AfterFind(tx *gorm.DB) (err error) {
	if m.Message != "" {
		decrypted, err := encryption.Decrypt(m.Message)
		if err != nil {
			m.Message = decrypted
		}

	}
	return nil
}

// MessageMention represents a mention in a thread message.
type MessageMention struct {
	MessageID       uuid.UUID `gorm:"primaryKey"`
	MentionedUserID uuid.UUID `gorm:"primaryKey"`
	CreatedAt       time.Time
}

// MessageReaction represents a reaction to a thread message.
type MessageReaction struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	MessageID uuid.UUID `gorm:"not null"`
	UserID    uuid.UUID `gorm:"not null"`
	Reaction  string    `gorm:"not null"`
	CreatedAt time.Time
}
