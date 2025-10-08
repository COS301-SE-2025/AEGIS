package unit_tests

import (
	"errors"
	"fmt"
	"testing"

	"aegis-api/services_/admin/get_collaborators"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepo implements the mock repository for testing
type MockCollaboratorsRepo struct {
	mock.Mock
}

func (m *MockCollaboratorsRepo) GetCollaboratorsByCaseID(caseID uuid.UUID) ([]get_collaborators.Collaborator, error) {
	args := m.Called(caseID)
	return args.Get(0).([]get_collaborators.Collaborator), args.Error(1)
}

func TestGetCollaborators_SuccessSingle(t *testing.T) {
	mockRepo := new(MockCollaboratorsRepo)
	service := get_collaborators.NewService(mockRepo)

	caseID := uuid.New()
	expected := []get_collaborators.Collaborator{
		{
			ID:       uuid.New(),
			FullName: "Alice Smith",
			Email:    "alice@example.com",
			Role:     "incident_responder",
		},
	}

	mockRepo.On("GetCollaboratorsByCaseID", caseID).Return(expected, nil)

	result, err := service.GetCollaborators(caseID)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	mockRepo.AssertExpectations(t)
}

func TestGetCollaborators_EmptyResult(t *testing.T) {
	mockRepo := new(MockCollaboratorsRepo)
	service := get_collaborators.NewService(mockRepo)

	caseID := uuid.New()
	mockRepo.On("GetCollaboratorsByCaseID", caseID).Return([]get_collaborators.Collaborator{}, nil)

	result, err := service.GetCollaborators(caseID)
	assert.NoError(t, err)
	assert.Empty(t, result)
	mockRepo.AssertExpectations(t)
}

func TestGetCollaborators_RepoError(t *testing.T) {
	mockRepo := new(MockCollaboratorsRepo)
	service := get_collaborators.NewService(mockRepo)

	caseID := uuid.New()
	mockRepo.On("GetCollaboratorsByCaseID", caseID).Return([]get_collaborators.Collaborator(nil), errors.New("database error"))

	result, err := service.GetCollaborators(caseID)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "database error")
	mockRepo.AssertExpectations(t)
}

func TestGetCollaborators_MultipleResults(t *testing.T) {
	mockRepo := new(MockCollaboratorsRepo)
	service := get_collaborators.NewService(mockRepo)

	caseID := uuid.New()
	expected := []get_collaborators.Collaborator{
		{
			ID:       uuid.New(),
			FullName: "Alice Smith",
			Email:    "alice@example.com",
			Role:     "incident_responder",
		},
		{
			ID:       uuid.New(),
			FullName: "Bob Johnson",
			Email:    "bob@example.com",
			Role:     "forensics_analyst",
		},
	}

	mockRepo.On("GetCollaboratorsByCaseID", caseID).Return(expected, nil)

	result, err := service.GetCollaborators(caseID)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	mockRepo.AssertExpectations(t)
}

func TestGetCollaborators_NilUUID(t *testing.T) {
	mockRepo := new(MockCollaboratorsRepo)
	service := get_collaborators.NewService(mockRepo)

	var nilUUID uuid.UUID
	mockRepo.On("GetCollaboratorsByCaseID", nilUUID).Return([]get_collaborators.Collaborator{}, nil)

	result, err := service.GetCollaborators(nilUUID)
	assert.NoError(t, err)
	assert.Empty(t, result)
	mockRepo.AssertExpectations(t)
}

func TestGetCollaborators_LargeDataset(t *testing.T) {
	mockRepo := new(MockCollaboratorsRepo)
	service := get_collaborators.NewService(mockRepo)

	caseID := uuid.New()
	expected := make([]get_collaborators.Collaborator, 100)
	for i := 0; i < 100; i++ {
		expected[i] = get_collaborators.Collaborator{
			ID:       uuid.New(),
			FullName: fmt.Sprintf("User %d", i+1),
			Email:    fmt.Sprintf("user%d@example.com", i+1),
			Role:     "incident_responder",
		}
	}

	mockRepo.On("GetCollaboratorsByCaseID", caseID).Return(expected, nil)

	result, err := service.GetCollaborators(caseID)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	assert.Len(t, result, 100)
	mockRepo.AssertExpectations(t)
}

