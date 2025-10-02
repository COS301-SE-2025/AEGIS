package unit_tests

import (
	"testing"

	"aegis-api/services_/timeline"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TestHelper to setup mock database
func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open gorm database: %v", err)
	}

	return gormDB, mock
}

func TestNewRepository(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.DB()

	// Expect the AutoMigrate call
	mock.ExpectExec("CREATE TABLE.*timeline_events.*").WillReturnResult(sqlmock.NewResult(0, 0))

	repo := timeline.NewRepository(db)

	assert.NotNil(t, repo)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_AutoMigrate(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.DB()

	repo := timeline.NewRepository(db)

	// Expect the AutoMigrate call
	mock.ExpectExec("CREATE TABLE.*timeline_events.*").WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.AutoMigrate()

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_AutoMigrate_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.DB()

	repo := timeline.NewRepository(db)

	// Expect the AutoMigrate call to fail
	mock.ExpectExec("CREATE TABLE.*timeline_events.*").WillReturnError(assert.AnError)

	err := repo.AutoMigrate()

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Create_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.DB()

	repo := timeline.NewRepository(db)

	event := &timeline.TimelineEvent{
		ID:          "test-id",
		CaseID:      "case-123",
		Description: "Test event",
		Severity:    "High",
		AnalystName: "John Doe",
		Order:       1,
	}

	// Expect INSERT query
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "timeline_events"`).
		WithArgs(event.ID, event.CaseID, event.Description, event.Severity, event.AnalystName, event.Order, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(event.ID))
	mock.ExpectCommit()

	err := repo.Create(event)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Create_AutoOrder(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.DB()

	repo := timeline.NewRepository(db)

	event := &timeline.TimelineEvent{
		ID:          "test-id",
		CaseID:      "case-123",
		Description: "Test event",
		Severity:    "High",
		AnalystName: "John Doe",
		Order:       0, // Order not set
	}

	// Expect SELECT for max order
	mock.ExpectQuery(`SELECT COALESCE\(MAX\("order"\), 0\) FROM "timeline_events" WHERE case_id = \$1`).
		WithArgs(event.CaseID).
		WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(2))

	// Expect INSERT query
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "timeline_events"`).
		WithArgs(event.ID, event.CaseID, event.Description, event.Severity, event.AnalystName, 3, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(event.ID))
	mock.ExpectCommit()

	err := repo.Create(event)

	assert.NoError(t, err)
	assert.Equal(t, 3, event.Order) // Should be set to max + 1
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Create_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.DB()

	repo := timeline.NewRepository(db)

	event := &timeline.TimelineEvent{
		ID:          "test-id",
		CaseID:      "case-123",
		Description: "Test event",
	}

	// Expect INSERT query to fail
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "timeline_events"`).WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err := repo.Create(event)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetByID_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.DB()

	repo := timeline.NewRepository(db)

	expectedEvent := &timeline.TimelineEvent{
		ID:          "test-id",
		CaseID:      "case-123",
		Description: "Test event",
		Severity:    "High",
		AnalystName: "John Doe",
		Order:       1,
	}

	// Expect SELECT query
	mock.ExpectQuery(`SELECT \* FROM "timeline_events" WHERE id = \$1 ORDER BY "timeline_events"."id" LIMIT \$2`).
		WithArgs("test-id", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "case_id", "description", "severity", "analyst_name", "order"}).
			AddRow(expectedEvent.ID, expectedEvent.CaseID, expectedEvent.Description, expectedEvent.Severity, expectedEvent.AnalystName, expectedEvent.Order))

	result, err := repo.GetByID("test-id")

	assert.NoError(t, err)
	assert.Equal(t, expectedEvent.ID, result.ID)
	assert.Equal(t, expectedEvent.CaseID, result.CaseID)
	assert.Equal(t, expectedEvent.Description, result.Description)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetByID_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.DB()

	repo := timeline.NewRepository(db)

	// Expect SELECT query to return no rows
	mock.ExpectQuery(`SELECT \* FROM "timeline_events" WHERE id = \$1 ORDER BY "timeline_events"."id" LIMIT \$2`).
		WithArgs("nonexistent", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	result, err := repo.GetByID("nonexistent")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_ListByCase_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.DB()

	repo := timeline.NewRepository(db)

	expectedEvents := []*timeline.TimelineEvent{
		{
			ID:          "event-1",
			CaseID:      "case-123",
			Description: "First event",
			Severity:    "High",
			Order:       1,
		},
		{
			ID:          "event-2",
			CaseID:      "case-123",
			Description: "Second event",
			Severity:    "Medium",
			Order:       2,
		},
	}

	// Expect SELECT query
	mock.ExpectQuery(`SELECT \* FROM "timeline_events" WHERE case_id = \$1 ORDER BY created_at ASC`).
		WithArgs("case-123").
		WillReturnRows(sqlmock.NewRows([]string{"id", "case_id", "description", "severity", "order"}).
			AddRow(expectedEvents[0].ID, expectedEvents[0].CaseID, expectedEvents[0].Description, expectedEvents[0].Severity, expectedEvents[0].Order).
			AddRow(expectedEvents[1].ID, expectedEvents[1].CaseID, expectedEvents[1].Description, expectedEvents[1].Severity, expectedEvents[1].Order))

	result, err := repo.ListByCase("case-123")

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, expectedEvents[0].ID, result[0].ID)
	assert.Equal(t, expectedEvents[1].ID, result[1].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_ListByCase_Empty(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.DB()

	repo := timeline.NewRepository(db)

	// Expect SELECT query to return no rows
	mock.ExpectQuery(`SELECT \* FROM "timeline_events" WHERE case_id = \$1 ORDER BY created_at ASC`).
		WithArgs("case-456").
		WillReturnRows(sqlmock.NewRows([]string{"id", "case_id", "description", "severity", "order"}))

	result, err := repo.ListByCase("case-456")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_ListByCase_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.DB()

	repo := timeline.NewRepository(db)

	// Expect SELECT query to fail
	mock.ExpectQuery(`SELECT \* FROM "timeline_events" WHERE case_id = \$1 ORDER BY created_at ASC`).
		WithArgs("case-123").
		WillReturnError(assert.AnError)

	result, err := repo.ListByCase("case-123")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Update_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.DB()

	repo := timeline.NewRepository(db)

	event := &timeline.TimelineEvent{
		ID:          "test-id",
		CaseID:      "case-123",
		Description: "Updated event",
		Severity:    "Low",
		AnalystName: "Jane Doe",
	}

	// Expect UPDATE query
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "timeline_events" SET .* WHERE id = \$1`).
		WithArgs(event.ID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.Update(event)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Update_MissingID(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.DB()

	repo := timeline.NewRepository(db)

	event := &timeline.TimelineEvent{
		CaseID:      "case-123",
		Description: "Updated event",
	}

	err := repo.Update(event)

	assert.Error(t, err)
	assert.Equal(t, "missing id on update", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Update_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.DB()

	repo := timeline.NewRepository(db)

	event := &timeline.TimelineEvent{
		ID:          "test-id",
		CaseID:      "case-123",
		Description: "Updated event",
	}

	// Expect UPDATE query to fail
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "timeline_events" SET .* WHERE id = \$1`).
		WithArgs(event.ID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err := repo.Update(event)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Delete_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.DB()

	repo := timeline.NewRepository(db)

	// Expect DELETE query
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "timeline_events" WHERE id = \$1`).
		WithArgs("test-id").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.Delete("test-id")

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Delete_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.DB()

	repo := timeline.NewRepository(db)

	// Expect DELETE query to fail
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "timeline_events" WHERE id = \$1`).
		WithArgs("test-id").
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err := repo.Delete("test-id")

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_UpdateOrder_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.DB()

	repo := timeline.NewRepository(db)

	caseID := "case-123"
	orderedIDs := []string{"event-3", "event-1", "event-2"}

	// Expect transaction and multiple UPDATE queries
	mock.ExpectBegin()
	for i, id := range orderedIDs {
		mock.ExpectExec(`UPDATE "timeline_events" SET "order" = \$1 WHERE id = \$2 AND case_id = \$3`).
			WithArgs(i+1, id, caseID).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}
	mock.ExpectCommit()

	err := repo.UpdateOrder(caseID, orderedIDs)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_UpdateOrder_EventNotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.DB()

	repo := timeline.NewRepository(db)

	caseID := "case-123"
	orderedIDs := []string{"event-1", "event-2"}

	// Expect transaction and first UPDATE succeeds, second fails
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "timeline_events" SET "order" = \$1 WHERE id = \$2 AND case_id = \$3`).
		WithArgs(1, "event-1", caseID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE "timeline_events" SET "order" = \$1 WHERE id = \$2 AND case_id = \$3`).
		WithArgs(2, "event-2", caseID).
		WillReturnResult(sqlmock.NewResult(0, 0)) // No rows affected
	mock.ExpectRollback()

	err := repo.UpdateOrder(caseID, orderedIDs)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "event event-2 not found for case case-123")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_UpdateOrder_EmptyList(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.DB()

	repo := timeline.NewRepository(db)

	err := repo.UpdateOrder("case-123", []string{})

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_UpdateOrder_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.DB()

	repo := timeline.NewRepository(db)

	caseID := "case-123"
	orderedIDs := []string{"event-1"}

	// Expect transaction to fail on first UPDATE
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "timeline_events" SET "order" = \$1 WHERE id = \$2 AND case_id = \$3`).
		WithArgs(1, "event-1", caseID).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err := repo.UpdateOrder(caseID, orderedIDs)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_FindByID_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.DB()

	repo := timeline.NewRepository(db)

	expectedEvent := &timeline.TimelineEvent{
		ID:          "test-id",
		CaseID:      "case-123",
		Description: "Test event",
		Severity:    "High",
	}

	// Expect SELECT query
	mock.ExpectQuery(`SELECT \* FROM "timeline_events" WHERE id = \$1 ORDER BY "timeline_events"."id" LIMIT \$2`).
		WithArgs("test-id", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "case_id", "description", "severity"}).
			AddRow(expectedEvent.ID, expectedEvent.CaseID, expectedEvent.Description, expectedEvent.Severity))

	result, err := repo.FindByID("test-id")

	assert.NoError(t, err)
	assert.Equal(t, expectedEvent.ID, result.ID)
	assert.Equal(t, expectedEvent.CaseID, result.CaseID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_FindByID_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.DB()

	repo := timeline.NewRepository(db)

	// Expect SELECT query to return no rows
	mock.ExpectQuery(`SELECT \* FROM "timeline_events" WHERE id = \$1 ORDER BY "timeline_events"."id" LIMIT \$2`).
		WithArgs("nonexistent", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	result, err := repo.FindByID("nonexistent")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}
