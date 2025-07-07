package registration

import "gorm.io/gorm"

type UserRepository interface {
	CreateUser(user *User) error
	GetUserByEmail(email string) (*User, error)
	UpdateUser(user *User) error
	GetUserByFullName(fullName string) (*User, error)
	GetUserByID(userID string) (*User, error)
	UpdateUserTokenInfo(user *User) error
	GetDB() *gorm.DB // Returns the underlying database connection, if needed
	FindAll() ([]User, error)
}
