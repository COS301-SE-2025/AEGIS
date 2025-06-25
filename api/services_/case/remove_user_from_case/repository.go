package remove_user_from_case

import "github.com/google/uuid"

type Repository interface {
	IsAdmin(userID uuid.UUID) (bool, error)
	RemoveUserFromCase(userID, caseID uuid.UUID) error
}
