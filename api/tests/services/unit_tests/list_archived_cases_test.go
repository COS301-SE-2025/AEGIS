package unit_tests

import (
	"aegis-api/services_/case/listArchiveCases"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ========== MOCKS ==========

// Mock for ArchiveCaseLister interface
type MockArchiveCaseLister struct {
	mock.Mock
}

func (m *MockArchiveCaseLister) ListArchivedCases(userID, tenantID, teamID string) ([]listArchiveCases.ArchivedCase, error) {
	args := m.Called(userID, tenantID, teamID)
	return args.Get(0).([]listArchiveCases.ArchivedCase), args.Error(1)
}

// ========== HELPER FUNCTIONS ==========

func createTestArchivedCase(title, description string) listArchiveCases.ArchivedCase {
	return listArchiveCases.ArchivedCase{
		ID:          uuid.New(),
		Title:       title,
		Description: description,
		Status:      "archived",
		Priority:    "medium",
		CreatedBy:   uuid.New(),
		TenantID:    uuid.New(),
		TeamID:      uuid.New(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		ArchivedAt:  time.Now(),
	}
}

func setupMockDBArchiveCases(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to create gorm DB: %v", err)
	}

	return gormDB, mock
}

// ========== SERVICE LAYER TESTS ==========

func TestNewArchiveCaseService(t *testing.T) {
	mockRepo := &MockArchiveCaseLister{}
	service := listArchiveCases.NewArchiveCaseService(mockRepo)

	assert.NotNil(t, service)
}

func TestArchiveCaseService_ListArchivedCases_Success(t *testing.T) {
	mockRepo := &MockArchiveCaseLister{}
	service := listArchiveCases.NewArchiveCaseService(mockRepo)

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"

	expectedCases := []listArchiveCases.ArchivedCase{
		createTestArchivedCase("Archived Security Incident #1", "First archived case"),
		createTestArchivedCase("Archived Data Breach Investigation", "Second archived case"),
		createTestArchivedCase("Archived Malware Detection", "Third archived case"),
	}

	mockRepo.On("ListArchivedCases", userID, tenantID, teamID).Return(expectedCases, nil)

	result, err := service.ListArchivedCases(userID, tenantID, teamID)

	assert.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, "Archived Security Incident #1", result[0].Title)
	assert.Equal(t, "Archived Data Breach Investigation", result[1].Title)
	assert.Equal(t, "Archived Malware Detection", result[2].Title)

	mockRepo.AssertExpectations(t)
}

func TestArchiveCaseService_ListArchivedCases_EmptyResult(t *testing.T) {
	mockRepo := &MockArchiveCaseLister{}
	service := listArchiveCases.NewArchiveCaseService(mockRepo)

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"

	emptyCases := []listArchiveCases.ArchivedCase{}

	mockRepo.On("ListArchivedCases", userID, tenantID, teamID).Return(emptyCases, nil)

	result, err := service.ListArchivedCases(userID, tenantID, teamID)

	assert.NoError(t, err)
	assert.Len(t, result, 0)

	mockRepo.AssertExpectations(t)
}

func TestArchiveCaseService_ListArchivedCases_RepositoryError(t *testing.T) {
	mockRepo := &MockArchiveCaseLister{}
	service := listArchiveCases.NewArchiveCaseService(mockRepo)

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"

	expectedError := errors.New("database connection failed")

	// Return nil slice instead of empty slice when there's an error
	mockRepo.On("ListArchivedCases", userID, tenantID, teamID).Return([]listArchiveCases.ArchivedCase(nil), expectedError)

	result, err := service.ListArchivedCases(userID, tenantID, teamID)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, result)

	mockRepo.AssertExpectations(t)
}

func TestArchiveCaseService_ListArchivedCases_InvalidParameters(t *testing.T) {
	testCases := []struct {
		name     string
		userID   string
		tenantID string
		teamID   string
	}{
		{
			name:     "Empty UserID",
			userID:   "",
			tenantID: "test-tenant-456",
			teamID:   "test-team-789",
		},
		{
			name:     "Empty TenantID",
			userID:   "test-user-123",
			tenantID: "",
			teamID:   "test-team-789",
		},
		{
			name:     "Empty TeamID",
			userID:   "test-user-123",
			tenantID: "test-tenant-456",
			teamID:   "",
		},
		{
			name:     "All Empty",
			userID:   "",
			tenantID: "",
			teamID:   "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := &MockArchiveCaseLister{}
			service := listArchiveCases.NewArchiveCaseService(mockRepo)

			emptyCases := []listArchiveCases.ArchivedCase{}

			mockRepo.On("ListArchivedCases", tc.userID, tc.tenantID, tc.teamID).Return(emptyCases, nil)

			result, err := service.ListArchivedCases(tc.userID, tc.tenantID, tc.teamID)

			assert.NoError(t, err)
			assert.Len(t, result, 0)

			mockRepo.AssertExpectations(t)
		})
	}
}

