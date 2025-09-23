package main

import (
	"log"
	"os"
	"time"

	"aegis-api/db"
	"fmt"

	"aegis-api/handlers"
	"aegis-api/middleware"
	"aegis-api/pkg/websocket"
	"aegis-api/routes"
	graphicalmapping "aegis-api/services_/GraphicalMapping"
	"aegis-api/services_/admin/get_collaborators"
	"aegis-api/services_/annotation_threads/messages"
	annotationthreads "aegis-api/services_/annotation_threads/threads"
	"aegis-api/services_/auditlog"
	"aegis-api/services_/auth/login"
	"aegis-api/services_/auth/registration"
	"aegis-api/services_/auth/reset_password"
	"aegis-api/services_/case/ListActiveCases"
	"aegis-api/services_/case/ListCases"
	"aegis-api/services_/case/ListClosedCases"
	"aegis-api/services_/case/ListUsers"
	"aegis-api/services_/case/case_assign"
	"aegis-api/services_/case/case_creation"
	"aegis-api/services_/case/case_evidence_totals"
	"aegis-api/services_/case/case_tags"
	update_case "aegis-api/services_/case/case_update"
	"aegis-api/services_/chain_of_custody"
	"aegis-api/services_/chat"
	evidencecount "aegis-api/services_/evidence/evidence_count"
	"aegis-api/services_/evidence/evidence_download"
	"aegis-api/services_/evidence/evidence_tag"
	"aegis-api/services_/evidence/evidence_viewer"
	"aegis-api/services_/evidence/metadata"
	"aegis-api/services_/evidence/upload"
	"aegis-api/services_/notification"

	"aegis-api/services_/report"
	report_ai_assistance "aegis-api/services_/report/report_ai_assistance"
	"aegis-api/services_/report/update_status"

	"aegis-api/services_/timeline"

	"aegis-api/pkg/encryption"
	"aegis-api/services_/user/profile"
	"aegis-api/services_/health"

	//"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
)

func InitCollections(db *mongo.Database) {
	chat.MessageCollection = db.Collection("chat_messages")
}

