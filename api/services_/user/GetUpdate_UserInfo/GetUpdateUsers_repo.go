package GetUpdate_UserInfo

import(
	"github.com/google/uuid"
)

type UserRepository interface {
	GetUserByID(userID uuid.UUID) (*User, error)
	UpdateUser(userID uuid.UUID, updates map[string]interface{}) error
	GetUserByEmail(email string) (*User, error)
	GetUserRoles(userID uuid.UUID) ([]string, error)
}
