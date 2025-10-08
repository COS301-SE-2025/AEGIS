package unit_tests

import (
	"aegis-api/services_/case/ListCases"
	"aegis-api/services_/case/case_creation"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockListCasesQueryRepository implements the mock repository for testing
type MockListCasesQueryRepository struct {
	mock.Mock
}

func (m *MockListCasesQueryRepository) GetAllCases(tenantID string) ([]case_creation.Case, error) {
	args := m.Called(tenantID)
	return args.Get(0).([]case_creation.Case), args.Error(1)
}

func (m *MockListCasesQueryRepository) GetCasesByUser(userID string, tenantID string) ([]case_creation.Case, error) {
	args := m.Called(userID, tenantID)
	return args.Get(0).([]case_creation.Case), args.Error(1)
}

func (m *MockListCasesQueryRepository) QueryCases(filter ListCases.CaseFilter) ([]ListCases.Case, error) {
	args := m.Called(filter)
	return args.Get(0).([]ListCases.Case), args.Error(1)
}

func (m *MockListCasesQueryRepository) GetCaseByID(caseID string, tenantID string) (*case_creation.Case, error) {
	args := m.Called(caseID, tenantID)
	return args.Get(0).(*case_creation.Case), args.Error(1)
}

// Helper function to create test cases with proper progress calculation
func createTestCase(id, tenantID uuid.UUID, title, status, investigationStage string, createdBy ...uuid.UUID) case_creation.Case {
	caseItem := case_creation.Case{
		ID:                 id,
		TenantID:           tenantID,
		Title:              title,
		Status:             status,
		InvestigationStage: investigationStage,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// Set CreatedBy if provided
	if len(createdBy) > 0 {
		caseItem.CreatedBy = createdBy[0]
	}

	return caseItem
}

// Helper function to calculate expected progress based on investigation stage
// Using the same stage names as your GetProgressForStage function
func getExpectedProgress(investigationStage string) int {
	switch investigationStage {
	case "Triage":
		return 10
	case "Evidence Collection":
		return 25
	case "Analysis":
		return 40
	case "Correlation & Threat Intelligence":
		return 55
	case "Containment & Eradication":
		return 70
	case "Recovery":
		return 85
	case "Reporting & Documentation":
		return 95
	case "Case Closure & Review":
		return 100
	default:
		return 0
	}
}

// TestListActiveCases tests the ListActiveCases method
func TestListActiveCases(t *testing.T) {
	// Test 1: Normal case with active cases
	t.Run("returns active cases only", func(t *testing.T) {
		mockRepo := new(MockListCasesQueryRepository)
		service := ListCases.NewListCasesService(mockRepo)

		tenantID := uuid.New()
		tenantIDStr := tenantID.String()
		caseID1 := uuid.New()
		caseID2 := uuid.New()
		caseID3 := uuid.New()

		expectedCases := []case_creation.Case{
			createTestCase(caseID1, tenantID, "Case 1", "active", "Analysis"),
			createTestCase(caseID2, tenantID, "Case 2", "closed", "Case Closure & Review"),
			createTestCase(caseID3, tenantID, "Case 3", "active", "Recovery"),
		}

		mockRepo.On("GetAllCases", tenantIDStr).Return(expectedCases, nil).Once()

		cases, err := service.ListActiveCases(tenantIDStr)

		assert.NoError(t, err)
		assert.Equal(t, 2, len(cases), "Expected 2 active cases")
		assert.Equal(t, caseID1, cases[0].ID)
		assert.Equal(t, caseID3, cases[1].ID)

		// Check progress calculation
		expectedProgress1 := getExpectedProgress("Analysis")
		expectedProgress3 := getExpectedProgress("Recovery")
		assert.Equal(t, expectedProgress1, cases[0].Progress, "Expected correct Progress for case-1")
		assert.Equal(t, expectedProgress3, cases[1].Progress, "Expected correct Progress for case-3")
		mockRepo.AssertExpectations(t)
	})

	// Test 2: Empty results from repository
	t.Run("handles empty repository results", func(t *testing.T) {
		mockRepo := new(MockListCasesQueryRepository)
		service := ListCases.NewListCasesService(mockRepo)

		tenantIDStr := uuid.New().String()
		mockRepo.On("GetAllCases", tenantIDStr).Return([]case_creation.Case{}, nil).Once()

		cases, err := service.ListActiveCases(tenantIDStr)
		assert.NoError(t, err, "Service should handle empty repository results gracefully")
		assert.Empty(t, cases, "Expected empty cases when repository returns empty")
		mockRepo.AssertExpectations(t)
	})

	// Test 3: No active cases (only closed)
	t.Run("returns empty when no active cases", func(t *testing.T) {
		mockRepo := new(MockListCasesQueryRepository)
		service := ListCases.NewListCasesService(mockRepo)

		tenantIDNoActive := uuid.New()
		tenantIDNoActiveStr := tenantIDNoActive.String()
		mockRepo.On("GetAllCases", tenantIDNoActiveStr).Return([]case_creation.Case{
			createTestCase(uuid.New(), tenantIDNoActive, "Case 4", "closed", "Case Closure & Review"),
		}, nil).Once()

		cases, err := service.ListActiveCases(tenantIDNoActiveStr)
		assert.NoError(t, err)
		assert.Empty(t, cases, "Expected no active cases")
		mockRepo.AssertExpectations(t)
	})
}

// TestGetAllCases tests the GetAllCases method
func TestGetAllCases(t *testing.T) {
	// Test 1: Normal case with multiple cases
	t.Run("returns all cases", func(t *testing.T) {
		mockRepo := new(MockListCasesQueryRepository)
		service := ListCases.NewListCasesService(mockRepo)

		tenantID := uuid.New()
		tenantIDStr := tenantID.String()
		caseID1 := uuid.New()
		caseID2 := uuid.New()

		expectedCases := []case_creation.Case{
			createTestCase(caseID1, tenantID, "Case 1", "active", "Analysis"),
			createTestCase(caseID2, tenantID, "Case 2", "closed", "Case Closure & Review"),
		}

		mockRepo.On("GetAllCases", tenantIDStr).Return(expectedCases, nil).Once()

		cases, err := service.GetAllCases(tenantIDStr)

		assert.NoError(t, err)
		assert.Equal(t, 2, len(cases), "Expected 2 cases")
		assert.Equal(t, caseID1, cases[0].ID)
		assert.Equal(t, caseID2, cases[1].ID)
		assert.Equal(t, tenantID, cases[0].TenantID, "Expected TenantID to be set")

		// Check progress calculation
		expectedProgress1 := getExpectedProgress("Analysis")
		expectedProgress2 := getExpectedProgress("Case Closure & Review")
		assert.Equal(t, expectedProgress1, cases[0].Progress, "Expected Progress to be set for case-1")
		assert.Equal(t, expectedProgress2, cases[1].Progress, "Expected Progress to be set for case-2")
		mockRepo.AssertExpectations(t)
	})

	// Test 2: Empty results from repository
	t.Run("handles empty repository results", func(t *testing.T) {
		mockRepo := new(MockListCasesQueryRepository)
		service := ListCases.NewListCasesService(mockRepo)

		tenantIDStr := uuid.New().String()
		mockRepo.On("GetAllCases", tenantIDStr).Return([]case_creation.Case{}, nil).Once()

		cases, err := service.GetAllCases(tenantIDStr)
		assert.NoError(t, err, "Service should handle empty repository results gracefully")
		assert.Empty(t, cases, "Expected empty cases when repository returns empty")
		mockRepo.AssertExpectations(t)
	})
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

		tenantID := uuid.New()
		tenantIDStr := tenantID.String()
		caseID1 := uuid.New()
		caseID2 := uuid.New()

		expectedCases := []ListCases.Case{
			{
				ID:                 caseID1,
				TenantID:           tenantID,
				Title:              "Filtered Case 1",
				Status:             "active",
				InvestigationStage: "Analysis",
			},
			{
				ID:                 caseID2,
				TenantID:           tenantID,
				Title:              "Filtered Case 2",
				Status:             "active",
				InvestigationStage: "Recovery",
			},
		}

		tenantUUID, _ := uuid.Parse(tenantIDStr)
		filter := ListCases.CaseFilter{TenantID: tenantUUID, Status: "active"}
		mockRepo.On("QueryCases", filter).Return(expectedCases, nil).Once()

		cases, err := service.GetFilteredCases(tenantIDStr, "active", "", "", "", "", "", "", "", "")

		assert.NoError(t, err)
		assert.Equal(t, 2, len(cases), "Expected 2 cases")
		assert.Equal(t, "Filtered Case 1", cases[0].Title)

		// Check progress - it should be set by the service now
		expectedProgress := getExpectedProgress("Analysis")
		assert.Equal(t, expectedProgress, cases[0].Progress, "Expected Progress to be set for case-1")
		mockRepo.AssertExpectations(t)
	})

	// Test 2: Invalid tenantID
	t.Run("returns error for invalid tenantID", func(t *testing.T) {
		mockRepo := new(MockListCasesQueryRepository)
		service := ListCases.NewListCasesService(mockRepo)

		_, err := service.GetFilteredCases("invalid-uuid", "", "", "", "", "", "", "", "", "")
		assert.Error(t, err, "Expected error for invalid tenantID")
	})

	// Test 3: Invalid teamID
	t.Run("returns error for invalid teamID", func(t *testing.T) {
		mockRepo := new(MockListCasesQueryRepository)
		service := ListCases.NewListCasesService(mockRepo)

		tenantIDStr := uuid.New().String()
		_, err := service.GetFilteredCases(tenantIDStr, "", "", "", "", "", "", "", "", "invalid-uuid")
		assert.Error(t, err, "Expected error for invalid teamID")
	})

	// Test 4: Filter by userID and teamID
	t.Run("filters by userID and teamID", func(t *testing.T) {
		mockRepo := new(MockListCasesQueryRepository)
		service := ListCases.NewListCasesService(mockRepo)

		tenantID := uuid.New()
		tenantIDStr := tenantID.String()
		userIDStr := uuid.New().String()
		teamIDStr := uuid.New().String()
		caseID1 := uuid.New()
		caseID2 := uuid.New()

		expectedCases := []ListCases.Case{
			{
				ID:                 caseID1,
				TenantID:           tenantID,
				Title:              "Filtered Case 1",
				Status:             "active",
				InvestigationStage: "Analysis",
			},
			{
				ID:                 caseID2,
				TenantID:           tenantID,
				Title:              "Filtered Case 2",
				Status:             "active",
				InvestigationStage: "Recovery",
			},
		}

		tenantUUID, _ := uuid.Parse(tenantIDStr)
		teamUUID, _ := uuid.Parse(teamIDStr)
		filter := ListCases.CaseFilter{TenantID: tenantUUID, UserID: userIDStr, TeamID: teamUUID}
		mockRepo.On("QueryCases", filter).Return(expectedCases, nil).Once()

		cases, err := service.GetFilteredCases(tenantIDStr, "", "", "", "", "", "", "", userIDStr, teamIDStr)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(cases), "Expected 2 cases")
		mockRepo.AssertExpectations(t)
	})

	// Test 5: Error from repository
	t.Run("propagates repository error", func(t *testing.T) {
		mockRepo := new(MockListCasesQueryRepository)
		service := ListCases.NewListCasesService(mockRepo)

		tenantID := uuid.New()
		tenantIDStr := tenantID.String()
		tenantUUID, _ := uuid.Parse(tenantIDStr)
		filter := ListCases.CaseFilter{TenantID: tenantUUID}
		mockRepo.On("QueryCases", filter).Return([]ListCases.Case{}, errors.New("db error")).Once()

		cases, err := service.GetFilteredCases(tenantIDStr, "", "", "", "", "", "", "", "", "")
		assert.Error(t, err, "Expected error from repository")
		assert.Empty(t, cases, "Expected empty cases on error")
		mockRepo.AssertExpectations(t)
	})

	// Test 6: Empty filter (tenantID only)
	t.Run("handles empty filter results", func(t *testing.T) {
		mockRepo := new(MockListCasesQueryRepository)
		service := ListCases.NewListCasesService(mockRepo)

		tenantID := uuid.New()
		tenantIDStr := tenantID.String()
		tenantUUID, _ := uuid.Parse(tenantIDStr)
		filter := ListCases.CaseFilter{TenantID: tenantUUID}
		mockRepo.On("QueryCases", filter).Return([]ListCases.Case{}, nil).Once()

		cases, err := service.GetFilteredCases(tenantIDStr, "", "", "", "", "", "", "", "", "")
		assert.NoError(t, err)
		assert.Empty(t, cases, "Expected no cases with empty filter")
		mockRepo.AssertExpectations(t)
	})
}

