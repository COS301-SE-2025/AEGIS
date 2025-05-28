package reset_password

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormResetTokenRepository struct {
	db *gorm.DB
}

func NewGormResetTokenRepository(db *gorm.DB) *GormResetTokenRepository {
	return &GormResetTokenRepository{db: db}
}

type PasswordResetToken struct {
	Token     string    `gorm:"primaryKey"`
	UserID    uuid.UUID `gorm:"index"`
	ExpiresAt time.Time
	Used      bool
}

func (r *GormResetTokenRepository) CreateToken(userID uuid.UUID, token string, expiresAt time.Time) error {
	prt := PasswordResetToken{
		Token:     token,
		UserID:    userID,
		ExpiresAt: expiresAt,
		Used:      false,
	}
	return r.db.Create(&prt).Error
}
type passwordResetToken struct {
    Token     string    `gorm:"primaryKey"`
    UserID    uuid.UUID `gorm:"index"`
    ExpiresAt time.Time
    Used      bool
}
func (r *GormResetTokenRepository) GetUserIDByToken(token string) (uuid.UUID, time.Time, error) {
	var prt PasswordResetToken
	err := r.db.Where("token = ? AND used = ?", token, false).First(&prt).Error
	if err != nil {
		return uuid.Nil, time.Time{}, err
	}
	return prt.UserID, prt.ExpiresAt, nil
}

func (r *GormResetTokenRepository) MarkTokenUsed(token string) error {
	return r.db.Model(&PasswordResetToken{}).Where("token = ?", token).Update("used", true).Error
}
