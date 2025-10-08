// package middleware

// import (
// 	"aegis-api/structs"
// 	"context"
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"os"
// 	"strings"
// 	"time"

// 	"github.com/gin-gonic/gin"
// 	"github.com/golang-jwt/jwt/v5"
// 	"github.com/redis/go-redis/v9"
// )

// // Granular limit config: map[method][path]limit
// type EndpointLimitConfig map[string]map[string]int

// // IPThrottleMiddleware applies rate limiting based on client IP, endpoint, and method
// func IPThrottleMiddleware(defaultLimit int, window time.Duration, config EndpointLimitConfig) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		ip := c.ClientIP()
// 		method := c.Request.Method
// 		path := c.FullPath()
// 		limit := defaultLimit
// 		if config != nil {
// 			if m, ok := config[method]; ok {
// 				if l, ok := m[path]; ok {
// 					limit = l
// 				}
// 			}
// 		}
// 		key := fmt.Sprintf("ip_sliding_window:%s:%s:%s", ip, method, path)
// 		now := time.Now().Unix()
// 		windowSec := int64(window.Seconds())
// 		// Remove old timestamps
// 		if err := RedisClient.ZRemRangeByScore(ctx, key, "-inf", fmt.Sprintf("%d", now-windowSec)).Err(); err != nil {
// 			c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 				Error:   "internal_error",
// 				Message: "Redis error",
// 			})
// 			c.Abort()
// 			return
// 		}
// 		// Count requests in window
// 		count, err := RedisClient.ZCard(ctx, key).Result()
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 				Error:   "internal_error",
// 				Message: "Redis error",
// 			})
// 			c.Abort()
// 			return
// 		}
// 		if int(count) >= limit {
// 			fmt.Fprintf(os.Stderr, "[THROTTLE] IP %s hit limit for %s %s at %v\n", ip, method, path, now)
// 			c.JSON(http.StatusTooManyRequests, structs.ErrorResponse{
// 				Error:   "rate_limited",
// 				Message: "Too many requests from this IP, slow down",
// 			})
// 			c.Abort()
// 			return
// 		}
// 		// Add current timestamp
// 		if err := RedisClient.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now}).Err(); err != nil {
// 			c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 				Error:   "internal_error",
// 				Message: "Redis error",
// 			})
// 			c.Abort()
// 			return
// 		}
// 		// Set expiry
// 		RedisClient.Expire(ctx, key, window)
// 		c.Next()
// 	}
// }

// // SlidingWindowRedisClient allows mocking for tests (includes sorted set ops)
// type SlidingWindowRedisClient interface {
// 	ZAdd(ctx context.Context, key string, members ...redis.Z) *redis.IntCmd
// 	ZCard(ctx context.Context, key string) *redis.IntCmd
// 	ZRemRangeByScore(ctx context.Context, key, min, max string) *redis.IntCmd
// 	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
// }

// var ctx = context.Background()

// // Setup Redis client (you can move this into a config/init file)
// var RedisClient SlidingWindowRedisClient = redis.NewClient(&redis.Options{
// 	Addr:     getRedisAddr(),
// 	Password: "", // set if needed
// 	DB:       0,
// })

// func getRedisAddr() string {
// 	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
// 		return addr
// 	}
// 	return "localhost:6379"
// }

