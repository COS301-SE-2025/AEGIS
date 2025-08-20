package sharecase

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func CreateCaseShareToken(db *gorm.DB, userID uuid.UUID, caseID uuid.UUID, expiresIn time.Duration, maxUses int) (string, error) {
	token := uuid.New().String()
	exp := time.Now().Add(expiresIn)

	entry := Token{
		ID:        uuid.New(),
		UserID:    userID,
		CaseID:    &caseID,
		Token:     token,
		Type:      "CASE_SHARE",
		ExpiresAt: &exp,
		//MaxUses:   &maxUses,
		CreatedAt: time.Now(),
	}
	if err := db.Create(&entry).Error; err != nil {
		return "", err
	}
	return token, nil
}

func GetValidCaseToken(db *gorm.DB, rawToken string) (*Token, error) {
	var token Token
	err := db.Where("token = ? AND type = ?", rawToken, "CASE_SHARE").First(&token).Error
	if err != nil {
		return nil, errors.New("invalid or expired token")
	}

	if token.ExpiresAt != nil && time.Now().After(*token.ExpiresAt) {
		return nil, errors.New("token expired")
	}

	// if token.MaxUses != nil && token.Uses >= *token.MaxUses {
	// 	return nil, errors.New("token usage limit reached")
	// }

	return &token, nil
}
