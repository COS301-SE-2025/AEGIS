package unit_tests

import (
	"aegis-api/services_/case/ListActiveCases"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ========== SERVICE LAYER TESTS ==========

// MockActiveCaseRepository implements the mock repository for testing
type MockActiveCaseRepository struct {
	mock.Mock
}

func (m *MockActiveCaseRepository) GetActiveCasesByUserID(ctx context.Context, userID, tenantID, teamID string) ([]ListActiveCases.ActiveCase, error) {
	args := m.Called(ctx, userID, tenantID, teamID)
	return args.Get(0).([]ListActiveCases.ActiveCase), args.Error(1)
}

// Test Service Constructor
func TestNewActiveCaseService(t *testing.T) {
	mockRepo := new(MockActiveCaseRepository)
	service := ListActiveCases.NewService(mockRepo)

	assert.NotNil(t, service)
}

// Test Service.ListActiveCases - Success scenarios
func TestService_ListActiveCases_Success(t *testing.T) {
	mockRepo := new(MockActiveCaseRepository)
	service := ListActiveCases.NewService(mockRepo)

	userID := "user-123"
	tenantID := "tenant-456"
	teamID := "team-789"

	expectedCases := []ListActiveCases.ActiveCase{
		{
			ID:                 uuid.New(),
			Title:              "Active Security Incident",
			Description:        "Ongoing security breach investigation",
			Status:             "open",
			InvestigationStage: "analysis",
			Priority:           "high",
			CreatedBy:          uuid.New(),
			CreatedAt:          time.Now(),
		},
		{
			ID:                 uuid.New(),
			Title:              "Active Network Issue",
			Description:        "Network performance degradation",
			Status:             "in_progress",
			InvestigationStage: "containment",
			Priority:           "medium",
			CreatedBy:          uuid.New(),
			CreatedAt:          time.Now(),
		},
	}

	mockRepo.On("GetActiveCasesByUserID", mock.Anything, userID, tenantID, teamID).Return(expectedCases, nil)

	result, err := service.ListActiveCases(userID, tenantID, teamID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Active Security Incident", result[0].Title)
	assert.Equal(t, "Active Network Issue", result[1].Title)
	mockRepo.AssertExpectations(t)
}

func TestService_ListActiveCases_EmptyResult(t *testing.T) {
	mockRepo := new(MockActiveCaseRepository)
	service := ListActiveCases.NewService(mockRepo)

	userID := "user-with-no-active-cases"
	tenantID := "tenant-123"
	teamID := "team-456"

	mockRepo.On("GetActiveCasesByUserID", mock.Anything, userID, tenantID, teamID).Return([]ListActiveCases.ActiveCase{}, nil)

	result, err := service.ListActiveCases(userID, tenantID, teamID)

	assert.NoError(t, err)
	assert.Empty(t, result)
	mockRepo.AssertExpectations(t)
}

func TestService_ListActiveCases_RepositoryError(t *testing.T) {
	mockRepo := new(MockActiveCaseRepository)
	service := ListActiveCases.NewService(mockRepo)

	userID := "user-123"
	tenantID := "tenant-456"
	teamID := "team-789"

	mockRepo.On("GetActiveCasesByUserID", mock.Anything, userID, tenantID, teamID).Return([]ListActiveCases.ActiveCase(nil), errors.New("database timeout"))

	result, err := service.ListActiveCases(userID, tenantID, teamID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "database timeout")
	mockRepo.AssertExpectations(t)
}

func TestService_ListActiveCases_EmptyParameters(t *testing.T) {
	mockRepo := new(MockActiveCaseRepository)
	service := ListActiveCases.NewService(mockRepo)

	expectedCases := []ListActiveCases.ActiveCase{
		{
			ID:     uuid.New(),
			Title:  "Global Active Case",
			Status: "open",
		},
	}

	mockRepo.On("GetActiveCasesByUserID", mock.Anything, "", "", "").Return(expectedCases, nil)

	result, err := service.ListActiveCases("", "", "")

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Global Active Case", result[0].Title)
	mockRepo.AssertExpectations(t)
}

func TestService_ListActiveCases_OnlyUserID(t *testing.T) {
	mockRepo := new(MockActiveCaseRepository)
	service := ListActiveCases.NewService(mockRepo)

	userID := "user-only"
	expectedCases := []ListActiveCases.ActiveCase{
		{
			ID:     uuid.New(),
			Title:  "User-specific Active Case",
			Status: "in_progress",
		},
	}

	mockRepo.On("GetActiveCasesByUserID", mock.Anything, userID, "", "").Return(expectedCases, nil)

	result, err := service.ListActiveCases(userID, "", "")

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "User-specific Active Case", result[0].Title)
	mockRepo.AssertExpectations(t)
}

func TestService_ListActiveCases_OnlyTenantID(t *testing.T) {
	mockRepo := new(MockActiveCaseRepository)
	service := ListActiveCases.NewService(mockRepo)

	tenantID := "tenant-only"
	expectedCases := []ListActiveCases.ActiveCase{
		{
			ID:     uuid.New(),
			Title:  "Tenant-specific Active Case",
			Status: "open",
		},
	}

	mockRepo.On("GetActiveCasesByUserID", mock.Anything, "", tenantID, "").Return(expectedCases, nil)

	result, err := service.ListActiveCases("", tenantID, "")

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Tenant-specific Active Case", result[0].Title)
	mockRepo.AssertExpectations(t)
}

func TestService_ListActiveCases_OnlyTeamID(t *testing.T) {
	mockRepo := new(MockActiveCaseRepository)
	service := ListActiveCases.NewService(mockRepo)

	teamID := "team-only"
	expectedCases := []ListActiveCases.ActiveCase{
		{
			ID:     uuid.New(),
			Title:  "Team-specific Active Case",
			Status: "open",
		},
	}

	mockRepo.On("GetActiveCasesByUserID", mock.Anything, "", "", teamID).Return(expectedCases, nil)

	result, err := service.ListActiveCases("", "", teamID)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Team-specific Active Case", result[0].Title)
	mockRepo.AssertExpectations(t)
}

func TestService_ListActiveCases_LargeDataset(t *testing.T) {
	mockRepo := new(MockActiveCaseRepository)
	service := ListActiveCases.NewService(mockRepo)

	userID := "user-with-many-cases"
	tenantID := "tenant-123"
	teamID := "team-456"

	// Generate 100 active cases
	expectedCases := make([]ListActiveCases.ActiveCase, 100)
	for i := 0; i < 100; i++ {
		expectedCases[i] = ListActiveCases.ActiveCase{
			ID:     uuid.New(),
			Title:  fmt.Sprintf("Active Case %d", i+1),
			Status: "open",
		}
	}

	mockRepo.On("GetActiveCasesByUserID", mock.Anything, userID, tenantID, teamID).Return(expectedCases, nil)

	result, err := service.ListActiveCases(userID, tenantID, teamID)

	assert.NoError(t, err)
	assert.Len(t, result, 100)
	assert.Equal(t, "Active Case 1", result[0].Title)
	assert.Equal(t, "Active Case 100", result[99].Title)
	mockRepo.AssertExpectations(t)
}

func TestService_ListActiveCases_DifferentStatuses(t *testing.T) {
	mockRepo := new(MockActiveCaseRepository)
	service := ListActiveCases.NewService(mockRepo)

	userID := "user-123"
	tenantID := "tenant-456"
	teamID := "team-789"

	expectedCases := []ListActiveCases.ActiveCase{
		{
			ID:     uuid.New(),
			Title:  "Open Case",
			Status: "open",
		},
		{
			ID:     uuid.New(),
			Title:  "In Progress Case",
			Status: "in_progress",
		},
		{
			ID:     uuid.New(),
			Title:  "Under Review Case",
			Status: "under_review",
		},
	}

	mockRepo.On("GetActiveCasesByUserID", mock.Anything, userID, tenantID, teamID).Return(expectedCases, nil)

	result, err := service.ListActiveCases(userID, tenantID, teamID)

	assert.NoError(t, err)
	assert.Len(t, result, 3)
	statuses := []string{result[0].Status, result[1].Status, result[2].Status}
	assert.Contains(t, statuses, "open")
	assert.Contains(t, statuses, "in_progress")
	assert.Contains(t, statuses, "under_review")
	mockRepo.AssertExpectations(t)
}

// ========== REPOSITORY LAYER TESTS ==========

func TestNewActiveCaseRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)

	repo := ListActiveCases.NewActiveCaseRepository(gdb)

	assert.NotNil(t, repo)
}

