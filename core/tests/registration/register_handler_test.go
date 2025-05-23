package registration_test

import (
	"bytes"
	"encoding/json" 
	"net/http"
	"net/http/httptest"
	"testing"

	"aegis-api/services/registration"
)

func TestRegisterHandler_Success(t *testing.T) {
	// Arrange: Set up the in-memory repo and service
	repo := registration.NewInMemoryUserRepository()
	service := registration.NewRegistrationService(repo)
	handler := registration.RegisterHandler(service)

	// Create a valid JSON request
	body := []byte(`{
		"name": "Tibose",
		"surname": "Mokwena",
		"email": "test@example.com",
		"password": "securePassword123"
	}`)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Use ResponseRecorder to capture the response
	rr := httptest.NewRecorder()

	// Act: Call the handler
	handler(rr, req)

	// Assert: Check response
	if rr.Code != http.StatusCreated {
		t.Errorf("expected status 201 Created, got %d", rr.Code)
	}

var userResp registration.UserResponse
err := json.Unmarshal(rr.Body.Bytes(), &userResp)
if err != nil {
	t.Fatalf("invalid JSON response: %v", err)
}

if userResp.Email != "test@example.com" {
	t.Errorf("expected email test@example.com, got %s", userResp.Email)
}

}

func TestRegisterHandler_ResponseStructure(t *testing.T) {
	repo := registration.NewInMemoryUserRepository()
	service := registration.NewRegistrationService(repo)
	handler := registration.RegisterHandler(service)

	body := []byte(`{
		"name": "Ofentse",
		"surname": "Mokwena",
		"email": "test2@example.com",
		"password": "Str0ngPass123"
	}`)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", rr.Code)
	}

	// Optional: Decode the response into UserResponse
	var userResp registration.UserResponse
	err := json.Unmarshal(rr.Body.Bytes(), &userResp)
	if err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Check structure
	if userResp.ID == "" || userResp.Email != "test2@example.com" {
		t.Errorf("unexpected response: %+v", userResp)
	}
}


func TestRegisterHandler_InvalidEmail(t *testing.T) {
	repo := registration.NewInMemoryUserRepository()
	service := registration.NewRegistrationService(repo)
	handler := registration.RegisterHandler(service)

	body := []byte(`{
    "name": "Tibose",
    "surname": "Mokwena",
    "email": "not-an-email",
    "password": "securePassword123" 
}`)


	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}
func TestRegisterHandler_WeakPassword(t *testing.T) {
	repo := registration.NewInMemoryUserRepository()
	service := registration.NewRegistrationService(repo)
	handler := registration.RegisterHandler(service)

	body := []byte(`{
		"name": "Weak",
		"surname": "Password",
		"email": "weakpass@example.com",
		"password": "password"
	}`)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for weak password, got %d", rr.Code)
	}
}
