package auth

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	ID         string `json:"id"`
	Token      string `json:"token"`
	Email      string `json:"email"`
	IsVerified bool   `json:"is_verified"`
	Role       string `json:"role,omitempty"`
}

type RegenerateTokenRequest struct {
	ExpiresInDays int `json:"expires_in_days"` // how many days until it expires
}

// type User struct {
// 	ID                  string     `gorm:"primaryKey" json:"id"`
// 	FullName            string     `json:"full_name"`
// 	Email               string     `json:"email"`
// 	PasswordHash        string     `json:"-"` // Do not expose in JSON responses
// 	Role                string     `json:"role"`
// 	TokenVersion        int        `json:"token_version"`
// 	IsVerified          bool       `json:"is_verified"`
// 	ExternalTokenStatus string     `json:"external_token_status"` // "active", "revoked", etc.
// 	ExternalTokenExpiry *time.Time `json:"external_token_expiry"` // nullable
// 	CreatedAt           time.Time  `json:"created_at"`
// 	UpdatedAt           time.Time  `json:"updated_at"`
// }
