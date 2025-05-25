package services

import (
	"testing"
	"aegis-api/db"
	"aegis-api/services/login/auth"
)

func init() {
	if err := db.Connect(); err != nil {
		panic("❌ DB connection failed: " + err.Error())
	}
}

func TestLoginSuccess(t *testing.T) {
	email := "roy@aegis.dev"
	password := "Fireal@chemist123" // This must match the stored hashed password

	resp, err := auth.Login(email, password)
	if err != nil {
		t.Fatalf("❌ Login failed: %v", err)
	}

	t.Logf("✅ Logged in user: %s with token: %s", resp.Email, resp.Token)
}

func TestLoginInvalidPassword(t *testing.T) {
	email := "roy@aegis.dev"
	password := "wrongpassword"

	_, err := auth.Login(email, password)
	if err == nil {
		t.Fatalf("❌ Login should have failed with wrong password")
	}

	t.Logf("✅ Correctly failed login with invalid password: %v", err)
}

func TestLoginUnknownUser(t *testing.T) {
	email := "unknown@aegis.dev"
	password := "doesntmatter"

	_, err := auth.Login(email, password)
	if err == nil {
		t.Fatalf("❌ Login should have failed for non-existent user")
	}

	t.Logf("✅ Correctly failed login for unknown user: %v", err)
}

func TestLoginEmptyPassword(t *testing.T) {
	email := "roy@aegis.dev"
	password := ""

	_, err := auth.Login(email, password)
	if err == nil {
		t.Fatalf("❌ Should have failed with empty password")
	}
	t.Logf("✅ Correctly failed login with empty password")
}
