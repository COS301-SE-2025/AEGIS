// file: report_repository_sqlmock_test.go
package unit_tests

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	report "aegis-api/services_/report"
	"database/sql/driver"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// replace the helper with this one
// helper to provide a full row for SELECT * scans into report.Report
func fullReportRow(id uuid.UUID, overrides map[string]any) []driver.Value {
	now := time.Now()

	// Build defaults using driver.Value-friendly types:
	// - UUIDs as strings
	// - ints as int64
	// - timestamps as time.Time
	// - nullable text as empty string
	base := map[string]driver.Value{
		"id":            id.String(),
		"case_id":       uuid.Nil.String(),
		"examiner_id":   uuid.Nil.String(),
		"tenant_id":     uuid.Nil.String(),
		"team_id":       uuid.Nil.String(),
		"name":          "n",
		"mongo_id":      "",
		"report_number": "",
		"status":        "draft",
		"version":       int64(1),
		"date_examined": now,
		"file_path":     "/f",
		"created_at":    now,
		"updated_at":    now,
	}

	// Apply overrides (convert a few common types on the fly)
	for k, v := range overrides {
		switch vv := v.(type) {
		case uuid.UUID:
			base[k] = vv.String()
		case int:
			base[k] = int64(vv)
		default:
			base[k] = vv
		}
	}

	// Match the column order used in allReportCols
	return []driver.Value{
		base["id"],
		base["case_id"],
		base["examiner_id"],
		base["tenant_id"],
		base["team_id"],
		base["name"],
		base["mongo_id"],
		base["report_number"],
		base["status"],
		base["version"],
		base["date_examined"],
		base["file_path"],
		base["created_at"],
		base["updated_at"],
	}
}

var allReportCols = []string{
	"id", "case_id", "examiner_id", "tenant_id", "team_id", "name",
	"mongo_id", "report_number", "status", "version", "date_examined",
	"file_path", "created_at", "updated_at",
}

/* -------------------- test harness -------------------- */

func newPGMock(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()
	std, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)

	gdb, err := gorm.Open(
		postgres.New(postgres.Config{Conn: std}),
		&gorm.Config{
			SkipDefaultTransaction: true, // 👈 turn off BEGIN/COMMIT for tests
		},
	)
	require.NoError(t, err)
	return gdb, mock, std
}

/* -------------------- tests: SaveReport -------------------- */

