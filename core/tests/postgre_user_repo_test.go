// postgres_user_repository_test.go
package tests

import (
	"database/sql"
	
	"testing"
	"time"
	"aegis-api/services/GetUpdate_Users"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestNewPostgresUserRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	
	repo := GetUpdate_Users.NewPostgresUserRepository(db)
	
	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.DB)
}

func TestPostgresUserRepository_GetUserByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	
	repo := GetUpdate_Users.NewPostgresUserRepository(db)
	
	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "full_name", "email", "password_hash", "role", 
		"is_verified", "verification_token", "created_at",
	}).AddRow("123", "John Doe", "john@example.com", "hashedpass", 
		"user", true, "token123", now)
	
	mock.ExpectQuery(`SELECT id, full_name, email, password_hash, role, is_verified, verification_token, created_at FROM users WHERE id = \$1`).
		WithArgs("123").
		WillReturnRows(rows)
	
	result, err := repo.GetUserByID("123")
	
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "123", result.ID)
	assert.Equal(t, "John Doe", result.FullName)
	assert.Equal(t, "john@example.com", result.Email)
	assert.Equal(t, "hashedpass", result.PasswordHash)
	assert.Equal(t, "user", result.Role)
	assert.True(t, result.IsVerified)
	assert.Equal(t, "token123", result.VerificationToken)
	assert.Equal(t, now, result.CreatedAt)
	
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresUserRepository_GetUserByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	
	repo :=GetUpdate_Users.NewPostgresUserRepository(db)
	
	mock.ExpectQuery(`SELECT id, full_name, email, password_hash, role, is_verified, verification_token, created_at FROM users WHERE id = \$1`).
		WithArgs("123").
		WillReturnError(sql.ErrNoRows)
	
	result, err := repo.GetUserByID("123")
	
	assert.NoError(t, err)
	assert.Nil(t, result)
	
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresUserRepository_GetUserByEmail_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	
	repo :=GetUpdate_Users.NewPostgresUserRepository(db)
	
	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "full_name", "email", "password_hash", "role", 
		"is_verified", "verification_token", "created_at",
	}).AddRow("123", "John Doe", "john@example.com", "hashedpass", 
		"user", true, "token123", now)
	
	mock.ExpectQuery(`SELECT id, full_name, email, password_hash, role, is_verified, verification_token, created_at FROM users WHERE email = \$1`).
		WithArgs("john@example.com").
		WillReturnRows(rows)
	
	result, err := repo.GetUserByEmail("john@example.com")
	
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "123", result.ID)
	assert.Equal(t, "john@example.com", result.Email)
	
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresUserRepository_UpdateUser_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	
	repo :=GetUpdate_Users.NewPostgresUserRepository(db)
	
	updates := map[string]interface{}{
		"full_name": "Jane Doe",
		"email":     "jane@example.com",
	}
	
	mock.ExpectExec(`UPDATE users SET .+ WHERE id = \$3`).
		WithArgs("Jane Doe", "jane@example.com", "123").
		WillReturnResult(sqlmock.NewResult(1, 1))
	
	err = repo.UpdateUser("123", updates)
	
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresUserRepository_UpdateUser_EmptyUpdates(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	
	repo := GetUpdate_Users.NewPostgresUserRepository(db)
	
	updates := map[string]interface{}{}
	
	err = repo.UpdateUser("123", updates)
	
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresUserRepository_GetUserRoles_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	
	repo :=GetUpdate_Users.NewPostgresUserRepository(db)
	
	rows := sqlmock.NewRows([]string{"role"}).
		AddRow("admin").
		AddRow("user")
	
	mock.ExpectQuery(`SELECT role FROM user_roles WHERE user_id = \$1`).
		WithArgs("123").
		WillReturnRows(rows)
	
	result, err := repo.GetUserRoles("123")
	
	assert.NoError(t, err)
	assert.Equal(t, []string{"admin", "user"}, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresUserRepository_GetUserRoles_NoRoles(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	
	repo :=GetUpdate_Users.NewPostgresUserRepository(db)
	
	rows := sqlmock.NewRows([]string{"role"})
	
	mock.ExpectQuery(`SELECT role FROM user_roles WHERE user_id = \$1`).
		WithArgs("123").
		WillReturnRows(rows)
	
	result, err := repo.GetUserRoles("123")
	
	assert.NoError(t, err)
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}