func TestGetActiveCasesByUserID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)

	repo := ListActiveCases.NewActiveCaseRepository(gdb)

	// Use proper UUIDs
	userID := "550e8400-e29b-41d4-a716-446655440000"
	tenantID := "660e8400-e29b-41d4-a716-446655440000"
	teamID := "770e8400-e29b-41d4-a716-446655440000"
	createdByUUID := "880e8400-e29b-41d4-a716-446655440000"

	mock.ExpectQuery(`SELECT DISTINCT cases.* FROM "cases" LEFT JOIN case_user_roles`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "status", "investigation_stage", "priority", "created_by", "created_at",
		}).AddRow(
			"990e8400-e29b-41d4-a716-446655440000", "Case A", "Desc", "open", "analysis", "medium", createdByUUID, time.Now(),
		))

	cases, err := repo.GetActiveCasesByUserID(context.Background(), userID, tenantID, teamID)
	assert.NoError(t, err)
	assert.Len(t, cases, 1)
	assert.Equal(t, "Case A", cases[0].Title)
	assert.Equal(t, "open", cases[0].Status)
}

func TestGetActiveCasesByUserID_EmptyUserID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)

	repo := ListActiveCases.NewActiveCaseRepository(gdb)

	tenantID := "660e8400-e29b-41d4-a716-446655440000"
	teamID := "770e8400-e29b-41d4-a716-446655440000"
	createdByUUID := "880e8400-e29b-41d4-a716-446655440000"

	mock.ExpectQuery(`SELECT DISTINCT cases.* FROM "cases" LEFT JOIN case_user_roles`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "status", "investigation_stage", "priority", "created_by", "created_at",
		}).AddRow(
			"990e8400-e29b-41d4-a716-446655440000", "Public Case", "Public Desc", "open", "initial", "low", createdByUUID, time.Now(),
		))

	cases, err := repo.GetActiveCasesByUserID(context.Background(), "", tenantID, teamID)
	assert.NoError(t, err)
	assert.Len(t, cases, 1)
	assert.Equal(t, "Public Case", cases[0].Title)
}

