package unit_tests

import (
	"aegis-api/services_/case/ListCases"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setupTestDB_repo creates a mocked database connection for testing
func setupTestDB_repo(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err, "Failed to create sqlmock")

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	require.NoError(t, err, "Failed to create gorm DB")

	cleanup := func() {
		sqlDB.Close()
	}

	return gormDB, mock, cleanup
}

// TestGormCaseQueryRepository_GetAllCases tests the GetAllCases repository method
func TestGormCaseQueryRepository_GetAllCases(t *testing.T) {
	t.Run("returns all cases for tenant", func(t *testing.T) {
		db, mock, cleanup := setupTestDB_repo(t)
		defer cleanup()

		repo := ListCases.NewGormCaseQueryRepository(db)

		tenantID := uuid.New()
		caseID1 := uuid.New()
		caseID2 := uuid.New()
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "title", "status", "investigation_stage",
			"created_by", "created_at", "updated_at", "priority",
		}).
			AddRow(caseID1, tenantID, "Case 1", "active", "Analysis", uuid.Nil, now, now, "high").
			AddRow(caseID2, tenantID, "Case 2", "closed", "Case Closure & Review", uuid.Nil, now, now, "medium")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cases" WHERE tenant_id = $1`)).
			WithArgs(tenantID.String()).
			WillReturnRows(rows)

		cases, err := repo.GetAllCases(tenantID.String())

		assert.NoError(t, err)
		assert.Equal(t, 2, len(cases))
		assert.Equal(t, caseID1, cases[0].ID)
		assert.Equal(t, caseID2, cases[1].ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns empty slice for non-existent tenant", func(t *testing.T) {
		db, mock, cleanup := setupTestDB_repo(t)
		defer cleanup()

		repo := ListCases.NewGormCaseQueryRepository(db)
		tenantID := uuid.New()

		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "title", "status", "investigation_stage",
			"created_by", "created_at", "updated_at", "priority",
		})

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cases" WHERE tenant_id = $1`)).
			WithArgs(tenantID.String()).
			WillReturnRows(rows)

		cases, err := repo.GetAllCases(tenantID.String())

		assert.NoError(t, err)
		assert.Empty(t, cases)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns error on database failure", func(t *testing.T) {
		db, mock, cleanup := setupTestDB_repo(t)
		defer cleanup()

		repo := ListCases.NewGormCaseQueryRepository(db)
		tenantID := uuid.New()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cases" WHERE tenant_id = $1`)).
			WithArgs(tenantID.String()).
			WillReturnError(sql.ErrConnDone)

		cases, err := repo.GetAllCases(tenantID.String())

		assert.Error(t, err)
		assert.Nil(t, cases)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestGormCaseQueryRepository_GetCasesByUser tests the GetCasesByUser repository method
func TestGormCaseQueryRepository_GetCasesByUser(t *testing.T) {
	t.Run("returns cases created by user", func(t *testing.T) {
		db, mock, cleanup := setupTestDB_repo(t)
		defer cleanup()

		repo := ListCases.NewGormCaseQueryRepository(db)

		tenantID := uuid.New()
		userID := uuid.New()
		caseID1 := uuid.New()
		caseID2 := uuid.New()
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "title", "status", "investigation_stage",
			"created_by", "created_at", "updated_at", "priority",
		}).
			AddRow(caseID1, tenantID, "User Case 1", "active", "Analysis", userID, now, now, "high").
			AddRow(caseID2, tenantID, "User Case 2", "active", "Recovery", userID, now, now, "medium")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cases" WHERE created_by = $1 AND tenant_id = $2`)).
			WithArgs(userID.String(), tenantID.String()).
			WillReturnRows(rows)

		cases, err := repo.GetCasesByUser(userID.String(), tenantID.String())

		assert.NoError(t, err)
		assert.Equal(t, 2, len(cases))
		assert.Equal(t, userID, cases[0].CreatedBy)
		assert.Equal(t, userID, cases[1].CreatedBy)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns empty for user with no cases", func(t *testing.T) {
		db, mock, cleanup := setupTestDB_repo(t)
		defer cleanup()

		repo := ListCases.NewGormCaseQueryRepository(db)
		tenantID := uuid.New()
		userID := uuid.New()

		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "title", "status", "investigation_stage",
			"created_by", "created_at", "updated_at", "priority",
		})

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cases" WHERE created_by = $1 AND tenant_id = $2`)).
			WithArgs(userID.String(), tenantID.String()).
			WillReturnRows(rows)

		cases, err := repo.GetCasesByUser(userID.String(), tenantID.String())

		assert.NoError(t, err)
		assert.Empty(t, cases)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns error on database failure", func(t *testing.T) {
		db, mock, cleanup := setupTestDB_repo(t)
		defer cleanup()

		repo := ListCases.NewGormCaseQueryRepository(db)
		tenantID := uuid.New()
		userID := uuid.New()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cases" WHERE created_by = $1 AND tenant_id = $2`)).
			WithArgs(userID.String(), tenantID.String()).
			WillReturnError(sql.ErrConnDone)

		cases, err := repo.GetCasesByUser(userID.String(), tenantID.String())

		assert.Error(t, err)
		assert.Nil(t, cases)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestGormCaseQueryRepository_QueryCases tests the QueryCases repository method
