package unit_tests

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

func setupEvidenceTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
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

func TestAddTagsToEvidence(t *testing.T) {
	db, mock, closeFn := setupEvidenceTestDB(t)
	defer closeFn()

	repo := repositories.NewEvidenceTagRepository(db)

	evidenceID := uuid.New()
	userID := uuid.New()
	tagName := "Urgent"

	mock.ExpectBegin()

	// Tag not found
	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT * FROM "tags" WHERE "tags"."name" = $1 ORDER BY "tags"."id" LIMIT $2`,
	)).
		WithArgs("urgent", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"})) // simulate not found

	// Insert new tag
	mock.ExpectQuery(regexp.QuoteMeta(
		`INSERT INTO "tags" ("name") VALUES ($1) RETURNING "id"`,
	)).
		WithArgs("urgent").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Insert into evidence_tags
	mock.ExpectExec(regexp.QuoteMeta(
		`INSERT INTO "evidence_tags" ("evidence_id","tag_id") VALUES ($1,$2) ON CONFLICT DO NOTHING`,
	)).
		WithArgs(evidenceID, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err := repo.AddTagsToEvidence(context.Background(), userID, evidenceID, []string{tagName})
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRemoveTagsFromEvidence(t *testing.T) {
	db, mock, closeFn := setupEvidenceTestDB(t)
	defer closeFn()

	repo := repositories.NewEvidenceTagRepository(db)

	evidenceID := uuid.New()
	userID := uuid.New()
	tagName := "duplicate"

	// Expect lookup
	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT * FROM "tags" WHERE "tags"."name" = $1 ORDER BY "tags"."id" LIMIT $2`,
	)).
		WithArgs("duplicate", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(3, "duplicate"))

	// Expect delete
	mock.ExpectExec(regexp.QuoteMeta(
		`DELETE FROM "evidence_tags" WHERE evidence_id = $1 AND tag_id = $2`,
	)).
		WithArgs(evidenceID, 3).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.RemoveTagsFromEvidence(context.Background(), userID, evidenceID, []string{tagName})
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetTagsForEvidence(t *testing.T) {
	db, mock, closeFn := setupEvidenceTestDB(t)
	defer closeFn()

	repo := repositories.NewEvidenceTagRepository(db)
	evidenceID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT tags.name FROM tags JOIN evidence_tags ON tags.id = evidence_tags.tag_id WHERE evidence_tags.evidence_id = $1`,
	)).
		WithArgs(evidenceID).
		WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("urgent").AddRow("legal"))

	tags, err := repo.GetTagsForEvidence(context.Background(), evidenceID)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"urgent", "legal"}, tags)
	assert.NoError(t, mock.ExpectationsWereMet())
}
