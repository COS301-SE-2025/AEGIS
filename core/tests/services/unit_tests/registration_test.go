package unit_tests

import (
   
    "testing"

    "aegis-api/services/registration"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "gorm.io/gorm"
)

type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) CreateUser(user *registration.User) error {
    args := m.Called(user)
    return args.Error(0)
}

func (m *MockUserRepository) GetUserByEmail(email string) (*registration.User, error) {
    args := m.Called(email)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*registration.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByToken(token string) (*registration.User, error) {
    args := m.Called(token)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*registration.User), args.Error(1)
}

func (m *MockUserRepository) UpdateUser(user *registration.User) error {
    args := m.Called(user)
    return args.Error(0)
}


func TestRegisterValidUser(t *testing.T) {
    mockRepo := new(MockUserRepository)
    service := registration.NewRegistrationService(mockRepo)

    req := registration.RegistrationRequest{
        FullName: "Test User",
        Email:    "testuser@aegis.dev",
        Password: "Secure123",
        Role:     "Generic user",
    }

    err := req.Validate()
    assert.NoError(t, err)

    mockRepo.On("GetUserByEmail", req.Email).Return(nil, gorm.ErrRecordNotFound)
    mockRepo.On("CreateUser", mock.AnythingOfType("*registration.User")).Return(nil)

    user, err := service.Register(req)

    assert.NoError(t, err)
    assert.Equal(t, req.Email, user.Email)

    mockRepo.AssertExpectations(t)
}


func TestRegisterDuplicateUser(t *testing.T) {
    mockRepo := new(MockUserRepository)
    service := registration.NewRegistrationService(mockRepo)

    req := registration.RegistrationRequest{
        FullName: "Roy Mustang",
        Email:    "roy@aegis.dev",
        Password: "Fireal@chemist123",
        Role:     "DFIR Manager",
    }

    existing := &registration.User{Email: req.Email}
    mockRepo.On("GetUserByEmail", req.Email).Return(existing, nil)

    _, err := service.Register(req)

    assert.Error(t, err)
    assert.Equal(t, "user already exists", err.Error())

    mockRepo.AssertNotCalled(t, "CreateUser", mock.Anything)
    mockRepo.AssertExpectations(t)
}


func TestRegisterWeakPassword(t *testing.T) {
    req := registration.RegistrationRequest{
        FullName: "Weak Pass",
        Email:    "weakpass@aegis.dev",
        Password: "abc",
        Role:     "Incident Responder",
    }

    err := req.Validate()
    assert.Error(t, err)
}

func TestRegisterInvalidRole(t *testing.T) {
    req := registration.RegistrationRequest{
        FullName: "Jane Unknown",
        Email:    "jane@aegis.dev",
        Password: "Valid123",
        Role:     "God",
    }

    err := req.Validate()
    assert.Error(t, err)
}