func TestGetActiveCasesByUserID_MultipleResults(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)

	repo := ListActiveCases.NewActiveCaseRepository(gdb)

	userID := "550e8400-e29b-41d4-a716-446655440000"
	tenantID := "660e8400-e29b-41d4-a716-446655440000"
	teamID := "770e8400-e29b-41d4-a716-446655440000"
	createdByUUID := "880e8400-e29b-41d4-a716-446655440000"

	rows := sqlmock.NewRows([]string{
		"id", "title", "description", "status", "investigation_stage", "priority", "created_by", "created_at",
	}).
		AddRow("990e8400-e29b-41d4-a716-446655440000", "First Active Case", "First Desc", "open", "analysis", "high", createdByUUID, time.Now()).
		AddRow("991e8400-e29b-41d4-a716-446655440000", "Second Active Case", "Second Desc", "in_progress", "containment", "medium", createdByUUID, time.Now()).
		AddRow("992e8400-e29b-41d4-a716-446655440000", "Third Active Case", "Third Desc", "under_review", "evaluation", "low", createdByUUID, time.Now())

	mock.ExpectQuery(`SELECT DISTINCT cases.* FROM "cases" LEFT JOIN case_user_roles`).WillReturnRows(rows)

	cases, err := repo.GetActiveCasesByUserID(context.Background(), userID, tenantID, teamID)
	assert.NoError(t, err)
	assert.Len(t, cases, 3)
	assert.Equal(t, "First Active Case", cases[0].Title)
	assert.Equal(t, "Second Active Case", cases[1].Title)
	assert.Equal(t, "Third Active Case", cases[2].Title)
}

