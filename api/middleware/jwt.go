package middleware

import "github.com/golang-jwt/jwt/v5"

var jwtSecret []byte

type Claims struct {
	UserID string `json:"UserID"`
	Email  string `json:"Email"`
	Role   string `json:"Role"`
	jwt.RegisteredClaims
}

func SetJWTSecret(secret []byte) {
	jwtSecret = secret
}

func GetJWTSecret() []byte {
	return jwtSecret
}
