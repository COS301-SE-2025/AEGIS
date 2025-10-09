package mfa

import (
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"go.uber.org/zap"
)

// MFAService handles Multi-Factor Authentication operations
type MFAService struct {
	db     *sqlx.DB
	logger *zap.Logger
}

// MFASetupResponse represents the response when setting up MFA
type MFASetupResponse struct {
	Secret      string   `json:"secret"`
	QRCodeURL   string   `json:"qr_code_url"`
	BackupCodes []string `json:"backup_codes"`
}

// MFAVerificationRequest represents an MFA verification request
type MFAVerificationRequest struct {
	UserID string `json:"user_id"`
	Code   string `json:"code"`
}

// BackupCode represents a backup code for MFA
type BackupCode struct {
	ID        uuid.UUID  `db:"id" json:"id"`
	UserID    uuid.UUID  `db:"user_id" json:"user_id"`
	Code      string     `db:"code" json:"code"`
	Used      bool       `db:"used" json:"used"`
	UsedAt    *time.Time `db:"used_at" json:"used_at,omitempty"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
}

// NewMFAService creates a new MFA service instance
func NewMFAService(db *sqlx.DB, logger *zap.Logger) *MFAService {
	return &MFAService{
		db:     db,
		logger: logger,
	}
}

// GenerateSecret generates a new TOTP secret for a user
func (s *MFAService) GenerateSecret(userID uuid.UUID, userEmail string) (*MFASetupResponse, error) {
	// Generate a random secret
	secret := make([]byte, 20)
	_, err := rand.Read(secret)
	if err != nil {
		s.logger.Error("Failed to generate MFA secret", zap.Error(err))
		return nil, fmt.Errorf("failed to generate MFA secret: %w", err)
	}

	secretBase32 := base32.StdEncoding.EncodeToString(secret)

	// Generate QR code URL
	key, err := otp.NewKeyFromURL(fmt.Sprintf(
		"otpauth://totp/AEGIS:%s?secret=%s&issuer=AEGIS",
		userEmail, secretBase32,
	))
	if err != nil {
		s.logger.Error("Failed to generate OTP key", zap.Error(err))
		return nil, fmt.Errorf("failed to generate OTP key: %w", err)
	}

	// Generate backup codes
	backupCodes, err := s.generateBackupCodes(userID)
	if err != nil {
		s.logger.Error("Failed to generate backup codes", zap.Error(err))
		return nil, fmt.Errorf("failed to generate backup codes: %w", err)
	}

	// Store the secret in database (but don't enable MFA yet)
	err = s.storeMFASecret(userID, secretBase32)
	if err != nil {
		s.logger.Error("Failed to store MFA secret", zap.Error(err))
		return nil, fmt.Errorf("failed to store MFA secret: %w", err)
	}

	return &MFASetupResponse{
		Secret:      secretBase32,
		QRCodeURL:   key.URL(),
		BackupCodes: backupCodes,
	}, nil
}

// VerifyAndEnableMFA verifies the initial TOTP code and enables MFA for the user
func (s *MFAService) VerifyAndEnableMFA(userID uuid.UUID, code string) error {
	// Get the stored secret
	var secret string
	err := s.db.Get(&secret, `
        SELECT COALESCE(mfa_secret, '') 
        FROM users 
        WHERE id = $1
    `, userID)
	if err != nil {
		s.logger.Error("Failed to get MFA secret", zap.Error(err))
		return fmt.Errorf("failed to get MFA secret: %w", err)
	}

	if secret == "" {
		return errors.New("no MFA secret found for user")
	}

	// Verify the code
	valid := totp.Validate(code, secret)
	if !valid {
		s.logger.Warn("Invalid MFA verification code", zap.String("user_id", userID.String()))
		return errors.New("invalid MFA code")
	}

	// Enable MFA for the user
	_, err = s.db.Exec(`
        UPDATE users 
        SET mfa_enabled = true, mfa_setup_completed_at = NOW()
        WHERE id = $1
    `, userID)
	if err != nil {
		s.logger.Error("Failed to enable MFA", zap.Error(err))
		return fmt.Errorf("failed to enable MFA: %w", err)
	}

	s.logger.Info("MFA enabled for user", zap.String("user_id", userID.String()))
	return nil
}

// VerifyTOTP verifies a TOTP code for an MFA-enabled user
func (s *MFAService) VerifyTOTP(userID uuid.UUID, code string) (bool, error) {
	// Get user's MFA details
	var user struct {
		MFAEnabled bool   `db:"mfa_enabled"`
		MFASecret  string `db:"mfa_secret"`
	}

	err := s.db.Get(&user, `
        SELECT 
            COALESCE(mfa_enabled, false) as mfa_enabled,
            COALESCE(mfa_secret, '') as mfa_secret
        FROM users 
        WHERE id = $1
    `, userID)
	if err != nil {
		s.logger.Error("Failed to get user MFA details", zap.Error(err))
		return false, fmt.Errorf("failed to get user MFA details: %w", err)
	}

	if !user.MFAEnabled {
		return false, errors.New("MFA is not enabled for this user")
	}

	if user.MFASecret == "" {
		return false, errors.New("no MFA secret configured for user")
	}

	// First, try to verify as TOTP code
	if totp.Validate(code, user.MFASecret) {
		return true, nil
	}

	// If TOTP fails, try backup codes
	valid, err := s.verifyBackupCode(userID, code)
	if err != nil {
		s.logger.Error("Failed to verify backup code", zap.Error(err))
		return false, err
	}

	return valid, nil
}

// DisableMFA disables MFA for a user
func (s *MFAService) DisableMFA(userID uuid.UUID) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Disable MFA and clear secret
	_, err = tx.Exec(`
        UPDATE users 
        SET mfa_enabled = false, mfa_secret = NULL, mfa_setup_completed_at = NULL
        WHERE id = $1
    `, userID)
	if err != nil {
		s.logger.Error("Failed to disable MFA", zap.Error(err))
		return fmt.Errorf("failed to disable MFA: %w", err)
	}

	// Delete all backup codes
	_, err = tx.Exec(`DELETE FROM mfa_backup_codes WHERE user_id = $1`, userID)
	if err != nil {
		s.logger.Error("Failed to delete backup codes", zap.Error(err))
		return fmt.Errorf("failed to delete backup codes: %w", err)
	}

	if err = tx.Commit(); err != nil {
		s.logger.Error("Failed to commit MFA disable transaction", zap.Error(err))
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Info("MFA disabled for user", zap.String("user_id", userID.String()))
	return nil
}

// GetMFAStatus returns the MFA status for a user
func (s *MFAService) GetMFAStatus(userID uuid.UUID) (bool, error) {
	var enabled bool
	err := s.db.Get(&enabled, `
        SELECT COALESCE(mfa_enabled, false) 
        FROM users 
        WHERE id = $1
    `, userID)
	if err != nil {
		s.logger.Error("Failed to get MFA status", zap.Error(err))
		return false, fmt.Errorf("failed to get MFA status: %w", err)
	}

	return enabled, nil
}

// generateBackupCodes generates backup codes for MFA
func (s *MFAService) generateBackupCodes(userID uuid.UUID) ([]string, error) {
	const numCodes = 10
	codes := make([]string, numCodes)

	// Delete existing backup codes
	_, err := s.db.Exec(`DELETE FROM mfa_backup_codes WHERE user_id = $1`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete existing backup codes: %w", err)
	}

	// Generate new codes
	for i := 0; i < numCodes; i++ {
		code, err := s.generateRandomCode(8)
		if err != nil {
			return nil, fmt.Errorf("failed to generate backup code: %w", err)
		}
		codes[i] = code

		// Store in database
		_, err = s.db.Exec(`
            INSERT INTO mfa_backup_codes (id, user_id, code, used, created_at)
            VALUES ($1, $2, $3, false, NOW())
        `, uuid.New(), userID, code)
		if err != nil {
			return nil, fmt.Errorf("failed to store backup code: %w", err)
		}
	}

	return codes, nil
}

// verifyBackupCode verifies and marks a backup code as used
func (s *MFAService) verifyBackupCode(userID uuid.UUID, code string) (bool, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return false, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Check if backup code exists and is unused
	var backupCodeID uuid.UUID
	err = tx.Get(&backupCodeID, `
        SELECT id FROM mfa_backup_codes 
        WHERE user_id = $1 AND code = $2 AND used = false
    `, userID, code)
	if err != nil {
		// Code not found or already used
		return false, nil
	}

	// Mark backup code as used
	_, err = tx.Exec(`
        UPDATE mfa_backup_codes 
        SET used = true, used_at = NOW()
        WHERE id = $1
    `, backupCodeID)
	if err != nil {
		s.logger.Error("Failed to mark backup code as used", zap.Error(err))
		return false, fmt.Errorf("failed to mark backup code as used: %w", err)
	}

	if err = tx.Commit(); err != nil {
		s.logger.Error("Failed to commit backup code verification", zap.Error(err))
		return false, fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Info("Backup code used",
		zap.String("user_id", userID.String()),
		zap.String("backup_code_id", backupCodeID.String()))

	return true, nil
}

// storeMFASecret stores the MFA secret for a user (without enabling MFA yet)
func (s *MFAService) storeMFASecret(userID uuid.UUID, secret string) error {
	_, err := s.db.Exec(`
        UPDATE users 
        SET mfa_secret = $1
        WHERE id = $2
    `, secret, userID)
	return err
}

// generateRandomCode generates a random alphanumeric code
func (s *MFAService) generateRandomCode(length int) (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)

	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		b[i] = charset[num.Int64()]
	}

	return string(b), nil
}
