package unit_tests

import (
	"aegis-api/services/registration"
	"github.com/stretchr/testify/mock"
	"errors"
	"testing"
	
	"aegis-api/services/login/auth"

	"github.com/stretchr/testify/assert"
)

type MockLoginUserRepository struct {
	mock.Mock
}

func (m *MockLoginUserRepository) GetUserByEmail(email string) (*registration.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*registration.User), args.Error(1)
}



func TestLoginSuccess(t *testing.T) {
    mockRepo := new(MockLoginUserRepository)
    service := auth.NewAuthService(mockRepo)

    email := "roy@aegis.dev"
    password := "Fireal@chemist123"

    hashed, _ := registration.HashPassword(password)
    user := &registration.User{
        Email:             email,
        PasswordHash:      hashed,
        VerificationToken: "mock-token-123",
    }

    mockRepo.On("GetUserByEmail", email).Return(user, nil)

    resp, err := service.Login(email, password)

    assert.NoError(t, err)
    assert.Equal(t, email, resp.Email)
    assert.NotEmpty(t, resp.Token)

    mockRepo.AssertExpectations(t)
}


func TestLoginInvalidPassword(t *testing.T) {
	mockRepo := new(MockLoginUserRepository)
	service := auth.NewAuthService(mockRepo)

	email := "roy@aegis.dev"
	password := "wrongpassword"

	// Real stored password hash
	hashed, _ := registration.HashPassword("Fireal@chemist123")
	user := &registration.User{Email: email, PasswordHash: hashed}

	mockRepo.On("GetUserByEmail", email).Return(user, nil)

	_, err := service.Login(email, password)

	assert.Error(t, err)
	assert.Equal(t, "invalid credentials", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestLoginUnknownUser(t *testing.T) {
	mockRepo := new(MockLoginUserRepository)
	service := auth.NewAuthService(mockRepo)

	email := "unknown@aegis.dev"
	password := "doesntmatter"

	mockRepo.On("GetUserByEmail", email).Return(nil, errors.New("not found"))

	_, err := service.Login(email, password)

	assert.Error(t, err)
	assert.Equal(t, "invalid credentials", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestLoginEmptyPassword(t *testing.T) {
	mockRepo := new(MockLoginUserRepository)
	service := auth.NewAuthService(mockRepo)

	email := "roy@aegis.dev"
	password := ""

	hashed, _ := registration.HashPassword("something")
	user := &registration.User{Email: email, PasswordHash: hashed}

	mockRepo.On("GetUserByEmail", email).Return(user, nil)

	_, err := service.Login(email, password)

	assert.Error(t, err)
	assert.Equal(t, "invalid credentials", err.Error())
	mockRepo.AssertExpectations(t)
}
