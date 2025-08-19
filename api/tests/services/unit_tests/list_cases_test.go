package unit_tests

import (
	"aegis-api/services_/case/ListCases"
	"aegis-api/services_/case/case_creation"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Add missing QueryCases method to satisfy ListCases.CaseQueryRepository interface
func (m *MockListCasesQueryRepository) QueryCases(filter ListCases.CaseFilter) ([]ListCases.Case, error) {
	args := m.Called(filter)
	return args.Get(0).([]ListCases.Case), args.Error(1)
}

// MockCaseQueryRepository implements the mock repository for testing
type MockListCasesQueryRepository struct {
	mock.Mock
}

func (m *MockListCasesQueryRepository) GetAllCases(arg string) ([]case_creation.Case, error) {
	args := m.Called(arg)
	return args.Get(0).([]case_creation.Case), args.Error(1)
}

func (m *MockListCasesQueryRepository) GetCasesByUser(userID string, anotherArg string) ([]case_creation.Case, error) {
	args := m.Called(userID)
	return args.Get(0).([]case_creation.Case), args.Error(1)
}

// Add missing method to satisfy ListCases.CaseQueryRepository interface
func (m *MockListCasesQueryRepository) GetCaseByID(caseID string, anotherArg string) (*case_creation.Case, error) {
	args := m.Called(caseID, anotherArg)
	return args.Get(0).(*case_creation.Case), args.Error(1)
}

func TestGetAllCases(t *testing.T) {
	mockRepo := new(MockListCasesQueryRepository)
	service := ListCases.NewListCasesService(mockRepo)

	expectedCases := []case_creation.Case{
		{Title: "Case 1"},
		{Title: "Case 2"},
	}

	mockRepo.On("GetAllCases", "test-user-id").Return(expectedCases, nil)

	cases, err := service.GetAllCases("test-user-id")

	assert.NoError(t, err)
	assert.Equal(t, 2, len(cases))
	mockRepo.AssertExpectations(t)
}

func TestGetCasesByUser(t *testing.T) {
	mockRepo := new(MockListCasesQueryRepository)
	service := ListCases.NewListCasesService(mockRepo)

	userID := "8fb89568-3c52-4535-af33-d2f1266def52"
	expected := []case_creation.Case{
		{Title: "User Case A"},
		{Title: "User Case B"},
	}

	mockRepo.On("GetCasesByUser", userID).Return(expected, nil)

	cases, err := service.GetCasesByUser(userID, "")

	assert.NoError(t, err)
	assert.Equal(t, 2, len(cases))
	mockRepo.AssertExpectations(t)
}

func TestGetCasesByNonexistentUser(t *testing.T) {
	mockRepo := new(MockListCasesQueryRepository)
	service := ListCases.NewListCasesService(mockRepo)

	nonexistentUserID := "00000000-0000-0000-0000-000000000999"
	mockRepo.On("GetCasesByUser", nonexistentUserID).Return([]case_creation.Case{}, nil)

	cases, err := service.GetCasesByUser(nonexistentUserID, "")

	assert.NoError(t, err)
	assert.Equal(t, 0, len(cases))
}