// Ensures SaveReport issues an INSERT ... RETURNING "id".
func TestRepo_SaveReport_SQLMock(t *testing.T) {
	db, mock, std := newPGMock(t)
	defer std.Close()

	repo := report.NewReportRepository(db)
	r := &report.Report{
		ID:         uuid.New(), // we provide ID so we don't care about DB defaults
		CaseID:     uuid.New(),
		ExaminerID: uuid.New(),
		TenantID:   uuid.New(),
		TeamID:     uuid.New(),
		Name:       "new",
		Status:     "draft",
		Version:    1,
		FilePath:   "/x",
	}

	// GORM uses INSERT ... RETURNING "id" on Postgres; we don't assert args order to avoid brittleness.
	mock.ExpectQuery(`INSERT\s+INTO\s+"reports"\s*\(.*\)\s*VALUES\s*\(.*\)\s*RETURNING\s+"id"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(r.ID))

	err := repo.SaveReport(context.Background(), r)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

/* -------------------- tests: GetByID / DownloadReport -------------------- */

func TestRepo_GetByID_SQLMock(t *testing.T) {
	db, mock, std := newPGMock(t)
	defer std.Close()

	repo := report.NewReportRepository(db)
	id := uuid.New()

	// GORM often emits: SELECT * FROM "reports" WHERE id = $1 ORDER BY "reports"."id" LIMIT $2
	mock.ExpectQuery(`SELECT .* FROM "reports" WHERE id = \$1 .* LIMIT (?:\$2|1)`).
		WithArgs(id, 1).
		WillReturnRows(sqlmock.NewRows(allReportCols).AddRow(fullReportRow(id, map[string]any{"name": "got"})...))

	got, err := repo.GetByID(context.Background(), id.String())
	require.NoError(t, err)
	assert.Equal(t, "got", got.Name)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepo_DownloadReport_SQLMock(t *testing.T) {
	db, mock, std := newPGMock(t)
	defer std.Close()

	repo := report.NewReportRepository(db)
	id := uuid.New()

	mock.ExpectQuery(`SELECT .* FROM "reports" WHERE id = \$1 .* LIMIT (?:\$2|1)`).
		WithArgs(id, 1).
		WillReturnRows(sqlmock.NewRows(allReportCols).AddRow(fullReportRow(id, map[string]any{"file_path": "/dl"})...))

	got, err := repo.DownloadReport(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, "/dl", got.FilePath)

	require.NoError(t, mock.ExpectationsWereMet())
}

/* -------------------- tests: GetAllReports -------------------- */

// Ensures a plain SELECT over reports returns a slice.
func TestRepo_GetAllReports_SQLMock(t *testing.T) {
	db, mock, std := newPGMock(t)
	defer std.Close()

	repo := report.NewReportRepository(db)
	id1, id2 := uuid.New(), uuid.New()

	mock.ExpectQuery(`SELECT .* FROM "reports"`).
		WillReturnRows(sqlmock.NewRows(allReportCols).
			AddRow(fullReportRow(id1, nil)...).
			AddRow(fullReportRow(id2, nil)...),
		)

	rows, err := repo.GetAllReports(context.Background())
	require.NoError(t, err)
	assert.Len(t, rows, 2)

	require.NoError(t, mock.ExpectationsWereMet())
}

/* -------------------- tests: DeleteReportByID -------------------- */

// Ensures a DELETE where id = $1 returns rows affected.
func TestRepo_DeleteReportByID_SQLMock(t *testing.T) {
	db, mock, std := newPGMock(t)
	defer std.Close()

	repo := report.NewReportRepository(db)
	id := uuid.New()

	mock.ExpectExec(`DELETE\s+FROM\s+"reports"\s+WHERE\s+id\s*=\s*\$1`).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.DeleteReportByID(context.Background(), id)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

/* -------------------- tests: GetReportsByEvidenceID -------------------- */

// Ensures Where("evidence_id = ?") scan works (no schema needed in mocks).
func TestRepo_GetReportsByEvidenceID_SQLMock(t *testing.T) {
	db, mock, std := newPGMock(t)
	defer std.Close()

	repo := report.NewReportRepository(db)
	evID := uuid.New()

	// We don’t rely on exact binding format; allow any arg shape.
	mock.ExpectQuery(`SELECT .* FROM "reports" WHERE evidence_id = \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(allReportCols).AddRow(fullReportRow(uuid.New(), map[string]any{"name": "ev"})...))

	out, err := repo.GetReportsByEvidenceID(context.Background(), evID)
	require.NoError(t, err)
	require.Len(t, out, 1)
	assert.Equal(t, "ev", out[0].Name)

	require.NoError(t, mock.ExpectationsWereMet())
}

/* -------------------- tests: UpdateReportName -------------------- */

// Verifies UPDATE ... NOW() and a subsequent reload SELECT.
func TestRepo_UpdateReportName_SQLMock(t *testing.T) {
	db, mock, std := newPGMock(t)
	defer std.Close()

	repo := report.NewReportRepository(db)
	id := uuid.New()
	newName := "Renamed"

	mock.ExpectExec(`UPDATE\s+reports\s+SET\s+name\s*=\s*\$1,\s*version\s*=\s*COALESCE\(version,\s*0\)\s*\+\s*1,\s*updated_at\s*=\s*NOW\(\)\s*WHERE\s+id\s*=\s*\$2`).
		WithArgs(newName, id).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectQuery(`SELECT .* FROM "reports" WHERE id = \$1 .* LIMIT (?:\$2|1)`).
		WithArgs(id, 1).
		WillReturnRows(sqlmock.NewRows(allReportCols).AddRow(fullReportRow(id, map[string]any{"name": newName, "version": 2})...))

	got, err := repo.(*report.ReportsRepoImpl).UpdateReportName(context.Background(), id, newName)
	require.NoError(t, err)
	assert.Equal(t, newName, got.Name)
	assert.Equal(t, 2, got.Version)

	require.NoError(t, mock.ExpectationsWereMet())
}

