package unit_tests

import (
	"testing"
	"errors"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"aegis-api/services/get_collaborators"
)

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) GetCollaboratorsByCaseID(caseID uuid.UUID) ([]get_collaborators.Collaborator, error) {
	args := m.Called(caseID)
	return args.Get(0).([]get_collaborators.Collaborator), args.Error(1)
}

func TestGetCollaborators_SuccessSingle(t *testing.T) {
	mockRepo := new(MockRepo)
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
	mockRepo := new(MockRepo)
	service := get_collaborators.NewService(mockRepo)

	caseID := uuid.New()
	mockRepo.On("GetCollaboratorsByCaseID", caseID).Return([]get_collaborators.Collaborator{}, nil)

	result, err := service.GetCollaborators(caseID)
	assert.NoError(t, err)
	assert.Empty(t, result)
	mockRepo.AssertExpectations(t)
}

func TestGetCollaborators_RepoError(t *testing.T) {
	mockRepo := new(MockRepo)
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
	mockRepo := new(MockRepo)
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
