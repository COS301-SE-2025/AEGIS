package main

import (
	"log"
	"os"

	"aegis-api/db"
	"aegis-api/handlers"
	"aegis-api/pkg/websocket"
	"aegis-api/routes"
	"aegis-api/services_/annotation_threads/messages"
	"aegis-api/services_/evidence/evidence_download"
	"aegis-api/services_/evidence/metadata"
	"aegis-api/services_/evidence/upload"

	// Services
	"aegis-api/middleware"
	"aegis-api/services_/auth/login"
	"aegis-api/services_/auth/registration"
	"aegis-api/services_/auth/reset_password"
	"aegis-api/services_/case/case_creation"

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

	// ─── Email Sender (Mock) ────────────────────────────────────
	emailSender := reset_password.NewMockEmailSender()

	// ─── Services ───────────────────────────────────────────────
	regService := registration.NewRegistrationService(userRepo)
	authService := login.NewAuthService(userRepo)
	resetService := reset_password.NewPasswordResetService(resetTokenRepo, userRepo, emailSender)
	caseService := case_creation.NewCaseService(caseRepo, userRepo)

	// ─── Handlers ───────────────────────────────────────────────
	adminHandler := handlers.NewAdminService(regService, nil, nil, nil)
	authHandler := handlers.NewAuthHandler(authService, resetService, userRepo)
	caseHandler := handlers.NewCaseHandler(caseService)
	// Repositories
	metadataRepo := metadata.NewGormRepository(db.DB)
	ipfsClient := upload.NewIPFSClient("")

	// Services
	uploadService := upload.NewEvidenceService(ipfsClient)
	metadataService := metadata.NewService(metadataRepo, ipfsClient)
	downloadService := evidence_download.NewService(metadataRepo, ipfsClient)

	// Handlers
	uploadHandler := handlers.NewUploadHandler(uploadService)
	metadataHandler := handlers.NewMetadataHandler(metadataService)
	downloadHandler := handlers.NewDownloadHandler(downloadService)

	// Messages
	messageRepo := messages.NewMessageRepository(db.DB)
	messageHub := websocket.NewHub()
	go messageHub.Run() // Starts the WebSocket hub in a goroutine
	messageService := messages.NewMessageService(*messageRepo, messageHub)

	// ─── Compose Handler Struct ─────────────────────────────────
	mainHandler := handlers.NewHandler(
		adminHandler,
		authHandler,
		caseService,   // CaseServiceInterface (used internally by handler wiring)
		nil,           // evidenceHandler
		nil,           // userHandler
		caseHandler,   // HTTP handler (Gin)
		uploadHandler, // Upload handler for file uploads
		downloadHandler,
		metadataHandler, // Metadata handler for evidence metadata
		messageService,
	)

	// ─── Set Up Router and Launch ───────────────────────────────
	router := routes.SetUpRouter(mainHandler)

	log.Println("🚀 Starting AEGIS API on :8080...")
	log.Println("📚 Swagger docs: http://localhost:8080/swagger/index.html")

	if err := router.Run(":8080"); err != nil {
		log.Fatal("❌ Failed to start server:", err)
	}
}

// import (
// 	database "aegis-api/db"
// 	"aegis-api/handlers"

// 	"aegis-api/pkg/websocket"

// 	"aegis-api/routes"
// 	"log"

// 	// services & repos
// 	"aegis-api/services_/case/case_assign"
// 	"aegis-api/services_/case/case_creation"
// 	"aegis-api/services/case_status_update"
// 	"aegis-api/services/delete_user"
// 	//"aegis-api/services_/evidence"
// 	"aegis-api/services/get_collaborators"
// 	"aegis-api/services_/case/listcases"
// 	"aegis-api/services/listclosedcases"
// 	"aegis-api/services/listusers"

// 	"aegis-api/services/login/auth"
// 	"aegis-api/services/registration"
// 	"aegis-api/services/reset_password"
// 	"aegis-api/services/update_user_role"

// 	"aegis-core/services/getupdate_users"

// )

