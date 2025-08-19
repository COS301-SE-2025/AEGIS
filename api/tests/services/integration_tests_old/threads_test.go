package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	// baseURL = "http://localhost:8080/api/v1"
	// token   = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Im5ld2FkbWluQGV4YW1wbGUuY29tIiwiZXhwIjoxNzUyMjMxMzUwLCJpYXQiOjE3NTIxNDQ5NTAsInJvbGUiOiJBZG1pbiIsInRva2VuX3ZlcnNpb24iOjEsInVzZXJfaWQiOiI3NmE1NTUxMC01M2QyLTQ0ZTYtODFjNC1kMzI2ZTcyNTNiMmMifQ.o2N0Iv_XfzHtoH7AS5tqvE3Npl1JV3FviKHhB0qVsCM" // your JWT
	// userID  = "27031538-2795-4095-9adf-59bb7bd3fc19"
	caseID = "08bffdb7-a74c-47c8-8bbf-f4df30b6bd54"
	fileID = "7b0aaeee-aad5-43d9-b826-9c1dba392628"
)

func TestCreateThreadIntegration(t *testing.T) {
	payload := map[string]interface{}{
		"case_id":  caseID,
		"file_id":  fileID,
		"user_id":  userID,
		"title":    "Integration Test Thread",
		"tags":     []string{"test", "integration"},
		"priority": "medium",
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", baseURL+"/threads", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	defer resp.Body.Close()
	var response map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&response)

	threadID := response["id"].(string)
	require.NotEmpty(t, threadID)

	// follow-up tests on this thread
	testUpdateStatus(t, threadID)
	testUpdatePriority(t, threadID)
	testAddParticipant(t, threadID)
	testGetParticipants(t, threadID)
}

func testUpdateStatus(t *testing.T, threadID string) {
	payload := map[string]interface{}{
		"status":  "resolved",
		"user_id": userID,
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("PATCH", baseURL+"/threads/"+threadID+"/status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func testUpdatePriority(t *testing.T, threadID string) {
	payload := map[string]interface{}{
		"priority": "high",
		"user_id":  userID,
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("PATCH", baseURL+"/threads/"+threadID+"/priority", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func testAddParticipant(t *testing.T, threadID string) {
	payload := map[string]interface{}{
		"user_id": userID,
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", baseURL+"/threads/"+threadID+"/participants", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func testGetParticipants(t *testing.T, threadID string) {
	req, _ := http.NewRequest("GET", baseURL+"/threads/"+threadID+"/participants", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	defer resp.Body.Close()
	var participants []map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&participants)
	require.GreaterOrEqual(t, len(participants), 1)
}