func TestGormCaseQueryRepository_QueryCases(t *testing.T) {
	t.Run("filters by status", func(t *testing.T) {
		db, mock, cleanup := setupTestDB_repo(t)
		defer cleanup()

		repo := ListCases.NewGormCaseQueryRepository(db)

		tenantID := uuid.New()
		caseID := uuid.New()
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "title", "status", "investigation_stage",
			"created_by", "created_at", "updated_at", "priority",
		}).
			AddRow(caseID, tenantID, "Active Case", "active", "Analysis", uuid.Nil, now, now, "high")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cases" WHERE status = $1`)).
			WithArgs("active").
			WillReturnRows(rows)

		filter := ListCases.CaseFilter{
			TenantID: tenantID,
			Status:   "active",
		}

		cases, err := repo.QueryCases(filter)

		assert.NoError(t, err)
		assert.Equal(t, 1, len(cases))
		assert.Equal(t, "active", cases[0].Status)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("filters by priority", func(t *testing.T) {
		db, mock, cleanup := setupTestDB_repo(t)
		defer cleanup()

		repo := ListCases.NewGormCaseQueryRepository(db)

		tenantID := uuid.New()
		caseID := uuid.New()
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "title", "status", "investigation_stage",
			"created_by", "created_at", "updated_at", "priority",
		}).
			AddRow(caseID, tenantID, "High Priority Case", "active", "Analysis", uuid.Nil, now, now, "high")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cases" WHERE priority = $1`)).
			WithArgs("high").
			WillReturnRows(rows)

		filter := ListCases.CaseFilter{
			TenantID: tenantID,
			Priority: "high",
		}

		cases, err := repo.QueryCases(filter)

		assert.NoError(t, err)
		assert.Equal(t, 1, len(cases))
		assert.Equal(t, "high", cases[0].Priority)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("filters by title term with ILIKE", func(t *testing.T) {
		db, mock, cleanup := setupTestDB_repo(t)
		defer cleanup()

		repo := ListCases.NewGormCaseQueryRepository(db)

		tenantID := uuid.New()
		caseID := uuid.New()
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "title", "status", "investigation_stage",
			"created_by", "created_at", "updated_at", "priority",
		}).
			AddRow(caseID, tenantID, "Phishing Attack Investigation", "active", "Analysis", uuid.Nil, now, now, "high")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cases" WHERE title ILIKE $1`)).
			WithArgs("%phishing%").
			WillReturnRows(rows)

		filter := ListCases.CaseFilter{
			TenantID:  tenantID,
			TitleTerm: "phishing",
		}

		cases, err := repo.QueryCases(filter)

		assert.NoError(t, err)
		assert.Equal(t, 1, len(cases))
		assert.Contains(t, cases[0].Title, "Phishing")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("filters by created_by", func(t *testing.T) {
		db, mock, cleanup := setupTestDB_repo(t)
		defer cleanup()

		repo := ListCases.NewGormCaseQueryRepository(db)

		tenantID := uuid.New()
		userID := uuid.New()
		caseID := uuid.New()
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "title", "status", "investigation_stage",
			"created_by", "created_at", "updated_at", "priority",
		}).
			AddRow(caseID, tenantID, "User Case", "active", "Analysis", userID, now, now, "high")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cases" WHERE created_by = $1`)).
			WithArgs(userID.String()).
			WillReturnRows(rows)

		filter := ListCases.CaseFilter{
			TenantID:  tenantID,
			CreatedBy: userID.String(),
		}

		cases, err := repo.QueryCases(filter)

		assert.NoError(t, err)
		assert.Equal(t, 1, len(cases))
		assert.Equal(t, userID, cases[0].CreatedBy)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("applies sorting ascending", func(t *testing.T) {
		db, mock, cleanup := setupTestDB_repo(t)
		defer cleanup()

		repo := ListCases.NewGormCaseQueryRepository(db)

		tenantID := uuid.New()
		caseID1 := uuid.New()
		caseID2 := uuid.New()
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "title", "status", "investigation_stage",
			"created_by", "created_at", "updated_at", "priority",
		}).
			AddRow(caseID1, tenantID, "Case A", "active", "Analysis", uuid.Nil, now, now, "high").
			AddRow(caseID2, tenantID, "Case B", "active", "Analysis", uuid.Nil, now, now, "medium")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cases" ORDER BY title asc`)).
			WillReturnRows(rows)

		filter := ListCases.CaseFilter{
			TenantID:  tenantID,
			SortBy:    "title",
			SortOrder: "asc",
		}

		cases, err := repo.QueryCases(filter)

		assert.NoError(t, err)
		assert.Equal(t, 2, len(cases))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns empty for no matches", func(t *testing.T) {
		db, mock, cleanup := setupTestDB_repo(t)
		defer cleanup()

		repo := ListCases.NewGormCaseQueryRepository(db)
		tenantID := uuid.New()

		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "title", "status", "investigation_stage",
			"created_by", "created_at", "updated_at", "priority",
		})

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cases" WHERE status = $1`)).
			WithArgs("nonexistent").
			WillReturnRows(rows)

		filter := ListCases.CaseFilter{
			TenantID: tenantID,
			Status:   "nonexistent",
		}

		cases, err := repo.QueryCases(filter)

		assert.NoError(t, err)
		assert.Empty(t, cases)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("combines multiple filters", func(t *testing.T) {
		db, mock, cleanup := setupTestDB_repo(t)
		defer cleanup()

		repo := ListCases.NewGormCaseQueryRepository(db)

		tenantID := uuid.New()
		userID := uuid.New()
		caseID := uuid.New()
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "title", "status", "investigation_stage",
			"created_by", "created_at", "updated_at", "priority",
		}).
			AddRow(caseID, tenantID, "High Priority Active Case", "active", "Analysis", userID, now, now, "high")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cases" WHERE status = $1 AND priority = $2 AND created_by = $3`)).
			WithArgs("active", "high", userID.String()).
			WillReturnRows(rows)

		filter := ListCases.CaseFilter{
			TenantID:  tenantID,
			Status:    "active",
			Priority:  "high",
			CreatedBy: userID.String(),
		}

		cases, err := repo.QueryCases(filter)

		assert.NoError(t, err)
		assert.Equal(t, 1, len(cases))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns error on database failure", func(t *testing.T) {
		db, mock, cleanup := setupTestDB_repo(t)
		defer cleanup()

		repo := ListCases.NewGormCaseQueryRepository(db)
		tenantID := uuid.New()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cases" WHERE status = $1`)).
			WithArgs("active").
			WillReturnError(sql.ErrConnDone)

		filter := ListCases.CaseFilter{
			TenantID: tenantID,
			Status:   "active",
		}

		cases, err := repo.QueryCases(filter)

		assert.Error(t, err)
		assert.Nil(t, cases)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestGormCaseQueryRepository_GetCaseByID tests the GetCaseByID repository method