// Negative case: UPDATE affects 0 rows → "not found" error.
func TestRepo_UpdateReportName_NotFound_SQLMock(t *testing.T) {
	db, mock, std := newPGMock(t)
	defer std.Close()

	repo := report.NewReportRepository(db)
	id := uuid.New()

	mock.ExpectExec(`UPDATE\s+reports\s+SET\s+name\s*=\s*\$1,\s*version\s*=\s*COALESCE\(version,\s*0\)\s*\+\s*1,\s*updated_at\s*=\s*NOW\(\)\s*WHERE\s+id\s*=\s*\$2`).
		WithArgs("X", id).
		WillReturnResult(sqlmock.NewResult(0, 0))

	_, err := repo.(*report.ReportsRepoImpl).UpdateReportName(context.Background(), id, "X")
	require.Error(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

/* -------------------- tests: ListRecentCandidates -------------------- */

// Default limit path (candidateLimit <= 0 → 60). We just ensure query runs and returns rows.
func TestRepo_ListRecentCandidates_DefaultLimit_SQLMock(t *testing.T) {
	db, mock, std := newPGMock(t)
	defer std.Close()

	repo := report.NewReportRepository(db)

	// Expect SELECT of the projected columns with ORDER BY updated_at DESC and LIMIT.
	re := regexp.MustCompile(`SELECT\s+id,\s*case_id,\s*examiner_id,\s*name,\s*status,\s*updated_at\s+FROM\s+"reports".*ORDER BY updated_at DESC.*LIMIT`)
	mock.ExpectQuery(re.String()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "case_id", "examiner_id", "name", "status", "updated_at"}).
			AddRow(uuid.New(), uuid.New(), uuid.New(), "A", "draft", time.Now()),
		)

	out, err := repo.ListRecentCandidates(context.Background(), report.RecentReportsOptions{}, 0)
	require.NoError(t, err)
	require.Len(t, out, 1)
	assert.Equal(t, "A", out[0].Name)

	require.NoError(t, mock.ExpectationsWereMet())
}

// MineOnly + ExaminerID filter adds WHERE examiner_id = $1
func TestRepo_ListRecentCandidates_MineOnly_SQLMock(t *testing.T) {
	db, mock, std := newPGMock(t)
	defer std.Close()

	repo := report.NewReportRepository(db)
	examinerID := uuid.New()

	mock.ExpectQuery(`SELECT\s+id,\s*case_id,\s*examiner_id,\s*name,\s*status,\s*updated_at\s+FROM\s+"reports"\s+WHERE\s+examiner_id\s*=\s*\$1.*ORDER BY updated_at DESC.*LIMIT`).
		WithArgs(examinerID, sqlmock.AnyArg()). // GORM often passes LIMIT as a bind, accept as AnyArg
		WillReturnRows(sqlmock.NewRows([]string{"id", "case_id", "examiner_id", "name", "status", "updated_at"}).
			AddRow(uuid.New(), uuid.New(), examinerID, "Mine", "draft", time.Now()),
		)

	opts := report.RecentReportsOptions{MineOnly: true, ExaminerID: examinerID}
	out, err := repo.ListRecentCandidates(context.Background(), opts, 100)
	require.NoError(t, err)
	require.Len(t, out, 1)
	assert.Equal(t, examinerID, out[0].ExaminerID)

	require.NoError(t, mock.ExpectationsWereMet())
}

// CaseID filter adds WHERE case_id = $1
func TestRepo_ListRecentCandidates_CaseFilter_SQLMock(t *testing.T) {
	db, mock, std := newPGMock(t)
	defer std.Close()

	repo := report.NewReportRepository(db)
	caseID := uuid.New()

	mock.ExpectQuery(`SELECT\s+id,\s*case_id,\s*examiner_id,\s*name,\s*status,\s*updated_at\s+FROM\s+"reports"\s+WHERE\s+case_id\s*=\s*\$1.*ORDER BY updated_at DESC.*LIMIT`).
		WithArgs(caseID, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "case_id", "examiner_id", "name", "status", "updated_at"}).
			AddRow(uuid.New(), caseID, uuid.New(), "ByCase", "review", time.Now()),
		)

	opts := report.RecentReportsOptions{CaseID: &caseID}
	out, err := repo.ListRecentCandidates(context.Background(), opts, 10)
	require.NoError(t, err)
	require.Len(t, out, 1)
	assert.Equal(t, caseID, out[0].CaseID)

	require.NoError(t, mock.ExpectationsWereMet())
}

