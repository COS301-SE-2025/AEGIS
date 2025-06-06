package auth

type User struct {
	ID         string
	Email      string
	Password   string // hashed
	IsVerified bool
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	ID         string `json:"id"`
	Token      string `json:"token"`
	Email      string `json:"email"`
	IsVerified bool   `json:"is_verified"`
}
