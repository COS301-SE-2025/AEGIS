package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestCreateCase_Success(t *testing.T) {
	payload := map[string]interface{}{
		"title":                "Case from Benjamin Full Name",
		"description":          "Testing created_by_full_name mapping",
		"status":               "open",
		"priority":             "high",
		"investigation_stage":  "analysis",
		"created_by_full_name": "Benjamin Modika", // ✅ correct field
		"team_name":            "Team Benjamin",
	}

	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "http://localhost:8080/api/v1/cases", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestCreateCase_UserNotFound(t *testing.T) {
	payload := map[string]interface{}{
		"title":                "Case With Missing User",
		"description":          "Should trigger internal server error (user not found)",
		"status":               "open",
		"priority":             "low",
		"investigation_stage":  "initial",
		"created_by_full_name": "Nonexistent User", // This user doesn't exist
		"team_name":            "Team Beta",
	}

	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "http://localhost:8080/api/v1/cases", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
