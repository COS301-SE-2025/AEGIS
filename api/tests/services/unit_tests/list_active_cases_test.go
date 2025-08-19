package unit_tests

import (
	"aegis-api/services_/case/ListActiveCases"
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestGetActiveCasesByUserID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)

	repo := ListActiveCases.NewActiveCaseRepository(gdb)
	userID := "some-user-id"

	mock.ExpectQuery(`SELECT .* FROM "cases"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "status", "investigation_stage", "priority", "created_by", "created_at",
		}).AddRow(
			"550e8400-e29b-41d4-a716-446655440000", "Case A", "Desc", "open", "analysis", "medium", "123e4567-e89b-12d3-a456-426614174000", time.Now(),
		))

	cases, err := repo.GetActiveCasesByUserID(context.Background(), userID, "", "")
	assert.NoError(t, err)
	assert.Len(t, cases, 1)
	assert.Equal(t, "Case A", cases[0].Title)
}
