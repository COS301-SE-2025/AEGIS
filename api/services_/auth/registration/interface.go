package registration

import "gorm.io/gorm"

type UserRepository interface {
	CreateUser(user *User) error
	GetUserByEmail(email string) (*User, error)
	UpdateUser(user *User) error
	GetUserByToken(token string) (*User, error)
	GetDB() *gorm.DB // Returns the underlying database connection, if needed
}
