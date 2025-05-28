package GetUpdate_Users

import "aegis-api/models"

type UserRepository interface {
	GetUserByID(userID string) (*models.UserDTO, error)
	GetUserByEmail(email string) (*models.UserDTO, error)
	UpdateUser(userID string, updates map[string]interface{}) error
	GetUserRoles(userID string) ([]string, error)
}
