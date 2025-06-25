package accept_terms

import (
	"errors"

	"gorm.io/gorm"
)

func AcceptTerms(db *gorm.DB, userID string) error {
	IsVerified, err := IsUserVerified(db, userID)
	if err != nil {
		return err
	}
	if !IsVerified {
		return errors.New("user must verify email before accepting terms")
	}
	return SetTermsAccepted(db, userID)
}