// // @title AEGIS Platform API
// // @version 1.0
// // @description API for collaborative digital forensics investigations.
// // @contact.name    AEGIS Support
// // @contact.email   support@aegis-dfir.com
// // @license.name    Apache 2.0
// // @host            localhost:8080
// // @BasePath        /api/v1
// // @securityDefinitions.apikey  BearerAuth
// // @in                          header
// // @name                        Authorization
// func main() {
// 	if err := database.InitDB(); err != nil {
// 		log.Fatalf("Database connection failed: %v", err)
// 	}

// 	// case_creation repos
// 	regRepo := registration.NewGormUserRepository(database.DB)
// 	listUserRepo := listusers.NewUserRepository(database.DB)

// 	updateRoleRepo := update_user_role.NewGormUserRepo(database.DB)
// 	deleteUserRepo := delete_user.NewGormUserRepository(database.DB)

// 	resetTokenRepo := reset_password.NewGormResetTokenRepository(database.DB)
// 	userRepo := registration.NewGormUserRepository(database.DB)
// 	emailSender := reset_password.NewMockEmailSender()

// 	// Services
// 	regService := registration.NewRegistrationService(regRepo)
// 	listUserService := listusers.NewListUserService(listUserRepo)

// 	updateRoleService := update_user_role.NewUserService(updateRoleRepo)
// 	deleteUserService := delete_user.NewUserDeleteService(deleteUserRepo)

// 	resetService := reset_password.NewPasswordResetService(resetTokenRepo, userRepo, emailSender)
// 	authService := auth.NewAuthService(userRepo)

// 	adminService := handlers.NewAdminServices(regService, listUserService, updateRoleService, deleteUserService)
// 	authHandler := handlers.NewAuthHandler(authService, resetService)

// 	// Case-related services initialization
// 	caseRepo := case_creation.NewGormCaseRepository(database.DB)
// 	caseService := case_creation.NewCaseService(caseRepo)

// 	listCasesRepo := listcases.NewGormCaseQueryRepository(database.DB)
// 	listCasesService := listcases.NewListCasesService(listCasesRepo)

// 	caseAssignRepo := case_assign.NewGormCaseAssignmentRepo(database.DB)
// 	caseAssignService := case_assign.NewCaseAssignmentService(caseAssignRepo)

// 	caseStatusRepo := case_status_update.NewGormRepository(database.DB)
// 	caseStatusService := case_status_update.NewCaseStatusService(caseStatusRepo)

// 	collaboratorsRepo := get_collaborators.NewGormRepository(database.DB)
// 	collaboratorsService := get_collaborators.NewService(collaboratorsRepo)

// 	closedCasesRepo := listclosedcases.NewGormRepository(database.DB)

// 	// Evidence repos and services initialization
// 	evidence.InitIPFSClient() // Initialize IPFS first
// 	evidenceRepo := evidence.NewMongoEvidenceRepository()
// 	evidenceService := evidence.NewEvidenceService(evidenceRepo)

// 	// User handling services
// 	updateUserRepo := getupdate_users.NewUserRepository(database.DB)
// 	updateUserService := getupdate_users.NewUserService(updateUserRepo)

// 	// Initialize handlers with all required services
// 	caseHandler := handlers.NewCaseServices(
// 		caseService,          // For case creation
// 		listCasesService,     // For listing cases
// 		caseStatusService,    // For updating case status
// 		collaboratorsService, // For collaborator operations
// 		caseAssignService,    // For assigning users
// 		caseAssignService,    // Same service handles unassignment
// 		closedCasesRepo,      // For closed cases operations
// 	)

// 	evidenceHandler := handlers.NewEvidenceHandler(evidenceService)
// 	userHandler := handlers.NewUserService(updateUserService, listCasesService)

// 	// Build main handler struct
// 	handler := handlers.NewHandler(
// 		adminService,
// 		authHandler, // Changed from authServices to authHandler
// 		caseHandler,
// 		evidenceHandler,
// 		userHandler,
// 	)
// 	// Set up the router with the main handler

// 	router := routes.SetUpRouter(handler)

// 	log.Println("Starting AEGIS server on :8080...")
// 	log.Println("Docs available at http://localhost:8080/swagger/index.html")

// 	if err := router.Run(":8080"); err != nil {
// 		log.Fatal("Failed to start server:", err)
// 	}

// }
