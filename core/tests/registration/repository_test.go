package registration_test

import (
	"testing"

	"aegis-api/services/registration"
)

func TestInMemoryUserRepository_CreateAndGetUser(t *testing.T) {
	repo := registration.NewInMemoryUserRepository()

	user := &registration.UserEntity{
		ID:           "test-id",
		Name:         "Test",
		Surname:      "User",
		Email:        "test@example.com",
		PasswordHash: "hashed-password",
	}

	err := repo.CreateUser(user)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	fetched, err := repo.GetUserByEmail("test@example.com")
	if err != nil {
		t.Fatalf("GetUserByEmail failed: %v", err)
	}

	if fetched.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, fetched.Email)
	}
}
