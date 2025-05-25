package services

import (
	"testing"
	"aegis-api/db"
	"aegis-api/services/registration"
)

func init() {
	// Ensure DB connection for GormUserRepository
	if err := db.Connect(); err != nil {
		panic("❌ Failed to connect to DB: " + err.Error())
	}
}

func TestRegisterValidUser(t *testing.T) {
	repo := registration.NewGormUserRepository(db.DB)
	service := registration.NewRegistrationService(repo)

	req := registration.RegistrationRequest{
		FullName: "Test User",
		Email:    "testuser@aegis.dev",
		Password: "Secure123",
		Role:     "Generic user",
	}

	err := req.Validate()
	if err != nil {
		t.Fatalf("❌ Validation failed unexpectedly: %v", err)
	}

	user, err := service.Register(req)

	if err != nil {
		if err.Error() == "user already exists" {
			t.Logf("ℹ️ User already registered, which is acceptable for this test.")
			return
		}
		t.Fatalf("❌ Registration failed: %v", err)
	}

	t.Logf("✅ Successfully registered user: %s (%s)", user.Email, user.ID)
}

func TestRegisterDuplicateUser(t *testing.T) {
	repo := registration.NewGormUserRepository(db.DB)
	service := registration.NewRegistrationService(repo)

	req := registration.RegistrationRequest{
		FullName: "Roy Mustang",
		Email:    "roy@aegis.dev",
		Password: "Fireal@chemist123",
		Role:     "DFIR Manager",
	}

	// First attempt (if not already registered)
	_, _ = service.Register(req)

	// Second attempt should fail
	_, err := service.Register(req)
	if err == nil {
		t.Fatalf("❌ Expected failure due to duplicate email")
	}

	t.Logf("✅ Correctly failed duplicate registration: %v", err)
}

func TestRegisterWeakPassword(t *testing.T) {
	req := registration.RegistrationRequest{
		FullName: "Weak Pass",
		Email:    "weakpass@aegis.dev",
		Password: "abc", // too weak
		Role:     "Incident Responder",
	}

	err := req.Validate()
	if err == nil {
		t.Fatalf("❌ Expected validation error for weak password")
	}

	t.Logf("✅ Correctly rejected weak password: %v", err)
}

func TestRegisterInvalidRole(t *testing.T) {
	req := registration.RegistrationRequest{
		FullName: "Jane Unknown",
		Email:    "jane@aegis.dev",
		Password: "Valid123",
		Role:     "God", // Invalid role
	}

	err := req.Validate()
	if err == nil {
		t.Fatalf("❌ Expected validation error for invalid role")
	}

	t.Logf("✅ Correctly rejected invalid role: %v", err)
}
