package middleware

import (
	"aegis-api/structs"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

// Granular limit config: map[method][path]limit
type EndpointLimitConfig map[string]map[string]int

// IPThrottleMiddleware applies rate limiting based on client IP, endpoint, and method
func IPThrottleMiddleware(defaultLimit int, window time.Duration, config EndpointLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		method := c.Request.Method
		path := c.FullPath()
		limit := defaultLimit
		if config != nil {
			if m, ok := config[method]; ok {
				if l, ok := m[path]; ok {
					limit = l
				}
			}
		}
		key := fmt.Sprintf("ip_rate_limit:%s:%s:%s", ip, method, path)
		count, err := RedisClient.Incr(ctx, key).Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
				Error:   "server_error",
				Message: "Could not check IP rate limit",
			})
			c.Abort()
			return
		}
		if count == 1 {
			RedisClient.Expire(ctx, key, window)
		}
		if count > int64(limit) {
			c.JSON(http.StatusTooManyRequests, structs.ErrorResponse{
				Error:   "rate_limited",
				Message: "Too many requests from this IP, slow down",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RedisClientInterface allows mocking for tests
type RedisClientInterface interface {
	Incr(ctx context.Context, key string) *redis.IntCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
}

var ctx = context.Background()

// Setup Redis client (you can move this into a config/init file)
var RedisClient RedisClientInterface = redis.NewClient(&redis.Options{
	Addr:     getRedisAddr(),
	Password: "", // set if needed
	DB:       0,
})

func getRedisAddr() string {
	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		return addr
	}
	return "localhost:6379"
}

// RateLimitMiddleware with granular endpoint/method limits
func RateLimitMiddleware(defaultLimit int, window time.Duration, config EndpointLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
				Error:   "unauthorized",
				Message: "User not authenticated",
			})
			c.Abort()
			return
		}
		method := c.Request.Method
		path := c.FullPath()
		roleLimit := defaultLimit
		roleName := "user"
		if role, ok := c.Get("role"); ok {
			roleName, _ = role.(string)
		}
		// Tenant Admins/DFIR Admins get higher limits
		if roleName == "Tenant Admin" || roleName == "DFIR Admin" {
			roleLimit = defaultLimit * 5
		}
		// Granular override
		if config != nil {
			if m, ok := config[method]; ok {
				if l, ok := m[path]; ok {
					roleLimit = l
				}
			}
		}
		key := fmt.Sprintf("rate_limit:%s:%s:%s", userID, method, path)
		count, err := RedisClient.Incr(ctx, key).Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
				Error:   "server_error",
				Message: "Could not check rate limit",
			})
			c.Abort()
			return
		}
		if count == 1 {
			RedisClient.Expire(ctx, key, window)
		}
		if count > int64(roleLimit) {
			c.JSON(http.StatusTooManyRequests, structs.ErrorResponse{
				Error:   "rate_limited",
				Message: fmt.Sprintf("Too many requests, slow down (role: %s)", roleName),
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

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

		// Extract string claims
		userID, ok1 := getStringClaim(claims, "user_id")
		email, ok2 := getStringClaim(claims, "email")
		role, ok3 := getStringClaim(claims, "role")
		fullName, _ := getStringClaim(claims, "full_name")
		tenantID, ok4 := getStringClaim(claims, "tenant_id")

		teamID, _ := getStringClaim(claims, "team_id")

		if !ok1 || !ok2 || !ok3 || !ok4 || userID == "" || email == "" || role == "" || tenantID == "" {
			c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
				Error:   "unauthorized",
				Message: "Missing required token claims",
			})
			c.Abort()
			return
		}

		// Only enforce teamID for roles that must belong to a team
		if teamID == "" && (role == "DFIR Admin" || role == "DFIR User") {
			c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
				Error:   "unauthorized",
				Message: "Team ID required for this role",
			})
			c.Abort()
			return
		}

		// ✅ Attach claims to context
		c.Set("userID", userID)
		c.Set("email", email)
		c.Set("userRole", role)
		c.Set("fullName", fullName)
		c.Set("tenantID", tenantID)
		c.Set("teamID", teamID) // may be empty for Tenant Admin

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
		// Extract token from query parameter
		tokenString := c.Query("token")
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, structs.ErrorResponse{
				Error:   "unauthorized",
				Message: "Missing token in query string",
			})
			return
		}

		// Parse the JWT with custom claims
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return GetJWTSecret(), nil
		})

		// Check token validity
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, structs.ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid or expired token",
			})
			return
		}

		// Type assert custom claims
		claims, ok := token.Claims.(*Claims)
		if !ok || claims.UserID == "" || claims.TenantID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, structs.ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid token claims",
			})
			return
		}

		// ✅ Inject user data into Gin context
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("userRole", claims.Role)
		c.Set("fullName", claims.FullName)
		c.Set("tenantID", claims.TenantID)
		c.Set("teamID", claims.TeamID)

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
