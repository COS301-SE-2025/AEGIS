package accept_terms

import (
	"time"

	"gorm.io/gorm"
)

func IsUserVerified(db *gorm.DB, userID string) (bool, error) {
	var verified bool
	err := db.Model(&User{}).
		Select("Isverified").
		Where("id = ?", userID).
		Scan(&verified).Error
	if err != nil {
		return false, err
	}
	return verified, nil
}

func SetTermsAccepted(db *gorm.DB, userID string) error {
	return db.Model(&User{}).
		Where("id = ?", userID).
		Update("AcceptedTermsAt", time.Now()).Error
}
