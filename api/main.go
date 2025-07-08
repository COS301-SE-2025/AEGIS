package main

import (
	"log"
	"os"

	"aegis-api/db"

	"aegis-api/handlers"
	"aegis-api/middleware"
	"aegis-api/pkg/websocket"
	"aegis-api/routes"
	"aegis-api/services_/annotation_threads/messages"
	annotationthreads "aegis-api/services_/annotation_threads/threads"
	"aegis-api/services_/auth/login"
	"aegis-api/services_/auth/registration"
	"aegis-api/services_/auth/reset_password"
	"aegis-api/services_/case/ListActiveCases"
	"aegis-api/services_/case/ListCases"
	"aegis-api/services_/case/ListUsers"
	"aegis-api/services_/case/case_assign"
	"aegis-api/services_/case/case_creation"
	"aegis-api/services_/evidence/evidence_download"
	"aegis-api/services_/evidence/metadata"
	"aegis-api/services_/evidence/upload"

	"github.com/joho/godotenv"
)

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

	// ─── Debug Logging ──────────────────────────────────────────
	log.Println("📨 Using SMTP server:", os.Getenv("SMTP_HOST"))

	// ─── Repositories ───────────────────────────────────────────
	userRepo := registration.NewGormUserRepository(db.DB)
	resetTokenRepo := reset_password.NewGormResetTokenRepository(db.DB)
	caseRepo := case_creation.NewGormCaseRepository(db.DB)
	caseAssignRepo := case_assign.NewGormCaseAssignmentRepo(db.DB)
	listActiveCasesRepo := ListActiveCases.NewActiveCaseRepository(db.DB)
	listCasesRepo := ListCases.NewGormCaseQueryRepository(db.DB)

	// ─── Email Sender (Mock) ────────────────────────────────────
	emailSender := reset_password.NewMockEmailSender()

	// ─── Services ───────────────────────────────────────────────
	regService := registration.NewRegistrationService(userRepo)
	authService := login.NewAuthService(userRepo)
	resetService := reset_password.NewPasswordResetService(resetTokenRepo, userRepo, emailSender)
	caseService := case_creation.NewCaseService(caseRepo)
	caseAssignService := case_assign.NewCaseAssignmentService(caseAssignRepo)
	listActiveCasesService := ListActiveCases.NewService(listActiveCasesRepo)
	listCasesService := ListCases.NewListCasesService(listCasesRepo)

	// ─── Unified Case Services ─────────────────────────────────
	caseServices := handlers.NewCaseServices(
		caseService,
		listCasesService,
		listActiveCasesService,
		caseAssignService,
	)

	// ─── List Users ─────────────────────────────────────────────
	listUserRepo := ListUsers.NewUserRepository(db.DB)
	listUserService := ListUsers.NewListUserService(listUserRepo)

	// ─── Handlers ───────────────────────────────────────────────
	adminHandler := handlers.NewAdminService(regService, listUserService, nil, nil)
	authHandler := handlers.NewAuthHandler(authService, resetService, userRepo)

	// 🔥 ✅ Updated to pass separate services explicitly
	caseHandler := handlers.NewCaseHandler(
		caseServices,           // CaseServiceInterface
		listCasesService,       // ListCasesService
		listActiveCasesService, // ListActiveCasesService
	)

	// ─── Evidence Upload/Download/Metadata ──────────────────────
	metadataRepo := metadata.NewGormRepository(db.DB)
	ipfsClient := upload.NewIPFSClient("")

	uploadService := upload.NewEvidenceService(ipfsClient)
	metadataService := metadata.NewService(metadataRepo, ipfsClient)
	downloadService := evidence_download.NewService(metadataRepo, ipfsClient)

	uploadHandler := handlers.NewUploadHandler(uploadService)
	metadataHandler := handlers.NewMetadataHandler(metadataService)
	downloadHandler := handlers.NewDownloadHandler(downloadService)

	// ─── Messages / WebSocket ───────────────────────────────────
	messageRepo := messages.NewMessageRepository(db.DB)
	messageHub := websocket.NewHub()
	go messageHub.Run()

	messageService := messages.NewMessageService(*messageRepo, messageHub)

	// ─── Annotation Threads ─────────────────────────────────────
	annotationRepo := annotationthreads.NewAnnotationThreadRepository(db.DB)
	annotationService := annotationthreads.NewAnnotationThreadService(*annotationRepo, messageHub)
	annotationThreadHandler := handlers.NewAnnotationThreadHandler(annotationService)

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
		messageService,
		annotationThreadHandler,
	)

	// ─── Set Up Router and Launch ───────────────────────────────
	router := routes.SetUpRouter(mainHandler)

	log.Println("🚀 Starting AEGIS API on :8080...")
	log.Println("📚 Swagger docs: http://localhost:8080/swagger/index.html")

	if err := router.Run(":8080"); err != nil {
		log.Fatal("❌ Failed to start server:", err)
	}
}