func main() {
	// Granular endpoint/method limit config for rate limiting
	granularLimits := middleware.EndpointLimitConfig{
		"POST": {
			"/api/v1/auth/login": 100,
			"/api/v1/register":   50,
		},
		"GET": {
			"/api/v1/teams": 200,
		},
	}
	// ─── Load Environment Variables ──────────────────────────────
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found. Using system environment variables.")
	}

	// ─── Set JWT Secret ─────────────────────────────────────────
	jwtSecret := os.Getenv("JWT_SECRET_KEY")
	if jwtSecret == "" {
		log.Fatal("❌ JWT_SECRET_KEY not set in environment")
	}
	middleware.SetJWTSecret([]byte(jwtSecret))

	// ─── Initialize Database ────────────────────────────────────
	if err := db.InitDB(); err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}
	log.Println("✅ Connected to the database")

	// Initialize Mongo
	if err := db.ConnectMongo(); err != nil {
		log.Fatal("❌ Failed to connect to MongoDB:", err)
	}
	mongoDatabase := db.MongoDatabase
	InitCollections(mongoDatabase) // Initialize MongoDB collections
	// ─── Debug Logging ──────────────────────────────────────────
	log.Println("📨 Using SMTP server:", os.Getenv("SMTP_HOST"))

	// ─── permission checker──────────────────────
	sqlDB, err := db.DB.DB()
	if err != nil {
		log.Fatalf("❌ Failed to extract SQL DB: %v", err)
	}
	permChecker := &middleware.DBPermissionChecker{DB: sqlDB}

	// ─── Initialize Encryption ──────────────────────────────────
	// Initialize encryption with master key
	if err := encryption.Init(); err != nil {
		log.Fatal("encryption init failed:", err)
	}

	// Test encryption
	enc, _ := encryption.Encrypt("secret123")
	fmt.Println("Encrypted:", enc)

	dec, _ := encryption.Decrypt(enc)
	fmt.Println("Decrypted:", string(dec))

	// ─── websocket ─────────────────────────────────

	// r := gin.Default()

	// // Apply middleware to inject userID into the context
	// r.Use(middleware.AuthMiddleware())

	// Create and start WebSocket hub
	notificationService := notification.NewNotificationService(db.DB)

	hub := websocket.NewHub(notificationService)
	go hub.Run()

	// ─── Repositories ───────────────────────────────────────────
	userRepo := registration.NewGormUserRepository(db.DB)
	tenantRepo := registration.NewGormTenantRepository(db.DB)
	teamRepo := registration.NewGormTeamRepository(db.DB)
	resetTokenRepo := reset_password.NewGormResetTokenRepository(db.DB)
	caseRepo := case_creation.NewGormCaseRepository(db.DB)
	caseAssignRepo := case_assign.NewGormCaseAssignmentRepo(db.DB) //caseAssignRepo := case_assign.NewCaseAssignmentRepo(db.DB) // Use the ne
	userAdapter := case_assign.NewUserAdapter(userRepo)
	adminChecker := case_assign.NewContextAdminChecker()

	listActiveCasesRepo := ListActiveCases.NewActiveCaseRepository(db.DB)
	listClosedCasesRepo := ListClosedCases.NewClosedCaseRepository(db.DB)
	listCasesRepo := ListCases.NewGormCaseQueryRepository(db.DB)
	iocRepo := graphicalmapping.NewIOCRepository(db.DB)

	//timeline
	timelineRepo := timeline.NewRepository(db.DB)
	if err := timelineRepo.AutoMigrate(); err != nil {
		log.Fatalf("failed migrating timeline: %v", err)
	}
	evidenceCountRepo := evidencecount.NewEvidenceRepository(db.DB)
	chainOfCustodyRepo := chain_of_custody.NewChainOfCustodyRepository(db.DB)
	if chainOfCustodyRepo == nil {
		log.Fatal("Failed to create chain of custody repository")
	}
	// ─── Email Sender (Mock) ────────────────────────────────────
	emailSender := reset_password.NewMockEmailSender()

	// ─── Services ───────────────────────────────────────────────
	regService := registration.NewRegistrationService(userRepo, tenantRepo, teamRepo)
	authService := login.NewAuthService(userRepo)
	resetService := reset_password.NewPasswordResetService(resetTokenRepo, userRepo, emailSender)
	caseService := case_creation.NewCaseService(caseRepo, notificationService, hub)
	caseAssignService := case_assign.NewCaseAssignmentService(caseAssignRepo, adminChecker, userAdapter, notificationService, hub)

	listActiveCasesService := ListActiveCases.NewService(listActiveCasesRepo)
	listClosedCasesService := ListClosedCases.NewService(listClosedCasesRepo)
	listCasesService := ListCases.NewListCasesService(listCasesRepo)

	// ─── Audit Logging ──────────────────────────────────────────
	mongoLogger := auditlog.NewMongoLogger(mongoDatabase) // mongoDB is your *mongo.Database
	zapLogger := auditlog.NewZapLogger()
	auditLogger := auditlog.NewAuditLogger(mongoLogger, zapLogger)

	// Build the get_collaborators repository & service
	collabRepo := get_collaborators.NewGormRepository(db.DB)
	collabService := get_collaborators.NewService(collabRepo)
	getCollaboratorsHandler := handlers.NewGetCollaboratorsHandler(collabService, auditLogger)

	// ─── Unified Case Services ─────────────────────────────────
	updateCaseRepo := update_case.NewGormUpdateCaseRepository(db.DB)
	updateCaseService := update_case.NewService(
		updateCaseRepo,
		collabService,       // ✅ GetCollaborators service
		notificationService, // ✅ Notification service
		hub,                 // ✅ WebSocket Hub
	)
	caseServices := handlers.NewCaseServices(
		caseService,
		listCasesService,
		listActiveCasesService,
		caseAssignService,
		listClosedCasesService,
		updateCaseService,
	)

	// ─── List Users ─────────────────────────────────────────────
	listUserRepo := ListUsers.NewUserRepository(db.DB)
	listUserService := ListUsers.NewListUserService(listUserRepo)
	// ioc
	iocService := graphicalmapping.NewIOCService(iocRepo)
	//timeline
	timelineService := timeline.NewService(timelineRepo)

	evidenceCountService := evidencecount.NewEvidenceService(evidenceCountRepo)

	// ─── Handlers ───────────────────────────────────────────────
	adminHandler := handlers.NewAdminService(regService, listUserService, nil, nil, auditLogger)
	authHandler := handlers.NewAuthHandler(authService, resetService, userRepo, auditLogger)

	//pass separate services explicitly
	caseHandler := handlers.NewCaseHandler(
		caseServices,
		listCasesService,
		listActiveCasesService,
		listClosedCasesService,
		auditLogger, // AuditLogger
		userAdapter, // UserRepo
		updateCaseService,
	)
	//ioc
	iocHandler := handlers.NewIOCHandler(iocService)
	//timeline
	timelineHandler := handlers.NewTimelineHandler(timelineService)
	// ─── Evidence Upload/Download/Metadata ──────────────────────
	evidenceHandler := handlers.NewEvidenceHandler(evidenceCountService)
	metadataRepo := metadata.NewGormRepository(db.DB)
	ipfsClient := upload.NewIPFSClient("")

	uploadService := upload.NewEvidenceService(ipfsClient)
	metadataService := metadata.NewService(metadataRepo, ipfsClient)
	downloadService := evidence_download.NewService(metadataRepo, ipfsClient)

	uploadHandler := handlers.NewUploadHandler(uploadService, auditLogger)
	metadataHandler := handlers.NewMetadataHandler(metadataService, auditLogger)
	downloadHandler := handlers.NewDownloadHandler(downloadService, auditLogger)

	// ─── Chain of Custody ─────────────────────────────────────
	chainOfCustodyService := chain_of_custody.NewChainOfCustodyService(chainOfCustodyRepo)
	if chainOfCustodyService == nil {
		log.Fatal("Failed to create chain of custody service")
	}
	chainOfCustodyHandler := handlers.NewChainOfCustodyHandler(chainOfCustodyService)

	// ─── Messages / WebSocket ───────────────────────────────────
	messageRepo := messages.NewMessageRepository(db.DB)
	messageHub := websocket.NewHub(notificationService)
	go messageHub.Run()

	messageService := messages.NewMessageService(*messageRepo, messageHub)

	//actual MessageHandler
	messageHandler := handlers.NewMessageHandler(messageService, auditLogger)
	// ─── Annotation Threads ─────────────────────────────────────
	annotationRepo := annotationthreads.NewAnnotationThreadRepository(db.DB)
	annotationService := annotationthreads.NewAnnotationThreadService(*annotationRepo, messageHub)
	annotationThreadHandler := handlers.NewAnnotationThreadHandler(annotationService, *auditLogger)

	// ─── Chat Service ───────────────────────────────────────────
	// Initialize chat repository, user service, IPFS uploader, WebSocket manager, and chat
	chatRepo := chat.NewChatRepository(mongoDatabase, db.DB, hub, notificationService)
	userService := chat.NewUserService(mongoDatabase)
	ipfsUploader := chat.NewIPFSUploader("http://ipfs:5001", "")
	wsManager := chat.NewWebSocketManager(userService, chatRepo)
	chatService := chat.NewChatService(chatRepo, ipfsUploader, wsManager)
	chatHandler := handlers.NewChatHandler(chatService, auditLogger)

	// User Profile Service
	profileRepo := profile.NewGormProfileRepository(db.DB)
	profileService := profile.NewProfileService(profileRepo)
	profileIPFSClient := upload.NewIPFSClient("")
	profileHandler := handlers.NewProfileHandler(profileService, auditLogger, profileIPFSClient)

	// ─── Evidence Tagging ─────────────────────────────
	evidenceTagRepo := evidence_tag.NewEvidenceTagRepository(db.DB)
	evidenceTagService := evidence_tag.NewEvidenceTagService(evidenceTagRepo)
	evidenceTagHandler := &handlers.EvidenceTagHandler{
		Service: evidenceTagService,
	}

	// ─── Evidence Viewer ─────────────────────────────
	viewerIPFSClient := evidence_viewer.NewIPFSClient()
	evidenceViewerRepo := evidence_viewer.NewPostgresEvidenceRepository(db.DB, viewerIPFSClient)
	evidenceViewerService := evidence_viewer.NewEvidenceService(evidenceViewerRepo)
	evidenceViewerHandler := &handlers.EvidenceViewerHandler{
		Service: evidenceViewerService,
	}

	// ─── Case Tagging ─────────────────────────────
	caseTagRepo := case_tags.NewCaseTagRepository(db.DB)
	caseTagService := case_tags.NewCaseTagService(caseTagRepo)
	caseTagHandler := &handlers.CaseTagHandler{
		Service: caseTagService,
	}

	// ─── Case Evidence Totals ─────────────────────────────
	caseEviRepo := case_evidence_totals.NewCaseEviRepository(db.DB)
	dashboardService := case_evidence_totals.NewDashboardService(caseEviRepo)
	caseEviTotalsHandler := handlers.NewCaseEvidenceTotalsHandler(dashboardService)

	// ─── AuditLog Service and Handler ─────────────────────────────
	auditLogService := auditlog.NewAuditLogService(mongoDatabase, userRepo)

	recentActivityHandler := handlers.NewRecentActivityHandler(auditLogService)

	notificationService = &notification.NotificationService{
		DB: db.DB,
	}

	// ─── Chain of Custody (CoC) ─────────────────────────────
	// Create an adapter for the AuditLogger to fit the coc.Auditor interface
	// auditAdapter := &coc.AuditLogAdapter{
	// 	AuditLogger: auditLogger, // Use the existing AuditLogger
	// }

	// Initialize the CoC service (pass it as a value, not a pointer)
	// cocSvc := coc.Service{
	// 	Repo:      coc.GormRepo{DB: db.DB}, // Initialize repository (GORM)
	// 	Authz:     coc.SimpleAuthz{},       // Placeholder for RBAC (role-based access control)
	// 	Audit:     auditAdapter,            // Use the adapter for AuditLogger
	// 	DedupeWin: 3 * time.Second,         // Deduplication window (optional)
	// }

	// Initialize the handler, passing a pointer to cocSvc to avoid copying sync.Mutex
	// cocHandler := handlers.NewCoCHandler(cocSvc, auditLogger) // Pass the service pointer

	// ─── Report Service Initialization ─────────────────────

	// Initialize the repository and service for report generation and management
	// ─── Report Service Initialization ─────────────────────────────

	// ...existing code...
	reportRepo := report.NewReportRepository(db.DB)
	reportContentCollection := mongoDatabase.Collection("report_contents")
	reportMongoRepo := report.NewReportMongoRepo(reportContentCollection)
	pgSectionRepo := report_ai_assistance.NewGormReportSectionRepo(db.DB)
	reportService := report.NewReportService(
		reportRepo,
		reportMongoRepo,
		pgSectionRepo,
	)

	// Evidence metadata service for context autofill
	metadataRepo = metadata.NewGormRepository(db.DB)
	ipfsClient = upload.NewIPFSClient("")
	metadataService = metadata.NewService(metadataRepo, ipfsClient)

	// Timeline service for context autofill
	timelineRepo = timeline.NewRepository(db.DB)
	timelineService = timeline.NewService(timelineRepo)

	// Use the new handler constructor with dependencies
	reportHandler := handlers.NewReportHandlerWithDeps(
		reportService,
		metadataService, // implements FindEvidenceByCaseID
		timelineService, // implements ListEvents
		caseService,     // implements GetCaseByID
		iocService,      // implements ListIOCsByCase
	)

	// Instantiate Report AI Service
	mongoSectionRepo := report_ai_assistance.NewMongoSectionRepositoryWithPg(mongoDatabase, db.DB)
	aiSuggestionRepo := report_ai_assistance.NewGormAISuggestionRepo(db.DB)
	sectionRefsRepo := report_ai_assistance.NewGormSectionRefsRepo(db.DB)
	aiFeedbackRepo := report_ai_assistance.NewGormAIFeedbackRepo(db.DB)
	// Ensure AIClient implementation matches the expected interface signature
	aiClient := report_ai_assistance.NewAIClientLocalAI("")
	reportAIService := report_ai_assistance.NewReportService(
		mongoSectionRepo,
		aiSuggestionRepo,
		sectionRefsRepo,
		aiFeedbackRepo,
		aiClient, // Your AI client (e.g., OpenAI wrapper)
	)

	// Instantiate Report AI Handler
	reportAIHandler := handlers.NewReportAIHandler(reportAIService, reportService)
	// ─── Report Status Update ─────────────────────────────

	reportStatusRepo := update_status.NewReportStatusRepository(db.DB)
	reportStatusService := update_status.NewReportStatusService(reportStatusRepo)
	reportStatusHandler := handlers.NewReportStatusHandler(reportStatusService)

	// ─── Health Check Service and Handler ─────────────────────────────
	
	repo := &health.Repository{
		Mongo:    db.MongoClient,
		Postgres: sqlDB,
		IPFS:     viewerIPFSClient,
	}
	healthService:= &health.Service{Repo: repo}
	healthHandler := &handlers.HealthHandler{Service: healthService}


	// ─── Compose Handler Struct ─────────────────────────────────
	mainHandler := handlers.NewHandler(
		adminHandler,
		authHandler,
		caseServices,
		nil, // evidenceHandler
		nil, // userHandler
		caseHandler,
		uploadHandler,
		downloadHandler,
		metadataHandler,
		messageHandler,
		annotationThreadHandler,
		chatHandler, // New ChatHandler
		profileHandler,
		getCollaboratorsHandler, // New GetCollaboratorsHandler
		evidenceViewerHandler,
		evidenceTagHandler,
		permChecker,
		caseTagHandler,
		caseEviTotalsHandler,
		hub,
		recentActivityHandler,
		teamRepo,   // Pass the team repository
		tenantRepo, // Pass the tenant repository
		userRepo,   // Pass the user repository
		notificationService,
		reportHandler,
		reportStatusHandler,
		reportAIHandler,

		iocHandler,
		timelineHandler,
		evidenceHandler,
		chainOfCustodyHandler,
		healthHandler,
		
	)

	// ─── Set Up Router and Launch ───────────────────────────────
	//router := routes.SetUpRouter(mainHandler)
	// ─── Set Up Router and Launch ───────────────────────────────
	router := routes.SetUpRouter(mainHandler)
	router.Use(middleware.AuthMiddleware())
	router.Use(middleware.RateLimitMiddleware(100, time.Minute, granularLimits)) // 100 requests per minute per user, granular config

	// ─── websocket ─────────────────────────────────
	wsGroup := router.Group("/ws")
	wsGroup.Use(middleware.WebSocketAuthMiddleware()) // ✅ For ws://.../cases/:id?token=...
	websocket.RegisterWebSocketRoutes(wsGroup, hub)

	//load balance port
	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // default
	}

	log.Println("🚀 Starting AEGIS API on :" + port + "...")
	log.Println("📚 Swagger docs: http://localhost:" + port + "/swagger/index.html")

	if err := router.Run(":" + port); err != nil {
		log.Fatal("❌ Failed to start server:", err)
	}

}
