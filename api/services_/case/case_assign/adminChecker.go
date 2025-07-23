// file: case_assign/jwt_admin_checker.go

package case_assign

import (
	"errors"

	"github.com/gin-gonic/gin"
)

type ContextAdminChecker struct{}

func NewContextAdminChecker() *ContextAdminChecker {
	return &ContextAdminChecker{}
}

// This version extracts role from Gin context rather than DB
func (c *ContextAdminChecker) IsAdminFromContext(ctx *gin.Context) (bool, error) {
	roleVal, exists := ctx.Get("userRole")
	if !exists {
		return false, errors.New("missing role in context")
	}
	role, ok := roleVal.(string)
	if !ok {
		return false, errors.New("invalid role type in context")
	}
	return role == "Admin", nil
}
