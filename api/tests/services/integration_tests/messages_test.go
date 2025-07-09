package integration

// import (
// 	"bytes"
// 	"encoding/json"
// 	"net/http"
// 	"testing"

// 	"github.com/stretchr/testify/require"
// )

// var (
// 	baseURL  = "http://localhost:8080/api/v1"
// 	token    = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InUyMjY0MDAyMEB0dWtzLmNvLnphIiwiZXhwIjoxNzUyMDk1MjgzLCJpYXQiOjE3NTIwMDg4ODMsInJvbGUiOiJBZG1pbiIsInRva2VuX3ZlcnNpb24iOjEsInVzZXJfaWQiOiJkZGUzOTQwYS1jZTAzLTQwNjItODQ5Mi0wN2JhZGVmMmRmMGUifQ.wcNzEIAS9-oPtAO2zxG75TKjtZvlnub58Qs9k8vCGyo" // your real JWT
// 	userID   = "27031538-2795-4095-9adf-59bb7bd3fc19"
// 	threadID = "4b06ec77-3959-4e7b-9ec9-8d5941347a65"
// )

// func TestThreadMessagesIntegration(t *testing.T) {
// 	// 1. Send message
// 	messageID := testSendMessage(t)

// 	// 2. Get messages by thread
// 	testGetMessages(t)

// 	// 3. Approve message
// 	testApproveMessage(t, messageID)

// 	// 4. Add reaction
// 	testAddReaction(t, messageID)

// 	// 5. Remove reaction
// 	testRemoveReaction(t, messageID)
// }

// func testSendMessage(t *testing.T) string {
// 	payload := map[string]interface{}{
// 		"user_id":  userID,
// 		"message":  "Integration test message from Go test",
// 		"mentions": []string{},
// 	}
// 	body, _ := json.Marshal(payload)

// 	req, _ := http.NewRequest("POST", baseURL+"/threads/"+threadID+"/messages", bytes.NewBuffer(body))
// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("Authorization", "Bearer "+token)

// 	resp, err := http.DefaultClient.Do(req)
// 	require.NoError(t, err)
// 	require.Equal(t, http.StatusOK, resp.StatusCode)

// 	var result map[string]interface{}
// 	_ = json.NewDecoder(resp.Body).Decode(&result)
// 	resp.Body.Close()

// 	messageID := result["ID"].(string)
// 	require.NotEmpty(t, messageID)
// 	t.Logf("✅ Created message ID: %s", messageID)
// 	return messageID
// }

// func testGetMessages(t *testing.T) {
// 	req, _ := http.NewRequest("GET", baseURL+"/threads/"+threadID+"/messages", nil)
// 	req.Header.Set("Authorization", "Bearer "+token)

// 	resp, err := http.DefaultClient.Do(req)
// 	require.NoError(t, err)
// 	require.Equal(t, http.StatusOK, resp.StatusCode)

// 	var messages []map[string]interface{}
// 	_ = json.NewDecoder(resp.Body).Decode(&messages)
// 	resp.Body.Close()

// 	require.Greater(t, len(messages), 0)
// 	t.Logf("✅ Retrieved %d messages", len(messages))
// }

// func testApproveMessage(t *testing.T, messageID string) {
// 	payload := map[string]interface{}{
// 		"approver_id": userID,
// 	}
// 	body, _ := json.Marshal(payload)

// 	req, _ := http.NewRequest("POST", baseURL+"/messages/"+messageID+"/approve", bytes.NewBuffer(body))
// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("Authorization", "Bearer "+token)

// 	resp, err := http.DefaultClient.Do(req)
// 	require.NoError(t, err)
// 	require.Equal(t, http.StatusOK, resp.StatusCode)
// 	t.Logf("✅ Approved message %s", messageID)
// }

// func testAddReaction(t *testing.T, messageID string) {
// 	payload := map[string]interface{}{
// 		"user_id":  userID,
// 		"reaction": "🎉",
// 	}
// 	body, _ := json.Marshal(payload)

// 	req, _ := http.NewRequest("POST", baseURL+"/messages/"+messageID+"/reactions", bytes.NewBuffer(body))
// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("Authorization", "Bearer "+token)

// 	resp, err := http.DefaultClient.Do(req)
// 	require.NoError(t, err)
// 	require.Equal(t, http.StatusOK, resp.StatusCode)
// 	t.Logf("✅ Added reaction to message %s", messageID)
// }

// func testRemoveReaction(t *testing.T, messageID string) {
// 	req, _ := http.NewRequest("DELETE", baseURL+"/messages/"+messageID+"/reactions/"+userID, nil)
// 	req.Header.Set("Authorization", "Bearer "+token)

// 	resp, err := http.DefaultClient.Do(req)
// 	require.NoError(t, err)
// 	require.Equal(t, http.StatusOK, resp.StatusCode)
// 	t.Logf("✅ Removed reaction from message %s", messageID)
// }
