package routes

import (
	"net/http"
	"time"

	"aegis-api/handlers"
	"aegis-api/middleware"
	"aegis-api/services_/annotation_threads/messages"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetUpRouter(h *handlers.Handler) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// ─── CORS Config ──────────────────────────────────────
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://127.0.0.1:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// ─── Health Check ─────────────────────────────────────
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// ─── API Routes ───────────────────────────────────────
	api := router.Group("/api/v1")
	{
		// ─── Auth ─────────────────────────────────────────
		auth := api.Group("/auth")
		{
			auth.POST("/login", h.AuthService.LoginHandler)
			auth.POST("/request-password-reset", h.AuthService.RequestPasswordReset)
			auth.POST("/reset-password", h.AuthService.ResetPasswordHandler)
			auth.GET("/verify", h.AdminService.VerifyEmail)
		}

		// ─── Registration ────────────────────────────────
		api.POST("/register", h.AdminService.RegisterUser)

		// ─── Public Evidence Upload/Download ─────────────
		api.POST("/upload", h.UploadHandler.Upload)
		api.GET("/download/:cid", h.DownloadHandler.Download)

		// ─── Protected Routes ────────────────────────────
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			// ─── Case Management ──────────────────────────
			protected.POST("/cases", h.CaseHandler.CreateCase)
			protected.GET("/cases/active", h.CaseHandler.ListActiveCasesHandler)
			protected.POST("/cases/assign", h.CaseHandler.AssignUserToCase)

			// ─── New List / Filter Cases ──────────────────
			protected.GET("/cases/all", h.CaseHandler.GetAllCasesHandler)
			protected.GET("/cases/user/:user_id", h.CaseHandler.GetCasesByUserHandler)
			protected.GET("/cases/filter", h.CaseHandler.GetFilteredCasesHandler)

			// ─── Metadata Evidence Upload ────────────────
			protected.POST("/evidence", h.MetadataHandler.UploadEvidence)

			// ─── Admin: Users ────────────────────────────
			protected.GET("/users", h.AdminService.ListUsers)

			// ─── Thread Messaging ────────────────────────
			RegisterMessageRoutes(protected, h.MessageService)
		}
	}

	return router
}

func RegisterMessageRoutes(r *gin.RouterGroup, svc messages.MessageService) {
	h := handlers.NewMessageHandler(svc)

	r.POST("/threads/:threadID/messages", h.SendMessage)
	r.GET("/threads/:threadID/messages", h.GetMessagesByThread)
	r.POST("/messages/:messageID/approve", h.ApproveMessage)
	r.POST("/messages/:messageID/reactions", h.AddReaction)
	r.DELETE("/messages/:messageID/reactions/:userID", h.RemoveReaction)
}
