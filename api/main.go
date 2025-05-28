package main

import (
	database "aegis-api/db"
	"aegis-api/routes"
	"aegis-api/handlers"
	"log"


	// services & repos
    "aegis-api/services/registration"
    "aegis-api/services/ListUsers"
    "aegis-api/services/update_user_role"
    "aegis-api/services/delete_user"
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

// Now database.DB should be initialized and non-nil

// ...

	  // 2. Build repositories
    // ↳ Make sure these constructors exist in their packages!
    regRepo         := registration.NewGormUserRepository(database.DB)         // or NewGormUserRepo
    listUserRepo    := ListUsers.NewUserRepository(database.DB)
    updateRoleRepo  := update_user_role.NewGormUserRepo(database.DB)
    deleteUserRepo  := delete_user.NewGormUserRepository(database.DB)

    // 3. Build services
    regService      := registration.NewRegistrationService(regRepo)
    listUserService := ListUsers.NewListUserService(listUserRepo)
    updateRoleSvc   := update_user_role.NewUserService(updateRoleRepo)
    deleteUserSvc   := delete_user.NewUserDeleteService(deleteUserRepo)

    // 4. Build your AdminService (implements AdminServiceInterface)
    adminSvc := handlers.NewAdminService(
        regService,
        listUserService,
        updateRoleSvc,
        deleteUserSvc,
    )

    // 5. Build the other (mock or real) services
    authSvc := &handlers.MockAuthService{}
    caseSvc := &handlers.MockCaseService{}
    evidSvc := &handlers.MockEvidenceService{}
    userSvc := &handlers.MockUserService{}

    // 6. Assemble the top-level Handler
    handler := handlers.NewHandler(
        adminSvc,
        authSvc,
        caseSvc,
        evidSvc,
        userSvc,
    )

    // 7. Wire it into your router
    router := routes.SetUpRouter(handler)
	log.Println("Starting AEGIS server on :8080...")
	log.Println("Docs available at http://localhost:8080/swagger/index.html")

	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
