package reset_password

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormResetTokenRepository is a GORM-based implementation of the ResetTokenRepository interface.

type GormResetTokenRepository struct {
	db *gorm.DB
}

// Token represents a row in the unified 'tokens' table.
// It includes general-purpose fields and is filtered by token type (e.g., RESET_PASSWORD).

type Token struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;index"`
	Type      string    `gorm:"type:token_type"`
	Token     string    `gorm:"uniqueIndex"`
	CreatedAt time.Time
	ExpiresAt time.Time
	Used      bool
}

// CreateToken creates a new RESET_PASSWORD token entry for the specified user.
func NewGormResetTokenRepository(db *gorm.DB) *GormResetTokenRepository {
	return &GormResetTokenRepository{db: db}
}

// Create a token of type RESET_PASSWORD
func (r *GormResetTokenRepository) CreateToken(userID uuid.UUID, token string, expiresAt time.Time) error {
	t := Token{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      "RESET_PASSWORD",
		Token:     token,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
		Used:      false,
	}
	return r.db.Create(&t).Error
}

// GetUserIDByToken fetches the user ID and expiration for a valid, unused RESET_PASSWORD token.
func (r *GormResetTokenRepository) GetUserIDByToken(token string) (uuid.UUID, time.Time, error) {
	var t Token
	err := r.db.
		Where("token = ? AND used = FALSE AND type = ?", token, "RESET_PASSWORD").
		First(&t).Error
	if err != nil {
		return uuid.Nil, time.Time{}, err
	}
	return t.UserID, t.ExpiresAt, nil
}

// MarkTokenUsed sets the used flag to true for a given RESET_PASSWORD token, preventing reuse.
func (r *GormResetTokenRepository) MarkTokenUsed(token string) error {
	return r.db.
		Model(&Token{}).
		Where("token = ? AND type = ?", token, "RESET_PASSWORD").
		Update("used", true).Error
}
