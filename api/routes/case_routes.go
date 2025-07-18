package routes

import (
	"net/http"
	"time"

	"aegis-api/handlers"
	"aegis-api/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	//"aegis-api/pkg/websocket"
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
			protected.GET("/cases/:case_id/collaborators", h.GetCollaboratorsHandler.GetCollaboratorsByCaseID)

			// ─── New List / Filter Cases ──────────────────
			protected.GET("/cases/all", h.CaseHandler.GetAllCasesHandler)
			protected.GET("/cases/user/:user_id", h.CaseHandler.GetCasesByUserHandler)
			protected.GET("/cases/filter", h.CaseHandler.GetFilteredCasesHandler)
			protected.GET("/cases/:case_id", h.CaseHandler.GetCaseByIDHandler)
			// ─── Metadata Evidence Upload ────────────────
			protected.POST("/evidence", h.MetadataHandler.UploadEvidence)

			// ─── Admin: Users ────────────────────────────
			protected.GET("/users", h.AdminService.ListUsers)

			// ─── Profile Routes ──────────────────────────
			protected.GET("/profile/:userID", h.ProfileHandler.GetProfileHandler)
			protected.PATCH("/profile", h.ProfileHandler.UpdateProfileHandler)

			// RegisterMessageRoutes(protected, h.MessageService, auditLogger)
			// ─── Thread Messaging ────────────────────────

			RegisterMessageRoutes(protected, h.MessageHandler)
			// ─── Thread Routes ─────────────────────────────────────
			RegisterThreadRoutes(protected, h.AnnotationThreadHandler)

			// ─── Chat_message Routes ────────────────────────────────
			RegisterChatRoutes(protected, h.ChatHandler)

			// ─── Evidence Viewer + Tagging ────────────────
  			RegisterEvidenceRoutes(protected, h.EvidenceViewerHandler, h.EvidenceTagHandler, h.PermissionChecker)

			RegisterCaseTagRoutes(protected, h.CaseTagHandler, h.PermissionChecker)

			// protected.GET("/ws/cases/:case_id", func(c *gin.Context) {
			// caseID := c.Param("case_id")

			// // Extract userID from JWT claims in context
			// 	userIDVal, exists := c.Get("userID")
			// 	if !exists {
			// 		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			// 		return
			// 	}

			// 	userID := userIDVal.(string)

			// 	websocket.ServeWS(wsHub, c.Writer, c.Request, userID, caseID)
			// })



		}
	}

	return router
}
