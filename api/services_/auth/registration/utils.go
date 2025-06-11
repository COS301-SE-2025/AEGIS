package registration

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// generateID generates a new UUID string
func generateID() string {
	return uuid.New().String()
}
func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// HashPassword hashes a plain-text password using bcrypt.
func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashed), err
}

func (r RegistrationRequest) Validate() error {
	if strings.TrimSpace(r.FullName) == "" {
		return errors.New("full name is required")
	}

	matched, _ := regexp.MatchString(`^[\w\.-]+@[\w\.-]+\.\w+$`, r.Email)
	if !matched {
		return errors.New("invalid email address format")
	}

	if len(r.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	if !isStrongPassword(r.Password) {
		return errors.New("password must contain uppercase, lowercase, and a digit")
	}
	if matched, _ := regexp.MatchString(`^[\w\.-]+@[\w\.-]+\.\w+$`, r.Email); !matched {
		return errors.New("invalid email address format")
	}
	// Validate ENUM role
	validRoles := map[string]bool{
		"Incident Responder": true, "Forensic Analyst": true, "Malware Analyst": true,
		"Threat Intelligent Analyst": true, "DFIR Manager": true, "Legal/Compliance Liaison": true,
		"Detection Engineer": true, "Generic user": true,
	}
	if _, ok := validRoles[r.Role]; !ok {
		return errors.New("invalid user role")
	}
	return nil
}

func isStrongPassword(password string) bool {
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	return hasUpper && hasLower && hasDigit
}
