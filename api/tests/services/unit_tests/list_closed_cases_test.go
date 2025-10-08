package unit_tests

import (
	"aegis-api/services_/case/ListClosedCases"
	"context"
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

// setupClosedCasesTestDB creates a mocked database connection for testing
func setupClosedCasesTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
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

// TestClosedCaseRepository_GetClosedCasesByUserID tests the GetClosedCasesByUserID repository method
func TestClosedCaseRepository_GetClosedCasesByUserID(t *testing.T) {
	t.Run("returns closed cases where user is creator", func(t *testing.T) {
		db, mock, cleanup := setupClosedCasesTestDB(t)
		defer cleanup()

		repo := ListClosedCases.NewClosedCaseRepository(db)

		userID := uuid.New()
		tenantID := uuid.New()
		teamID := uuid.New()
		caseID1 := uuid.New()
		caseID2 := uuid.New()
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "team_id", "title", "status", "investigation_stage",
			"created_by", "created_at", "updated_at", "priority",
		}).
			AddRow(caseID1, tenantID, teamID, "Closed Case 1", "closed", "Case Closure & Review", userID, now, now, "high").
			AddRow(caseID2, tenantID, teamID, "Closed Case 2", "closed", "Case Closure & Review", userID, now, now, "medium")

		expectedQuery := `SELECT DISTINCT cases.* FROM "cases" LEFT JOIN case_user_roles ON case_user_roles.case_id = cases.id ` +
			`WHERE (case_user_roles.user_id = $1 OR cases.created_by = $2) AND cases.status = $3 AND cases.tenant_id = $4 AND cases.team_id = $5`

		mock.ExpectQuery(regexp.QuoteMeta(expectedQuery)).
			WithArgs(userID.String(), userID.String(), "closed", tenantID.String(), teamID.String()).
			WillReturnRows(rows)

		cases, err := repo.GetClosedCasesByUserID(context.Background(), userID.String(), tenantID.String(), teamID.String())

		assert.NoError(t, err)
		assert.Equal(t, 2, len(cases))
		assert.Equal(t, caseID1, cases[0].ID)
		assert.Equal(t, caseID2, cases[1].ID)
		assert.Equal(t, "closed", cases[0].Status)
		assert.Equal(t, "closed", cases[1].Status)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns closed cases where user has role via case_user_roles", func(t *testing.T) {
		db, mock, cleanup := setupClosedCasesTestDB(t)
		defer cleanup()

		repo := ListClosedCases.NewClosedCaseRepository(db)

		userID := uuid.New()
		creatorID := uuid.New() // Different from userID
		tenantID := uuid.New()
		teamID := uuid.New()
		caseID := uuid.New()
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "team_id", "title", "status", "investigation_stage",
			"created_by", "created_at", "updated_at", "priority",
		}).
			AddRow(caseID, tenantID, teamID, "Assigned Closed Case", "closed", "Case Closure & Review", creatorID, now, now, "high")

		expectedQuery := `SELECT DISTINCT cases.* FROM "cases" LEFT JOIN case_user_roles ON case_user_roles.case_id = cases.id ` +
			`WHERE (case_user_roles.user_id = $1 OR cases.created_by = $2) AND cases.status = $3 AND cases.tenant_id = $4 AND cases.team_id = $5`

		mock.ExpectQuery(regexp.QuoteMeta(expectedQuery)).
			WithArgs(userID.String(), userID.String(), "closed", tenantID.String(), teamID.String()).
			WillReturnRows(rows)

		cases, err := repo.GetClosedCasesByUserID(context.Background(), userID.String(), tenantID.String(), teamID.String())

		assert.NoError(t, err)
		assert.Equal(t, 1, len(cases))
		assert.Equal(t, caseID, cases[0].ID)
		assert.Equal(t, "closed", cases[0].Status)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns DISTINCT cases (no duplicates)", func(t *testing.T) {
		db, mock, cleanup := setupClosedCasesTestDB(t)
		defer cleanup()

		repo := ListClosedCases.NewClosedCaseRepository(db)

		userID := uuid.New()
		tenantID := uuid.New()
		teamID := uuid.New()
		caseID := uuid.New()
		now := time.Now()

		// Even if a case matches both conditions (created_by AND has role),
		// DISTINCT should ensure it appears only once
		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "team_id", "title", "status", "investigation_stage",
			"created_by", "created_at", "updated_at", "priority",
		}).
			AddRow(caseID, tenantID, teamID, "Case", "closed", "Case Closure & Review", userID, now, now, "high")

		expectedQuery := `SELECT DISTINCT cases.* FROM "cases" LEFT JOIN case_user_roles ON case_user_roles.case_id = cases.id ` +
			`WHERE (case_user_roles.user_id = $1 OR cases.created_by = $2) AND cases.status = $3 AND cases.tenant_id = $4 AND cases.team_id = $5`

		mock.ExpectQuery(regexp.QuoteMeta(expectedQuery)).
			WithArgs(userID.String(), userID.String(), "closed", tenantID.String(), teamID.String()).
			WillReturnRows(rows)

		cases, err := repo.GetClosedCasesByUserID(context.Background(), userID.String(), tenantID.String(), teamID.String())

		assert.NoError(t, err)
		assert.Equal(t, 1, len(cases), "DISTINCT should prevent duplicates")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("only returns closed cases, not active ones", func(t *testing.T) {
		db, mock, cleanup := setupClosedCasesTestDB(t)
		defer cleanup()

		repo := ListClosedCases.NewClosedCaseRepository(db)

		userID := uuid.New()
		tenantID := uuid.New()
		teamID := uuid.New()
		caseID := uuid.New()
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "team_id", "title", "status", "investigation_stage",
			"created_by", "created_at", "updated_at", "priority",
		}).
			AddRow(caseID, tenantID, teamID, "Closed Case", "closed", "Case Closure & Review", userID, now, now, "high")

		expectedQuery := `SELECT DISTINCT cases.* FROM "cases" LEFT JOIN case_user_roles ON case_user_roles.case_id = cases.id ` +
			`WHERE (case_user_roles.user_id = $1 OR cases.created_by = $2) AND cases.status = $3 AND cases.tenant_id = $4 AND cases.team_id = $5`

		mock.ExpectQuery(regexp.QuoteMeta(expectedQuery)).
			WithArgs(userID.String(), userID.String(), "closed", tenantID.String(), teamID.String()).
			WillReturnRows(rows)

		cases, err := repo.GetClosedCasesByUserID(context.Background(), userID.String(), tenantID.String(), teamID.String())

		assert.NoError(t, err)
		assert.Equal(t, 1, len(cases))
		assert.Equal(t, "closed", cases[0].Status)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("respects tenant isolation", func(t *testing.T) {
		db, mock, cleanup := setupClosedCasesTestDB(t)
		defer cleanup()

		repo := ListClosedCases.NewClosedCaseRepository(db)

		userID := uuid.New()
		tenantID := uuid.New()
		teamID := uuid.New()

		// Return no rows because cases from other tenants should be filtered out
		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "team_id", "title", "status", "investigation_stage",
			"created_by", "created_at", "updated_at", "priority",
		})

		expectedQuery := `SELECT DISTINCT cases.* FROM "cases" LEFT JOIN case_user_roles ON case_user_roles.case_id = cases.id ` +
			`WHERE (case_user_roles.user_id = $1 OR cases.created_by = $2) AND cases.status = $3 AND cases.tenant_id = $4 AND cases.team_id = $5`

		mock.ExpectQuery(regexp.QuoteMeta(expectedQuery)).
			WithArgs(userID.String(), userID.String(), "closed", tenantID.String(), teamID.String()).
			WillReturnRows(rows)

		cases, err := repo.GetClosedCasesByUserID(context.Background(), userID.String(), tenantID.String(), teamID.String())

		assert.NoError(t, err)
		assert.Empty(t, cases, "Should not return cases from other tenants")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("respects team isolation", func(t *testing.T) {
		db, mock, cleanup := setupClosedCasesTestDB(t)
		defer cleanup()

		repo := ListClosedCases.NewClosedCaseRepository(db)

		userID := uuid.New()
		tenantID := uuid.New()
		teamID := uuid.New()

		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "team_id", "title", "status", "investigation_stage",
			"created_by", "created_at", "updated_at", "priority",
		})

		expectedQuery := `SELECT DISTINCT cases.* FROM "cases" LEFT JOIN case_user_roles ON case_user_roles.case_id = cases.id ` +
			`WHERE (case_user_roles.user_id = $1 OR cases.created_by = $2) AND cases.status = $3 AND cases.tenant_id = $4 AND cases.team_id = $5`

		mock.ExpectQuery(regexp.QuoteMeta(expectedQuery)).
			WithArgs(userID.String(), userID.String(), "closed", tenantID.String(), teamID.String()).
			WillReturnRows(rows)

		cases, err := repo.GetClosedCasesByUserID(context.Background(), userID.String(), tenantID.String(), teamID.String())

		assert.NoError(t, err)
		assert.Empty(t, cases, "Should not return cases from other teams")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns empty slice when user has no closed cases", func(t *testing.T) {
		db, mock, cleanup := setupClosedCasesTestDB(t)
		defer cleanup()

		repo := ListClosedCases.NewClosedCaseRepository(db)

		userID := uuid.New()
		tenantID := uuid.New()
		teamID := uuid.New()

		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "team_id", "title", "status", "investigation_stage",
			"created_by", "created_at", "updated_at", "priority",
		})

		expectedQuery := `SELECT DISTINCT cases.* FROM "cases" LEFT JOIN case_user_roles ON case_user_roles.case_id = cases.id ` +
			`WHERE (case_user_roles.user_id = $1 OR cases.created_by = $2) AND cases.status = $3 AND cases.tenant_id = $4 AND cases.team_id = $5`

		mock.ExpectQuery(regexp.QuoteMeta(expectedQuery)).
			WithArgs(userID.String(), userID.String(), "closed", tenantID.String(), teamID.String()).
			WillReturnRows(rows)

		cases, err := repo.GetClosedCasesByUserID(context.Background(), userID.String(), tenantID.String(), teamID.String())

		assert.NoError(t, err)
		assert.Empty(t, cases)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns error on database failure", func(t *testing.T) {
		db, mock, cleanup := setupClosedCasesTestDB(t)
		defer cleanup()

		repo := ListClosedCases.NewClosedCaseRepository(db)

		userID := uuid.New()
		tenantID := uuid.New()
		teamID := uuid.New()

		expectedQuery := `SELECT DISTINCT cases.* FROM "cases" LEFT JOIN case_user_roles ON case_user_roles.case_id = cases.id ` +
			`WHERE (case_user_roles.user_id = $1 OR cases.created_by = $2) AND cases.status = $3 AND cases.tenant_id = $4 AND cases.team_id = $5`

		mock.ExpectQuery(regexp.QuoteMeta(expectedQuery)).
			WithArgs(userID.String(), userID.String(), "closed", tenantID.String(), teamID.String()).
			WillReturnError(sql.ErrConnDone)

		cases, err := repo.GetClosedCasesByUserID(context.Background(), userID.String(), tenantID.String(), teamID.String())

		assert.Error(t, err)
		assert.Nil(t, cases)
		assert.Equal(t, sql.ErrConnDone, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("handles context cancellation", func(t *testing.T) {
		db, mock, cleanup := setupClosedCasesTestDB(t)
		defer cleanup()

		repo := ListClosedCases.NewClosedCaseRepository(db)

		userID := uuid.New()
		tenantID := uuid.New()
		teamID := uuid.New()

		// Create a cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		expectedQuery := `SELECT DISTINCT cases.* FROM "cases" LEFT JOIN case_user_roles ON case_user_roles.case_id = cases.id ` +
			`WHERE (case_user_roles.user_id = $1 OR cases.created_by = $2) AND cases.status = $3 AND cases.tenant_id = $4 AND cases.team_id = $5`

		mock.ExpectQuery(regexp.QuoteMeta(expectedQuery)).
			WithArgs(userID.String(), userID.String(), "closed", tenantID.String(), teamID.String()).
			WillReturnError(context.Canceled)

		cases, err := repo.GetClosedCasesByUserID(ctx, userID.String(), tenantID.String(), teamID.String())

		assert.Error(t, err)
		assert.Nil(t, cases)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("handles multiple cases with different priorities", func(t *testing.T) {
		db, mock, cleanup := setupClosedCasesTestDB(t)
		defer cleanup()

		repo := ListClosedCases.NewClosedCaseRepository(db)

		userID := uuid.New()
		tenantID := uuid.New()
		teamID := uuid.New()
		caseID1 := uuid.New()
		caseID2 := uuid.New()
		caseID3 := uuid.New()
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "team_id", "title", "status", "investigation_stage",
			"created_by", "created_at", "updated_at", "priority",
		}).
			AddRow(caseID1, tenantID, teamID, "High Priority Closed", "closed", "Case Closure & Review", userID, now, now, "high").
			AddRow(caseID2, tenantID, teamID, "Medium Priority Closed", "closed", "Case Closure & Review", userID, now, now, "medium").
			AddRow(caseID3, tenantID, teamID, "Low Priority Closed", "closed", "Case Closure & Review", userID, now, now, "low")

		expectedQuery := `SELECT DISTINCT cases.* FROM "cases" LEFT JOIN case_user_roles ON case_user_roles.case_id = cases.id ` +
			`WHERE (case_user_roles.user_id = $1 OR cases.created_by = $2) AND cases.status = $3 AND cases.tenant_id = $4 AND cases.team_id = $5`

		mock.ExpectQuery(regexp.QuoteMeta(expectedQuery)).
			WithArgs(userID.String(), userID.String(), "closed", tenantID.String(), teamID.String()).
			WillReturnRows(rows)

		cases, err := repo.GetClosedCasesByUserID(context.Background(), userID.String(), tenantID.String(), teamID.String())

		assert.NoError(t, err)
		assert.Equal(t, 3, len(cases))
		assert.Equal(t, "high", cases[0].Priority)
		assert.Equal(t, "medium", cases[1].Priority)
		assert.Equal(t, "low", cases[2].Priority)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
