package unit_tests

import (
	"testing"
	"time"

	"aegis-api/services/reset_password"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ─────────────────────────────────────────────
// Mock implementations
// ─────────────────────────────────────────────

type MockTokenRepo struct{ mock.Mock }
type MockUserRepo struct{ mock.Mock }
type MockEmailSender struct{ mock.Mock }

func (m *MockTokenRepo) CreateToken(userID uuid.UUID, token string, expiresAt time.Time) error {
	args := m.Called(userID, token, expiresAt)
	return args.Error(0)
}

func (m *MockTokenRepo) GetUserIDByToken(token string) (uuid.UUID, time.Time, error) {
	args := m.Called(token)
	return args.Get(0).(uuid.UUID), args.Get(1).(time.Time), args.Error(2)
}

func (m *MockTokenRepo) MarkTokenUsed(token string) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockUserRepo) UpdatePassword(userID uuid.UUID, hashedPassword string) error {
	args := m.Called(userID, hashedPassword)
	return args.Error(0)
}

func (m *MockEmailSender) SendPasswordResetEmail(email string, token string) error {
	args := m.Called(email, token)
	return args.Error(0)
}

// ─────────────────────────────────────────────
// Test cases
// ─────────────────────────────────────────────

func TestRequestPasswordReset_Success(t *testing.T) {
	repo := new(MockTokenRepo)
	users := new(MockUserRepo)
	emailer := new(MockEmailSender)
	service := reset_password.NewPasswordResetService(repo, users, emailer)

	userID := uuid.New()

	repo.On("CreateToken", userID, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(nil)
	emailer.On("SendPasswordResetEmail", "user@example.com", mock.AnythingOfType("string")).Return(nil)

	err := service.RequestPasswordReset(userID, "user@example.com")
	assert.NoError(t, err)

	repo.AssertExpectations(t)
	emailer.AssertExpectations(t)
}

func TestResetPassword_Success(t *testing.T) {
	repo := new(MockTokenRepo)
	users := new(MockUserRepo)
	emailer := new(MockEmailSender)
	service := reset_password.NewPasswordResetService(repo, users, emailer)

	token := "valid-token"
	userID := uuid.New()
	expires := time.Now().Add(10 * time.Minute)

	repo.On("GetUserIDByToken", token).Return(userID, expires, nil)
	users.On("UpdatePassword", userID, mock.AnythingOfType("string")).Return(nil)
	repo.On("MarkTokenUsed", token).Return(nil)

	err := service.ResetPassword(token, "newSecurePassword123")
	assert.NoError(t, err)

	repo.AssertExpectations(t)
	users.AssertExpectations(t)
}
