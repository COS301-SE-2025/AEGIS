package structs

type RegistrationRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	FullName string `json:"full_name" binding:"required"`
	Role     string `json:"role" binding:"required"`
}

// UpdateUserRoleRequest defines the structure for updating a user's role
type UpdateUserRoleRequest struct {
	Role string `json:"role" binding:"required"`
}
