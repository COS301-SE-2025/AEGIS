package unit_tests

import (
	"aegis-api/middleware"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestRateLimitMiddleware_RoleBasedLimits(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	limit := 2
	window := time.Second
	// Simulate Tenant Admin (should get higher limit)
	router.Use(func(c *gin.Context) {
		c.Set("userID", "adminuser")
		c.Set("role", "Tenant Admin")
		c.Next()
	})
	router.Use(middleware.RateLimitMiddleware(limit, window, nil))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})
	req, _ := http.NewRequest("GET", "/test", nil)
	// Should allow 10 requests (limit * 5)
	for i := 0; i < limit*5; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, 429, w.Code)
}

func TestRateLimitMiddleware_NormalUserStrictLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	limit := 2
	window := time.Second
	// Simulate normal user
	router.Use(func(c *gin.Context) {
		c.Set("userID", "normaluser")
		c.Set("role", "user")
		c.Next()
	})
	router.Use(middleware.RateLimitMiddleware(limit, window, nil))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})
	req, _ := http.NewRequest("GET", "/test", nil)
	for i := 0; i < limit; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, 429, w.Code)
}

func TestIPThrottleMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	limit := 3
	window := time.Second
	router.Use(middleware.IPThrottleMiddleware(limit, window, nil))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "1.2.3.4:12345"
	for i := 0; i < limit; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, 429, w.Code)
}

// errorRedisClient mocks Redis and always returns an error
type errorRedisClient struct{}

// errorRedisClient mocks Redis and always returns an error

func (e *errorRedisClient) Incr(ctx context.Context, key string) *redis.IntCmd {
	cmd := redis.NewIntCmd(ctx)
	cmd.SetErr(errors.New("redis error"))
	return cmd
}

func (e *errorRedisClient) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	cmd := redis.NewBoolCmd(ctx)
	cmd.SetErr(errors.New("redis error"))
	return cmd
}

// Use the real RateLimitMiddleware from the middleware package

func TestRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	limit := 5
	window := time.Second
	router.Use(func(c *gin.Context) {
		c.Set("userID", "testuser")
		c.Next()
	})
	router.Use(middleware.RateLimitMiddleware(limit, window, nil))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)

	for i := 0; i < limit; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, 429, w.Code)

	time.Sleep(window)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}
func TestRateLimitMiddleware_MissingUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RateLimitMiddleware(5, time.Second, nil))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, 401, w.Code)
}

func TestRateLimitMiddleware_RedisFailure(t *testing.T) {
	// Save original client and replace with a mock that always errors
	orig := middleware.RedisClient
	middleware.RedisClient = &errorRedisClient{}
	defer func() { middleware.RedisClient = orig }()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", "testuser")
		c.Next()
	})
	router.Use(middleware.RateLimitMiddleware(5, time.Second, nil))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, 500, w.Code)
}

func TestRateLimitMiddleware_ConcurrentRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	limit := 5
	window := time.Second
	router.Use(func(c *gin.Context) {
		c.Set("userID", "concurrentuser")
		c.Next()
	})
	router.Use(middleware.RateLimitMiddleware(limit, window, nil))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})
	req, _ := http.NewRequest("GET", "/test", nil)

	results := make(chan int, limit+2)
	for i := 0; i < limit+2; i++ {
		go func() {
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			results <- w.Code
		}()
	}
	codes := []int{}
	for i := 0; i < limit+2; i++ {
		codes = append(codes, <-results)
	}
	success, throttled := 0, 0
	for _, code := range codes {
		if code == 200 {
			success++
		} else if code == 429 {
			throttled++
		}
	}
	assert.Equal(t, limit, success)
	assert.GreaterOrEqual(t, throttled, 1)
}