// Status filter adds WHERE status = $1
func TestRepo_ListRecentCandidates_StatusFilter_SQLMock(t *testing.T) {
	db, mock, std := newPGMock(t)
	defer std.Close()

	repo := report.NewReportRepository(db)
	status := "review"

	mock.ExpectQuery(`SELECT\s+id,\s*case_id,\s*examiner_id,\s*name,\s*status,\s*updated_at\s+FROM\s+"reports"\s+WHERE\s+status\s*=\s*\$1.*ORDER BY updated_at DESC.*LIMIT`).
		WithArgs(status, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "case_id", "examiner_id", "name", "status", "updated_at"}).
			AddRow(uuid.New(), uuid.New(), uuid.New(), "ByStatus", status, time.Now()),
		)

	opts := report.RecentReportsOptions{Status: &status}
	out, err := repo.ListRecentCandidates(context.Background(), opts, 10)
	require.NoError(t, err)
	require.Len(t, out, 1)
	assert.Equal(t, status, out[0].Status)

	require.NoError(t, mock.ExpectationsWereMet())
}

/* -------------------- tests: Raw SQL (team/case) via sqlmock -------------------- */

// GetReportsByTeamID uses a large Raw query with joins and to_char(... AT TIME ZONE ...).
// We don’t validate the entire SQL, just that the right bindings are used and the scan works.
func TestRepo_GetReportsByTeamID_SQLMock(t *testing.T) {
	db, mock, std := newPGMock(t)
	defer std.Close()

	repo := report.NewReportRepository(db)
	tenantID, teamID := uuid.New(), uuid.New()

	cols := []string{
		"id", "case_id", "team_id", "name", "type", "status", "version",
		"last_modified", "file_path", "author", "collaborators", "case_name", "team_name",
	}

	mock.ExpectQuery(`SELECT\s+r\.id.*FROM\s+reports\s+r.*WHERE\s+r\.tenant_id\s*=\s*\$1\s+AND\s+r\.team_id\s*=\s*\$2.*ORDER BY\s+r\.updated_at\s+DESC`).
		WithArgs(tenantID, teamID).
		WillReturnRows(sqlmock.NewRows(cols).AddRow(
			uuid.New(), uuid.New(), teamID, "R1", "", "draft", 1, "2025-08-19T06:00:00Z",
			"/files/r1.pdf", "Alice", 2, "Case A", "Team X",
		))

	out, err := repo.GetReportsByTeamID(context.Background(), tenantID, teamID)
	require.NoError(t, err)
	require.Len(t, out, 1)
	assert.Equal(t, "R1", out[0].Name)
	assert.Equal(t, "Alice", out[0].Author)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepo_GetReportsByCaseID_SQLMock(t *testing.T) {
	db, mock, std := newPGMock(t)
	defer std.Close()

	repo := report.NewReportRepository(db)
	caseID := uuid.New()

	cols := []string{
		"id", "case_id", "team_id", "name", "type", "status", "version",
		"last_modified", "file_path", "author", "collaborators", "case_name", "team_name",
	}

	mock.ExpectQuery(`SELECT\s+r\.id.*FROM\s+reports\s+r.*WHERE\s+r\.case_id\s*=\s*\$1.*ORDER BY\s+r\.updated_at\s+DESC`).
		WithArgs(caseID).
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow(uuid.New(), caseID, uuid.New(), "R1", "", "review", 2, "2025-08-19T06:01:00Z", "/f1.pdf", "Bob", 3, "Case A", "Team Z").
			AddRow(uuid.New(), caseID, uuid.New(), "R0", "", "draft", 1, "2025-08-18T22:00:00Z", "/f0.pdf", "Bob", 3, "Case A", "Team Z"),
		)

	out, err := repo.GetReportsByCaseID(context.Background(), caseID)
	require.NoError(t, err)
	require.Len(t, out, 2)
	assert.Equal(t, "R1", out[0].Name)

	require.NoError(t, mock.ExpectationsWereMet())
}
