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

func (r *GormResetTokenRepository) CreateToken(userID uuid.UUID, token string, expiresAt time.Time) error {
	prt := PasswordResetToken{
		Token:     token,
		UserID:    userID,
		ExpiresAt: expiresAt,
		Used:      false,
	}
	return r.db.Create(&prt).Error
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
