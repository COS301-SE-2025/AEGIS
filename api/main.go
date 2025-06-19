package main

import (
	database "aegis-api/db"
	"aegis-api/handlers"
	"aegis-api/routes"
	"log"

	// services & repos
	"aegis-api/services/case_assign"
	"aegis-api/services/case_creation"
	"aegis-api/services/case_status_update"
	"aegis-api/services/delete_user"
	"aegis-api/services/evidence"
	"aegis-api/services/get_collaborators"
	"aegis-api/services/listcases"
	"aegis-api/services/listclosedcases"
	"aegis-api/services/listusers"
	"aegis-api/services/login/auth"
	"aegis-api/services/registration"
	"aegis-api/services/reset_password"
	"aegis-api/services/update_user_role"
	"aegis-core/services/getupdate_users"
)

// @title AEGIS Platform API
// @version 1.0
// @description API for collaborative digital forensics investigations.
// @contact.name    AEGIS Support
// @contact.email   support@aegis-dfir.com
// @license.name    Apache 2.0
// @host            localhost:8080
// @BasePath        /api/v1
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
func main() {
	if err := database.InitDB(); err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	// Create repos
	regRepo := registration.NewGormUserRepository(database.DB)
	listUserRepo := listusers.NewUserRepository(database.DB)
	updateRoleRepo := update_user_role.NewGormUserRepo(database.DB)
	deleteUserRepo := delete_user.NewGormUserRepository(database.DB)

	resetTokenRepo := reset_password.NewGormResetTokenRepository(database.DB)
	userRepo := registration.NewGormUserRepository(database.DB)
	emailSender := reset_password.NewMockEmailSender()

	// Services
	regService := registration.NewRegistrationService(regRepo)
	listUserService := listusers.NewListUserService(listUserRepo)
	updateRoleService := update_user_role.NewUserService(updateRoleRepo)
	deleteUserService := delete_user.NewUserDeleteService(deleteUserRepo)

	resetService := reset_password.NewPasswordResetService(resetTokenRepo, userRepo, emailSender)
	authService := auth.NewAuthService(userRepo)

	adminService := handlers.NewAdminServices(regService, listUserService, updateRoleService, deleteUserService)
	authHandler := handlers.NewAuthHandler(authService, resetService)

	// Case-related services initialization
	caseRepo := case_creation.NewGormCaseRepository(database.DB)
	caseService := case_creation.NewCaseService(caseRepo)

	listCasesRepo := listcases.NewGormCaseQueryRepository(database.DB)
	listCasesService := listcases.NewListCasesService(listCasesRepo)

	caseAssignRepo := case_assign.NewGormCaseAssignmentRepo(database.DB)
	caseAssignService := case_assign.NewCaseAssignmentService(caseAssignRepo)

	caseStatusRepo := case_status_update.NewGormRepository(database.DB)
	caseStatusService := case_status_update.NewCaseStatusService(caseStatusRepo)

	collaboratorsRepo := get_collaborators.NewGormRepository(database.DB)
	collaboratorsService := get_collaborators.NewService(collaboratorsRepo)

	closedCasesRepo := ListClosedCases.NewGormRepository(database.DB)

	// Evidence repos and services initialization
	evidence.InitIPFSClient() // Initialize IPFS first
	evidenceRepo := evidence.NewMongoEvidenceRepository()
	evidenceService := evidence.NewEvidenceService(evidenceRepo)

	// User handling services
	updateUserRepo := getupdate_users.NewUserRepository(database.DB)
	updateUserService := getupdate_users.NewUserService(updateUserRepo)

	// Initialize handlers with all required services
	caseHandler := handlers.NewCaseServices(
		caseService,          // For case creation
		listCasesService,     // For listing cases
		caseStatusService,    // For updating case status
		collaboratorsService, // For collaborator operations
		caseAssignService,    // For assigning users
		caseAssignService,    // Same service handles unassignment
		closedCasesRepo,      // For closed cases operations
	)

	evidenceHandler := handlers.NewEvidenceHandler(evidenceService)
	userHandler := handlers.NewUserService(updateUserService, listCasesService)

	// Build main handler struct
	handler := handlers.NewHandler(
		adminService,
		authHandler, // Changed from authServices to authHandler
		caseHandler,
		evidenceHandler,
		userHandler,
	)
	// Set up the router with the main handler
	router := routes.SetUpRouter(handler)

	log.Println("Starting AEGIS server on :8080...")
	log.Println("Docs available at http://localhost:8080/swagger/index.html")

	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
