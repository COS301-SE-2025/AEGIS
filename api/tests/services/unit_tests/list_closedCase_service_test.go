package unit_tests

import (
	"aegis-api/services_/case/ListClosedCases"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockListClosedCasesRepository is a mock implementation of the repository
type MockListClosedCasesRepository struct {
	mock.Mock
}

func (m *MockListClosedCasesRepository) GetClosedCasesByUserID(ctx context.Context, userID string, tenantID string, teamID string) ([]ListClosedCases.ClosedCase, error) {
	args := m.Called(ctx, userID, tenantID, teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]ListClosedCases.ClosedCase), args.Error(1)
}

// TestListClosedCasesService_ListClosedCases tests the service layer
func TestListClosedCasesService_ListClosedCases(t *testing.T) {
	t.Run("successfully returns closed cases", func(t *testing.T) {
		mockRepo := new(MockListClosedCasesRepository)
		service := ListClosedCases.NewService(mockRepo)

		userID := uuid.New().String()
		tenantID := uuid.New().String()
		teamID := uuid.New().String()

		expectedCases := []ListClosedCases.ClosedCase{
			{
				ID:                 uuid.New(),
				TenantID:           uuid.MustParse(tenantID),
				TeamID:             uuid.MustParse(teamID),
				Title:              "Closed Case 1",
				Status:             "closed",
				InvestigationStage: "Case Closure & Review",
				CreatedBy:          uuid.MustParse(userID),
				Priority:           "high",
				CreatedAt:          time.Now(),
			},
			{
				ID:                 uuid.New(),
				TenantID:           uuid.MustParse(tenantID),
				TeamID:             uuid.MustParse(teamID),
				Title:              "Closed Case 2",
				Status:             "closed",
				InvestigationStage: "Case Closure & Review",
				CreatedBy:          uuid.MustParse(userID),
				Priority:           "medium",
				CreatedAt:          time.Now(),
			},
		}

		mockRepo.On("GetClosedCasesByUserID", mock.Anything, userID, tenantID, teamID).
			Return(expectedCases, nil).
			Once()

		cases, err := service.ListClosedCases(userID, tenantID, teamID)

		assert.NoError(t, err)
		assert.NotNil(t, cases)
		assert.Equal(t, 2, len(cases))
		assert.Equal(t, "Closed Case 1", cases[0].Title)
		assert.Equal(t, "Closed Case 2", cases[1].Title)
		assert.Equal(t, "closed", cases[0].Status)
		assert.Equal(t, "closed", cases[1].Status)
		mockRepo.AssertExpectations(t)
	})

	t.Run("returns empty slice when no closed cases exist", func(t *testing.T) {
		mockRepo := new(MockListClosedCasesRepository)
		service := ListClosedCases.NewService(mockRepo)

		userID := uuid.New().String()
		tenantID := uuid.New().String()
		teamID := uuid.New().String()

		mockRepo.On("GetClosedCasesByUserID", mock.Anything, userID, tenantID, teamID).
			Return([]ListClosedCases.ClosedCase{}, nil).
			Once()

		cases, err := service.ListClosedCases(userID, tenantID, teamID)

		assert.NoError(t, err)
		assert.NotNil(t, cases)
		assert.Empty(t, cases)
		mockRepo.AssertExpectations(t)
	})

	t.Run("propagates repository error", func(t *testing.T) {
		mockRepo := new(MockListClosedCasesRepository)
		service := ListClosedCases.NewService(mockRepo)

		userID := uuid.New().String()
		tenantID := uuid.New().String()
		teamID := uuid.New().String()

		expectedError := errors.New("database connection failed")

		mockRepo.On("GetClosedCasesByUserID", mock.Anything, userID, tenantID, teamID).
			Return(nil, expectedError).
			Once()

		cases, err := service.ListClosedCases(userID, tenantID, teamID)

		assert.Error(t, err)
		assert.Nil(t, cases)
		assert.Equal(t, expectedError, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("passes correct parameters to repository", func(t *testing.T) {
		mockRepo := new(MockListClosedCasesRepository)
		service := ListClosedCases.NewService(mockRepo)

		userID := uuid.New().String()
		tenantID := uuid.New().String()
		teamID := uuid.New().String()

		mockRepo.On("GetClosedCasesByUserID", mock.Anything, userID, tenantID, teamID).
			Return([]ListClosedCases.ClosedCase{}, nil).
			Once()

		service.ListClosedCases(userID, tenantID, teamID)

		// Verify the repository was called with the correct parameters
		mockRepo.AssertCalled(t, "GetClosedCasesByUserID", mock.Anything, userID, tenantID, teamID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("handles repository returning nil with error", func(t *testing.T) {
		mockRepo := new(MockListClosedCasesRepository)
		service := ListClosedCases.NewService(mockRepo)

		userID := uuid.New().String()
		tenantID := uuid.New().String()
		teamID := uuid.New().String()

		expectedError := errors.New("record not found")

		mockRepo.On("GetClosedCasesByUserID", mock.Anything, userID, tenantID, teamID).
			Return(nil, expectedError).
			Once()

		cases, err := service.ListClosedCases(userID, tenantID, teamID)

		assert.Error(t, err)
		assert.Nil(t, cases)
		assert.Equal(t, expectedError, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("returns cases with different priorities", func(t *testing.T) {
		mockRepo := new(MockListClosedCasesRepository)
		service := ListClosedCases.NewService(mockRepo)

		userID := uuid.New().String()
		tenantID := uuid.New().String()
		teamID := uuid.New().String()

		expectedCases := []ListClosedCases.ClosedCase{
			{
				ID:                 uuid.New(),
				TenantID:           uuid.MustParse(tenantID),
				TeamID:             uuid.MustParse(teamID),
				Title:              "High Priority Closed",
				Status:             "closed",
				InvestigationStage: "Case Closure & Review",
				Priority:           "high",
				CreatedAt:          time.Now(),
			},
			{
				ID:                 uuid.New(),
				TenantID:           uuid.MustParse(tenantID),
				TeamID:             uuid.MustParse(teamID),
				Title:              "Medium Priority Closed",
				Status:             "closed",
				InvestigationStage: "Case Closure & Review",
				Priority:           "medium",
				CreatedAt:          time.Now(),
			},
			{
				ID:                 uuid.New(),
				TenantID:           uuid.MustParse(tenantID),
				TeamID:             uuid.MustParse(teamID),
				Title:              "Low Priority Closed",
				Status:             "closed",
				InvestigationStage: "Case Closure & Review",
				Priority:           "low",
				CreatedAt:          time.Now(),
			},
		}

		mockRepo.On("GetClosedCasesByUserID", mock.Anything, userID, tenantID, teamID).
			Return(expectedCases, nil).
			Once()

		cases, err := service.ListClosedCases(userID, tenantID, teamID)

		assert.NoError(t, err)
		assert.Equal(t, 3, len(cases))
		assert.Equal(t, "high", cases[0].Priority)
		assert.Equal(t, "medium", cases[1].Priority)
		assert.Equal(t, "low", cases[2].Priority)
		mockRepo.AssertExpectations(t)
	})

	t.Run("handles large number of closed cases", func(t *testing.T) {
		mockRepo := new(MockListClosedCasesRepository)
		service := ListClosedCases.NewService(mockRepo)

		userID := uuid.New().String()
		tenantID := uuid.New().String()
		teamID := uuid.New().String()

		// Create 100 closed cases
		expectedCases := make([]ListClosedCases.ClosedCase, 100)
		for i := 0; i < 100; i++ {
			expectedCases[i] = ListClosedCases.ClosedCase{
				ID:                 uuid.New(),
				TenantID:           uuid.MustParse(tenantID),
				TeamID:             uuid.MustParse(teamID),
				Title:              "Closed Case",
				Status:             "closed",
				InvestigationStage: "Case Closure & Review",
				Priority:           "medium",
				CreatedAt:          time.Now(),
			}
		}

		mockRepo.On("GetClosedCasesByUserID", mock.Anything, userID, tenantID, teamID).
			Return(expectedCases, nil).
			Once()

		cases, err := service.ListClosedCases(userID, tenantID, teamID)

		assert.NoError(t, err)
		assert.Equal(t, 100, len(cases))
		mockRepo.AssertExpectations(t)
	})

	t.Run("service passes nil context to repository", func(t *testing.T) {
		mockRepo := new(MockListClosedCasesRepository)
		service := ListClosedCases.NewService(mockRepo)

		userID := uuid.New().String()
		tenantID := uuid.New().String()
		teamID := uuid.New().String()

		// Use mock.MatchedBy to verify nil context is passed
		mockRepo.On("GetClosedCasesByUserID", 
			mock.MatchedBy(func(ctx context.Context) bool {
				return ctx == nil
			}), 
			userID, tenantID, teamID).
			Return([]ListClosedCases.ClosedCase{}, nil).
			Once()

		service.ListClosedCases(userID, tenantID, teamID)

		mockRepo.AssertExpectations(t)
	})
}