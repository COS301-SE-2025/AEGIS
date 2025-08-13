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

	// Serve /uploads as static file directory
	router.Static("/uploads", "/app/uploads")

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
			auth.POST("/accept-terms", h.AdminService.AcceptTerms)
		}

		// ─── Registration ────────────────────────────────
		api.POST("/register", middleware.AuthMiddleware(), middleware.RequireRole("DFIR Admin"), h.AdminService.RegisterUser)
		api.POST("/register/tenant", h.AdminService.RegisterTenantUser)
		api.POST("/register/team", middleware.AuthMiddleware(), middleware.RequireRole("Tenant Admin"), h.AdminService.RegisterTeamUser)
		api.POST("/tenant", h.AdminService.CreateTenant)
		api.POST("/team", h.AdminService.CreateTeam)
		api.GET("/teams/:id", h.GetTeamByID)
		api.GET("/teams", h.GetTeamsByTenant)
		api.GET("/tenants", h.GetAllTenants)

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
			protected.POST("/cases/assign", middleware.AuthMiddleware(), h.CaseHandler.AssignUserToCase)
			protected.GET("/cases/:case_id/collaborators", h.GetCollaboratorsHandler.GetCollaboratorsByCaseID)
			protected.POST("/cases/unassign", h.CaseHandler.UnassignUserFromCase)
			protected.GET("/cases/closed", h.CaseHandler.ListClosedCasesHandler)
			protected.PATCH("/cases/:case_id", h.CaseHandler.UpdateCaseHandler)

			// ─── New List / Filter Cases ──────────────────
			protected.GET("/cases/all", h.CaseHandler.GetAllCasesHandler)
			protected.GET("/cases/user/:user_id", h.CaseHandler.GetCasesByUserHandler)
			protected.GET("/cases/filter", h.CaseHandler.GetFilteredCasesHandler)
			protected.GET("/cases/:case_id", h.CaseHandler.GetCaseByIDHandler)

			// Graphical Mapping - IOC Graph endpoints
			protected.GET("/tenants/:tenantId/ioc-graph", middleware.AuthMiddleware(), h.IOCHandler.GetTenantIOCGraph)
			protected.GET("/tenants/:tenantId/cases/:case_id/ioc-graph", middleware.AuthMiddleware(), h.IOCHandler.GetCaseIOCGraph)

			protected.GET("/cases/:case_id/iocs", middleware.AuthMiddleware(), h.IOCHandler.GetIOCsByCase)
			protected.POST("/cases/:case_id/iocs", middleware.AuthMiddleware(), h.IOCHandler.AddIOCToCase)
			// ______timeline routes______________
			// List all events for a case
			protected.GET("/cases/:case_id/timeline", middleware.AuthMiddleware(), h.TimelineHandler.ListByCase)
			// Create new event for a case
			protected.POST("/cases/:case_id/timeline", middleware.AuthMiddleware(), h.TimelineHandler.Create)
			// Update a timeline event by ID
			protected.PATCH("/timeline/:event_id", middleware.AuthMiddleware(), h.TimelineHandler.Update)
			// Delete a timeline event by ID
			protected.DELETE("/timeline/:event_id", middleware.AuthMiddleware(), h.TimelineHandler.Delete)
			// Reorder events for a case
			protected.POST("/cases/:case_id/timeline/reorder", h.TimelineHandler.Reorder)

			// ─── Metadata Evidence Upload ────────────────
			protected.POST("/evidence", h.MetadataHandler.UploadEvidence)
			// ─── Metadata Evidence Retrieval ─────────────
			protected.GET("/evidence-metadata/:id", h.MetadataHandler.GetEvidenceByID)
			protected.GET("/evidence-metadata/case/:case_id", h.MetadataHandler.GetEvidenceByCaseID)

			// ─── Admin: Users ────────────────────────────
			protected.GET("/users", h.AdminService.ListUsers)

			// ─── Profile Routes ──────────────────────────
			protected.GET("/profile/:userID", h.ProfileHandler.GetProfileHandler)
			protected.PATCH("/profile", h.ProfileHandler.UpdateProfileHandler)

			// ─── case and evidence totals ──────────────────────────
			protected.GET("/dashboard/totals", h.CaseEvidenceTotalsHandler.GetDashboardTotals)
			// ─── Recent Activities ───────────────────────────────
			protected.GET("/auditlogs/recent/:userId", h.RecentActivityHandler.GetRecentActivities)

			// ─── Notification Routes ──────────────────────────────
			protected.GET("/notifications", h.GetNotifications)
			protected.POST("/notifications/read", h.MarkNotificationsRead)
			protected.DELETE("/notifications/delete", h.DeleteNotifications)
			protected.POST("/notifications/archive", h.ArchiveNotifications)

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

		}
	}

	return router
}
