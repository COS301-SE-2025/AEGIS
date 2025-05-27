package ListClosedCases

type ListClosedCasesRequest struct {
	UserID string `json:"user_id"`
}

type ListClosedCasesResponse struct {
	ClosedCases []ClosedCase `json:"closed_cases"`
}

