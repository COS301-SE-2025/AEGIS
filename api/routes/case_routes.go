package routes

import (
	"time"

	"aegis-api/handlers"
	x3dh "aegis-api/internal/x3dh"
	"aegis-api/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetUpRouter(h *handlers.Handler) *gin.Engine {
	// Granular endpoint/method limit config
	granularLimits := middleware.EndpointLimitConfig{
		"POST": {
			"/api/v1/auth/login": 10,
			"/api/v1/register":   5,
			"/api/v1/upload":     8,
		},
		"GET": {
			"/api/v1/download/:cid": 20,
		},
	}
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://127.0.0.1:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	api := router.Group("/api/v1")
	// ─── Auth ─────────────────────────────────────────
	auth := api.Group("/auth")
	auth.Use(middleware.IPThrottleMiddleware(20, time.Minute, granularLimits)) // 20 req/min per IP for unauthenticated
	auth.POST("/login", h.AuthService.LoginHandler)
	auth.POST("/request-password-reset", h.AuthService.RequestPasswordReset)
	auth.POST("/reset-password", h.AuthService.ResetPasswordHandler)
	auth.GET("/verify", h.AdminService.VerifyEmail)
	auth.POST("/accept-terms", h.AdminService.AcceptTerms)

	// ─── Registration ────────────────────────────────
	api.POST("/register", middleware.AuthMiddleware(), middleware.IPThrottleMiddleware(20, time.Minute, granularLimits), h.AdminService.RegisterUser)
	api.POST("/register/tenant", middleware.IPThrottleMiddleware(20, time.Minute, granularLimits), h.AdminService.RegisterTenantUser)
	api.POST("/register/team", middleware.AuthMiddleware(), middleware.RequireRole("Tenant Admin"), h.AdminService.RegisterTeamUser)
	api.GET("/teams/:id", h.GetTeamByID)
	api.GET("/teams", h.GetTeamsByTenant)
	api.GET("/tenants", h.GetAllTenants)

	// ─── Public Evidence Upload/Download ─────────────
	api.POST("/upload", middleware.IPThrottleMiddleware(20, time.Minute, granularLimits), h.UploadHandler.Upload)
	api.GET("/download/:id", middleware.IPThrottleMiddleware(20, time.Minute, granularLimits), h.DownloadHandler.Download)

	//________AI Routes________
	timelineAIGroup := api.Group("/ai")
	{

		timelineAIGroup.POST("/suggestions", middleware.AuthMiddleware(), h.TimelineAIHandler.GetEventSuggestions)
		timelineAIGroup.POST("/severity", middleware.AuthMiddleware(), h.TimelineAIHandler.GetSeverityRecommendation)
		timelineAIGroup.POST("/tags", middleware.AuthMiddleware(), h.TimelineAIHandler.GetTagSuggestions)
		timelineAIGroup.POST("/iocs", middleware.AuthMiddleware(), h.TimelineAIHandler.ExtractIOCs)
		timelineAIGroup.POST("/analyze-event", middleware.AuthMiddleware(), h.TimelineAIHandler.AnalyzeEvent)
		timelineAIGroup.GET("/cases/:case_id/next-steps", middleware.AuthMiddleware(), h.TimelineAIHandler.GetNextSteps)
		timelineAIGroup.GET("/analyze-event", middleware.AuthMiddleware(), h.TimelineAIHandler.AnalyzeEvent)
		timelineAIGroup.POST("/correlate-evidence", middleware.AuthMiddleware(), h.TimelineAIHandler.CorrelateEvidence)
		//timelineAIGroup.POST("/feedback", middleware.AuthMiddleware(), h.TimelineAIHandler.SubmitFeedback)
		//timelineAIGroup.GET("/model-status", middleware.AuthMiddleware(), h.TimelineAIHandler.GetModelStatus)
		//timelineAIGroup.POST("/update-model-config", middleware.AuthMiddleware(), middleware.RequireRole("DFIR Admin"), h.TimelineAIHandler.UpdateModelConfig)

	}

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
		// Archive case (move to archived tab)
		protected.PATCH("/cases/:case_id/archive", h.CaseDeletionHandler.ArchiveCaseHandler)
		// List archived cases
		protected.GET("/cases/archived", h.CaseHandler.ListArchivedCasesHandler)
		protected.POST("/auth/verify-admin", h.VerificationHandler.VerifyAdminGin) // Move here
		protected.POST("/auth/logout", h.AuthService.LogoutHandler)
		protected.POST("/auth/change-password", h.AuthService.ChangePasswordHandler)
		// ─── New List / Filter Cases ──────────────────
		protected.GET("/cases/all", h.CaseHandler.GetAllCasesHandler)
		protected.GET("/cases/user/:user_id", h.CaseHandler.GetCasesByUserHandler)
		protected.GET("/cases/filter", h.CaseHandler.GetFilteredCasesHandler)
		protected.GET("/cases/:case_id", h.CaseHandler.GetCaseByIDHandler)

		protected.GET("/tenants/:tenantId/cases/:case_id/ioc-graph", middleware.AuthMiddleware(), h.IOCHandler.GetCaseIOCGraph)
		protected.GET("tenants/:tenantId/ioc-graph", middleware.AuthMiddleware(), h.IOCHandler.GetTenantIOCGraph)
		protected.POST("/cases/:case_id/iocs", middleware.AuthMiddleware(), h.IOCHandler.AddIOCToCase)
		protected.GET("/cases/:case_id/iocs", middleware.AuthMiddleware(), h.IOCHandler.GetIOCsByCase)
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
		//chain of custody
		protected.POST("/cases/:case_id/chain_of_custody", h.ChainOfCustodyHandler.AddEntry)
		protected.PUT("/cases/:case_id/chain_of_custody/:id", h.ChainOfCustodyHandler.UpdateEntry)
		protected.GET("/cases/:case_id/chain_of_custody/:id", h.ChainOfCustodyHandler.GetEntry)
		protected.GET("/cases/:case_id/chain_of_custody", h.ChainOfCustodyHandler.GetEntries)
		// ─── Metadata Evidence Upload ────────────────
		protected.POST("/evidence", h.MetadataHandler.UploadEvidence)
		// ─── Metadata Evidence Retrieval ─────────────
		protected.GET("/evidence-metadata/:id", h.MetadataHandler.GetEvidenceByID)
		protected.GET("/evidence-metadata/case/:case_id", h.MetadataHandler.GetEvidenceByCaseID)
		protected.GET("/evidence/count/:tenantId", h.EvidenceHandler.GetEvidenceCount)
		// ─── Admin: Users ────────────────────────────
		protected.GET("/users", h.AdminService.ListUsers)
		protected.GET("tenants/:tenantId/users", middleware.AuthMiddleware(), h.AdminService.ListUsersByTenant)
		protected.DELETE("/users/:userId", h.AdminService.DeleteUserHandler)
		protected.GET("/audit-logs", h.AdminService.GetAuditLogs)
		protected.GET("/audit-logs/export", h.AdminService.ExportAuditLogs) // Add this line

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

		x3dhGroup := api.Group("/x3dh")
		x3dh.RegisterX3DHHandlers(x3dhGroup, h.X3DHService)
		// RegisterMessageRoutes(protected, h.MessageService, auditLogger)
		// ─── Thread Messaging ────────────────────────

		RegisterMessageRoutes(protected, h.MessageHandler)
		// ─── Thread Routes ─────────────────────────────────────
		RegisterThreadRoutes(protected, h.AnnotationThreadHandler)

		// ─── Chat_message Routes ────────────────────────────────
		RegisterChatRoutes(protected, h.ChatHandler)

		// ─── Evidence Viewer + Tagging ────────────────
		RegisterEvidenceRoutes(protected, h.EvidenceViewerHandler, h.EvidenceTagHandler, h.MetadataHandler, h.PermissionChecker)

		RegisterCaseTagRoutes(protected, h.CaseTagHandler, h.PermissionChecker)

		// ─── Report Generation ──────────────────────────────
		RegisterReportRoutes(protected, h.ReportHandler)

		// // ─── Report Status Update ─────────────────────────────
		RegisterReportStatusRoutes(protected, h.ReportStatusHandler)
		// ─── Report AI Assistance ─────────────────────────────
		RegisterReportAIRoutes(protected, h.ReportAIHandler)

		// ─── Health Checks ──────────────────────────────
		RegisterHealthRoutes(protected, h.HealthHandler)

	}
	return router
}
