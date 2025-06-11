package login

type User struct {
	ID       string
	Email    string
	Password string // hashed
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	ID    string `json:"id"`
	Token string `json:"token"`
	Email string `json:"email"`
}

type AuthService struct {
	repo UserRepository
}
