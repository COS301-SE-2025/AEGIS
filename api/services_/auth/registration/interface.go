package registration

type UserRepository interface {
	CreateUser(user *User) error
	GetUserByEmail(email string) (*User, error)
	UpdateUser(user *User) error
	GetUserByToken(token string) (*User, error)
}
