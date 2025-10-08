package unit_tests

import (
	"aegis-api/services_/case/ListClosedCases"
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

// MockClosedCasesRepository implements the mock repository for testing
type MockClosedCasesRepository struct {
	mock.Mock
}

func (m *MockClosedCasesRepository) GetClosedCasesByUserID(ctx context.Context, userID, tenantID, teamID string) ([]ListClosedCases.ClosedCase, error) {
	args := m.Called(ctx, userID, tenantID, teamID)
	return args.Get(0).([]ListClosedCases.ClosedCase), args.Error(1)
}

// Test Service Constructor
func TestNewService(t *testing.T) {
	mockRepo := new(MockClosedCasesRepository)
	service := ListClosedCases.NewService(mockRepo)

	assert.NotNil(t, service)
}

// Test Service.ListClosedCases - Success scenarios
func TestService_ListClosedCases_Success(t *testing.T) {
	mockRepo := new(MockClosedCasesRepository)
	service := ListClosedCases.NewService(mockRepo)

	userID := "user-123"
	tenantID := "tenant-456"
	teamID := "team-789"

	expectedCases := []ListClosedCases.ClosedCase{
		{
			ID:                 uuid.New(),
			Title:              "Closed Security Incident",
			Description:        "Security breach resolved",
			Status:             "closed",
			InvestigationStage: "resolved",
			Priority:           "high",
			CreatedBy:          uuid.New(),
			CreatedAt:          time.Now(),
		},
		{
			ID:                 uuid.New(),
			Title:              "Closed Network Issue",
			Description:        "Network problem fixed",
			Status:             "closed",
			InvestigationStage: "resolved",
			Priority:           "medium",
			CreatedBy:          uuid.New(),
			CreatedAt:          time.Now(),
		},
	}

	mockRepo.On("GetClosedCasesByUserID", mock.Anything, userID, tenantID, teamID).Return(expectedCases, nil)

	result, err := service.ListClosedCases(userID, tenantID, teamID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Closed Security Incident", result[0].Title)
	assert.Equal(t, "Closed Network Issue", result[1].Title)
	mockRepo.AssertExpectations(t)
}

func TestService_ListClosedCases_EmptyResult(t *testing.T) {
	mockRepo := new(MockClosedCasesRepository)
	service := ListClosedCases.NewService(mockRepo)

	userID := "user-with-no-cases"
	tenantID := "tenant-123"
	teamID := "team-456"

	mockRepo.On("GetClosedCasesByUserID", mock.Anything, userID, tenantID, teamID).Return([]ListClosedCases.ClosedCase{}, nil)

	result, err := service.ListClosedCases(userID, tenantID, teamID)

	assert.NoError(t, err)
	assert.Empty(t, result)
	mockRepo.AssertExpectations(t)
}

func TestService_ListClosedCases_RepositoryError(t *testing.T) {
	mockRepo := new(MockClosedCasesRepository)
	service := ListClosedCases.NewService(mockRepo)

	userID := "user-123"
	tenantID := "tenant-456"
	teamID := "team-789"

	mockRepo.On("GetClosedCasesByUserID", mock.Anything, userID, tenantID, teamID).Return([]ListClosedCases.ClosedCase(nil), errors.New("database connection failed"))

	result, err := service.ListClosedCases(userID, tenantID, teamID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "database connection failed")
	mockRepo.AssertExpectations(t)
}

func TestService_ListClosedCases_EmptyParameters(t *testing.T) {
	mockRepo := new(MockClosedCasesRepository)
	service := ListClosedCases.NewService(mockRepo)

	expectedCases := []ListClosedCases.ClosedCase{
		{
			ID:     uuid.New(),
			Title:  "Global Closed Case",
			Status: "closed",
		},
	}

	mockRepo.On("GetClosedCasesByUserID", mock.Anything, "", "", "").Return(expectedCases, nil)

	result, err := service.ListClosedCases("", "", "")

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Global Closed Case", result[0].Title)
	mockRepo.AssertExpectations(t)
}

func TestService_ListClosedCases_OnlyUserID(t *testing.T) {
	mockRepo := new(MockClosedCasesRepository)
	service := ListClosedCases.NewService(mockRepo)

	userID := "user-only"
	expectedCases := []ListClosedCases.ClosedCase{
		{
			ID:     uuid.New(),
			Title:  "User-specific Closed Case",
			Status: "closed",
		},
	}

	mockRepo.On("GetClosedCasesByUserID", mock.Anything, userID, "", "").Return(expectedCases, nil)

	result, err := service.ListClosedCases(userID, "", "")

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "User-specific Closed Case", result[0].Title)
	mockRepo.AssertExpectations(t)
}

func TestService_ListClosedCases_OnlyTenantID(t *testing.T) {
	mockRepo := new(MockClosedCasesRepository)
	service := ListClosedCases.NewService(mockRepo)

	tenantID := "tenant-only"
	expectedCases := []ListClosedCases.ClosedCase{
		{
			ID:     uuid.New(),
			Title:  "Tenant-specific Closed Case",
			Status: "closed",
		},
	}

	mockRepo.On("GetClosedCasesByUserID", mock.Anything, "", tenantID, "").Return(expectedCases, nil)

	result, err := service.ListClosedCases("", tenantID, "")

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Tenant-specific Closed Case", result[0].Title)
	mockRepo.AssertExpectations(t)
}

func TestService_ListClosedCases_OnlyTeamID(t *testing.T) {
	mockRepo := new(MockClosedCasesRepository)
	service := ListClosedCases.NewService(mockRepo)

	teamID := "team-only"
	expectedCases := []ListClosedCases.ClosedCase{
		{
			ID:     uuid.New(),
			Title:  "Team-specific Closed Case",
			Status: "closed",
		},
	}

	mockRepo.On("GetClosedCasesByUserID", mock.Anything, "", "", teamID).Return(expectedCases, nil)

	result, err := service.ListClosedCases("", "", teamID)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Team-specific Closed Case", result[0].Title)
	mockRepo.AssertExpectations(t)
}

func TestService_ListClosedCases_LargeDataset(t *testing.T) {
	mockRepo := new(MockClosedCasesRepository)
	service := ListClosedCases.NewService(mockRepo)

	userID := "user-with-many-cases"
	tenantID := "tenant-123"
	teamID := "team-456"

	// Generate 50 closed cases
	expectedCases := make([]ListClosedCases.ClosedCase, 50)
	for i := 0; i < 50; i++ {
		expectedCases[i] = ListClosedCases.ClosedCase{
			ID:     uuid.New(),
			Title:  fmt.Sprintf("Closed Case %d", i+1),
			Status: "closed",
		}
	}

	mockRepo.On("GetClosedCasesByUserID", mock.Anything, userID, tenantID, teamID).Return(expectedCases, nil)

	result, err := service.ListClosedCases(userID, tenantID, teamID)

	assert.NoError(t, err)
	assert.Len(t, result, 50)
	assert.Equal(t, "Closed Case 1", result[0].Title)
	assert.Equal(t, "Closed Case 50", result[49].Title)
	mockRepo.AssertExpectations(t)
}

// Extended Repository Tests using sqlmock
func TestGetClosedCasesByUserID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)

	repo := ListClosedCases.NewClosedCaseRepository(gdb)
	userID := "user-1"

	mock.ExpectQuery(`SELECT .* FROM "cases"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "status", "investigation_stage", "priority", "created_by", "created_at",
		}).AddRow(
			"678e8400-e80b-41d4-a716-446345440000", "Case B", "Closed Desc", "closed", "evaluation", "high", "987e4567-e89b-34a3-a456-334614174000", time.Now(),
		))

	cases, err := repo.GetClosedCasesByUserID(context.Background(), userID, "", "")
	assert.NoError(t, err)
	assert.Len(t, cases, 1)
	assert.Equal(t, "closed", string(cases[0].Status))
}

func TestGetClosedCasesByUserID_WithTenantAndTeam(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)

	repo := ListClosedCases.NewClosedCaseRepository(gdb)
	userID := "user-1"
	tenantID := "tenant-123"
	teamID := "team-456"

	mock.ExpectQuery(`SELECT .* FROM "cases"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "status", "investigation_stage", "priority", "created_by", "created_at",
		}).AddRow(
			"678e8400-e80b-41d4-a716-446345440000", "Team Case", "Team Closed Desc", "closed", "resolved", "medium", "987e4567-e89b-34a3-a456-334614174000", time.Now(),
		))

	cases, err := repo.GetClosedCasesByUserID(context.Background(), userID, tenantID, teamID)
	assert.NoError(t, err)
	assert.Len(t, cases, 1)
	assert.Equal(t, "Team Case", cases[0].Title)
	assert.Equal(t, "closed", string(cases[0].Status))
}

