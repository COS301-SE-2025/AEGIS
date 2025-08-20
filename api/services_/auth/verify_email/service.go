package verifyemail

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

func VerifyEmail(db *gorm.DB, rawToken string) error {
	token, err := GetValidToken(db, rawToken)
	if err != nil {
		return err
	}

	err = db.Model(&User{}).
		Where("id = ?", token.UserID).
		Updates(map[string]interface{}{
			"is_verified":       true,
			"email_verified_at": time.Now(),
		}).Error
	if err != nil {
		return errors.New("failed to update user verification")
	}

	return IncrementTokenUse(db, token)
}
