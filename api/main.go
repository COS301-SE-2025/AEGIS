package main

import (
	database "aegis-api/db"
	"aegis-api/routes"
	"log"
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
	router := routes.SetUpRouter()

	log.Println("Starting AEGIS server on :8080...")
	log.Println("Docs available at http://localhost:8080/swagger/index.html")

	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
