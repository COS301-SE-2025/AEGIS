package postgres

import (
	"database/sql"
	"errors"
	"fmt"

	"aegis-api/models"
)

type PostgresUserRepository struct {
	DB *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{DB: db}
}

func (r *PostgresUserRepository) GetUserByID(userID string) (*models.UserDTO, error) {
	var user models.UserDTO
	query := `SELECT id, full_name, email, password_hash, role, is_verified, verification_token, created_at FROM users WHERE id = $1`
	err := r.DB.QueryRow(query, userID).Scan(
		&user.ID,
		&user.FullName,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsVerified,
		&user.VerificationToken,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // or return an appropriate custom error
		}
		return nil, err
	}
	return &user, nil
}

func (r *PostgresUserRepository) GetUserByEmail(email string) (*models.UserDTO, error) {
	var user models.UserDTO
	query := `SELECT id, full_name, email, password_hash, role, is_verified, verification_token, created_at FROM users WHERE email = $1`
	err := r.DB.QueryRow(query, email).Scan(
		&user.ID,
		&user.FullName,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsVerified,
		&user.VerificationToken,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *PostgresUserRepository) UpdateUser(userID string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	query := "UPDATE users SET "
	params := []interface{}{}
	i := 1

	for field, value := range updates {
		if i > 1 {
			query += ", "
		}
		query += fmt.Sprintf("%s = $%d", field, i)
		params = append(params, value)
		i++
	}

	query += fmt.Sprintf(" WHERE id = $%d", i)
	params = append(params, userID)

	_, err := r.DB.Exec(query, params...)
	return err
}

func (r *PostgresUserRepository) GetUserRoles(userID string) ([]string, error) {
	// This assumes a separate roles table; adjust if role is stored in the `users` table
	query := `SELECT role FROM user_roles WHERE user_id = $1`
	rows, err := r.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, rows.Err()
}