// // RateLimitMiddleware with granular endpoint/method limits
// func RateLimitMiddleware(defaultLimit int, window time.Duration, config EndpointLimitConfig) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		userID, exists := c.Get("userID")
// 		tenantID, tenantExists := c.Get("tenantID")
// 		if !exists || !tenantExists {
// 			c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
// 				Error:   "unauthorized",
// 				Message: "User or tenant not authenticated",
// 			})
// 			c.Abort()
// 			return
// 		}
// 		method := c.Request.Method
// 		path := c.FullPath()
// 		roleLimit := defaultLimit
// 		roleName := "user"
// 		if role, ok := c.Get("role"); ok {
// 			roleName, _ = role.(string)
// 		}
// 		// Tenant Admins/DFIR Admins get higher limits
// 		if roleName == "Tenant Admin" || roleName == "DFIR Admin" {
// 			roleLimit = defaultLimit * 5
// 		}
// 		// Granular override
// 		if config != nil {
// 			if m, ok := config[method]; ok {
// 				if l, ok := m[path]; ok {
// 					roleLimit = l
// 				}
// 			}
// 		}
// 		// Sliding window for user
// 		userKey := fmt.Sprintf("user_sliding_window:%s:%s:%s", userID, method, path)
// 		tenantKey := fmt.Sprintf("tenant_sliding_window:%s:%s:%s", tenantID, method, path)
// 		now := time.Now().Unix()
// 		windowSec := int64(window.Seconds())
// 		// Remove old timestamps
// 		if err := RedisClient.ZRemRangeByScore(ctx, userKey, "-inf", fmt.Sprintf("%d", now-windowSec)).Err(); err != nil {
// 			c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 				Error:   "internal_error",
// 				Message: "Redis error",
// 			})
// 			c.Abort()
// 			return
// 		}
// 		if err := RedisClient.ZRemRangeByScore(ctx, tenantKey, "-inf", fmt.Sprintf("%d", now-windowSec)).Err(); err != nil {
// 			c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 				Error:   "internal_error",
// 				Message: "Redis error",
// 			})
// 			c.Abort()
// 			return
// 		}
// 		// Count requests in window
// 		userCount, err := RedisClient.ZCard(ctx, userKey).Result()
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 				Error:   "internal_error",
// 				Message: "Redis error",
// 			})
// 			c.Abort()
// 			return
// 		}
// 		tenantCount, err := RedisClient.ZCard(ctx, tenantKey).Result()
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 				Error:   "internal_error",
// 				Message: "Redis error",
// 			})
// 			c.Abort()
// 			return
// 		}
// 		if int(userCount) >= roleLimit || int(tenantCount) >= roleLimit*20 {
// 			fmt.Fprintf(os.Stderr, "[THROTTLE] User %v or Tenant %v hit limit for %s %s at %v\n", userID, tenantID, method, path, now)
// 			c.JSON(http.StatusTooManyRequests, structs.ErrorResponse{
// 				Error:   "rate_limited",
// 				Message: fmt.Sprintf("Too many requests, slow down (role: %s)", roleName),
// 			})
// 			c.Abort()
// 			return
// 		}
// 		// Add current timestamp
// 		if err := RedisClient.ZAdd(ctx, userKey, redis.Z{Score: float64(now), Member: now}).Err(); err != nil {
// 			c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 				Error:   "internal_error",
// 				Message: "Redis error",
// 			})
// 			c.Abort()
// 			return
// 		}
// 		if err := RedisClient.ZAdd(ctx, tenantKey, redis.Z{Score: float64(now), Member: now}).Err(); err != nil {
// 			c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 				Error:   "internal_error",
// 				Message: "Redis error",
// 			})
// 			c.Abort()
// 			return
// 		}
// 		// Set expiry
// 		RedisClient.Expire(ctx, userKey, window)
// 		RedisClient.Expire(ctx, tenantKey, window)
// 		c.Next()
// 	}
// }

// func AuthMiddleware() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		fmt.Println("Authorization Header:", c.GetHeader("Authorization"))
// 		authHeader := c.GetHeader("Authorization")

// 		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
// 			c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
// 				Error:   "unauthorized",
// 				Message: "Authorization token required",
// 			})
// 			c.Abort()
// 			return
// 		}

// 		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
// 		fmt.Println("Parsed Token String:", tokenString)

// 		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
// 			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
// 				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
// 			}
// 			return GetJWTSecret(), nil
// 		})

// 		if err != nil || !token.Valid {
// 			c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
// 				Error:   "unauthorized",
// 				Message: "Invalid or expired token",
// 			})
// 			c.Abort()
// 			return
// 		}

// 		claims, ok := token.Claims.(jwt.MapClaims)
// 		if !ok {
// 			c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
// 				Error:   "unauthorized",
// 				Message: "Invalid token claims format",
// 			})
// 			c.Abort()
// 			return
// 		}

// 		// Extract string claims
// 		userID, ok1 := getStringClaim(claims, "user_id")
// 		email, ok2 := getStringClaim(claims, "email")
// 		role, ok3 := getStringClaim(claims, "role")
// 		fullName, _ := getStringClaim(claims, "full_name")
// 		tenantID, ok4 := getStringClaim(claims, "tenant_id")

// 		teamID, _ := getStringClaim(claims, "team_id")

// 		if !ok1 || !ok2 || !ok3 || !ok4 || userID == "" || email == "" || role == "" || tenantID == "" {
// 			c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
// 				Error:   "unauthorized",
// 				Message: "Missing required token claims",
// 			})
// 			c.Abort()
// 			return
// 		}

// 		// Only enforce teamID for roles that must belong to a team
// 		if teamID == "" && (role == "DFIR Admin" || role == "DFIR User") {
// 			c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
// 				Error:   "unauthorized",
// 				Message: "Team ID required for this role",
// 			})
// 			c.Abort()
// 			return
// 		}

