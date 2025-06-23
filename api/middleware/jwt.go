package middleware

import "github.com/golang-jwt/jwt/v5"

var jwtSecret []byte
var isInitialized bool = false

type Claims struct {
	UserID string `json:"UserID"`
	Email  string `json:"Email"`
	Role   string `json:"Role"`
	jwt.RegisteredClaims
}

func SetJWTSecret(secret []byte) {
	if len(secret) == 0 {
		panic("JWT secret cannot be empty")
	}
	jwtSecret = secret
	isInitialized = true
}

func GetJWTSecret() []byte {
	if !isInitialized {
		panic("JWT secret not initialized. Call SetJWTSecret first.")
	}
	return jwtSecret
}
