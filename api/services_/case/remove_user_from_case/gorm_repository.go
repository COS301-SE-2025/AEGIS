package remove_user_from_case

import (
	"database/sql"
	"github.com/google/uuid"
)

type GormRepository struct {
	DB *sql.DB
}

func (r *GormRepository) IsAdmin(userID uuid.UUID) (bool, error) {
	var role string
	err := r.DB.QueryRow("SELECT role FROM users WHERE id = $1", userID).Scan(&role)
	if err != nil {
		return false, err
	}
	return role == "admin", nil
}

func (r *GormRepository) RemoveUserFromCase(userID, caseID uuid.UUID) error {
	_, err := r.DB.Exec("DELETE FROM case_user_roles WHERE user_id = $1 AND case_id = $2", userID, caseID)
	return err
}
