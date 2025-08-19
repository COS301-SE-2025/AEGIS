package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// ✅ Helper to insert case
func insertTestCase(t *testing.T) string {
	var caseID string
	err := db.QueryRow(`
		INSERT INTO cases (id, title, team_name)
		VALUES (gen_random_uuid(), 'Test Case', 'Test Team')
		RETURNING id
	`).Scan(&caseID)
	require.NoError(t, err)
	return caseID
}

// ✅ Helper to insert collaborator link
func insertCaseCollaborator(t *testing.T, caseID, userID string, role string) {
	_, err := db.Exec(`
		INSERT INTO case_user_roles (case_id, user_id, role)
		VALUES ($1, $2, $3)
	`, caseID, userID, role)
	require.NoError(t, err)
}

// ✅ The integration test
func TestGetCollaborators_LiveServer(t *testing.T) {
	cleanDB(t)

	// Insert data
	caseID := insertTestCase(t)
	user1ID := insertTestUserDirect(t, "Alice Example", "alice.collab@example.com")
	user2ID := insertTestUserDirect(t, "Bob Example", "bob.collab@example.com")

	insertCaseCollaborator(t, caseID, user1ID, "Incident Responder")
	insertCaseCollaborator(t, caseID, user2ID, "Forensic Analyst")

	// Login one of the users to get a token
	token := loginAndGetToken(t, "alice.collab@example.com", "password123")

	// Call the endpoint
	req, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:8080/api/v1/cases/%s/collaborators", caseID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var response struct {
		Success bool                     `json:"success"`
		Message string                   `json:"message"`
		Data    []map[string]interface{} `json:"data"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	fmt.Println("Response Body:", response)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.True(t, response.Success)
	require.Equal(t, "Collaborators retrieved successfully", response.Message)
	require.Len(t, response.Data, 2)

	// Check that IDs & roles are returned
	foundRoles := map[string]bool{}
	for _, collab := range response.Data {
		require.NotEmpty(t, collab["id"])
		require.NotEmpty(t, collab["full_name"])
		require.NotEmpty(t, collab["email"])
		role := collab["role"].(string)
		foundRoles[role] = true
	}

	require.True(t, foundRoles["Incident Responder"])
	require.True(t, foundRoles["Forensic Analyst"])
}

func TestGetCollaborators_InvalidCaseID(t *testing.T) {
	cleanDB(t)
	insertTestUserDirect(t, "Charlie Example", "charlie.collab@example.com") // no unused var
	token := loginAndGetToken(t, "charlie.collab@example.com", "password123")

	req, _ := http.NewRequest("GET", "http://localhost:8080/api/v1/cases/not-a-uuid/collaborators", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var response map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&response)
	fmt.Println("Response Body:", response)

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.Equal(t, "invalid_request", response["error"])
	require.Equal(t, "invalid case_id format", response["message"])
}
