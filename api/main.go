package main

import (
	"aegis-core/routes"
	"log"
)

func main() {
	router := routes.SetUpRouter()
	log.Println("Starting AEGIS server on :8080...")

	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