func TestGormCaseQueryRepository_GetCaseByID(t *testing.T) {
	t.Run("returns case by ID", func(t *testing.T) {
		db, mock, cleanup := setupTestDB_repo(t)
		defer cleanup()

		repo := ListCases.NewGormCaseQueryRepository(db)

		tenantID := uuid.New()
		caseID := uuid.New()
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "title", "status", "investigation_stage",
			"created_by", "created_at", "updated_at", "priority",
		}).
			AddRow(caseID, tenantID, "Test Case", "active", "Analysis", uuid.Nil, now, now, "high")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cases" WHERE id = $1 AND tenant_id = $2 ORDER BY "cases"."id" LIMIT $3`)).
			WithArgs(caseID.String(), tenantID.String(), 1).
			WillReturnRows(rows)

		result, err := repo.GetCaseByID(caseID.String(), tenantID.String())

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, caseID, result.ID)
		assert.Equal(t, tenantID, result.TenantID)
		assert.Equal(t, "Test Case", result.Title)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns error for non-existent case", func(t *testing.T) {
		db, mock, cleanup := setupTestDB_repo(t)
		defer cleanup()

		repo := ListCases.NewGormCaseQueryRepository(db)

		caseID := uuid.New()
		tenantID := uuid.New()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cases" WHERE id = $1 AND tenant_id = $2 ORDER BY "cases"."id" LIMIT $3`)).
			WithArgs(caseID.String(), tenantID.String(), 1).
			WillReturnError(gorm.ErrRecordNotFound)

		result, err := repo.GetCaseByID(caseID.String(), tenantID.String())

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns error on database failure", func(t *testing.T) {
		db, mock, cleanup := setupTestDB_repo(t)
		defer cleanup()

		repo := ListCases.NewGormCaseQueryRepository(db)

		caseID := uuid.New()
		tenantID := uuid.New()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cases" WHERE id = $1 AND tenant_id = $2 ORDER BY "cases"."id" LIMIT $3`)).
			WithArgs(caseID.String(), tenantID.String(), 1).
			WillReturnError(sql.ErrConnDone)

		result, err := repo.GetCaseByID(caseID.String(), tenantID.String())

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
