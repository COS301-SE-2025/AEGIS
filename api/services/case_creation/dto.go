package case_creation

type CreateCaseRequest struct {
	Title              string `json:"title" validate:"required"`
	Description        string `json:"description"`
	Status             string `json:"status"` // optional: default is handled by DB
	Priority           string `json:"priority"`
	InvestigationStage string `json:"investigation_stage"`
	CreatedByFullName  string `json:"created_by_full_name" binding:"required"`
	TeamName           string `json:"team_name" validate:"required"`
}
