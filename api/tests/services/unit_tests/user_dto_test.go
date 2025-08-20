package unit_tests

import (
	"encoding/json"
	"testing"
	"time"

	"aegis-api/services_/user/GetUpdate_UserInfo"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUserDTO_JSONSerialization(t *testing.T) {
	now := time.Now()
	id := uuid.New()

	user := GetUpdate_UserInfo.User{
		ID:           id,
		FullName:     "John Doe",
		Email:        "john@example.com",
		PasswordHash: "secrethash",
		Role:         "admin",
		IsVerified:   true,
		CreatedAt:    now,
	}

	jsonData, err := json.Marshal(user)
	assert.NoError(t, err)

	// PasswordHash should NOT be included in JSON output due to json:"-" tag
	assert.NotContains(t, string(jsonData), "secrethash")
	assert.NotContains(t, string(jsonData), "password_hash")

	// These fields should be present
	assert.Contains(t, string(jsonData), "John Doe")
	assert.Contains(t, string(jsonData), "john@example.com")
	assert.Contains(t, string(jsonData), "admin")
}

func TestUserDTO_JSONDeserialization(t *testing.T) {
	jsonStr := `{
		"id": "550e8400-e29b-41d4-a716-446655440000",
		"full_name": "John Doe",
		"email": "john@example.com",
		"role": "user",
		"is_verified": true,
		"created_at": "2023-01-01T12:00:00Z"
	}`

	var user GetUpdate_UserInfo.User
	err := json.Unmarshal([]byte(jsonStr), &user)
	assert.NoError(t, err)

	expectedID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	assert.Equal(t, expectedID, user.ID)
	assert.Equal(t, "John Doe", user.FullName)
	assert.Equal(t, "john@example.com", user.Email)
	assert.Equal(t, "user", user.Role)
	assert.True(t, user.IsVerified)
	assert.Equal(t, "", user.PasswordHash) // Should be empty as it's not in the JSON
}