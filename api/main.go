package main

import (
	"aegis-api/db"
	"aegis-api/services_/admin/delete_user"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"aegis-api/cache"
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
	"aegis-api/services_/case/case_deletion"
	"aegis-api/services_/case/case_evidence_totals"
	"aegis-api/services_/case/case_tags"
	update_case "aegis-api/services_/case/case_update"
	"aegis-api/services_/case/listArchiveCases"
	"aegis-api/services_/chain_of_custody"
	"aegis-api/services_/chat"
	evidencecount "aegis-api/services_/evidence/evidence_count"
	"aegis-api/services_/evidence/evidence_download"
	"aegis-api/services_/evidence/evidence_tag"
	"aegis-api/services_/evidence/evidence_viewer"
	"aegis-api/services_/evidence/metadata"
	"aegis-api/services_/evidence/upload"
	"aegis-api/services_/notification"
	timelineai "aegis-api/services_/timeline/timeline_ai"

	"aegis-api/services_/report"
	report_ai_assistance "aegis-api/services_/report/report_ai_assistance"
	"aegis-api/services_/report/update_status"

	"aegis-api/services_/timeline"

	"aegis-api/pkg/encryption"
	"aegis-api/services_/health"
	"aegis-api/services_/user/profile"

	"github.com/gin-gonic/gin"

	"aegis-api/internal/x3dh"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
)

func InitCollections(db *mongo.Database) {
	websocket.MessageCollection = db.Collection("chat_messages")
}

/*
Function to enforce HTTPS and add HSTS headers.
--ENCRYPTION IN TRANSIT HTTPS--
*/
// requireTLS is a middleware that rejects non-TLS requests.
func requireTLS() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.TLS == nil {
			// Option A: Respond with 426 Upgrade Required
			c.Header("Connection", "close")
			c.AbortWithStatusJSON(http.StatusUpgradeRequired, gin.H{
				"error": "HTTPS required",
			})
			return
		}
		// Add HSTS header on TLS requests
		// max-age=63072000 (2 years), includeSubDomains, preload candidate
		c.Header("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		c.Next()
	}
}

var rdb *redis.Client
var ctx = context.Background()

// Updated InitRedis function
func InitRedis() {
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisPassword := os.Getenv("REDIS_PASSWORD")

	// Default values if environment variables are not set
	if redisHost == "" {
		redisHost = "redis" // Docker service name
	}
	if redisPort == "" {
		redisPort = "6379"
	}

	addr := fmt.Sprintf("%s:%s", redisHost, redisPort)
	log.Printf("🔄 Attempting to connect to Redis at: %s", addr)

	rdb = redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     redisPassword, // Can be empty string for no password
		DB:           0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 1,
	})

	// Test the connection with retries
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err := rdb.Ping(ctx).Result()
		cancel()

		if err == nil {
			log.Println("✅ Connected to Redis successfully")
			// IMPORTANT: Set the Redis client for middleware
			middleware.SetRedisClient(rdb)
			return
		}

		log.Printf("⚠️  Redis connection attempt %d failed: %v", i+1, err)
		if i < maxRetries-1 {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}

	// Don't panic - just log the error and continue without Redis
	log.Println("❌ Failed to connect to Redis after multiple attempts - continuing without rate limiting")
}

