package middleware //jwt.go

import "github.com/golang-jwt/jwt/v5"

var jwtSecret []byte
var isInitialized bool = false

type Claims struct {
	UserID               string `json:"user_id"`
	Email                string `json:"email"`
	Role                 string `json:"role"`
	jwt.RegisteredClaims        // still needed if you want "exp", "iat" etc. validated
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