// 		// ✅ Attach claims to context
// 		c.Set("userID", userID)
// 		c.Set("email", email)
// 		c.Set("userRole", role)
// 		c.Set("fullName", fullName)
// 		c.Set("tenantID", tenantID)
// 		c.Set("teamID", teamID) // may be empty for Tenant Admin

// 		c.Next()
// 	}
// }
// func getStringClaim(claims jwt.MapClaims, key string) (string, bool) {
// 	val, ok := claims[key]
// 	if !ok {
// 		return "", false
// 	}
// 	str, ok := val.(string)
// 	return str, ok
// }

// func WebSocketAuthMiddleware() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		// ✅ FIRST check query parameter (for WebSockets)
// 		tokenString := c.Query("token")

// 		// ✅ SECOND check Authorization header (for HTTP fallback)
// 		if tokenString == "" {
// 			authHeader := c.GetHeader("Authorization")
// 			if strings.HasPrefix(authHeader, "Bearer ") {
// 				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
// 			}
// 		}

// 		// Log what we found
// 		log.Printf("🔐 Auth Middleware - Token from query: %t, Token from header: %t",
// 			c.Query("token") != "",
// 			c.GetHeader("Authorization") != "")

// 		if tokenString == "" {
// 			log.Println("❌ WebSocket auth failed: Missing token in both query and header")
// 			c.AbortWithStatusJSON(http.StatusUnauthorized, structs.ErrorResponse{
// 				Error:   "unauthorized",
// 				Message: "Missing token in query string or Authorization header",
// 			})
// 			return
// 		}

// 		log.Printf("🔐 Token found: %s...", tokenString[:min(20, len(tokenString))])

// 		// Parse the JWT with custom claims
// 		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
// 			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
// 				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
// 			}
// 			return GetJWTSecret(), nil
// 		})

// 		// Check token validity
// 		if err != nil || !token.Valid {
// 			log.Printf("❌ WebSocket auth failed: Invalid token - %v", err)
// 			c.AbortWithStatusJSON(http.StatusUnauthorized, structs.ErrorResponse{
// 				Error:   "unauthorized",
// 				Message: "Invalid or expired token",
// 			})
// 			return
// 		}

// 		// Type assert custom claims
// 		claims, ok := token.Claims.(*Claims)
// 		if !ok || claims.UserID == "" || claims.TenantID == "" {
// 			log.Println("❌ WebSocket auth failed: Invalid token claims")
// 			c.AbortWithStatusJSON(http.StatusUnauthorized, structs.ErrorResponse{
// 				Error:   "unauthorized",
// 				Message: "Invalid token claims",
// 			})
// 			return
// 		}

// 		// ✅ Inject user data into Gin context
// 		c.Set("userID", claims.UserID)
// 		c.Set("email", claims.Email)
// 		c.Set("userRole", claims.Role)
// 		c.Set("fullName", claims.FullName)
// 		c.Set("tenantID", claims.TenantID)
// 		c.Set("teamID", claims.TeamID)

// 		log.Printf("✅ WebSocket auth successful for user %s", claims.UserID)
// 		c.Next()
// 	}
// }

// func min(a, b int) int {
// 	if a < b {
// 		return a
// 	}
// 	return b
// }
// func RequireRole(allowedRoles ...string) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		userRole, exists := c.Get("userRole")
// 		if !exists {
// 			c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
// 				Error:   "unauthorized",
// 				Message: "Authentication required",
// 			})
// 			c.Abort()
// 			return
// 		}

// 		role := userRole.(string)
// 		for _, allowed := range allowedRoles {
// 			if role == allowed {
// 				c.Next()
// 				return
// 			}
// 		}

// 		c.JSON(http.StatusForbidden, structs.ErrorResponse{
// 			Error:   "forbidden",
// 			Message: "Insufficient permissions",
// 		})
// 		c.Abort()
// 	}
// }

// func GetTargetUserID(c *gin.Context) (string, bool) {
// 	targetUserID := c.Param("user_id")
// 	role, _ := c.Get("userRole")

// 	if targetUserID != "" && role == "Admin" {
// 		return targetUserID, true
// 	}

// 	currUserID, exists := c.Get("userID")
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
// 			Error:   "unauthorized",
// 			Message: "User not authenticated",
// 		})
// 		return "", false
// 	}

// 	return currUserID.(string), true
// }

package middleware

