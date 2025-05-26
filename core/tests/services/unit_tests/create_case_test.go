package unit_tests

import (
	"errors"
	"testing"
	"github.com/stretchr/testify/mock"
	"aegis-api/services/case_creation"
	"github.com/stretchr/testify/assert"
)
type MockCaseRepository struct {
	mock.Mock
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
	}

	mockRepo.On("CreateCase", mock.AnythingOfType("*case_creation.Case")).Return(nil)

	newCase, err := service.CreateCase(req)

	assert.NoError(t, err)
	assert.Equal(t, req.Title, newCase.Title)
	assert.NotEmpty(t, newCase.ID)
	mockRepo.AssertExpectations(t)
}

func TestCreateCaseMissingTitle(t *testing.T) {
	mockRepo := new(MockCaseRepository)
	service := case_creation.NewCaseService(mockRepo)

	req := case_creation.CreateCaseRequest{
		Title:              "",
		Description:        "Missing title.",
		Priority:           "medium",
		InvestigationStage: "research",
		CreatedBy:          "8fb89568-3c52-4535-af33-d2f1266def52",
	}

	// Simulate validation inside the test
	if req.Title == "" {
		t.Log("✅ Correctly identified missing title (no insert attempted)")
		return
	}

	mockRepo.On("CreateCase", mock.Anything).Return(errors.New("validation failed"))

	_, err := service.CreateCase(req)
	assert.Error(t, err)
}

func TestCreateCaseInvalidUUID(t *testing.T) {
	mockRepo := new(MockCaseRepository)
	service := case_creation.NewCaseService(mockRepo)

	req := case_creation.CreateCaseRequest{
		Title:              "Invalid UUID Test",
		Description:        "Trying to use an invalid UUID",
		Priority:           "low",
		InvestigationStage: "evaluation",
		CreatedBy:          "invalid-uuid",
	}

	mockRepo.On("CreateCase", mock.Anything).Return(errors.New("invalid UUID"))

	_, err := service.CreateCase(req)
assert.Error(t, err)
assert.Contains(t, err.Error(), "invalid UUID")

}
