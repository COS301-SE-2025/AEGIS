package unit_tests

import (
	"testing"
	"aegis-api/services/ListCases"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"aegis-api/services/case_creation"
)

// ─────────────────────────────────────────────────────────────
// MOCK
// ─────────────────────────────────────────────────────────────

type MockCaseQueryRepository struct {
	mock.Mock
}

func (m *MockCaseQueryRepository) QueryCases(filter ListCases.CaseFilter) ([]ListCases.Case, error) {
	args := m.Called(filter)
	return args.Get(0).([]ListCases.Case), args.Error(1)
}


func (m *MockCaseQueryRepository) GetAllCases() ([]case_creation.Case, error) {
	args := m.Called()
	return args.Get(0).([]case_creation.Case), args.Error(1)
}

func (m *MockCaseQueryRepository) GetCasesByUser(userID string) ([]case_creation.Case, error) {
	args := m.Called(userID)
	return args.Get(0).([]case_creation.Case), args.Error(1)
}
// ─────────────────────────────────────────────────────────────
// TEST CASES
// ─────────────────────────────────────────────────────────────

func TestGetFilteredCases_ByStatus(t *testing.T) {
	mockRepo := new(MockCaseQueryRepository)
	service := ListCases.NewListCasesService(mockRepo)

	expected := []ListCases.Case{{Title: "Unauthorized Access Detected", Status: "open"}}

	mockRepo.On("QueryCases", mock.MatchedBy(func(f ListCases.CaseFilter) bool {
		return f.Status == "open"
	})).Return(expected, nil)

	results, err := service.GetFilteredCases("open", "", "", "", "created_at", "desc")

	assert.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "open", results[0].Status)
	mockRepo.AssertExpectations(t)
}

func TestGetFilteredCases_ByPriority(t *testing.T) {
	mockRepo := new(MockCaseQueryRepository)
	service := ListCases.NewListCasesService(mockRepo)

	expected := []ListCases.Case{{Title: "Suspicious Network Activity", Priority: "medium"}}

	mockRepo.On("QueryCases", mock.MatchedBy(func(f ListCases.CaseFilter) bool {
		return f.Priority == "medium"
	})).Return(expected, nil)

	results, err := service.GetFilteredCases("", "medium", "", "", "created_at", "desc")

	assert.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "medium", results[0].Priority)
	mockRepo.AssertExpectations(t)
}

func TestGetFilteredCases_InvalidSortAndOrder(t *testing.T) {
	mockRepo := new(MockCaseQueryRepository)
	service := ListCases.NewListCasesService(mockRepo)

	mockRepo.On("QueryCases", mock.MatchedBy(func(f ListCases.CaseFilter) bool {
		return f.SortBy == "invalid_field" && f.SortOrder == "invalid_order"
	})).Return([]ListCases.Case{}, nil)

	results, err := service.GetFilteredCases("", "", "", "", "invalid_field", "invalid_order")

	assert.NoError(t, err)
	assert.NotNil(t, results)
}

func TestGetFilteredCases_NoFilters(t *testing.T) {
	mockRepo := new(MockCaseQueryRepository)
	service := ListCases.NewListCasesService(mockRepo)

	expected := []ListCases.Case{
		{Title: "Case 1"},
		{Title: "Case 2"},
	}

	mockRepo.On("QueryCases", mock.Anything).Return(expected, nil)

	results, err := service.GetFilteredCases("", "", "", "", "", "")

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 2)
}

func TestGetFilteredCases_CombinedFilters(t *testing.T) {
	mockRepo := new(MockCaseQueryRepository)
	service := ListCases.NewListCasesService(mockRepo)

	expected := []ListCases.Case{
		{
			Status:    "open",
			Priority:  "high",
			CreatedBy: uuid.MustParse("ded0a1b3-4712-46b5-8d01-fafbaf3f8236"),
		},
	}

	mockRepo.On("QueryCases", mock.MatchedBy(func(f ListCases.CaseFilter) bool {
		return f.Status == "open" &&
			f.Priority == "high" &&
			f.CreatedBy == "ded0a1b3-4712-46b5-8d01-fafbaf3f8236"
	})).Return(expected, nil)

	results, err := service.GetFilteredCases("open", "high", "ded0a1b3-4712-46b5-8d01-fafbaf3f8236", "", "created_at", "desc")

	assert.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "open", results[0].Status)
	assert.Equal(t, "high", results[0].Priority)
	assert.Equal(t, "ded0a1b3-4712-46b5-8d01-fafbaf3f8236", results[0].CreatedBy.String())
}

func TestGetFilteredCases_TitleSearchMatch(t *testing.T) {
	mockRepo := new(MockCaseQueryRepository)
	service := ListCases.NewListCasesService(mockRepo)

	tests := []struct {
		term     string
		expected int
	}{
		{"unauthorized", 1},
		{"suspicious", 1},
		{"", 2},
		{"notfound", 0},
	}

	for _, tt := range tests {
		mockRepo.On("QueryCases", mock.MatchedBy(func(f ListCases.CaseFilter) bool {
			return f.TitleTerm == tt.term
		})).Return(make([]ListCases.Case, tt.expected), nil)

		results, err := service.GetFilteredCases("", "", "", tt.term, "created_at", "desc")

		assert.NoError(t, err)
		assert.Equal(t, tt.expected, len(results), "Failed for term: %s", tt.term)
	}
}
