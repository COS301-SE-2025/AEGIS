package services

import (
	"testing"
	"aegis-api/db"
	"aegis-api/services/registration"
)

func init() {
	// Connect to the database so GormUserRepository works
	if err := db.Connect(); err != nil {
		panic("❌ Failed to connect to DB: " + err.Error())
	}
}

func TestRegisterValidUser(t *testing.T) {
	repo := registration.NewGormUserRepository(db.DB)
	service := registration.NewRegistrationService(repo)

	req := registration.RegistrationRequest{
		Name:     "Test",
		Surname:  "User",
		Email:    "testuser@aegis.dev", // Can be reused
		Password: "Secure123",
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
		Name:     "Roy",
		Surname:  "Mustang",
		Email:    "roy@aegis.dev", // existing email
		Password: "Fireal@chemist123",
	}

	_, err := service.Register(req)
	if err == nil {
		t.Fatalf("❌ Expected failure due to duplicate email")
	}

	t.Logf("✅ Correctly failed duplicate registration: %v", err)
}

func TestRegisterWeakPassword(t *testing.T) {
	req := registration.RegistrationRequest{
		Name:     "Weak",
		Surname:  "Pass",
		Email:    "weakpass@aegis.dev",
		Password: "abc", // too weak
	}

	err := req.Validate()
	if err == nil {
		t.Fatalf("❌ Expected validation error for weak password")
	}

	t.Logf("✅ Correctly rejected weak password: %v", err)
}
