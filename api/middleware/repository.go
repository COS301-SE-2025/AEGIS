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
			SELECT 1 FROM enum_role_permissions rp
			JOIN permissions p ON rp.permission_id = p.id
			WHERE rp.role = $1 AND p.name = $2
			)
	`, role, permission).Scan(&exists)
	return exists, err
}
