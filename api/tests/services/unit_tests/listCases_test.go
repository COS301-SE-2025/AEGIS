package unit_tests

import (
	"testing"
	"aegis-api/services_/case/ListCases"
	"aegis-api/services/case_creation"
	"github.com/stretchr/testify/assert"
	//"github.com/stretchr/testify/mock"
)




func TestGetAllCases(t *testing.T) {
	mockRepo := new(MockCaseQueryRepository)
	service := ListCases.NewListCasesService(mockRepo)

	expectedCases := []case_creation.Case{
		{Title: "Case 1"},
		{Title: "Case 2"},
	}

	mockRepo.On("GetAllCases").Return(expectedCases, nil)

	cases, err := service.GetAllCases()

	assert.NoError(t, err)
	assert.Equal(t, 2, len(cases))
	mockRepo.AssertExpectations(t)
}

func TestGetCasesByUser(t *testing.T) {
	mockRepo := new(MockCaseQueryRepository)
	service := ListCases.NewListCasesService(mockRepo)

	userID := "8fb89568-3c52-4535-af33-d2f1266def52"
	expected := []case_creation.Case{
		{Title: "User Case A"},
		{Title: "User Case B"},
	}

	mockRepo.On("GetCasesByUser", userID).Return(expected, nil)

	cases, err := service.GetCasesByUser(userID)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(cases))
	mockRepo.AssertExpectations(t)
}

func TestGetCasesByNonexistentUser(t *testing.T) {
	mockRepo := new(MockCaseQueryRepository)
	service := ListCases.NewListCasesService(mockRepo)

	nonexistentUserID := "00000000-0000-0000-0000-000000000999"
	mockRepo.On("GetCasesByUser", nonexistentUserID).Return([]case_creation.Case{}, nil)

	cases, err := service.GetCasesByUser(nonexistentUserID)

	assert.NoError(t, err)
	assert.Equal(t, 0, len(cases))
}