// Add this function to gracefully close Redis connection
func CloseRedis() {
	if rdb != nil {
		if err := rdb.Close(); err != nil {
			log.Printf("Error closing Redis connection: %v", err)
		}
	}
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
	} else {
		log.Println("✅ Loaded .env file")
	}

	// Initialize Redis
	InitRedis()

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
	// if err := encryption.Init(); err != nil {
	// 	log.Fatal("encryption init failed:", err)
	// }

	// Test encryption
	enc, _ := encryption.Encrypt("secret123")
	fmt.Println("Encrypted:", enc)

	dec, _ := encryption.Decrypt(enc)
	fmt.Println("Decrypted:", string(dec))

	//--Gin setup for HTTPS--
	r := gin.Default()
	// Enforce HTTPS and add HSTS headers
	r.Use(gin.Recovery())

	// ─── websocket ─────────────────────────────────

	// r := gin.Default()

	// // Apply middleware to inject userID into the context
	// r.Use(middleware.AuthMiddleware())

	// Create and start WebSocket hub
	notificationService := notification.NewNotificationService(db.DB)

	hub := websocket.NewHub(notificationService, mongoDatabase)
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
	listArchivedCasesRepo := listArchiveCases.NewArchiveCaseRepository(db.DB)
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
	listArchiveCasesService := listArchiveCases.NewArchiveCaseService(listArchivedCasesRepo)
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
		collabService,       //  GetCollaborators service
		notificationService, //  Notification service
		hub,                 //  WebSocket Hub
	)
	caseServices := handlers.NewCaseServices(
		caseService,
		listCasesService,
		listActiveCasesService,
		caseAssignService,
		listClosedCasesService,
		listArchiveCasesService,
		updateCaseService,
	)

	// ─── List Users ─────────────────────────────────────────────
	listUserRepo := ListUsers.NewUserRepository(db.DB)
	listUserService := ListUsers.NewListUserService(listUserRepo)

	// ─── User Delete Service (Admin) ───────────────────────────
	deleteUserGormRepo := delete_user.NewGormUserRepository(db.DB)
	userDeleteService := delete_user.NewUserDeleteService(deleteUserGormRepo)

	// ioc
	iocService := graphicalmapping.NewIOCService(iocRepo)
	//timeline
	timelineService := timeline.NewService(timelineRepo)

	evidenceCountService := evidencecount.NewEvidenceService(evidenceCountRepo)

	// ─── Handlers ───────────────────────────────────────────────
	adminHandler := handlers.NewAdminService(regService, listUserService, nil, userDeleteService, auditLogger)
	authHandler := handlers.NewAuthHandler(authService, resetService, userRepo, auditLogger)

	addr := os.Getenv("REDIS_ADDR") // "redis:6379" in compose
	pass := os.Getenv("REDIS_PASS")
	db1 := 0
	if v, err := strconv.Atoi(os.Getenv("REDIS_DB")); err == nil {
		db1 = v
	}

	var cacheClient cache.Client
	if addr != "" {
		cacheClient = cache.NewRedis(addr, pass, db1)
	} else {
		cacheClient = cache.NewMemory()
	}

	//pass separate services explicitly
	caseHandler := handlers.NewCaseHandler(
		caseServices,
		listCasesService,
		listActiveCasesService,
		listClosedCasesService,
		listArchiveCasesService,
		auditLogger, // AuditLogger
		userAdapter, // UserRepo
		updateCaseService,
		cacheClient, // Cache Client
	)
	//ioc
	iocHandler := handlers.NewIOCHandler(iocService)
	//timeline
	timelineHandler := handlers.NewTimelineHandler(timelineService)

	//Timeline
	// ─── Evidence Upload/Download/Metadata ──────────────────────
	evidenceHandler := handlers.NewEvidenceHandler(evidenceCountService, cacheClient)
	metadataRepo := metadata.NewGormRepository(db.DB)
	ipfsClient := upload.NewIPFSClient("")

	uploadService := upload.NewEvidenceService(ipfsClient)
	metadataService := metadata.NewService(metadataRepo, ipfsClient)
	downloadService := evidence_download.NewService(metadataRepo, ipfsClient)

	uploadHandler := handlers.NewUploadHandler(uploadService, auditLogger)
	metadataHandler := handlers.NewMetadataHandler(metadataService, auditLogger, cacheClient)
	downloadHandler := handlers.NewDownloadHandler(downloadService, auditLogger)

	// ─── Chain of Custody ─────────────────────────────────────
	chainOfCustodyService := chain_of_custody.NewChainOfCustodyService(chainOfCustodyRepo)
	if chainOfCustodyService == nil {
		log.Fatal("Failed to create chain of custody service")
	}
	chainOfCustodyHandler := handlers.NewChainOfCustodyHandler(chainOfCustodyService)

	// ─── Messages / WebSocket ───────────────────────────────────
	messageRepo := messages.NewMessageRepository(db.DB)
	messageHub := websocket.NewHub(notificationService, mongoDatabase)
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
	caseEviTotalsHandler := handlers.NewCaseEvidenceTotalsHandler(dashboardService, cacheClient)

	// ─── Case Deletion ──────────────────────────────────────
	caseDeletionRepo := case_deletion.NewGormCaseRepository(db.DB)
	caseDeletionService := case_deletion.NewCaseDeletionService(caseDeletionRepo)
	caseDeletionHandler := handlers.NewCaseDeletionHandler(caseDeletionService)

	// ─── AuditLog Service and Handler ─────────────────────────────
	auditLogService := auditlog.NewAuditLogService(mongoDatabase, userRepo)

	recentActivityHandler := handlers.NewRecentActivityHandler(auditLogService)

	notificationService = &notification.NotificationService{
		DB: db.DB,
	}

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
	timelineAIrepo := timelineai.NewAIRepository(db.DB)

	aiConfig := timelineai.AIModelConfig{
		ModelName:   "gpt2",
		MaxTokens:   1500,
		Temperature: 0.7,
		BaseURL:     os.Getenv("AI_BASE_URL"),
		Enabled:     os.Getenv("AI_ENABLED") == "true",
	}

	TimelineAIService := timelineai.NewAIService(timelineAIrepo, &aiConfig)

	// Instantiate Timeline AI Handler
	timelineAIHandler := handlers.NewTimelineAIHandler(TimelineAIService)
	log.Println("✅ Timeline AI service initialized")

	// ─── Report Handler Initialization ─────────────────────────────

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
	healthService := &health.Service{Repo: repo}
	healthHandler := &handlers.HealthHandler{Service: healthService}

	// --- X3DH: keystore + crypto + auditor + service -----------------
	x3dhStore := x3dh.NewPostgresKeyStore(sqlDB)

	// AES key (32 bytes for AES-256) – supply via env, base64-encoded
	aesKeyB64 := os.Getenv("X3DH_AES_KEY_B64")
	if aesKeyB64 == "" {
		log.Fatal("❌ X3DH_AES_KEY_B64 not set")
	}
	aesKey, err := base64.StdEncoding.DecodeString(aesKeyB64)
	if err != nil || len(aesKey) != 32 {
		log.Fatal("❌ X3DH_AES_KEY_B64 must be base64 for 32 bytes (AES-256)")
	}

	cryptoSvc, err := x3dh.NewAESGCMCryptoService(aesKey)
	if err != nil {
		log.Fatalf("❌ AESGCM init failed: %v", err)
	}

	x3dhAuditor := x3dh.NewMongoAuditLogger(mongoDatabase)
	x3dhService := x3dh.NewBundleService(x3dhStore, cryptoSvc, x3dhAuditor)

	// ─── Compose Handler Struct ─────────────────────────────────
	mainHandler := handlers.NewHandler(
		adminHandler,
		authHandler,
		caseServices,
		nil, // evidenceHandler
		nil, // userHandler
		caseHandler,
		caseDeletionHandler,
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
		timelineAIHandler,
		evidenceHandler,
		chainOfCustodyHandler,

		healthHandler,

		x3dhService, // X3DH Service

	)

	// ─── Set Up Router and Launch ───────────────────────────────
	//router := routes.SetUpRouter(mainHandler)
	// ─── Set Up Router and Launch ───────────────────────────────
	router := routes.SetUpRouter(mainHandler)
	router.Use(middleware.AuthMiddleware())
	router.Use(middleware.RateLimitMiddleware(100, time.Minute, granularLimits)) // 100 requests per minute per user, granular config

	// //───────────── ENCRYPTION IN TRANSIT HTTPS ────────────────────────
	// // Note: In production, use a proper TLS certificate from a trusted CA
	// // For local testing, you can generate self-signed certs or use mkcert

	// Example of setting up HTTPS with Gin (commented out for now)
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	//Creating a test endpoint without authentication for testing HTTPS
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"time":    time.Now().Format(time.RFC3339),
			"message": "This is a test endpoint over HTTPS!",
		})
	})

	// // Start HTTPS server (commented out to avoid accidental execution)
	// //
	// err = router.RunTLS(":8443", "certs/localhost.pem", "certs/localhost-key.pem")
	// if err != nil {
	// 	log.Fatal("Failed to start HTTPS server:", err)
	// }

	// // ─── websocket ─────────────────────────────────
	// wsGroup := router.Group("/ws")
	// wsGroup.Use(middleware.WebSocketAuthMiddleware()) // ✅ For ws://.../cases/:id?token=...
	// websocket.RegisterWebSocketRoutes(wsGroup, hub)

	//load balance port
	// Get port from environment variable or use default
	// port := os.Getenv("PORT")
	// if port == "" {
	// 	port = "8080" // default
	// }

	// log.Println("🚀 Starting AEGIS API on :" + port + "...")
	// log.Println("📚 Swagger docs: http://localhost:" + port + "/swagger/index.html")

	// if err := router.Run(":" + port); err != nil {
	// 	log.Fatal("❌ Failed to start server:", err)
	// }

	// HTTPS port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8443"
	}

	// HTTP port for redirects
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	// Certificate paths
	certFile := os.Getenv("SSL_CERT_FILE")
	keyFile := os.Getenv("SSL_KEY_FILE")
	if certFile == "" {
		certFile = "api/certs/localhost.pem"
	}
	if keyFile == "" {
		keyFile = "api/certs/localhost-key.pem"
	}

	// Start HTTP redirect server
	go func() {
		redirectHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			httpsURL := "https://" + r.Host + ":" + port + r.URL.Path
			if r.URL.RawQuery != "" {
				httpsURL += "?" + r.URL.RawQuery
			}
			http.Redirect(w, r, httpsURL, http.StatusMovedPermanently)
		})

		log.Printf("🔁 Starting HTTP redirect server on :%s...", httpPort)
		if err := http.ListenAndServe(":"+httpPort, redirectHandler); err != nil {
			log.Fatal("❌ Failed to start HTTP redirect server:", err)
		}
	}()

	log.Println("🚀 Starting AEGIS API on :" + port + " (HTTPS)...")
	log.Println("📚 Swagger docs: https://localhost:" + port + "/swagger/index.html")

	// Start your Gin router with HTTPS
	if err := router.RunTLS(":"+port, certFile, keyFile); err != nil {
		log.Fatal("❌ Failed to start HTTPS server:", err)
	}

	// ───────────── ENCRYPTION IN TRANSIT HTTPS ────────────────────────
	// Register websocket routes BEFORE starting the servers
	wsGroup := router.Group("/ws")
	wsGroup.Use(middleware.WebSocketAuthMiddleware()) // WebSocket auth for upgrades

	websocket.RegisterWebSocketRoutes(wsGroup, hub)

	// Enforce TLS and HSTS for all incoming requests handled by this router
	// (allows only TLS requests to be processed — non-TLS should be redirected)
	router.Use(requireTLS())

	// Apply auth and rate-limit middleware after requireTLS
	// (order matters: requireTLS first, then auth, then rate limiting)
	router.Use(middleware.AuthMiddleware())
	router.Use(middleware.RateLimitMiddleware(100, time.Minute, granularLimits))

	// Simple test endpoints already registered earlier: /ping and /health

	// Start an HTTP server that redirects all traffic to HTTPS.
	// This runs in a goroutine so it does not block.
	go func() {
		redirectHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Build redirect target to same host/path but with https
			target := "https://" + r.Host + r.URL.RequestURI()
			http.Redirect(w, r, target, http.StatusMovedPermanently)
		})

		srv := &http.Server{
			Addr:         ":8080", // HTTP port for redirect (dev). In prod use :80.
			Handler:      redirectHandler,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  30 * time.Second,
		}

		log.Println("HTTP -> HTTPS redirect server listening on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP redirect server failed: %v", err)
		}
	}()

	// Start the HTTPS server (blocking). Uses certs in certs/
	// certFile := "certs/localhost.pem"
	// keyFile := "certs/localhost-key.pem"

	// // Log and start HTTPS only
	// log.Println("🚀 Starting AEGIS API on :8443 (HTTPS only)")
	// if err := router.RunTLS(":8443", certFile, keyFile); err != nil {
	// 	log.Fatal("❌ Failed to start HTTPS server:", err)
	// }

	// Custom TLS configuration (force TLS 1.3)
	tlsCfg := &tls.Config{
		MinVersion: tls.VersionTLS13,
	}

	srv := &http.Server{
		Addr:      ":8443",
		Handler:   router,
		TLSConfig: tlsCfg,
	}

	log.Println("🚀 Starting AEGIS API on :8443 (TLS 1.3 only)")
	if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
		log.Fatal("❌ Failed to start HTTPS server:", err)
	}

}