func TestGetActiveCasesByUserID_EmptyResult(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)

	repo := ListActiveCases.NewActiveCaseRepository(gdb)
	userID := "nonexistent-user"
	tenantID := "tenant-123"
	teamID := "team-456"

	mock.ExpectQuery(`SELECT DISTINCT cases.* FROM "cases" LEFT JOIN case_user_roles`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "status", "investigation_stage", "priority", "created_by", "created_at",
		}))

	cases, err := repo.GetActiveCasesByUserID(context.Background(), userID, tenantID, teamID)
	assert.NoError(t, err)
	assert.Empty(t, cases)
}

func TestGetActiveCasesByUserID_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)

	repo := ListActiveCases.NewActiveCaseRepository(gdb)
	userID := "user-1"
	tenantID := "tenant-123"
	teamID := "team-456"

	mock.ExpectQuery(`SELECT DISTINCT cases.* FROM "cases" LEFT JOIN case_user_roles`).
		WillReturnError(errors.New("database connection lost"))

	cases, err := repo.GetActiveCasesByUserID(context.Background(), userID, tenantID, teamID)
	assert.Error(t, err)
	assert.Nil(t, cases)
	assert.Contains(t, err.Error(), "database connection lost")
}

func TestGetActiveCasesByUserID_WithContext(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)

	repo := ListActiveCases.NewActiveCaseRepository(gdb)

	userID := "550e8400-e29b-41d4-a716-446655440000"
	tenantID := "660e8400-e29b-41d4-a716-446655440000"
	teamID := "770e8400-e29b-41d4-a716-446655440000"
	createdByUUID := "880e8400-e29b-41d4-a716-446655440000"

	ctx := context.WithValue(context.Background(), "test", "value")

	mock.ExpectQuery(`SELECT DISTINCT cases.* FROM "cases" LEFT JOIN case_user_roles`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "status", "investigation_stage", "priority", "created_by", "created_at",
		}).AddRow(
			"990e8400-e29b-41d4-a716-446655440000", "Context Case", "Context Desc", "open", "analysis", "medium", createdByUUID, time.Now(),
		))

	cases, err := repo.GetActiveCasesByUserID(ctx, userID, tenantID, teamID)
	assert.NoError(t, err)
	assert.Len(t, cases, 1)
	assert.Equal(t, "Context Case", cases[0].Title)
}

func TestGetActiveCasesByUserID_NilContext(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)

	repo := ListActiveCases.NewActiveCaseRepository(gdb)

	userID := "550e8400-e29b-41d4-a716-446655440000"
	tenantID := "660e8400-e29b-41d4-a716-446655440000"
	teamID := "770e8400-e29b-41d4-a716-446655440000"
	createdByUUID := "880e8400-e29b-41d4-a716-446655440000"

	mock.ExpectQuery(`SELECT DISTINCT cases.* FROM "cases" LEFT JOIN case_user_roles`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "status", "investigation_stage", "priority", "created_by", "created_at",
		}).AddRow(
			"990e8400-e29b-41d4-a716-446655440000", "Nil Context Case", "Nil Context Desc", "open", "analysis", "medium", createdByUUID, time.Now(),
		))

	// Test with context.TODO() (service passes nil which becomes context.TODO())
	cases, err := repo.GetActiveCasesByUserID(context.TODO(), userID, tenantID, teamID)
	assert.NoError(t, err)
	assert.Len(t, cases, 1)
	assert.Equal(t, "Nil Context Case", cases[0].Title)
}

