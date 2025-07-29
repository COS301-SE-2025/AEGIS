package verifyemail

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VerifyEmailRepository struct {
	db *gorm.DB
}

// NewVerifyEmailRepository creates a new instance of verifyEmailRepository
func NewVerifyEmailRepository(db *gorm.DB) *VerifyEmailRepository {

	return &VerifyEmailRepository{db: db}

}

func GetValidToken(db *gorm.DB, rawToken string) (*Token, error) {
	var token Token
	err := db.Where("token = ? AND type = ?", rawToken, "EMAIL_VERIFY").First(&token).Error
	if err != nil {
		return nil, errors.New("token not found or invalid")
	}
	if token.ExpiresAt != nil && time.Now().After(*token.ExpiresAt) {
		return nil, errors.New("token expired")
	}
	// if token.MaxUses != nil && token.Uses >= *token.MaxUses {
	// 	return nil, errors.New("token usage limit reached")
	// }
	return &token, nil
}

func IncrementTokenUse(db *gorm.DB, token *Token) error {
	update := db.Model(token).Update("uses", gorm.Expr("uses + 1"))
	// if token.MaxUses != nil && *token.MaxUses == 1 {
	// 	update = update.Update("used", true)
	// }
	return update.Error
}

func CreateEmailVerificationToken(db *gorm.DB, userID uuid.UUID) (string, error) {
	token := uuid.NewString()
	entry := Token{
		UserID:    userID,
		Token:     token,
		Type:      "EMAIL_VERIFY",
		Used:      false,
		CreatedAt: time.Now(),
		ExpiresAt: nil, // for never expiring tokens (for registration)
		//MaxUses: ptrInt(1),
	}
	if err := db.Create(&entry).Error; err != nil {
		return "", err
	}
	return token, nil
}

func ptrInt(i int) *int {
	return &i
}
