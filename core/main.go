package main

import (
	"fmt"
	_ "github.com/gofiber/fiber/v2"
	"net/http"
)

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := fmt.Fprintf(w, `{"status": "UP"}`)
	if err != nil {
		return
	}
}

func main() {
	http.HandleFunc("/health", healthCheckHandler)
	port := ":8080"
	fmt.Printf("Listening on port %s...\n", port)

	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}
