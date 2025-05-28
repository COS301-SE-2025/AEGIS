package unit_tests

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	//"github.com/stretchr/testify/mock"

	"aegis-api/services/update_user_role"
)

// MockUserRepo mocks the UserRepository interface
// type MockUserRepo struct {
// 	mock.Mock
// }

func (m *MockUserRepo) UpdateRoleAndMirror(userID uuid.UUID, role string) error {
	args := m.Called(userID, role)
	return args.Error(0)
}

func TestUpdateUserRole_Success(t *testing.T) {
	repo := new(MockUserRepo)
	svc := update_user_role.NewUserService(repo)

	uid := uuid.New()
	newRole := "DFIR Manager"

	// Expectation: UpdateRoleAndMirror is called once with correct args
	repo.On("UpdateRoleAndMirror", uid, newRole).Return(nil)

	err := svc.UpdateUserRole(uid.String(), newRole)
	assert.NoError(t, err)

	repo.AssertExpectations(t)
}

func TestUpdateUserRole_InvalidUUID(t *testing.T) {
	repo := new(MockUserRepo)
	svc := update_user_role.NewUserService(repo)

	err := svc.UpdateUserRole("not-a-uuid", "Generic user")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid user ID")
}

func TestUpdateUserRole_InvalidRole(t *testing.T) {
	repo := new(MockUserRepo)
	svc := update_user_role.NewUserService(repo)

	uid := uuid.New()

	err := svc.UpdateUserRole(uid.String(), "Nonexistent Role")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid role")
}

func TestUpdateUserRole_RepoError(t *testing.T) {
	repo := new(MockUserRepo)
	svc := update_user_role.NewUserService(repo)

	uid := uuid.New()
	newRole := "Malware Analyst"

	// Simulate repository failure
	repo.On("UpdateRoleAndMirror", uid, newRole).Return(errors.New("db failure"))

	err := svc.UpdateUserRole(uid.String(), newRole)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update user role")

	repo.AssertExpectations(t)
}
