package mocks


import (
	"github.com/stretchr/testify/mock"
	"aegis-api/models"
)

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) GetUserByID(userID string) (*models.UserDTO, error) {
	args := m.Called(userID)
	if user := args.Get(0); user != nil {
		return user.(*models.UserDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepo) UpdateUser(userID string, updates map[string]interface{}) error {
	args := m.Called(userID, updates)
	return args.Error(0)
}

func (m *MockUserRepo) GetUserByEmail(email string) (*models.UserDTO, error) {
	args := m.Called(email)
	if user := args.Get(0); user != nil {
		return user.(*models.UserDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepo) GetUserRoles(userID string) ([]string, error) {
	args := m.Called(userID)
	return args.Get(0).([]string), args.Error(1)
}