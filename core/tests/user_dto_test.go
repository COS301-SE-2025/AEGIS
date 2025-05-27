// user_dto_test.go
package tests

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUserDTO_JSONSerialization(t *testing.T) {
	now := time.Now()
	user := UserDTO{
		ID:                "123",
		FullName:          "John Doe",
		Email:             "john@example.com",
		PasswordHash:      "secrethash",
		Role:              "admin",
		IsVerified:        true,
		VerificationToken: "token123",
		CreatedAt:         now,
	}
	
	jsonData, err := json.Marshal(user)
	assert.NoError(t, err)
	
	// Password hash should be excluded from JSON
	assert.NotContains(t, string(jsonData), "secrethash")
	assert.NotContains(t, string(jsonData), "password_hash")
	
	// Other fields should be present
	assert.Contains(t, string(jsonData), "John Doe")
	assert.Contains(t, string(jsonData), "john@example.com")
	assert.Contains(t, string(jsonData), "admin")
	assert.Contains(t, string(jsonData), "token123")
}

func TestUserDTO_JSONDeserialization(t *testing.T) {
	jsonStr := `{
		"id": "123",
		"full_name": "John Doe",
		"email": "john@example.com",
		"role": "user",
		"is_verified": true,
		"verification_token": "token123",
		"created_at": "2023-01-01T12:00:00Z"
	}`
	
	var user UserDTO
	err := json.Unmarshal([]byte(jsonStr), &user)
	assert.NoError(t, err)
	
	assert.Equal(t, "123", user.ID)
	assert.Equal(t, "John Doe", user.FullName)
	assert.Equal(t, "john@example.com", user.Email)
	assert.Equal(t, "user", user.Role)
	assert.True(t, user.IsVerified)
	assert.Equal(t, "token123", user.VerificationToken)
	assert.Equal(t, "", user.PasswordHash) // Should be empty as it's not in JSON
}

func TestUserDTO_EmptyVerificationToken(t *testing.T) {
	user := UserDTO{
		ID:                "123",
		FullName:          "John Doe",
		Email:             "john@example.com",
		Role:              "user",
		IsVerified:        true,
		VerificationToken: "", // Empty token
		CreatedAt:         time.Now(),
	}
	
	jsonData, err := json.Marshal(user)
	assert.NoError(t, err)
	
	// Empty verification token should be omitted from JSON
	assert.NotContains(t, string(jsonData), "verification_token")
}