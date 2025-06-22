package main

import (
	database "aegis-api/db"
	"aegis-api/handlers"
	"aegis-api/pkg/websocket"
	"aegis-api/routes"
	"log"

	// services & repos
	"aegis-api/services/ListUsers"
	"aegis-api/services/case_creation"
	"aegis-api/services/delete_user"
	"aegis-api/services/login/auth"
	"aegis-api/services/registration"
	"aegis-api/services/reset_password"
	"aegis-api/services/update_user_role"

	//"aegis-api/middleware"
	// "github.com/gin-gonic/gin"
	"aegis-api/services/case_assign"
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
	listUserRepo := ListUsers.NewUserRepository(database.DB)
	updateRoleRepo := update_user_role.NewGormUserRepo(database.DB)
	deleteUserRepo := delete_user.NewGormUserRepository(database.DB)

	resetTokenRepo := reset_password.NewGormResetTokenRepository(database.DB)
	userRepo := registration.NewGormUserRepository(database.DB)
	emailSender := reset_password.NewMockEmailSender()

	// Services
	regService := registration.NewRegistrationService(regRepo)
	listUserService := ListUsers.NewListUserService(listUserRepo)
	updateRoleService := update_user_role.NewUserService(updateRoleRepo)
	deleteUserService := delete_user.NewUserDeleteService(deleteUserRepo)

	resetService := reset_password.NewPasswordResetService(resetTokenRepo, userRepo, emailSender)
	authService := auth.NewAuthService(userRepo)

	adminService := handlers.NewAdminService(regService, listUserService, updateRoleService, deleteUserService)
	authHandler := handlers.NewAuthHandler(authService, resetService)

	// Case repos and services
	caseRepo := case_creation.NewGormCaseRepository(database.DB)
	caseService := case_creation.NewCaseService(caseRepo)

	caseAssignRepo := case_assign.NewGormCaseAssignmentRepo(database.DB)
	caseAssignService := case_assign.NewCaseAssignmentService(caseAssignRepo)

	caseHandler := handlers.NewCaseHandler(caseService, caseAssignService)

	evidenceService := &handlers.MockEvidenceService{}
	userService := &handlers.MockUserService{}

	// Build main handler struct
	handler := handlers.NewHandler(adminService, authHandler, caseHandler, evidenceService, userService)

	// Setup router from your routes package
	router := routes.SetUpRouter(handler)

	log.Println("Starting AEGIS server on :8080...")
	log.Println("Docs available at http://localhost:8080/swagger/index.html")

	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
	//websocket hub setup
	// Initialize the WebSocket hub and start its goroutine
	// This should be done after the router is set up, so it can handle WebSocket
	// connections properly.
	hub := websocket.NewHub()
	go hub.Run()
}
