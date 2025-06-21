package middleware

import (
	"aegis-api/structs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Simple auth middleware - just checks for any token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")

		//check if token exists and starts with "Bearer "
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Authorization token required",
			})
			c.Abort()
			return
		}

		// Extract token (remove "Bearer " prefix)
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// For PoC, set a mock user based on token
		// In real implementation, you'd decode JWT here
		var userID, userRole string
		//switch token {
		//case "admin-token":
		//	userID = "admin-123"
		//	userRole = "DFIR Manager"
		//case "analyst-token":
		//	userID = "analyst-456"
		//	userRole = "Forensic Analyst"
		//case "user-token":
		//	userID = "user-789"
		//	userRole = "Incident Responder"
		//default:
		//	//accept any token and set default user ****
		//	userID = "default-user-999"
		//	userRole = "Generic user"
		//}
		parts := strings.SplitN(token, ":", 2)
		if len(parts) != 2 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Invalid authorisation token.",
			})
			c.Abort()
			return
		}

		userID = strings.TrimSpace(parts[0])
		userRole = strings.TrimSpace(parts[1])

		if userID == "" || userRole == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Missing userID or userRole in token",
			})
			c.Abort()
			return
		}

		// Set context for handlers
		c.Set("userID", userID)
		c.Set("userRole", userRole)
		c.Next()
	}
}

// Role-based authorization middleware
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("userRole")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "User role not found",
			})
			c.Abort()
			return
		}

		role := userRole.(string)
		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error":   "forbidden",
			"message": "Insufficient permissions",
		})
		c.Abort()
	}
}

// CORS middleware -- vite proxy

// Request logging middleware
//func LoggingMiddleware() gin.HandlerFunc {
//	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
//		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
//			param.ClientIP,
//			param.TimeStamp.Format("02/Jan/2006:15:04:05 -0700"),
//			param.Method,
//			param.Path,
//			param.Request.Proto,
//			param.StatusCode,
//			param.Latency,
//			param.Request.UserAgent(),
//			param.ErrorMessage,
//		)
//	})
//}

// get user id based on which path was used
func GetTargetUserID(c *gin.Context) (string, bool) {
	targetUserID := c.Param("user_id")
	role, _ := c.Get("userRole")

	if targetUserID != "" && role == "Admin" { //admin to view any user profile
		return targetUserID, true
	}

	currUserID, exists := c.Get("userID") //user to view own profile
	if !exists {
		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return "", false
	}

	return currUserID.(string), true
}
