package login

import "aegis-api/services_/auth/registration"

type UserRepository interface {
	GetUserByEmail(email string) (*registration.User, error)
	GetUserByID(userID string) (*registration.User, error)
	UpdateUserTokenInfo(user *registration.User) error
}
