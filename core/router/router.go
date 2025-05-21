package router

import (
	"log"
	"net/http"

	"AEGIS/core/services/registration"
)

func StartServer() {
	
	http.HandleFunc("/api/register", registration.RegisterHandler(service))
	http.HandleFunc("/verify", registration.VerifyHandler(repo))

	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
