package structs

type UpdateProfileRequest struct {
	FullName string `json:"full_name" form:"email,omitempty" binding:"omitempty,min=1"`
	Email    string `json:"email" form:"email,omitempty" binding:"omitempty,email"`
}
