package middleware

import (
	"database/sql"
)

type DBPermissionChecker struct {
	DB *sql.DB
}

func (p *DBPermissionChecker) RoleHasPermission(role string, permission string) (bool, error) {
	var exists bool
	err := p.DB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM role_permissions
			WHERE role = $1 AND permission = $2
		)
	`, role, permission).Scan(&exists)
	return exists, err
}
