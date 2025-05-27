package unit_tests

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/google/uuid"
	"aegis-api/services/case_creation"
)

// MockCaseRepository mocks the repository for case creation.
type MockCaseRepository struct {
	mock.Mock
}
func (m *MockCaseAssignmentRepo) IsAdmin(userID uuid.UUID) (bool, error) {
	args := m.Called(userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockCaseRepository) CreateCase(c *case_creation.Case) error {
	args := m.Called(c)
	return args.Error(0)
}

func TestCreateValidCase(t *testing.T) {
	mockRepo := new(MockCaseRepository)
	service := case_creation.NewCaseService(mockRepo)

	req := case_creation.CreateCaseRequest{
		Title:              "Unauthorized Access Detected",
		Description:        "Anomalous login patterns observed.",
		Status:             "open",
		Priority:           "high",
		InvestigationStage: "analysis",
		CreatedBy:          "8fb89568-3c52-4535-af33-d2f1266def52",
		TeamName:           "SOC Team",
	}

	// Expect the repository to be called with a Case matching req fields and a non-zero ID
	mockRepo.On("CreateCase", mock.MatchedBy(func(c *case_creation.Case) bool {
		return c.Title == req.Title &&
			c.Description == req.Description &&
			c.Status == req.Status &&
			c.Priority == req.Priority &&
			c.InvestigationStage == req.InvestigationStage &&
			c.CreatedBy.String() == req.CreatedBy &&
			c.TeamName == req.TeamName &&
			c.ID.String() != ""
	})).Return(nil)

	newCase, err := service.CreateCase(req)

	assert.NoError(t, err)
	assert.Equal(t, req.Title, newCase.Title)
	assert.NotEmpty(t, newCase.ID.String())
	mockRepo.AssertExpectations(t)
}

func TestCreateCaseMissingTitle(t *testing.T) {
	mockRepo := new(MockCaseRepository)
	service := case_creation.NewCaseService(mockRepo)

	req := case_creation.CreateCaseRequest{
		Title:              "",
		Description:        "Missing title.",
		Status:             "open",
		Priority:           "medium",
		InvestigationStage: "research",
		CreatedBy:          "8fb89568-3c52-4535-af33-d2f1266def52",
		TeamName:           "SOC Team",
	}

	newCase, err := service.CreateCase(req)

	assert.Nil(t, newCase)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "title is required")
	mockRepo.AssertNotCalled(t, "CreateCase", mock.Anything)
}

func TestCreateCaseInvalidUUID(t *testing.T) {
	mockRepo := new(MockCaseRepository)
	service := case_creation.NewCaseService(mockRepo)

	req := case_creation.CreateCaseRequest{
		Title:              "Invalid UUID Test",
		Description:        "Trying to use an invalid UUID",
		Status:             "open",
		Priority:           "low",
		InvestigationStage: "evaluation",
		CreatedBy:          "invalid-uuid",
		TeamName:           "SOC Team",
	}

	newCase, err := service.CreateCase(req)

	assert.Nil(t, newCase)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid CreatedBy UUID")
	mockRepo.AssertNotCalled(t, "CreateCase", mock.Anything)
}

func TestCreateCase_RepositoryError(t *testing.T) {
	mockRepo := new(MockCaseRepository)
	service := case_creation.NewCaseService(mockRepo)

	req := case_creation.CreateCaseRequest{
		Title:              "Repo Error Test",
		Description:        "Simulate DB failure",
		Status:             "open",
		Priority:           "low",
		InvestigationStage: "evaluation",
		CreatedBy:          "8fb89568-3c52-4535-af33-d2f1266def52",
		TeamName:           "SOC Team",
	}

	// Repository returns an error
	mockRepo.On("CreateCase", mock.AnythingOfType("*case_creation.Case")).Return(errors.New("db error"))

	newCase, err := service.CreateCase(req)

	assert.Nil(t, newCase)
	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())
	mockRepo.AssertExpectations(t)
}
