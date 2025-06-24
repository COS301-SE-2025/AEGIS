package routes

import (
	"net/http"
	"time"

	"aegis-api/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetUpRouter configures Gin routes with CORS and login/registration handlers
func SetUpRouter(h *handlers.Handler) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// ─── CORS Config ─────────────────────────────
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5173",
			"http://127.0.0.1:5173",
		},

		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// ─── Health Check ────────────────────────────
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// ─── API v1 ──────────────────────────────────
	api := router.Group("/api/v1")
	{
		// Ping again under /api/v1
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "pong"})
		})

		// ─── Public Authentication Routes ──────────────
		auth := api.Group("/auth")
		{
			auth.POST("/login", h.AuthService.LoginHandler)
			auth.POST("/request-password-reset", h.AuthService.RequestPasswordReset)
			auth.POST("/reset-password", h.AuthService.ResetPasswordHandler)
			auth.POST("/resend-verification", h.AdminService.RegisterUser) // Optional: if AdminService handles resend
			auth.GET("/verify", h.AdminService.VerifyEmail)
		}

		// ─── Public Registration Route ────────────────
		api.POST("/register", h.AdminService.RegisterUser)
	}

	return router
}
