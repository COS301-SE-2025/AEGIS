package login

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func ComparePasswords(hashedPwd, plainPwd string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(plainPwd))
	return err == nil
}

func GetJWTSecret() []byte {
	return jwtSecret
}

// HashPassword hashes a plain-text password using bcrypt.
func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashed), err
}

// GenerateJWT issues a token with claims including version and expiration logic.
func GenerateJWT(userID, email, role string, tokenVersion int, customExpiry *time.Time) (string, error) {
	var exp time.Time

	// External users default to 10 days unless a custom expiry is provided
	if role == "External Collaborator" {
		if customExpiry != nil {
			exp = *customExpiry
		} else {
			exp = time.Now().Add(10 * 24 * time.Hour)
		}
	} else {
		// Default expiry for internal users: 24 hours
		exp = time.Now().Add(24 * time.Hour)
	}

	claims := jwt.MapClaims{
		"sub":           userID,
		"email":         email,
		"role":          role,
		"token_version": tokenVersion,
		"iat":           time.Now().Unix(),
		"exp":           exp.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