// TestGetCaseByID tests the GetCaseByID method
func TestGetCaseByID(t *testing.T) {
	// Test 1: Normal case
	t.Run("returns case by ID", func(t *testing.T) {
		mockRepo := new(MockListCasesQueryRepository)
		service := ListCases.NewListCasesService(mockRepo)

		caseID := uuid.New()
		caseIDStr := caseID.String()
		tenantID := uuid.New()
		tenantIDStr := tenantID.String()

		expectedCase := &case_creation.Case{
			ID:                 caseID,
			TenantID:           tenantID,
			Title:              "Test Case",
			Status:             "active",
			InvestigationStage: "Analysis",
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}

		mockRepo.On("GetCaseByID", caseIDStr, tenantIDStr).Return(expectedCase, nil).Once()

		result, err := service.GetCaseByID(caseIDStr, tenantIDStr)

		assert.NoError(t, err)
		assert.NotNil(t, result, "Expected non-nil result")
		assert.Equal(t, caseID, result.ID)
		assert.Equal(t, tenantID, result.TenantID, "Expected TenantID to be set")

		// Check progress calculation
		expectedProgress := getExpectedProgress("Analysis")
		assert.Equal(t, expectedProgress, result.Progress, "Expected Progress to be set")
		mockRepo.AssertExpectations(t)
	})

	// Test 2: Non-existent case
	t.Run("returns error for non-existent case", func(t *testing.T) {
		mockRepo := new(MockListCasesQueryRepository)
		service := ListCases.NewListCasesService(mockRepo)

		nonExistentCaseID := uuid.New().String()
		tenantIDStr := uuid.New().String()
		mockRepo.On("GetCaseByID", nonExistentCaseID, tenantIDStr).Return((*case_creation.Case)(nil), errors.New("not found")).Once()

		result, err := service.GetCaseByID(nonExistentCaseID, tenantIDStr)
		assert.Error(t, err, "Expected error for non-existent case")
		assert.Nil(t, result, "Expected nil result")
		mockRepo.AssertExpectations(t)
	})
}
