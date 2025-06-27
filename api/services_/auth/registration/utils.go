package registration

import (
	"errors"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

// Claims defines the structure for JWT claims.
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// generateID generates a new UUID string
func generateID() uuid.UUID {
	return uuid.New()
}

// HashPassword hashes a plain-text password using bcrypt.
func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashed), err
}

func GenerateJWT(userID, email, role string) (string, error) {
	claims := &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func VerifyJWT(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	return token.Claims.(*Claims), nil
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