// ========== REPOSITORY LAYER TESTS ==========

func TestNewArchiveCaseRepository(t *testing.T) {
	gormDB, mock := setupMockDBArchiveCases(t)
	defer func() {
		db, _ := gormDB.DB()
		db.Close()
	}()

	repo := listArchiveCases.NewArchiveCaseRepository(gormDB)

	assert.NotNil(t, repo)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestArchiveCaseRepository_ListArchivedCases_Success(t *testing.T) {
	gormDB, mock := setupMockDBArchiveCases(t)
	defer func() {
		db, _ := gormDB.DB()
		db.Close()
	}()

	repo := listArchiveCases.NewArchiveCaseRepository(gormDB)

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"

	// Mock the expected SQL query
	rows := sqlmock.NewRows([]string{
		"id", "title", "description", "status", "priority",
		"created_by", "tenant_id", "team_id", "created_at", "updated_at", "archived_at",
	}).
		AddRow(
			uuid.New(), "Archived Case 1", "Description 1", "archived", "high",
			uuid.New(), uuid.New(), uuid.New(), time.Now(), time.Now(), time.Now(),
		).
		AddRow(
			uuid.New(), "Archived Case 2", "Description 2", "archived", "medium",
			uuid.New(), uuid.New(), uuid.New(), time.Now(), time.Now(), time.Now(),
		)

	// Fix the SQL query pattern - use simpler pattern without escaping
	expectedQuery := `SELECT \* FROM "cases" WHERE status = \$1 AND created_by = \$2 AND tenant_id = \$3 AND team_id = \$4`
	mock.ExpectQuery(expectedQuery).
		WithArgs("archived", userID, tenantID, teamID).
		WillReturnRows(rows)

	result, err := repo.ListArchivedCases(userID, tenantID, teamID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	if len(result) > 0 {
		assert.Equal(t, "Archived Case 1", result[0].Title)
		assert.Equal(t, "archived", result[0].Status)
	}
	if len(result) > 1 {
		assert.Equal(t, "Archived Case 2", result[1].Title)
		assert.Equal(t, "archived", result[1].Status)
	}

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestArchiveCaseRepository_ListArchivedCases_EmptyResult(t *testing.T) {
	gormDB, mock := setupMockDBArchiveCases(t)
	defer func() {
		db, _ := gormDB.DB()
		db.Close()
	}()

	repo := listArchiveCases.NewArchiveCaseRepository(gormDB)

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"

	// Mock empty result
	rows := sqlmock.NewRows([]string{
		"id", "title", "description", "status", "priority",
		"created_by", "tenant_id", "team_id", "created_at", "updated_at", "archived_at",
	})

	expectedQuery := `SELECT \* FROM "cases" WHERE status = \$1 AND created_by = \$2 AND tenant_id = \$3 AND team_id = \$4`
	mock.ExpectQuery(expectedQuery).
		WithArgs("archived", userID, tenantID, teamID).
		WillReturnRows(rows)

	result, err := repo.ListArchivedCases(userID, tenantID, teamID)

	assert.NoError(t, err)
	assert.Len(t, result, 0)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestArchiveCaseRepository_ListArchivedCases_DatabaseError(t *testing.T) {
	gormDB, mock := setupMockDBArchiveCases(t)
	defer func() {
		db, _ := gormDB.DB()
		db.Close()
	}()

	repo := listArchiveCases.NewArchiveCaseRepository(gormDB)

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"

	expectedQuery := `SELECT \* FROM "cases" WHERE status = \$1 AND created_by = \$2 AND tenant_id = \$3 AND team_id = \$4`
	mock.ExpectQuery(expectedQuery).
		WithArgs("archived", userID, tenantID, teamID).
		WillReturnError(errors.New("database connection failed"))

	result, err := repo.ListArchivedCases(userID, tenantID, teamID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection failed")
	assert.Nil(t, result)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestArchiveCaseRepository_ListArchivedCases_SQLInjectionProtection(t *testing.T) {
	gormDB, mock := setupMockDBArchiveCases(t)
	defer func() {
		db, _ := gormDB.DB()
		db.Close()
	}()

	repo := listArchiveCases.NewArchiveCaseRepository(gormDB)

	// Attempt SQL injection
	userID := "test-user'; DROP TABLE cases; --"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"

	rows := sqlmock.NewRows([]string{
		"id", "title", "description", "status", "priority",
		"created_by", "tenant_id", "team_id", "created_at", "updated_at", "archived_at",
	})

	// The query should still be parameterized and safe
	expectedQuery := `SELECT \* FROM "cases" WHERE status = \$1 AND created_by = \$2 AND tenant_id = \$3 AND team_id = \$4`
	mock.ExpectQuery(expectedQuery).
		WithArgs("archived", userID, tenantID, teamID).
		WillReturnRows(rows)

	result, err := repo.ListArchivedCases(userID, tenantID, teamID)

	assert.NoError(t, err)
	assert.Len(t, result, 0)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestArchiveCaseService_Integration_Success(t *testing.T) {
	gormDB, mock := setupMockDBArchiveCases(t)
	defer func() {
		db, _ := gormDB.DB()
		db.Close()
	}()

	repo := listArchiveCases.NewArchiveCaseRepository(gormDB)
	service := listArchiveCases.NewArchiveCaseService(repo)

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"

	testID1 := uuid.New()
	testID2 := uuid.New()
	createdBy1 := uuid.New()
	createdBy2 := uuid.New()
	tenant1 := uuid.New()
	tenant2 := uuid.New()
	team1 := uuid.New()
	team2 := uuid.New()

	rows := sqlmock.NewRows([]string{
		"id", "title", "description", "status", "priority",
		"created_by", "tenant_id", "team_id", "created_at", "updated_at", "archived_at",
	}).
		AddRow(
			testID1, "Integration Test Case 1", "Description 1", "archived", "high",
			createdBy1, tenant1, team1, time.Now(), time.Now(), time.Now(),
		).
		AddRow(
			testID2, "Integration Test Case 2", "Description 2", "archived", "low",
			createdBy2, tenant2, team2, time.Now(), time.Now(), time.Now(),
		)

	expectedQuery := `SELECT \* FROM "cases" WHERE status = \$1 AND created_by = \$2 AND tenant_id = \$3 AND team_id = \$4`
	mock.ExpectQuery(expectedQuery).
		WithArgs("archived", userID, tenantID, teamID).
		WillReturnRows(rows)

	result, err := service.ListArchivedCases(userID, tenantID, teamID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	if len(result) > 0 {
		assert.Equal(t, "Integration Test Case 1", result[0].Title)
	}
	if len(result) > 1 {
		assert.Equal(t, "Integration Test Case 2", result[1].Title)
	}

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestArchiveCaseService_Integration_DatabaseError(t *testing.T) {
	gormDB, mock := setupMockDBArchiveCases(t)
	defer func() {
		db, _ := gormDB.DB()
		db.Close()
	}()

	repo := listArchiveCases.NewArchiveCaseRepository(gormDB)
	service := listArchiveCases.NewArchiveCaseService(repo)

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"

	expectedQuery := `SELECT \* FROM "cases" WHERE status = \$1 AND created_by = \$2 AND tenant_id = \$3 AND team_id = \$4`
	mock.ExpectQuery(expectedQuery).
		WithArgs("archived", userID, tenantID, teamID).
		WillReturnError(errors.New("connection timeout"))

	result, err := service.ListArchivedCases(userID, tenantID, teamID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection timeout")
	assert.Nil(t, result)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// ========== BENCHMARK TESTS ==========

func BenchmarkArchiveCaseService_ListArchivedCases(b *testing.B) {
	mockRepo := &MockArchiveCaseLister{}
	service := listArchiveCases.NewArchiveCaseService(mockRepo)

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"

	// Create a large dataset for benchmarking
	largeCaseSet := make([]listArchiveCases.ArchivedCase, 1000)
	for i := 0; i < 1000; i++ {
		largeCaseSet[i] = createTestArchivedCase(
			"Benchmark Case "+string(rune(i)),
			"Benchmark description "+string(rune(i)),
		)
	}

	mockRepo.On("ListArchivedCases", userID, tenantID, teamID).Return(largeCaseSet, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.ListArchivedCases(userID, tenantID, teamID)
		if err != nil {
			b.Fatal(err)
		}
	}
}
