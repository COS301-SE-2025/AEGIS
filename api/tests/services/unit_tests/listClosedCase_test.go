package unit_tests

import (
	"testing"
	"time"
	"context"
	"aegis-api/services_/case/ListClosedCases"
	"github.com/stretchr/testify/assert"
	"github.com/DATA-DOG/go-sqlmock"
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

	cases, err := repo.GetClosedCasesByUserID(context.Background(), userID)
	assert.NoError(t, err)
	assert.Len(t, cases, 1)
	assert.Equal(t, "closed", string(cases[0].Status))
}