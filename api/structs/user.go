package structs

type UpdateProfileRequest struct {
	FullName string `json:"full_name,omitempty" binding:"omitempty,min=1"`
	Email    string `json:"email,omitempty" binding:"omitempty,email"`
}
