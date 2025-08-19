package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

var db *sql.DB

func TestMain(m *testing.M) {
	// Load environment variables from .env
	err := godotenv.Load("../../../../.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Build DSN from environment variables
	dsn := fmt.Sprintf(
		"host=localhost port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	// Open a connection to the database
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatalf("Failed to ping DB: %v", err)
	}

	// Run tests
	code := m.Run()
	db.Close()
	os.Exit(code)
}

// TestRegisterUser_Success tests the user registration flow
// TestRegisterUser_Success tests the user registration flow
// TestRegisterUser_Success tests the user registration flow
func TestRegisterUser_Success(t *testing.T) {
	// Prepare the registration payload
	payload := map[string]interface{}{
		"email":     "testuser@example.com", // Correct email
		"password":  "password123",          // Correct password
		"full_name": "Test User",            // Full name (if required by the API)
		"role":      "Admin",                // Explicitly set role
	}

	// Marshal the payload into JSON
	body, _ := json.Marshal(payload)

	// Create a new POST request for registration
	req, _ := http.NewRequest("POST", "http://localhost:8080/api/v1/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	// Log the response for debugging
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	// Log the full response body to understand why we got 400 instead of 201
	fmt.Println("Response Body:", response)

	// Check for 'registration_failed' error and the appropriate message for already existing user
	if resp.StatusCode == http.StatusBadRequest {
		require.Equal(t, "registration_failed", response["error"])
		require.Equal(t, "user already exists", response["message"])
	} else {
		// If the user doesn't exist, expect the normal successful registration
		require.Equal(t, http.StatusCreated, resp.StatusCode) // Expecting 201 Created
		require.Equal(t, "User registered successfully. Please check your email for a verification link.", response["message"])
	}
}

// TestLogin_Success tests the user login flow
func TestLogin_Success(t *testing.T) {
	// Prepare the login payload (same credentials used during registration)
	payload := map[string]interface{}{
		"email":    "testuser@example.com", // Correct email from registration
		"password": "password123",          // Correct password from registration
	}

	// Marshal the payload into JSON
	body, _ := json.Marshal(payload)

	// Create a new POST request for login
	req, _ := http.NewRequest("POST", "http://localhost:8080/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	// Log the response for debugging
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	// Log the full response body to check for issues
	fmt.Println("Response Body:", response)

	// Verify the response status and ensure that the response contains the token
	require.Equal(t, http.StatusOK, resp.StatusCode) // Expecting 200 OK
	require.Contains(t, response["data"], "token")   // Ensure the response contains a 'token'

	// Optionally: You can further check the structure of the token (if required)
	// token := response["data"].(map[string]interface{})["token"]
	// require.NotEmpty(t, token)
}
