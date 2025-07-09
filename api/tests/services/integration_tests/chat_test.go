package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	baseURLChat = "http://localhost:8080/api/v1/chat"
	jwtToken    = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InUyMjY0MDAyMEB0dWtzLmNvLnphIiwiZXhwIjoxNzUyMDk1MjgzLCJpYXQiOjE3NTIwMDg4ODMsInJvbGUiOiJBZG1pbiIsInRva2VuX3ZlcnNpb24iOjEsInVzZXJfaWQiOiJkZGUzOTQwYS1jZTAzLTQwNjItODQ5Mi0wN2JhZGVmMmRmMGUifQ.wcNzEIAS9-oPtAO2zxG75TKjtZvlnub58Qs9k8vCGyo" // put a valid token here
	userEmail   = "u22640020@tuks.co.za"
)

func TestChatIntegration(t *testing.T) {
	var groupID string

	t.Run("Create Group", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":        "Integration Test Group",
			"description": "Group created by integration test",
			"type":        "group",
			"created_by":  userEmail,
			"members":     []map[string]interface{}{}, // important fix
		}

		body, _ := json.Marshal(payload)
		resp, err := doRequest("POST", baseURLChat+"/groups", body)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var respBody map[string]interface{}
		readJSON(resp, &respBody)
		groupID = respBody["id"].(string)
		require.NotEmpty(t, groupID)
	})

	t.Run("Get Group By ID", func(t *testing.T) {
		resp, err := doRequest("GET", fmt.Sprintf("%s/groups/%s", baseURLChat, groupID), nil)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var group map[string]interface{}
		readJSON(resp, &group)
		require.Equal(t, "Integration Test Group", group["name"])
	})

	t.Run("Add Member", func(t *testing.T) {
		payload := map[string]interface{}{
			"user_email": "newmember@example.com",
			"role":       "member",
		}

		body, _ := json.Marshal(payload)
		resp, err := doRequest("POST", fmt.Sprintf("%s/groups/%s/members", baseURLChat, groupID), body)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Send Message", func(t *testing.T) {
		payload := map[string]interface{}{
			"sender_email": userEmail,
			"sender_name":  "Integration Tester",
			"content":      "This is a message from integration test",
			"message_type": "text",
		}

		body, _ := json.Marshal(payload)
		resp, err := doRequest("POST", fmt.Sprintf("%s/groups/%s/messages", baseURLChat, groupID), body)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var message map[string]interface{}
		readJSON(resp, &message)
		require.Equal(t, "This is a message from integration test", message["content"])
	})

	t.Run("Get Messages", func(t *testing.T) {
		resp, err := doRequest("GET", fmt.Sprintf("%s/groups/%s/messages", baseURLChat, groupID), nil)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var messages []map[string]interface{}
		readJSON(resp, &messages)
		require.NotEmpty(t, messages)
	})
}

func doRequest(method, url string, body []byte) (*http.Response, error) {
	req, _ := http.NewRequest(method, url, bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+jwtToken)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	return client.Do(req)
}

func readJSON(resp *http.Response, target interface{}) {
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)
	_ = json.Unmarshal(data, target)
}
