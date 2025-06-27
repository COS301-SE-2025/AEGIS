package sharecase

import (
	"time"

	"gorm.io/gorm"

	"github.com/google/uuid"
)

func ShareCaseWithUser(db *gorm.DB, userID, caseID uuid.UUID, email string) error {
	token, err := CreateCaseShareToken(db, userID, caseID, 24*time.Hour, 1)
	if err != nil {
		return err
	}

	return SendCaseShareEmail(email, token)
}