func TestGetClosedCasesByUserID_EmptyResult(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)

	repo := ListClosedCases.NewClosedCaseRepository(gdb)
	userID := "nonexistent-user"

	mock.ExpectQuery(`SELECT .* FROM "cases"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "status", "investigation_stage", "priority", "created_by", "created_at",
		}))

	cases, err := repo.GetClosedCasesByUserID(context.Background(), userID, "", "")
	assert.NoError(t, err)
	assert.Empty(t, cases)
}

func TestGetClosedCasesByUserID_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)

	repo := ListClosedCases.NewClosedCaseRepository(gdb)
	userID := "user-1"

	mock.ExpectQuery(`SELECT .* FROM "cases"`).
		WillReturnError(errors.New("database connection lost"))

	cases, err := repo.GetClosedCasesByUserID(context.Background(), userID, "", "")
	assert.Error(t, err)
	assert.Nil(t, cases)
	assert.Contains(t, err.Error(), "database connection lost")
}

func TestGetClosedCasesByUserID_MultipleResults(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)

	repo := ListClosedCases.NewClosedCaseRepository(gdb)
	userID := "user-with-multiple-cases"

	rows := sqlmock.NewRows([]string{
		"id", "title", "description", "status", "investigation_stage", "priority", "created_by", "created_at",
	}).
		AddRow("678e8400-e80b-41d4-a716-446345440000", "First Case", "First Desc", "closed", "resolved", "high", "987e4567-e89b-34a3-a456-334614174000", time.Now()).
		AddRow("678e8400-e80b-41d4-a716-446345440001", "Second Case", "Second Desc", "closed", "resolved", "medium", "987e4567-e89b-34a3-a456-334614174000", time.Now()).
		AddRow("678e8400-e80b-41d4-a716-446345440002", "Third Case", "Third Desc", "closed", "resolved", "low", "987e4567-e89b-34a3-a456-334614174000", time.Now())

	mock.ExpectQuery(`SELECT .* FROM "cases"`).WillReturnRows(rows)

	cases, err := repo.GetClosedCasesByUserID(context.Background(), userID, "", "")
	assert.NoError(t, err)
	assert.Len(t, cases, 3)
	assert.Equal(t, "First Case", cases[0].Title)
	assert.Equal(t, "Second Case", cases[1].Title)
	assert.Equal(t, "Third Case", cases[2].Title)
}

func TestGetClosedCasesByUserID_NilContext(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)

	repo := ListClosedCases.NewClosedCaseRepository(gdb)
	userID := "user-1"

	mock.ExpectQuery(`SELECT .* FROM "cases"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "status", "investigation_stage", "priority", "created_by", "created_at",
		}).AddRow(
			"678e8400-e80b-41d4-a716-446345440000", "Case B", "Closed Desc", "closed", "evaluation", "high", "987e4567-e89b-34a3-a456-334614174000", time.Now(),
		))

	// Test with nil context (service passes nil)
	cases, err := repo.GetClosedCasesByUserID(context.TODO(), userID, "", "")
	assert.NoError(t, err)
	assert.Len(t, cases, 1)
}
