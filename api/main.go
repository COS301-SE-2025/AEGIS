package main

import (
	"log"
	"os"

	"aegis-api/db"

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
	"aegis-api/services_/timeline"
	"aegis-api/services_/user/profile"

	//"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
)

func InitCollections(db *mongo.Database) {
	chat.MessageCollection = db.Collection("chat_messages")
}

func main() {
	// ─── Load Environment Variables ──────────────────────────────
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found. Using system environment variables.")
	}

	// ─── Set JWT Secret ─────────────────────────────────────────
	jwtSecret := os.Getenv("JWT_SECRET_KEY")
	if jwtSecret == "" {
		log.Fatal("❌ JWT_SECRET not set in environment")
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
	// ─── Email Sender (Mock) ────────────────────────────────────
	emailSender := reset_password.NewMockEmailSender()

	// ─── Services ───────────────────────────────────────────────
	regService := registration.NewRegistrationService(userRepo, tenantRepo, teamRepo)
	authService := login.NewAuthService(userRepo)
	resetService := reset_password.NewPasswordResetService(resetTokenRepo, userRepo, emailSender)
	caseService := case_creation.NewCaseService(caseRepo, notificationService, hub)
	caseAssignService := case_assign.NewCaseAssignmentService(caseAssignRepo, adminChecker, userAdapter)

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
	chatRepo := chat.NewChatRepository(mongoDatabase)
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
		iocHandler,
		timelineHandler,
		evidenceHandler,
		chainOfCustodyHandler,
	)

	// ─── Set Up Router and Launch ───────────────────────────────
	//router := routes.SetUpRouter(mainHandler)
	// ─── Set Up Router and Launch ───────────────────────────────
	router := routes.SetUpRouter(mainHandler)

	// ─── websocket ─────────────────────────────────
	wsGroup := router.Group("/ws")
	wsGroup.Use(middleware.WebSocketAuthMiddleware()) // ✅ For ws://.../cases/:id?token=...
	websocket.RegisterWebSocketRoutes(wsGroup, hub)

	log.Println("🚀 Starting AEGIS API on :8080...")
	log.Println("📚 Swagger docs: http://localhost:8080/swagger/index.html")

	if err := router.Run(":8080"); err != nil {
		log.Fatal("❌ Failed to start server:", err)
	}
}
