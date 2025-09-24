package sharecase

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func CreateCaseShareToken(db *gorm.DB, senderID, caseID uuid.UUID, expiresIn time.Duration) (string, error) {
	token := uuid.New().String()
	exp := time.Now().Add(expiresIn)

	entry := Token{
		ID:        uuid.New(),
		UserID:    senderID,
		CaseID:    &caseID,
		Token:     token,
		Type:      "CASE_SHARE",
		ExpiresAt: &exp,
		Used:      false,
		CreatedAt: time.Now(),
	}
	if err := db.Create(&entry).Error; err != nil {
		return "", err
	}
	return token, nil
}

// GetValidCaseToken validates token and expiry
func GetValidCaseToken(db *gorm.DB, rawToken string) (*Token, error) {
	var token Token
	if err := db.Where("token = ? AND type = ?", rawToken, "CASE_SHARE").First(&token).Error; err != nil {
		return nil, errors.New("invalid token")
	}
	if token.Used {
		return nil, errors.New("token already used")
	}
	if token.ExpiresAt != nil && time.Now().After(*token.ExpiresAt) {
		return nil, errors.New("token expired")
	}
	return &token, nil
}

func ExpireOldCaseShares(db *gorm.DB) error {
	return db.Exec(`
		UPDATE case_collaborators cc
		SET status = 'expired'
		FROM tokens t
		WHERE t.case_id = cc.case_id
		  AND t.type = 'CASE_SHARE'
		  AND t.expires_at IS NOT NULL
		  AND t.expires_at < NOW()
	`).Error
}

func ListSharedCases(db *gorm.DB, userID uuid.UUID) ([]Case, error) {
	var cases []Case
	err := db.Raw(`
		SELECT c.* 
		FROM cases c
		JOIN case_collaborators cc ON cc.case_id = c.id
		WHERE cc.user_id = ? AND cc.status = 'active'
	`, userID).Scan(&cases).Error
	return cases, err
}

func RedeemCaseShareToken(db *gorm.DB, rawToken string, recipientID uuid.UUID, role string) error {
	var token Token
	err := db.Where("token = ? AND type = ?", rawToken, "CASE_SHARE").First(&token).Error
	if err != nil {
		return errors.New("invalid token")
	}

	if token.ExpiresAt != nil && time.Now().After(*token.ExpiresAt) {
		return errors.New("token expired")
	}

	// Add user to case_collaborators if not already
	collab := CaseCollaborator{
		ID:        uuid.New(),
		CaseID:    *token.CaseID,
		UserID:    recipientID,
		Role:      role,
		InvitedBy: token.UserID,
		Status:    "active",
		InvitedAt: time.Now(),
	}
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&collab).Error; err != nil {
		return err
	}

	// Mark token as used
	token.Used = true
	return db.Save(&token).Error
}

// AddUserToCase inserts a new collaborator
func AddUserToCase(db *gorm.DB, caseID, userID, invitedBy uuid.UUID, role string, expiresAt *time.Time) error {
	collab := CaseCollaborator{
		ID:        uuid.New(),
		CaseID:    caseID,
		UserID:    userID,
		Role:      role,
		InvitedBy: invitedBy,
		InvitedAt: time.Now(),
		ExpiresAt: expiresAt,
		Status:    "active",
	}
	return db.Create(&collab).Error
}
