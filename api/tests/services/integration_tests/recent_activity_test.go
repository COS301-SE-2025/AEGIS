package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// const token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InRlc3R1c2VyQGV4YW1wbGUuY29tIiwiZXhwIjoxNzUzMzkzMzM1LCJpYXQiOjE3NTMzMDY5MzUsInJvbGUiOiJBZG1pbiIsInRva2VuX3ZlcnNpb24iOjEsInVzZXJfaWQiOiI5ODZjNTA5OC1kZjhjLTQ0MTUtOTgzOC1hOTc3MjI1MDk2OTQifQ.vv-oHDjKqad02JX1S6BbYO2XL0mgOZpLQN0aZL9KsIo"
// const baseURL = "http://localhost:8080/api/v1"
// const userID = "986c5098-df8c-4415-9838-a97722509694"

func TestRecentActivitiesEndpoint(t *testing.T) {
	req, _ := http.NewRequest("GET", baseURL+"/auditlogs/recent/"+userID, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result struct {
		Success bool                     `json:"success"`
		Message string                   `json:"message"`
		Data    []map[string]interface{} `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()
	require.NoError(t, err)

	require.True(t, result.Success)
	require.NotEmpty(t, result.Data)
	t.Logf(" Retrieved %d recent activity logs", len(result.Data))
}