func TestGetCollaborators_DifferentRoles(t *testing.T) {
	mockRepo := new(MockCollaboratorsRepo)
	service := get_collaborators.NewService(mockRepo)

	caseID := uuid.New()
	expected := []get_collaborators.Collaborator{
		{
			ID:       uuid.New(),
			FullName: "Admin User",
			Email:    "admin@example.com",
			Role:     "admin",
		},
		{
			ID:       uuid.New(),
			FullName: "Manager User",
			Email:    "manager@example.com",
			Role:     "manager",
		},
		{
			ID:       uuid.New(),
			FullName: "Analyst User",
			Email:    "analyst@example.com",
			Role:     "analyst",
		},
	}

	mockRepo.On("GetCollaboratorsByCaseID", caseID).Return(expected, nil)

	result, err := service.GetCollaborators(caseID)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	assert.Len(t, result, 3)
	mockRepo.AssertExpectations(t)
}

func TestGetCollaborators_SpecialCharactersInNames(t *testing.T) {
	mockRepo := new(MockCollaboratorsRepo)
	service := get_collaborators.NewService(mockRepo)

	caseID := uuid.New()
	expected := []get_collaborators.Collaborator{
		{
			ID:       uuid.New(),
			FullName: "José María García-González",
			Email:    "jose@example.com",
			Role:     "incident_responder",
		},
		{
			ID:       uuid.New(),
			FullName: "李小明",
			Email:    "li@example.com",
			Role:     "forensics_analyst",
		},
	}

	mockRepo.On("GetCollaboratorsByCaseID", caseID).Return(expected, nil)

	result, err := service.GetCollaborators(caseID)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	mockRepo.AssertExpectations(t)
}

func TestGetCollaborators_DuplicateCollaborators(t *testing.T) {
	mockRepo := new(MockCollaboratorsRepo)
	service := get_collaborators.NewService(mockRepo)

	caseID := uuid.New()
	collaboratorID := uuid.New()
	expected := []get_collaborators.Collaborator{
		{
			ID:       collaboratorID,
			FullName: "Alice Smith",
			Email:    "alice@example.com",
			Role:     "incident_responder",
		},
		{
			ID:       collaboratorID,
			FullName: "Alice Smith",
			Email:    "alice@example.com",
			Role:     "incident_responder",
		},
	}

	mockRepo.On("GetCollaboratorsByCaseID", caseID).Return(expected, nil)

	result, err := service.GetCollaborators(caseID)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	assert.Len(t, result, 2)
	mockRepo.AssertExpectations(t)
}

func TestGetCollaborators_NetworkTimeoutError(t *testing.T) {
	mockRepo := new(MockCollaboratorsRepo)
	service := get_collaborators.NewService(mockRepo)

	caseID := uuid.New()
	mockRepo.On("GetCollaboratorsByCaseID", caseID).Return([]get_collaborators.Collaborator(nil), errors.New("network timeout"))

	result, err := service.GetCollaborators(caseID)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "network timeout")
	mockRepo.AssertExpectations(t)
}

func TestGetCollaborators_InvalidDataError(t *testing.T) {
	mockRepo := new(MockCollaboratorsRepo)
	service := get_collaborators.NewService(mockRepo)

	caseID := uuid.New()
	mockRepo.On("GetCollaboratorsByCaseID", caseID).Return([]get_collaborators.Collaborator(nil), errors.New("invalid data format"))

	result, err := service.GetCollaborators(caseID)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "invalid data format")
	mockRepo.AssertExpectations(t)
}

func TestGetCollaborators_PartiallyPopulatedCollaborator(t *testing.T) {
	mockRepo := new(MockCollaboratorsRepo)
	service := get_collaborators.NewService(mockRepo)

	caseID := uuid.New()
	expected := []get_collaborators.Collaborator{
		{
			ID:       uuid.New(),
			FullName: "",
			Email:    "noemail@example.com",
			Role:     "incident_responder",
		},
		{
			ID:       uuid.New(),
			FullName: "No Email User",
			Email:    "",
			Role:     "forensics_analyst",
		},
	}

	mockRepo.On("GetCollaboratorsByCaseID", caseID).Return(expected, nil)

	result, err := service.GetCollaborators(caseID)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	assert.Len(t, result, 2)
	mockRepo.AssertExpectations(t)
}
