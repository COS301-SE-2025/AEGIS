package main

import (
	"fmt"
	"log"
	"net/http"
	"aegis-api/router"
	database "aegis-api/db"

)

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, `{"status": "UP"}`)
}

func main() {
	// 🔌 Initialize DB connection
	if err := database.InitDB(); err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	http.HandleFunc("/health", healthCheckHandler)
	
	router.StartServer()
	// 🚀 Start server
	fmt.Println("✅ Server running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("❌ Server failed: %v", err)
	}
}
