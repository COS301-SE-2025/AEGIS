package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"aegis-api/services/login/auth" // Import the actual package where types are defined
	
)

func TestLoginHandler(t *testing.T) {
	// Create a mock request for login
	req := auth.LoginRequest{
		Email:    "roy@aegis.dev",
		Password: "Fireal@chemist123",
	}

	// Convert the request into JSON format
	reqBody, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new HTTP request with the login endpoint
	request := httptest.NewRequest("POST", "/login", bytes.NewReader(reqBody))
	recorder := httptest.NewRecorder()

	// Create the login handler
	handler := http.HandlerFunc(auth.LoginHandler)

	// Call the handler with the request and recorder
	handler.ServeHTTP(recorder, request)

	// Check the status code
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, recorder.Code)
	}

	// Check the response body
	var resp auth.LoginResponse
	err = json.NewDecoder(recorder.Body).Decode(&resp)
	if err != nil {
		t.Fatal(err)
	}

	// Check if the response contains the correct values
	if resp.Email != req.Email {
		t.Errorf("Expected email %s, got %s", req.Email, resp.Email)
	}
	if resp.Token == "" {
		t.Error("Expected non-empty token, got empty")
	}
}
