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
			"Isverified":      true,
			"EmailVerifiedAt": time.Now(),
		}).Error
	if err != nil {
		return errors.New("failed to update user verification")
	}

	return IncrementTokenUse(db, token)
}