import (
	"aegis-api/structs"
	"context"
	"fmt"
	"log"
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

// SlidingWindowRedisClient allows mocking for tests (includes sorted set ops)
type SlidingWindowRedisClient interface {
	ZAdd(ctx context.Context, key string, members ...redis.Z) *redis.IntCmd
	ZCard(ctx context.Context, key string) *redis.IntCmd
	ZRemRangeByScore(ctx context.Context, key, min, max string) *redis.IntCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
}

var ctx = context.Background()

// Global Redis client that will be set by main.go
var RedisClient SlidingWindowRedisClient

// SetRedisClient sets the shared Redis client for rate limiting
func SetRedisClient(client *redis.Client) {
	RedisClient = client
}

// IPThrottleMiddleware applies rate limiting based on client IP, endpoint, and method
func IPThrottleMiddleware(defaultLimit int, window time.Duration, config EndpointLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if RedisClient == nil {
			// If Redis is not available, allow the request to proceed
			log.Println("⚠️ Redis client not available, skipping IP throttling")
			c.Next()
			return
		}

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
		key := fmt.Sprintf("ip_sliding_window:%s:%s:%s", ip, method, path)
		now := time.Now().Unix()
		windowSec := int64(window.Seconds())

		// Remove old timestamps
		if err := RedisClient.ZRemRangeByScore(ctx, key, "-inf", fmt.Sprintf("%d", now-windowSec)).Err(); err != nil {
			log.Printf("Redis error in IP throttle: %v", err)
			c.Next() // Continue without rate limiting on Redis error
			return
		}

		// Count requests in window
		count, err := RedisClient.ZCard(ctx, key).Result()
		if err != nil {
			log.Printf("Redis error in IP throttle: %v", err)
			c.Next() // Continue without rate limiting on Redis error
			return
		}

		if int(count) >= limit {
			fmt.Fprintf(os.Stderr, "[THROTTLE] IP %s hit limit for %s %s at %v\n", ip, method, path, now)
			c.JSON(http.StatusTooManyRequests, structs.ErrorResponse{
				Error:   "rate_limited",
				Message: "Too many requests from this IP, slow down",
			})
			c.Abort()
			return
		}

		// Add current timestamp
		if err := RedisClient.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now}).Err(); err != nil {
			log.Printf("Redis error in IP throttle: %v", err)
			c.Next() // Continue without rate limiting on Redis error
			return
		}

		// Set expiry
		RedisClient.Expire(ctx, key, window)
		c.Next()
	}
}

