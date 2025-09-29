package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type PermissionChecker interface {
	RoleHasPermission(role, permission string) (bool, error)
}

func RequirePermission(permission string, checker PermissionChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := c.GetString("userRole") // e.g., extracted from JWT or session
		if userRole == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing role"})
			c.Abort()
			return
		}

		hasPermission, err := checker.RoleHasPermission(userRole, permission)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking permissions"})
			c.Abort()
			return
		}
		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}
