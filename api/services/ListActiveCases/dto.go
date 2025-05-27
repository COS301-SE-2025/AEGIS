package ListActiveCases

type RequestDTO struct {
	// UserID is the ID of the user for whom to list active cases.
	UserID string `json:"user_id" validate:"required,uuid"`
}

type ResponseDTO struct {
	// Cases is a list of active cases.
	Cases []ActiveCase `json:"cases"`
}