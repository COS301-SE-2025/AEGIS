package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	//DBUser "aegis-api/services_/user/profile"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type DBUser struct {
	ID                string
	FullName          string
	Email             string
	ProfilePictureURL string
}

func insertTestUserDirect(t *testing.T, name, email string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	var userID string
	err = db.QueryRow(`
		SELECT id FROM users WHERE email = $1
	`, email).Scan(&userID)

	if err == sql.ErrNoRows {
		err = db.QueryRow(`
			INSERT INTO users (full_name, email, password_hash, role, is_verified)
			VALUES ($1, $2, $3, 'Admin', TRUE)
			RETURNING id
		`, name, email, string(hashedPassword)).Scan(&userID)
		require.NoError(t, err)
	} else {
		require.NoError(t, err)
	}

	return userID
}

func cleanDB(t *testing.T) {
	_, err := db.Exec("TRUNCATE users RESTART IDENTITY CASCADE")
	require.NoError(t, err)
}

func getUserDirect(t *testing.T, userID string) DBUser {

	var user DBUser
	err := db.QueryRow(`
		SELECT id, full_name, email, profile_picture_url
		FROM users
		WHERE id = $1
	`, userID).Scan(&user.ID, &user.FullName, &user.Email, &user.ProfilePictureURL)
	require.NoError(t, err)
	return user
}
func loginAndGetToken(t *testing.T, email, password string) string {
	payload := map[string]interface{}{
		"email":    email,
		"password": password,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "http://localhost:8080/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var response struct {
		Success bool                   `json:"success"`
		Message string                 `json:"message"`
		Data    map[string]interface{} `json:"data"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	fmt.Println("Login Response Body:", response)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.True(t, response.Success)
	require.Equal(t, "Login successful", response.Message)

	token, ok := response.Data["token"].(string)
	require.True(t, ok, "token missing in response data")
	require.NotEmpty(t, token)

	return token
}

func TestGetProfile_LiveServer(t *testing.T) {
	// Insert user directly into DB
	cleanDB(t)
	userID := insertTestUserDirect(t, "Alice Live", "alice.live@example.com")

	// Get JWT by logging in
	token := loginAndGetToken(t, "alice.live@example.com", "password123")

	// Prepare GET request
	req, _ := http.NewRequest("GET", "http://localhost:8080/api/v1/profile/"+userID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Parse JSON response
	var response struct {
		Success bool                   `json:"success"`
		Message string                 `json:"message"`
		Data    map[string]interface{} `json:"data"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	// Debug output
	fmt.Println("Response Body:", response)

	// Assert HTTP & JSON structure
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.True(t, response.Success)
	require.Equal(t, "Profile retrieved successfully", response.Message)

	// Check returned data fields
	require.Equal(t, userID, response.Data["id"])
	require.Equal(t, "Alice Live", response.Data["name"])
	require.Equal(t, "alice.live@example.com", response.Data["email"])
}

func TestUpdateProfile_LiveServer(t *testing.T) {
	cleanDB(t)
	userID := insertTestUserDirect(t, "Bob Live", "bob.live@example.com")

	// Login to get valid token
	token := loginAndGetToken(t, "bob.live@example.com", "password123")

	payload := map[string]interface{}{
		"id":        userID,
		"name":      "Bob Updated",
		"email":     "bob.updated@example.com",
		"image_url": "http://live.example.com/avatar.jpg",
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("PATCH", "http://localhost:8080/api/v1/profile", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)
	fmt.Println("Response Body:", response)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.True(t, response["success"].(bool))
	require.Equal(t, "Profile updated successfully", response["message"])

	// Optionally check DB side-effects
	profile := getUserDirect(t, userID)
	require.Equal(t, "Bob Updated", profile.FullName)
	require.Equal(t, "bob.updated@example.com", profile.Email)
	require.Equal(t, "http://live.example.com/avatar.jpg", profile.ProfilePictureURL)
}
