package middleware

import (
	"aegis-api/structs"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
				Error:   "unauthorized",
				Message: "Authorization token required",
			})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Parse with custom Claims struct
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecret, nil //[]byte(jstSecret),nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid or expired token",
			})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*Claims)
		if !ok || claims.UserID == "" || claims.Email == "" || claims.Role == "" {
			c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid token format",
			})
			c.Abort()
			return
		}

		// Set context using Claims fields
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("userRole", claims.Role)
		c.Next()
	}
}

// Role-based authorization middleware
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("userRole")
		if !exists {
			c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
				Error:   "unauthorized",
				Message: "Authentication required",
			})
			c.Abort()
			return
		}

		role := userRole.(string)
		//if !ok {
		//	c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
		//		Error:   "server_error",
		//		Message: "Internal server error",
		//	})
		//	c.Abort()
		//	return
		//}

		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				c.Next()
				return
			}
		}

		//log access failure

		c.JSON(http.StatusForbidden, structs.ErrorResponse{
			Error:   "forbidden",
			Message: "Insufficient permissions",
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
