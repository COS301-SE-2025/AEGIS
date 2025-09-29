package sharecase

import (
	"time"

	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func ShareCaseWithUser(db *gorm.DB, senderID, caseID uuid.UUID, recipientEmail string) error {
	// 1️⃣ Fetch sender's role in this case
	var senderRole string
	err := db.Raw(`
		SELECT role
		FROM case_user_roles
		WHERE user_id = ? AND case_id = ?
	`, senderID, caseID).Scan(&senderRole).Error
	if err != nil {
		return errors.New("failed to fetch sender role")
	}

	// 2️⃣ Check sender's permissions
	var perms []Permission
	err = db.Raw(`
		SELECT p.*
		FROM permissions p
		JOIN enum_role_permissions rp ON rp.permission_id = p.id
		WHERE rp.role = ?
	`, senderRole).Scan(&perms).Error
	if err != nil {
		return errors.New("failed to fetch permissions for role")
	}

	allowed := false
	for _, p := range perms {
		if p.name == "case:share" || p.name == "case:update" {
			allowed = true
			break
		}
	}

	if !allowed {
		return errors.New("user does not have permission to share this case")
	}

	// 3️⃣ Create token
	token, err := CreateCaseShareToken(db, senderID, caseID, 24*time.Hour)
	if err != nil {
		return err
	}

	// 4️⃣ Send email
	return SendCaseShareEmail(recipientEmail, token)
}

func AcceptCaseShare(db *gorm.DB, userID uuid.UUID, tokenStr string) error {
	// 1️⃣ Validate token
	token, err := GetValidCaseToken(db, tokenStr)
	if err != nil {
		return err
	}

	// 2️⃣ Check if user is already a collaborator
	var existing CaseCollaborator
	err = db.Where("case_id = ? AND user_id = ?", token.CaseID, userID).First(&existing).Error
	if err == nil {
		return errors.New("user is already a collaborator")
	}

	// 3️⃣ Add user as collaborator with 'Full Collaborator' role	// 2️⃣ Add user to case as External Collaborator
	err = AddUserToCase(db, *token.CaseID, userID, token.UserID, "External Collaborator", token.ExpiresAt)
	if err != nil {
		return err
	}

	// 4️⃣ Mark token as used
	token.Used = true
	return db.Save(&token).Error
}

// GetSharedCasesForUser: for dashboard "Shared With Me"
func GetSharedCasesForUser(db *gorm.DB, userID uuid.UUID) ([]CaseCollaborator, error) {
	var sharedCases []CaseCollaborator

	err := db.
		Where("user_id = ? AND status = ?", userID, "active").
		Find(&sharedCases).Error
	if err != nil {
		return nil, err
	}

	return sharedCases, nil
}

// LockExpiredCases: marks expired shared cases as revoked (for dashboard display)
func LockExpiredCases(db *gorm.DB) error {
	return db.Model(&CaseCollaborator{}).
		Where("expires_at IS NOT NULL AND expires_at <= ? AND status = ?", time.Now(), "active").
		Update("status", "expired").Error
}
