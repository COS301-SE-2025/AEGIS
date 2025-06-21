package structs

type CreateCaseRequest struct {
	Title              string `json:"title" binding:"required"`
	Description        string `json:"description"`
	Status             string `json:"status" binding:"required"`
	Priority           string `json:"priority" binding:"required"`
	InvestigationStage string `json:"investigation_stage"`
	TeamName           string `json:"team_name" binding:"required"`
}

type UpdateCaseStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type AssignCaseRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required"` //might need to remove
}
