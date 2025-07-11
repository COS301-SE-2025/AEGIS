package case_status_update

type UpdateCaseStatusRequest struct {
	CaseID string `json:"case_id" validate:"required,uuid"`
	Status string `json:"status" validate:"required"`
}
