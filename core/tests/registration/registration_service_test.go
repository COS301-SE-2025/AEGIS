package registration_test

import (
	"testing"

	"aegis-api/services/registration"
)

func TestRegistrationService_Register(t *testing.T) {
	repo := registration.NewInMemoryUserRepository()
	service := registration.NewRegistrationService(repo)

	req := registration.RegistrationRequest{
		Name:     "Tibose",
		Surname:  "Mokwena",
		Email:    "tibose@example.com",
		Password: "securePassword123",
	}

	//  Correct: capture both return values
	user, err := service.Register(req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if user.Email != req.Email {
		t.Errorf("expected returned user email %s, got %s", req.Email, user.Email)
	}

	//  Also confirm the repository contains the user
	storedUser, err := repo.GetUserByEmail(req.Email)
	if err != nil {
		t.Fatalf("user should be stored, but got error: %v", err)
	}

	if storedUser.Email != req.Email {
		t.Errorf("expected stored email %s, got %s", req.Email, storedUser.Email)
	}

	if storedUser.PasswordHash == req.Password {
		t.Errorf("password should be hashed, but was stored in plain text")
	}
}


func TestRegistrationService_DuplicateEmail(t *testing.T) {
	repo := registration.NewInMemoryUserRepository()
	service := registration.NewRegistrationService(repo)

	req := registration.RegistrationRequest{
		Name:     "Tibose",
		Surname:  "Mokwena",
		Email:    "duplicate@example.com",
		Password: "strongpass123",
	}

	_, err := service.Register(req)
	if err != nil {
		t.Fatalf("first registration should succeed: %v", err)
	}

	_, err = service.Register(req)
	if err == nil {
		t.Fatal("expected error on duplicate registration, got nil")
	}
}
