package GetUpdate_UserInfo

import (
	"errors"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) GetUserByID(userID uuid.UUID) (*User, error) {
	args := m.Called(userID)
	if user := args.Get(0); user != nil {
		return user.(*User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepo) GetUserByEmail(email string) (*User, error) {
	args := m.Called(email)
	if user := args.Get(0); user != nil {
		return user.(*User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepo) UpdateUser(userID uuid.UUID, updates map[string]interface{}) error {
	args := m.Called(userID, updates)
	return args.Error(0)
}

func (m *MockUserRepo) GetUserRoles(userID uuid.UUID) ([]string, error) {
	args := m.Called(userID)

	roles, ok := args.Get(0).([]string)
	if !ok {
		return nil, errors.New("failed type assertion: expected []string")
	}

	return roles, args.Error(1)
}
