package middleware

import (
	"aegis-api/structs"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("Authorization Header:", c.GetHeader("Authorization"))
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
		fmt.Println("Parsed Token String:", tokenString)
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return GetJWTSecret(), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid or expired token",
			})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid token claims format",
			})
			c.Abort()
			return
		}

		// Extract claims from MapClaims
		userID, ok1 := getStringClaim(claims, "user_id")
		email, ok2 := getStringClaim(claims, "email")
		role, ok3 := getStringClaim(claims, "role")
		fullName, _ := getStringClaim(claims, "full_name")
		tenantID, ok4 := getStringClaim(claims, "tenant_id")
		teamID, ok5 := getStringClaim(claims, "team_id")

		if !ok1 || !ok2 || !ok3 || !ok4 || !ok5 || userID == "" || email == "" || role == "" || tenantID == "" || teamID == "" {
			c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
				Error:   "unauthorized",
				Message: "Missing required token claims",
			})
			c.Abort()
			return
		}

		// Attach claims to context
		c.Set("userID", userID)
		c.Set("email", email)
		c.Set("userRole", role)
		c.Set("fullName", fullName)
		c.Set("tenantID", tenantID)
		c.Set("teamID", teamID)

		c.Next()
	}
}
func getStringClaim(claims jwt.MapClaims, key string) (string, bool) {
	val, ok := claims[key]
	if !ok {
		return "", false
	}
	str, ok := val.(string)
	return str, ok
}

func WebSocketAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check token from query param
		tokenString := c.Query("token")
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, structs.ErrorResponse{
				Error:   "unauthorized",
				Message: "Missing token in query string",
			})
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return GetJWTSecret(), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, structs.ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid or expired token",
			})
			return
		}

		claims, ok := token.Claims.(*Claims)
		if !ok || claims.UserID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, structs.ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid token claims",
			})
			return
		}

		// ✅ Inject claims into context
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("userRole", claims.Role)
		c.Set("fullName", claims.FullName)

		c.Next()
	}
}

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
		for _, allowed := range allowedRoles {
			if role == allowed {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, structs.ErrorResponse{
			Error:   "forbidden",
			Message: "Insufficient permissions",
		})
		c.Abort()
	}
}

func GetTargetUserID(c *gin.Context) (string, bool) {
	targetUserID := c.Param("user_id")
	role, _ := c.Get("userRole")

	if targetUserID != "" && role == "Admin" {
		return targetUserID, true
	}

	currUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return "", false
	}

	return currUserID.(string), true
}
