package unit_tests
import (
	"testing"
	"time"
	"context"
	"aegis-api/services_/case/ListUsers"
	"github.com/stretchr/testify/assert"
	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

)
func TestListUsers(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)

	repo := ListUsers.NewUserRepository(gdb)

	mock.ExpectQuery(`SELECT .* FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "full_name", "email", "password_hash", "role", "is_verified", "verification_token", "created_at",
		}).AddRow(
			"123e4567-e89b-12d3-a456-426614174000", "Alice Sonders", "alice@example.com", "hashed_pwd", "Incident Responder", true, "token123", time.Now(),
		))

	users, err := repo.GetAllUsers(context.Background())
	assert.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, "Alice Sonders", users[0].Full_name)
}

