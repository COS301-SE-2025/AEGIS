package tests

import (
    "aegis-api/models"
    "aegis-api/services/GetUpdate_Users"
    "database/sql"
    "regexp"
    "testing"
    "time"

    "github.com/DATA-DOG/go-sqlmock"
    "github.com/stretchr/testify/assert"
)

func TestGetUserByID_Success(t *testing.T) {
    db, mock, _ := sqlmock.New()
    defer db.Close()

    repo := GetUpdate_Users.NewPostgresUserRepository(db)

    expected := models.UserDTO{
        ID:                "user123",
        FullName:          "Jane Doe",
        Email:             "jane@example.com",
        PasswordHash:      "hashedpass",
        Role:              "admin",
        IsVerified:        true,
        VerificationToken: "token123",
        CreatedAt:         time.Now(),
    }

    rows := sqlmock.NewRows([]string{
        "id", "full_name", "email", "password_hash", "role", "is_verified", "verification_token", "created_at",
    }).AddRow(expected.ID, expected.FullName, expected.Email, expected.PasswordHash, expected.Role, expected.IsVerified, expected.VerificationToken, expected.CreatedAt)

    mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, full_name, email, password_hash, role, is_verified, verification_token, created_at FROM users WHERE id = $1`)).
        WithArgs("user123").WillReturnRows(rows)

    actual, err := repo.GetUserByID("user123")
    assert.NoError(t, err)
    assert.Equal(t, expected.ID, actual.ID)
}

func TestGetUserByEmail_NoRows(t *testing.T) {
    db, mock, _ := sqlmock.New()
    defer db.Close()

    repo := GetUpdate_Users.NewPostgresUserRepository(db)

    mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, full_name, email, password_hash, role, is_verified, verification_token, created_at FROM users WHERE email = $1`)).
        WithArgs("missing@example.com").WillReturnError(sql.ErrNoRows)

    user, err := repo.GetUserByEmail("missing@example.com")
    assert.NoError(t, err)
    assert.Nil(t, user)
}

func TestUpdateUser(t *testing.T) {
    db, mock, _ := sqlmock.New()
    defer db.Close()

    repo := GetUpdate_Users.NewPostgresUserRepository(db)

    updates := map[string]interface{}{
        "full_name": "Updated Name",
        "role":      "user",
    }

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE users SET full_name = $1, role = $2 WHERE id = $3`)).
    WithArgs("Updated Name", "user", "user123").
    WillReturnResult(sqlmock.NewResult(1, 1))


    err := repo.UpdateUser("user123", updates)
    assert.NoError(t, err)
}

func TestGetUserRoles(t *testing.T) {
    db, mock, _ := sqlmock.New()
    defer db.Close()

    repo := GetUpdate_Users.NewPostgresUserRepository(db)

    rows := sqlmock.NewRows([]string{"role"}).AddRow("admin").AddRow("editor")
    mock.ExpectQuery(regexp.QuoteMeta("SELECT role FROM user_roles WHERE user_id = $1")).
        WithArgs("user123").WillReturnRows(rows)

    roles, err := repo.GetUserRoles("user123")
    assert.NoError(t, err)
    assert.ElementsMatch(t, []string{"admin", "editor"}, roles)
}

func TestUpdateUser_EmptyUpdate(t *testing.T) {
    db, _, _ := sqlmock.New()
    defer db.Close()

    repo := GetUpdate_Users.NewPostgresUserRepository(db)

    err := repo.UpdateUser("user123", map[string]interface{}{})
    assert.NoError(t, err)
}