// RateLimitMiddleware with granular endpoint/method limits
func RateLimitMiddleware(defaultLimit int, window time.Duration, config EndpointLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if RedisClient == nil {
			log.Println("⚠️ Redis client not available, skipping rate limiting")
			c.Next()
			return
		}

		userID, exists := c.Get("userID")
		tenantID, tenantExists := c.Get("tenantID")
		if !exists || !tenantExists {
			c.JSON(http.StatusServiceUnavailable, structs.ErrorResponse{
				Error:   "service_unavailable",
				Message: "User or tenant not authenticated",
			})
			c.Abort()
			return
		}

		method := c.Request.Method
		path := c.FullPath()
		roleLimit := defaultLimit
		roleName := "user"

		// Fix: Check for both "role" and "userRole" keys
		if role, ok := c.Get("role"); ok {
			roleName, _ = role.(string)
		} else if role, ok := c.Get("userRole"); ok {
			roleName, _ = role.(string)
		}

		// Apply role-based multipliers
		switch roleName {
		case "Tenant Admin":
			roleLimit = defaultLimit * 5
		case "Admin":
			roleLimit = defaultLimit * 3
		case "DFIR Admin":
			roleLimit = defaultLimit * 5
		default:
			roleLimit = defaultLimit
		}

		// Granular override
		if config != nil {
			if m, ok := config[method]; ok {
				if l, ok := m[path]; ok {
					roleLimit = l
				}
			}
		}

		// Use consistent key format - simplified for rate limiting
		key := fmt.Sprintf("rate_limit:%s:%s", userID, tenantID)
		now := time.Now().Unix()
		windowSec := int64(window.Seconds())

		// Remove old timestamps
		if err := RedisClient.ZRemRangeByScore(ctx, key, "-inf", fmt.Sprintf("%d", now-windowSec)).Err(); err != nil {
			log.Printf("Redis error in rate limiting: %v", err)
			c.Next()
			return
		}

		// Count requests in window
		count, err := RedisClient.ZCard(ctx, key).Result()
		if err != nil {
			log.Printf("Redis error in rate limiting: %v", err)
			c.Next()
			return
		}

		if int(count) >= roleLimit {
			fmt.Fprintf(os.Stderr, "[THROTTLE] User %v hit limit for %s %s at %v (role: %s, limit: %d)\n",
				userID, method, path, now, roleName, roleLimit)
			c.JSON(http.StatusTooManyRequests, structs.ErrorResponse{
				Error:   "rate_limited",
				Message: fmt.Sprintf("Too many requests, slow down (role: %s)", roleName),
			})
			c.Abort()
			return
		}

		// Add current timestamp with unique member to avoid duplicates
		member := fmt.Sprintf("%d_%d", now, time.Now().UnixNano())
		if err := RedisClient.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: member}).Err(); err != nil {
			log.Printf("Redis error in rate limiting: %v", err)
			c.Next()
			return
		}

		// Set expiry
		RedisClient.Expire(ctx, key, window)
		c.Next()
	}
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("[DEBUG] AuthMiddleware: Authorization Header: %s", c.GetHeader("Authorization"))
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			log.Printf("[ERROR] AuthMiddleware: Missing or invalid Authorization header")
			c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
				Error:   "unauthorized",
				Message: "Authorization token required",
			})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		log.Printf("[DEBUG] AuthMiddleware: Parsed Token String: %s", tokenString)

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				log.Printf("[ERROR] AuthMiddleware: Unexpected signing method: %v", token.Header["alg"])
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return GetJWTSecret(), nil
		})

		if err != nil || !token.Valid {
			log.Printf("[ERROR] AuthMiddleware: Invalid or expired token: %v", err)
			c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid or expired token",
			})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			log.Printf("[ERROR] AuthMiddleware: Invalid token claims format")
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

		log.Printf("[DEBUG] AuthMiddleware: Claims extracted userID=%s email=%s role=%s tenantID=%s teamID=%s", userID, email, role, tenantID, teamID)
		if !ok1 || !ok2 || !ok3 || !ok4 || userID == "" || email == "" || role == "" || tenantID == "" {
			log.Printf("[ERROR] AuthMiddleware: Missing required token claims")
			c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
				Error:   "unauthorized",
				Message: "Missing required token claims",
			})
			c.Abort()
			return
		}

		// Only enforce teamID for roles that must belong to a team
		if teamID == "" && (role == "DFIR Admin" || role == "DFIR User") {
			log.Printf("[ERROR] AuthMiddleware: Team ID required for role %s", role)
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
		// ✅ FIRST check query parameter (for WebSockets)
		tokenString := c.Query("token")

		// ✅ SECOND check Authorization header (for HTTP fallback)
		if tokenString == "" {
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		// Log what we found
		log.Printf("🔍 Auth Middleware - Token from query: %t, Token from header: %t",
			c.Query("token") != "",
			c.GetHeader("Authorization") != "")

		if tokenString == "" {
			log.Println("❌ WebSocket auth failed: Missing token in both query and header")
			c.AbortWithStatusJSON(http.StatusUnauthorized, structs.ErrorResponse{
				Error:   "unauthorized",
				Message: "Missing token in query string or Authorization header",
			})
			return
		}

		log.Printf("🔍 Token found: %s...", tokenString[:min(20, len(tokenString))])

		// Parse the JWT with custom claims
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return GetJWTSecret(), nil
		})

		// Check token validity
		if err != nil || !token.Valid {
			log.Printf("❌ WebSocket auth failed: Invalid token - %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, structs.ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid or expired token",
			})
			return
		}

		// Type assert custom claims
		claims, ok := token.Claims.(*Claims)
		if !ok || claims.UserID == "" || claims.TenantID == "" {
			log.Println("❌ WebSocket auth failed: Invalid token claims")
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

		log.Printf("✅ WebSocket auth successful for user %s", claims.UserID)
		c.Next()
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("userRole")
		if !exists {
			log.Printf("[ERROR] RequireRole: userRole not found in context")
			c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
				Error:   "unauthorized",
				Message: "Authentication required",
			})
			c.Abort()
			return
		}

		role := userRole.(string)
		log.Printf("[DEBUG] RequireRole: userRole=%s allowedRoles=%v", role, allowedRoles)
		for _, allowed := range allowedRoles {
			if role == allowed {
				log.Printf("[DEBUG] RequireRole: role %s is allowed", role)
				c.Next()
				return
			}
		}

		log.Printf("[ERROR] RequireRole: role %s is not allowed", role)
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
