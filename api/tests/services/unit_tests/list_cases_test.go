package unit_tests

import (
	"aegis-api/services_/case/ListCases"
	"aegis-api/services_/case/case_creation"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
	"time"

	"github.com/google/uuid"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCaseQueryRepository implements the mock repository for testing
type MockListCasesQueryRepository struct {
	mock.Mock
}

func (m *MockListCasesQueryRepository) GetAllCases(tenantID string) ([]case_creation.Case, error) {
	args := m.Called(tenantID)
func (m *MockListCasesQueryRepository) GetAllCases(tenantID string) ([]case_creation.Case, error) {
	args := m.Called(tenantID)
	return args.Get(0).([]case_creation.Case), args.Error(1)
}

func (m *MockListCasesQueryRepository) GetCasesByUser(userID string, tenantID string) ([]case_creation.Case, error) {
	args := m.Called(userID, tenantID)
func (m *MockListCasesQueryRepository) GetCasesByUser(userID string, tenantID string) ([]case_creation.Case, error) {
	args := m.Called(userID, tenantID)
	return args.Get(0).([]case_creation.Case), args.Error(1)
}

func (m *MockListCasesQueryRepository) GetCaseByID(caseID string, tenantID string) (*case_creation.Case, error) {
	args := m.Called(caseID, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*case_creation.Case), args.Error(1)
}

func (m *MockListCasesQueryRepository) QueryCases(filter ListCases.CaseFilter) ([]ListCases.Case, error) {
	args := m.Called(filter)
	return args.Get(0).([]ListCases.Case), args.Error(1)
}

// Test NewListCasesService constructor
func TestNewListCasesService(t *testing.T) {
	mockRepo := new(MockListCasesQueryRepository)
	service := ListCases.NewListCasesService(mockRepo)

	assert.NotNil(t, service)
}

func TestGetAllCases(t *testing.T) {
	mockRepo := new(MockListCasesQueryRepository)
	service := ListCases.NewListCasesService(mockRepo)

	expectedCases := []case_creation.Case{
		{
			ID:                 uuid.New(),
			Title:              "Case 1",
			Description:        "Description 1",
			Status:             "open",
			Priority:           "high",
			InvestigationStage: "initial",
			CreatedBy:          uuid.New(),
			TeamName:           "Team A",
			TenantID:           uuid.New(),
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		},
		{
			ID:                 uuid.New(),
			Title:              "Case 2",
			Description:        "Description 2",
			Status:             "closed",
			Priority:           "medium",
			InvestigationStage: "analysis",
			CreatedBy:          uuid.New(),
			TeamName:           "Team B",
			TenantID:           uuid.New(),
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		},
	}

	mockRepo.On("GetAllCases", "test-tenant-id").Return(expectedCases, nil)

	cases, err := service.GetAllCases("test-tenant-id")

	assert.NoError(t, err)
	assert.Equal(t, 2, len(cases))
	assert.Equal(t, "Case 1", cases[0].Title)
	assert.Equal(t, "Case 2", cases[1].Title)
	mockRepo.AssertExpectations(t)
}

func TestGetAllCases_Error(t *testing.T) {
	mockRepo := new(MockListCasesQueryRepository)
	service := ListCases.NewListCasesService(mockRepo)

	mockRepo.On("GetAllCases", "test-tenant-id").Return([]case_creation.Case(nil), errors.New("database error"))

	cases, err := service.GetAllCases("test-tenant-id")

	assert.Error(t, err)
	assert.Nil(t, cases)
	assert.EqualError(t, err, "database error")
	mockRepo.AssertExpectations(t)
}

// TestGetCasesByUser tests the GetCasesByUser method
func TestGetCasesByUser(t *testing.T) {
	// Test 1: Normal case with user cases
	t.Run("returns cases for user", func(t *testing.T) {
		mockRepo := new(MockListCasesQueryRepository)
		service := ListCases.NewListCasesService(mockRepo)

		userID := uuid.New()
		userIDStr := userID.String()
		tenantID := uuid.New()
		tenantIDStr := tenantID.String()
		caseID1 := uuid.New()
		caseID2 := uuid.New()

		expectedCases := []case_creation.Case{
			createTestCase(caseID1, tenantID, "User Case A", "active", "Analysis", userID),
			createTestCase(caseID2, tenantID, "User Case B", "active", "Recovery", userID),
		}

		mockRepo.On("GetCasesByUser", userIDStr, tenantIDStr).Return(expectedCases, nil).Once()

		cases, err := service.GetCasesByUser(userIDStr, tenantIDStr)

		assert.NoError(t, err)
		assert.Equal(t, 2, len(cases), "Expected 2 cases")
		assert.Equal(t, "User Case A", cases[0].Title)
		assert.Equal(t, "User Case B", cases[1].Title)
		assert.Equal(t, tenantID, cases[0].TenantID, "Expected TenantID to be set")

		// Check progress calculation
		expectedProgress1 := getExpectedProgress("Analysis")
		expectedProgress2 := getExpectedProgress("Recovery")
		assert.Equal(t, expectedProgress1, cases[0].Progress, "Expected Progress to be set for case-1")
		assert.Equal(t, expectedProgress2, cases[1].Progress, "Expected Progress to be set for case-2")
		mockRepo.AssertExpectations(t)
	})

	// Test 2: Non-existent user
	t.Run("returns empty for non-existent user", func(t *testing.T) {
		mockRepo := new(MockListCasesQueryRepository)
		service := ListCases.NewListCasesService(mockRepo)

		nonExistentUserID := uuid.New().String()
		tenantIDStr := uuid.New().String()
		mockRepo.On("GetCasesByUser", nonExistentUserID, tenantIDStr).Return([]case_creation.Case{}, nil).Once()

		cases, err := service.GetCasesByUser(nonExistentUserID, tenantIDStr)
		assert.NoError(t, err)
		assert.Empty(t, cases, "Expected no cases for non-existent user")
		mockRepo.AssertExpectations(t)
	})

	// Test 3: Empty results from repository
	t.Run("handles empty repository results", func(t *testing.T) {
		mockRepo := new(MockListCasesQueryRepository)
		service := ListCases.NewListCasesService(mockRepo)

		userIDStr := uuid.New().String()
		tenantIDStr := uuid.New().String()
		mockRepo.On("GetCasesByUser", userIDStr, tenantIDStr).Return([]case_creation.Case{}, nil).Once()

		cases, err := service.GetCasesByUser(userIDStr, tenantIDStr)
		assert.NoError(t, err, "Service should handle empty repository results gracefully")
		assert.Empty(t, cases, "Expected empty cases when repository returns empty")
		mockRepo.AssertExpectations(t)
	})
}

// TestGetFilteredCases tests the GetFilteredCases method
func TestGetFilteredCases(t *testing.T) {
	// Test 1: Filter by status and tenantID
	t.Run("filters by status", func(t *testing.T) {
		mockRepo := new(MockListCasesQueryRepository)
		service := ListCases.NewListCasesService(mockRepo)

	userID := "8fb89568-3c52-4535-af33-d2f1266def52"
	tenantID := "tenant-123"
	expected := []case_creation.Case{
		{
			ID:                 uuid.New(),
			Title:              "User Case A",
			Description:        "User case description A",
			Status:             "active",
			Priority:           "high",
			InvestigationStage: "containment",
			CreatedBy:          uuid.MustParse(userID),
			TeamName:           "Security Team",
			TenantID:           uuid.New(),
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		},
	}

	mockRepo.On("GetCasesByUser", userID, tenantID).Return(expected, nil)

	cases, err := service.GetCasesByUser(userID, tenantID)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(cases))
	assert.Equal(t, "User Case A", cases[0].Title)
	mockRepo.AssertExpectations(t)
}

func TestGetCasesByUser_Error(t *testing.T) {
	mockRepo := new(MockListCasesQueryRepository)
	service := ListCases.NewListCasesService(mockRepo)

	userID := "test-user-id"
	tenantID := "test-tenant-id"
	mockRepo.On("GetCasesByUser", userID, tenantID).Return([]case_creation.Case(nil), errors.New("user not found"))

	cases, err := service.GetCasesByUser(userID, tenantID)

	assert.Error(t, err)
	assert.Nil(t, cases)
	assert.EqualError(t, err, "user not found")
	mockRepo.AssertExpectations(t)
}

// TestGetCaseByID tests the GetCaseByID method
func TestGetCaseByID(t *testing.T) {
	// Test 1: Normal case
	t.Run("returns case by ID", func(t *testing.T) {
		mockRepo := new(MockListCasesQueryRepository)
		service := ListCases.NewListCasesService(mockRepo)

	nonexistentUserID := "00000000-0000-0000-0000-000000000999"
	tenantID := "test-tenant-id"
	mockRepo.On("GetCasesByUser", nonexistentUserID, tenantID).Return([]case_creation.Case{}, nil)

	cases, err := service.GetCasesByUser(nonexistentUserID, tenantID)

	assert.NoError(t, err)
	assert.Equal(t, 0, len(cases))
	mockRepo.AssertExpectations(t)
}

// Test ListActiveCases method
func TestListActiveCases_Success(t *testing.T) {
	mockRepo := new(MockListCasesQueryRepository)
	service := ListCases.NewListCasesService(mockRepo)

	tenantID := "test-tenant-id"
	cases := []case_creation.Case{
		{
			ID:                 uuid.New(),
			Title:              "Active Case 1",
			Status:             "active",
			Priority:           "high",
			InvestigationStage: "containment",
			CreatedBy:          uuid.New(),
			TeamName:           "Team A",
			TenantID:           uuid.New(),
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		},
		{
			ID:                 uuid.New(),
			Title:              "Inactive Case",
			Status:             "closed",
			Priority:           "low",
			InvestigationStage: "resolved",
			CreatedBy:          uuid.New(),
			TeamName:           "Team B",
			TenantID:           uuid.New(),
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		},
		{
			ID:                 uuid.New(),
			Title:              "Active Case 2",
			Status:             "active",
			Priority:           "medium",
			InvestigationStage: "analysis",
			CreatedBy:          uuid.New(),
			TeamName:           "Team C",
			TenantID:           uuid.New(),
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		},
	}

	mockRepo.On("GetAllCases", tenantID).Return(cases, nil)

	activeCases, err := service.ListActiveCases(tenantID)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(activeCases))
	assert.Equal(t, "Active Case 1", activeCases[0].Title)
	assert.Equal(t, "Active Case 2", activeCases[1].Title)
	mockRepo.AssertExpectations(t)
}

func TestListActiveCases_Error(t *testing.T) {
	mockRepo := new(MockListCasesQueryRepository)
	service := ListCases.NewListCasesService(mockRepo)

	tenantID := "test-tenant-id"
	mockRepo.On("GetAllCases", tenantID).Return([]case_creation.Case(nil), errors.New("database connection failed"))

	activeCases, err := service.ListActiveCases(tenantID)

	assert.Error(t, err)
	assert.Nil(t, activeCases)
	assert.EqualError(t, err, "database connection failed")
	mockRepo.AssertExpectations(t)
}

func TestListActiveCases_NoActiveCases(t *testing.T) {
	mockRepo := new(MockListCasesQueryRepository)
	service := ListCases.NewListCasesService(mockRepo)

	tenantID := "test-tenant-id"
	cases := []case_creation.Case{
		{
			ID:     uuid.New(),
			Title:  "Closed Case",
			Status: "closed",
		},
		{
			ID:     uuid.New(),
			Title:  "Resolved Case",
			Status: "resolved",
		},
	}

	mockRepo.On("GetAllCases", tenantID).Return(cases, nil)

	activeCases, err := service.ListActiveCases(tenantID)

	assert.NoError(t, err)
	assert.Equal(t, 0, len(activeCases))
	mockRepo.AssertExpectations(t)
}

// Test GetCaseByID method
func TestGetCaseByID_Success(t *testing.T) {
	mockRepo := new(MockListCasesQueryRepository)
	service := ListCases.NewListCasesService(mockRepo)

	caseID := "test-case-id"
	tenantID := "test-tenant-id"
	expectedCase := &case_creation.Case{
		ID:                 uuid.New(),
		Title:              "Test Case",
		Description:        "Test Description",
		Status:             "active",
		Priority:           "high",
		InvestigationStage: "analysis",
		CreatedBy:          uuid.New(),
		TeamName:           "Security Team",
		TenantID:           uuid.New(),
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	mockRepo.On("GetCaseByID", caseID, tenantID).Return(expectedCase, nil)

	result, err := service.GetCaseByID(caseID, tenantID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test Case", result.Title)
	assert.Equal(t, "Test Description", result.Description)
	mockRepo.AssertExpectations(t)
}

func TestGetCaseByID_Error(t *testing.T) {
	mockRepo := new(MockListCasesQueryRepository)
	service := ListCases.NewListCasesService(mockRepo)

	caseID := "nonexistent-case-id"
	tenantID := "test-tenant-id"
	mockRepo.On("GetCaseByID", caseID, tenantID).Return(nil, errors.New("case not found"))

	result, err := service.GetCaseByID(caseID, tenantID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "case not found")
	mockRepo.AssertExpectations(t)
}

// Test GetFilteredCases method
func TestGetFilteredCases_Success(t *testing.T) {
	mockRepo := new(MockListCasesQueryRepository)
	service := ListCases.NewListCasesService(mockRepo)

	tenantID := uuid.New().String()
	userID := uuid.New().String()
	teamID := uuid.New().String()

	expectedCases := []ListCases.Case{
		{
			ID:                 uuid.New(),
			Title:              "Filtered Case 1",
			Status:             "active",
			Priority:           "high",
			InvestigationStage: "containment",
		},
	}

	// Update the expected filter to include UserID and TeamID that will be set by the service
	expectedFilter := ListCases.CaseFilter{
		TenantID:  uuid.MustParse(tenantID),
		Status:    "active",
		Priority:  "high",
		CreatedBy: "user-123",
		TeamName:  "Security",
		TitleTerm: "incident",
		SortBy:    "created_at",
		SortOrder: "desc",
		UserID:    userID,                 // Include the userID
		TeamID:    uuid.MustParse(teamID), // Include the teamID as UUID
	}

	mockRepo.On("QueryCases", expectedFilter).Return(expectedCases, nil)

	result, err := service.GetFilteredCases(tenantID, "active", "high", "user-123", "Security", "incident", "created_at", "desc", userID, teamID)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "Filtered Case 1", result[0].Title)
	mockRepo.AssertExpectations(t)
}

func TestGetFilteredCases_InvalidTenantID(t *testing.T) {
	mockRepo := new(MockListCasesQueryRepository)
	service := ListCases.NewListCasesService(mockRepo)

	invalidTenantID := "invalid-uuid"

	result, err := service.GetFilteredCases(invalidTenantID, "active", "high", "", "", "", "", "", "", "")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid UUID")
	mockRepo.AssertNotCalled(t, "QueryCases")
}

func TestGetFilteredCases_EmptyTenantID(t *testing.T) {
	mockRepo := new(MockListCasesQueryRepository)
	service := ListCases.NewListCasesService(mockRepo)

	expectedCases := []ListCases.Case{
		{ID: uuid.New(), Title: "Case without tenant filter"},
	}

	expectedFilter := ListCases.CaseFilter{
		TenantID:  uuid.UUID{}, // Empty UUID
		Status:    "open",
		Priority:  "",
		CreatedBy: "",
		TeamName:  "",
		TitleTerm: "",
		SortBy:    "",
		SortOrder: "",
	}

	mockRepo.On("QueryCases", expectedFilter).Return(expectedCases, nil)

	result, err := service.GetFilteredCases("", "open", "", "", "", "", "", "", "", "")

	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
	mockRepo.AssertExpectations(t)
}

func TestGetFilteredCases_RepositoryError(t *testing.T) {
	mockRepo := new(MockListCasesQueryRepository)
	service := ListCases.NewListCasesService(mockRepo)

	tenantID := uuid.New().String()
	expectedFilter := ListCases.CaseFilter{
		TenantID: uuid.MustParse(tenantID),
	}

	mockRepo.On("QueryCases", expectedFilter).Return([]ListCases.Case(nil), errors.New("query failed"))

	result, err := service.GetFilteredCases(tenantID, "", "", "", "", "", "", "", "", "")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "query failed")
	mockRepo.AssertExpectations(t)
}
func TestGetFilteredCases_MinimalSuccess(t *testing.T) {
	mockRepo := new(MockListCasesQueryRepository)
	service := ListCases.NewListCasesService(mockRepo)

	tenantID := uuid.New().String()

	expectedCases := []ListCases.Case{
		{
			ID:    uuid.New(),
			Title: "Test Case",
		},
	}

	expectedFilter := ListCases.CaseFilter{
		TenantID: uuid.MustParse(tenantID),
	}

	mockRepo.On("QueryCases", expectedFilter).Return(expectedCases, nil)

	// Call with minimal parameters to isolate the issue
	result, err := service.GetFilteredCases(tenantID, "", "", "", "", "", "", "", "", "")

	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
	mockRepo.AssertExpectations(t)
}

// ========== PROGRESS CALCULATION TESTS ==========

func TestGetProgressForStage_AllStages(t *testing.T) {
	testCases := []struct {
		stage            string
		expectedProgress int
	}{
		{"Triage", 10},
		{"Evidence Collection", 25},
		{"Analysis", 40},
		{"Correlation & Threat Intelligence", 55},
		{"Containment & Eradication", 70},
		{"Recovery", 85},
		{"Reporting & Documentation", 95},
		{"Case Closure & Review", 100},
	}

	for _, tc := range testCases {
		t.Run(tc.stage, func(t *testing.T) {
			progress := ListCases.GetProgressForStage(tc.stage)
			assert.Equal(t, tc.expectedProgress, progress, "Stage '%s' should return progress %d", tc.stage, tc.expectedProgress)
		})
	}
}

func TestGetProgressForStage_InvalidStages(t *testing.T) {
	testCases := []struct {
		name  string
		stage string
	}{
		{"EmptyString", ""},
		{"UnknownStage", "Unknown Stage"},
		{"CaseSensitive", "triage"}, // lowercase should return 0
		{"PartialMatch", "Analysis Phase"},
		{"NullValue", "null"},
		{"SpecialCharacters", "Analysis & Review"},
		{"NumericString", "123"},
		{"WhitespaceOnly", "   "},
		{"TabsAndNewlines", "\t\n"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			progress := ListCases.GetProgressForStage(tc.stage)
			assert.Equal(t, 0, progress, "Invalid stage '%s' should return progress 0", tc.stage)
		})
	}
}

func TestGetProgressForStage_CaseSensitivity(t *testing.T) {
	// Test that the function is case-sensitive
	validStages := []string{"Triage", "Evidence Collection", "Analysis", "Correlation & Threat Intelligence", "Containment & Eradication", "Recovery", "Reporting & Documentation", "Case Closure & Review"}

	for _, stage := range validStages {
		t.Run("Lowercase_"+stage, func(t *testing.T) {
			lowercaseStage := strings.ToLower(stage)
			progress := ListCases.GetProgressForStage(lowercaseStage)
			assert.Equal(t, 0, progress, "Lowercase stage '%s' should return 0", lowercaseStage)
		})

		t.Run("Uppercase_"+stage, func(t *testing.T) {
			uppercaseStage := strings.ToUpper(stage)
			progress := ListCases.GetProgressForStage(uppercaseStage)
			assert.Equal(t, 0, progress, "Uppercase stage '%s' should return 0", uppercaseStage)
		})
	}
}

func TestGetProgressForStage_ProgressOrder(t *testing.T) {
	// Test that progress values are in ascending order
	stages := []string{
		"Triage",
		"Evidence Collection",
		"Analysis",
		"Correlation & Threat Intelligence",
		"Containment & Eradication",
		"Recovery",
		"Reporting & Documentation",
		"Case Closure & Review",
	}

	var previousProgress int
	for i, stage := range stages {
		progress := ListCases.GetProgressForStage(stage)
		if i > 0 {
			assert.Greater(t, progress, previousProgress,
				"Progress for stage '%s' (%d) should be greater than previous stage (%d)",
				stage, progress, previousProgress)
		}
		previousProgress = progress
	}
}

func TestGetProgressForStage_BoundaryValues(t *testing.T) {
	// Test minimum and maximum progress values
	minProgress := ListCases.GetProgressForStage("Triage")
	maxProgress := ListCases.GetProgressForStage("Case Closure & Review")

	assert.Equal(t, 10, minProgress, "Minimum progress should be 10")
	assert.Equal(t, 100, maxProgress, "Maximum progress should be 100")

	// Test that all valid stages return values between 10 and 100
	validStages := []string{
		"Triage", "Evidence Collection", "Analysis",
		"Correlation & Threat Intelligence", "Containment & Eradication",
		"Recovery", "Reporting & Documentation", "Case Closure & Review",
	}

	for _, stage := range validStages {
		progress := ListCases.GetProgressForStage(stage)
		assert.GreaterOrEqual(t, progress, 10, "Progress for '%s' should be >= 10", stage)
		assert.LessOrEqual(t, progress, 100, "Progress for '%s' should be <= 100", stage)
	}
}

// ========== SETPROGRESSFORCASES TESTS ==========

func TestSetProgressForCases_Success(t *testing.T) {
	cases := []ListCases.Case{
		{
			ID:                 uuid.New(),
			Title:              "Triage Case",
			InvestigationStage: "Triage",
		},
		{
			ID:                 uuid.New(),
			Title:              "Analysis Case",
			InvestigationStage: "Analysis",
		},
		{
			ID:                 uuid.New(),
			Title:              "Completed Case",
			InvestigationStage: "Case Closure & Review",
		},
	}

	result := ListCases.SetProgressForCases(cases)

	assert.Len(t, result, 3)
	assert.Equal(t, 10, result[0].Progress, "Triage case should have 10% progress")
	assert.Equal(t, 40, result[1].Progress, "Analysis case should have 40% progress")
	assert.Equal(t, 100, result[2].Progress, "Completed case should have 100% progress")
}

func TestSetProgressForCases_EmptySlice(t *testing.T) {
	cases := []ListCases.Case{}
	result := ListCases.SetProgressForCases(cases)

	assert.Empty(t, result, "Empty slice should remain empty")
}

func TestSetProgressForCases_InvalidStages(t *testing.T) {
	cases := []ListCases.Case{
		{
			ID:                 uuid.New(),
			Title:              "Invalid Stage Case",
			InvestigationStage: "Unknown Stage",
		},
		{
			ID:                 uuid.New(),
			Title:              "Empty Stage Case",
			InvestigationStage: "",
		},
	}

	result := ListCases.SetProgressForCases(cases)

	assert.Len(t, result, 2)
	assert.Equal(t, 0, result[0].Progress, "Unknown stage should have 0% progress")
	assert.Equal(t, 0, result[1].Progress, "Empty stage should have 0% progress")
}

func TestSetProgressForCases_MixedValidInvalid(t *testing.T) {
	cases := []ListCases.Case{
		{
			ID:                 uuid.New(),
			Title:              "Valid Case",
			InvestigationStage: "Recovery",
		},
		{
			ID:                 uuid.New(),
			Title:              "Invalid Case",
			InvestigationStage: "invalid",
		},
		{
			ID:                 uuid.New(),
			Title:              "Another Valid Case",
			InvestigationStage: "Evidence Collection",
		},
	}

	result := ListCases.SetProgressForCases(cases)

	assert.Len(t, result, 3)
	assert.Equal(t, 85, result[0].Progress, "Recovery should have 85% progress")
	assert.Equal(t, 0, result[1].Progress, "Invalid stage should have 0% progress")
	assert.Equal(t, 25, result[2].Progress, "Evidence Collection should have 25% progress")
}

func TestSetProgressForCases_LargeDataset(t *testing.T) {
	// Test performance with large dataset
	cases := make([]ListCases.Case, 1000)
	stages := []string{"Triage", "Analysis", "Recovery", "Case Closure & Review"}

	for i := 0; i < 1000; i++ {
		cases[i] = ListCases.Case{
			ID:                 uuid.New(),
			Title:              fmt.Sprintf("Case %d", i),
			InvestigationStage: stages[i%len(stages)],
		}
	}

	start := time.Now()
	result := ListCases.SetProgressForCases(cases)
	duration := time.Since(start)

	assert.Len(t, result, 1000)
	assert.Less(t, duration, 10*time.Millisecond, "Processing 1000 cases should complete within 10ms")

	// Verify progress values are set correctly
	expectedProgresses := []int{10, 40, 85, 100} // Corresponding to the stages array
	for i, expectedProgress := range expectedProgresses {
		assert.Equal(t, expectedProgress, result[i].Progress,
			"Case %d should have progress %d", i, expectedProgress)
	}
}

func TestSetProgressForCases_OriginalSliceUnmodified(t *testing.T) {
	// Test that the original slice is modified (not a copy)
	originalCases := []ListCases.Case{
		{
			ID:                 uuid.New(),
			Title:              "Test Case",
			InvestigationStage: "Analysis",
			Progress:           0, // Initial progress
		},
	}

	result := ListCases.SetProgressForCases(originalCases)

	// The function should modify the original slice
	assert.Equal(t, 40, originalCases[0].Progress, "Original slice should be modified")
	assert.Equal(t, 40, result[0].Progress, "Returned slice should have same progress")
	assert.Same(t, &originalCases[0], &result[0], "Should be the same slice, not a copy")
}

// ========== INTEGRATION TESTS WITH PROGRESS ==========

func TestGetAllCases_WithProgress(t *testing.T) {
	mockRepo := new(MockListCasesQueryRepository)
	service := ListCases.NewListCasesService(mockRepo)

	expectedCases := []case_creation.Case{
		{
			ID:                 uuid.New(),
			Title:              "Case with Progress",
			Description:        "Test case",
			Status:             "open",
			Priority:           "high",
			InvestigationStage: "Evidence Collection",
			CreatedBy:          uuid.New(),
			TeamName:           "Team A",
			TenantID:           uuid.New(),
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		},
	}

	mockRepo.On("GetAllCases", "test-tenant-id").Return(expectedCases, nil)

	cases, err := service.GetAllCases("test-tenant-id")

	assert.NoError(t, err)
	assert.Equal(t, 1, len(cases))
	assert.Equal(t, "Evidence Collection", cases[0].InvestigationStage)

	// Note: The service doesn't automatically set progress, so this would need to be done separately
	// This test documents the current behavior
	mockRepo.AssertExpectations(t)
}

func TestListActiveCases_ProgressCalculation(t *testing.T) {
	// Test that active cases can have their progress calculated
	mockRepo := new(MockListCasesQueryRepository)
	service := ListCases.NewListCasesService(mockRepo)

	tenantID := "test-tenant-id"
	cases := []case_creation.Case{
		{
			ID:                 uuid.New(),
			Title:              "Active Case",
			Status:             "active",
			Priority:           "high",
			InvestigationStage: "Containment & Eradication",
			CreatedBy:          uuid.New(),
			TeamName:           "Team A",
			TenantID:           uuid.New(),
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		},
	}

	mockRepo.On("GetAllCases", tenantID).Return(cases, nil)

	activeCases, err := service.ListActiveCases(tenantID)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(activeCases))

	// Manually calculate progress to verify the stage mapping
	progress := ListCases.GetProgressForStage(activeCases[0].InvestigationStage)
	assert.Equal(t, 70, progress, "Containment & Eradication should have 70% progress")

	mockRepo.AssertExpectations(t)
}

// ========== EDGE CASE TESTS ==========

func TestGetProgressForStage_UnicodeCharacters(t *testing.T) {
	unicodeStages := []string{
		"Triagé",       // With accent
		"Analysis 分析",  // With Chinese characters
		"Recovery🔄",    // With emoji
		"Triage\u200B", // With zero-width space
	}

	for _, stage := range unicodeStages {
		progress := ListCases.GetProgressForStage(stage)
		assert.Equal(t, 0, progress, "Unicode stage '%s' should return 0", stage)
	}
}

func TestGetProgressForStage_VeryLongStrings(t *testing.T) {
	longStage := strings.Repeat("Very Long Stage Name ", 1000)
	progress := ListCases.GetProgressForStage(longStage)
	assert.Equal(t, 0, progress, "Very long stage name should return 0")
}

func TestSetProgressForCases_NilSlice(t *testing.T) {
	// Test with nil slice (should not panic)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("SetProgressForCases should not panic with nil slice: %v", r)
		}
	}()

	var cases []ListCases.Case = nil
	result := ListCases.SetProgressForCases(cases)
	assert.Nil(t, result, "Nil slice should return nil")
}
