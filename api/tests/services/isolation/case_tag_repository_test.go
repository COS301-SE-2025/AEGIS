package isolation


import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"aegis-api/repositories"
)

func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	gdb, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	assert.NoError(t, err)

	return gdb, mock, func() {
		db.Close()
	}
}

func TestAddTagsToCase(t *testing.T) {
	db, mock, closeFn := setupTestDB(t)
	defer closeFn()

	repo := repositories.NewCaseTagRepository(db)

	caseID := uuid.New()
	userID := uuid.New()
	tagName := "urgent"

	mock.ExpectBegin()

	// FirstOrCreate: SELECT (tag doesn't exist yet)
	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT * FROM "tags" WHERE "tags"."name" = $1 ORDER BY "tags"."id" LIMIT $2`,
	)).
		WithArgs(tagName, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"})) // simulate not found

	// INSERT INTO tags ... RETURNING id — needs ExpectQuery
	mock.ExpectQuery(regexp.QuoteMeta(
		`INSERT INTO "tags" ("name") VALUES ($1) RETURNING "id"`,
	)).
		WithArgs(tagName).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1)) // tag ID = 1

	// Insert into case_tags
	mock.ExpectExec(regexp.QuoteMeta(
		`INSERT INTO "case_tags" ("case_id","tag_id") VALUES ($1,$2) ON CONFLICT DO NOTHING`,
	)).
		WithArgs(caseID, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err := repo.AddTagsToCase(context.Background(), userID, caseID, []string{tagName})
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
