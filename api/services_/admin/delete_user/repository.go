package delete_user

import "github.com/google/uuid"

type UserRepository interface {
	DeleteUserByID(userID uuid.UUID) error
}
