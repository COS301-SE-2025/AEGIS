package structs

type CreateCaseRequest struct {
	Title              string `json:"title" form:"title" binding:"required"`
	Description        string `json:"description" form:"description"`
	Status             string `json:"status" form:"status" binding:"required"`     //what are the different ones?
	Priority           string `json:"priority" form:"priority" binding:"required"` //same as above
	InvestigationStage string `json:"investigation_stage" form:"investigation_stage"`
	TeamName           string `json:"team_name" form:"team_name" binding:"required"`
}

type UpdateCaseStatusRequest struct {
	Status string `json:"status" form:"status" binding:"required"`
}

type AssignCaseRequest struct {
	UserID string `json:"user_id" form:"user_id" binding:"required"`
	Role   string `json:"role" form:"role" binding:"required"`
}
