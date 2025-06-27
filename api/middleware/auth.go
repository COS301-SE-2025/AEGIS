package middleware

// import (
// 	"net/http"
// 	"strings"

// 	"github.com/gin-gonic/gin"
// )

// func AuthMiddleware() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		authHeader := c.GetHeader("Authorization")
// 		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
// 			c.Abort()
// 			return
// 		}

// 		token := strings.TrimPrefix(authHeader, "Bearer ")
// 		parts := strings.SplitN(token, ":", 2) // TEMPORARY for PoC

// 		if len(parts) != 2 {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
// 			c.Abort()
// 			return
// 		}

// 		userID := strings.TrimSpace(parts[0])
// 		userRole := strings.TrimSpace(parts[1])

// 		if userID == "" || userRole == "" {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID or role"})
// 			c.Abort()
// 			return
// 		}

// 		c.Set("userID", userID)
// 		c.Set("userRole", userRole)
// 		c.Next()
// 	}
// }
