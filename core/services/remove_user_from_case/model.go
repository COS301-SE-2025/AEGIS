package remove_user_from_case

import "github.com/google/uuid"

type RemoveUserRequest struct {
	CaseID uuid.UUID `json:"case_id"`
	UserID uuid.UUID `json:"user_id"`
	AdminID uuid.UUID `json:"admin_id"` // used for authorization
}
