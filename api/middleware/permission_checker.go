package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PermissionChecker interface {
	RoleHasPermission(role, permission string) (bool, error)
}

func RequirePermission(permission string, checker PermissionChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("[DEBUG] RequirePermission: Checking permission '%s' for userRole '%s'", permission, c.GetString("userRole"))
		userRole := c.GetString("userRole") // e.g., extracted from JWT or session
		if userRole == "" {
			log.Printf("[ERROR] RequirePermission: Unauthorized, missing role")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing role"})
			c.Abort()
			return
		}

		hasPermission, err := checker.RoleHasPermission(userRole, permission)
		if err != nil {
			log.Printf("[ERROR] RequirePermission: Error checking permissions for role '%s' and permission '%s': %v", userRole, permission, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking permissions"})
			c.Abort()
			return
		}
		if !hasPermission {
			log.Printf("[ERROR] RequirePermission: Forbidden, role '%s' lacks permission '%s'", userRole, permission)
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: insufficient permissions"})
			c.Abort()
			return
		}

		log.Printf("[DEBUG] RequirePermission: Permission '%s' granted for role '%s'", permission, userRole)
		c.Next()
	}
}
