package unit_tests

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"aegis-api/services/case_status_update"
)

// --- Mocks ---

type MockCaseStatusRepo struct {
	mock.Mock
}

func (m *MockCaseStatusRepo) UpdateStatus(caseID string, newStatus string) error {
	args := m.Called(caseID, newStatus)
	return args.Error(0)
}

// --- Tests ---

func TestUpdateCaseStatus_Success(t *testing.T) {
	mockRepo := new(MockCaseStatusRepo)
	service := case_status_update.NewCaseStatusService(mockRepo)

	caseID := uuid.New().String()
	req := case_status_update.UpdateCaseStatusRequest{
		CaseID: caseID,
		Status: "closed",
	}

	mockRepo.On("UpdateStatus", req.CaseID, req.Status).Return(nil)

	err := service.UpdateCaseStatus(req, "Admin")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUpdateCaseStatus_Unauthorized(t *testing.T) {
	mockRepo := new(MockCaseStatusRepo)
	service := case_status_update.NewCaseStatusService(mockRepo)

	req := case_status_update.UpdateCaseStatusRequest{
		CaseID: uuid.New().String(),
		Status: "closed",
	}

	err := service.UpdateCaseStatus(req, "Analyst") // Not Admin

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized")
}

func TestUpdateCaseStatus_InvalidUUID(t *testing.T) {
	mockRepo := new(MockCaseStatusRepo)
	service := case_status_update.NewCaseStatusService(mockRepo)

	req := case_status_update.UpdateCaseStatusRequest{
		CaseID: "not-a-uuid",
		Status: "closed",
	}

	err := service.UpdateCaseStatus(req, "Admin")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid case UUID")
}

func TestUpdateCaseStatus_RepoFailure(t *testing.T) {
	mockRepo := new(MockCaseStatusRepo)
	service := case_status_update.NewCaseStatusService(mockRepo)

	caseID := uuid.New().String()
	req := case_status_update.UpdateCaseStatusRequest{
		CaseID: caseID,
		Status: "closed",
	}

	mockRepo.On("UpdateStatus", req.CaseID, req.Status).Return(errors.New("DB failure"))

	err := service.UpdateCaseStatus(req, "Admin")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DB failure")
	mockRepo.AssertExpectations(t)
}
