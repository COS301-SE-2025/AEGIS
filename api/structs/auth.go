package structs

type LoginRequest struct {
	Email    string `json:"email" form:"email" binding:"required,email"`
	Password string `json:"password" form:"password" binding:"required"`
}

type ResetPasswordRequest struct {
	Email string `json:"email" form:"email" binding:"required,email"`
}

type PasswordResetBody struct {
	Token       string `json:"token" form:"token" binding:"required"`
	NewPassword string `json:"new_password" form:"new_password" binding:"required,min=8"`
}