func TestGetActiveCasesByUserID_FilteringLogic(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)

	repo := ListActiveCases.NewActiveCaseRepository(gdb)

	// Use proper UUIDs instead of short strings
	userID := "550e8400-e29b-41d4-a716-446655440000"
	tenantID := "660e8400-e29b-41d4-a716-446655440000"
	teamID := "770e8400-e29b-41d4-a716-446655440000"
	createdByUUID := "880e8400-e29b-41d4-a716-446655440000"

	// FIXED: Update expected query to match the actual GORM-generated SQL
	// The actual implementation uses NOT IN with both "closed" and "archived" statuses
	expectedQuery := `SELECT DISTINCT cases\.\* FROM "cases" LEFT JOIN case_user_roles ON case_user_roles\.case_id = cases\.id WHERE \(case_user_roles\.user_id = \$1 OR cases\.created_by = \$2\) AND cases\.status NOT IN \(\$3, \$4\) AND cases\.tenant_id = \$5 AND cases\.team_id = \$6`

	mock.ExpectQuery(expectedQuery).
		WithArgs(userID, userID, "closed", "archived", tenantID, teamID). // FIXED: Added "archived" parameter
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "status", "investigation_stage", "priority", "created_by", "created_at",
		}).AddRow(
			"990e8400-e29b-41d4-a716-446655440000", "Filtered Case", "Filtered Desc", "open", "analysis", "medium", createdByUUID, time.Now(),
		))

	cases, err := repo.GetActiveCasesByUserID(context.Background(), userID, tenantID, teamID)

	assert.NoError(t, err)
	if assert.Len(t, cases, 1) { // Check length first to avoid panic
		assert.Equal(t, "Filtered Case", cases[0].Title)
		assert.NotEqual(t, "closed", cases[0].Status)   // Ensure closed cases are filtered out
		assert.NotEqual(t, "archived", cases[0].Status) // Ensure archived cases are filtered out
	}
}

// ========== ADDITIONAL COMPREHENSIVE TESTS ==========

// Test the actual filtering behavior with different status combinations

// Test parameter binding and SQL structure
func TestGetActiveCasesByUserID_SQLStructure(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)

	repo := ListActiveCases.NewActiveCaseRepository(gdb)

	userID := "user-123"
	tenantID := "tenant-456"
	teamID := "team-789"

	// Test that the SQL structure includes all required elements
	mock.ExpectQuery(`SELECT DISTINCT cases\.\*`). // SELECT DISTINCT
							WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "status", "investigation_stage", "priority", "created_by", "created_at",
		}))

	_, err = repo.GetActiveCasesByUserID(context.Background(), userID, tenantID, teamID)
	assert.NoError(t, err)
}

// Test with empty parameters to verify query structure
func TestGetActiveCasesByUserID_EmptyParametersQuery(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)

	repo := ListActiveCases.NewActiveCaseRepository(gdb)

	// Test with empty parameters - should still include status filtering
	expectedQuery := `SELECT DISTINCT cases\.\* FROM "cases" LEFT JOIN case_user_roles ON case_user_roles\.case_id = cases\.id WHERE \(case_user_roles\.user_id = \$1 OR cases\.created_by = \$2\) AND cases\.status NOT IN \(\$3, \$4\) AND cases\.tenant_id = \$5 AND cases\.team_id = \$6`

	mock.ExpectQuery(expectedQuery).
		WithArgs("", "", "closed", "archived", "", "").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "status", "investigation_stage", "priority", "created_by", "created_at",
		}))

	cases, err := repo.GetActiveCasesByUserID(context.Background(), "", "", "")
	assert.NoError(t, err)
	assert.Empty(t, cases) // Should return empty result with empty parameters
}

// Test user access logic - cases user has access to via role OR cases they create
// Test query performance with large datasets

func TestGetActiveCasesByUserID_ContextTimeout(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)

	repo := ListActiveCases.NewActiveCaseRepository(gdb)

	userID := "550e8400-e29b-41d4-a716-446655440000"
	tenantID := "660e8400-e29b-41d4-a716-446655440000"
	teamID := "770e8400-e29b-41d4-a716-446655440000"

	// Create a context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait for context to timeout
	time.Sleep(1 * time.Millisecond)

	mock.ExpectQuery(`SELECT DISTINCT cases\.\*`).
		WillReturnError(context.DeadlineExceeded)

	cases, err := repo.GetActiveCasesByUserID(ctx, userID, tenantID, teamID)

	assert.Error(t, err)
	assert.Nil(t, cases)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}
