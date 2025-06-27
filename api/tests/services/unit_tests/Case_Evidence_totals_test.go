package unit_tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	services "aegis-api/services_/case/case_evidence_totals"
)

// --- Mock definition ---

type MockStatsRepository struct {
	mock.Mock
}

func (m *MockStatsRepository) CountCases() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockStatsRepository) CountEvidence() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

// --- Unit Test ---

func TestDashboardService_GetCounts(t *testing.T) {
	mockRepo := new(MockStatsRepository)
	service := services.NewDashboardService(mockRepo)

	// Setup mock expectations
	mockRepo.On("CountCases").Return(int64(5), nil)
	mockRepo.On("CountEvidence").Return(int64(12), nil)

	// Call the service method
	caseCount, evidenceCount, err := service.GetCounts()

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, int64(5), caseCount)
	assert.Equal(t, int64(12), evidenceCount)

	// Verify expectations
	mockRepo.AssertExpectations(t)
}
