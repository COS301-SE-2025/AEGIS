package integration

// import (
// 	"bytes"
// 	"encoding/json"
// 	"net/http"
// 	"testing"

// 	"github.com/stretchr/testify/require"
// )

// func TestCreateCase_Success(t *testing.T) {
// 	payload := map[string]interface{}{
// 		"title":               "Case from Benjamin",
// 		"description":         "Testing create case endpoint",
// 		"status":              "open",
// 		"priority":            "high",
// 		"investigation_stage": "analysis",
// 		"created_by":          "b5e19c4b-69a0-4604-add1-6e603b85ea85", // your real user ID
// 		"team_name":           "Team Benjamin",
// 	}

// 	body, _ := json.Marshal(payload)

// 	req, _ := http.NewRequest("POST", "http://localhost:8080/api/v1/cases", bytes.NewBuffer(body))
// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InRlc3R1c2VyQGV4YW1wbGUuY29tIiwiZXhwIjoxNzUyMDkzNDM0LCJpYXQiOjE3NTIwMDcwMzQsInJvbGUiOiJBZG1pbiIsInRva2VuX3ZlcnNpb24iOjEsInVzZXJfaWQiOiJiNWUxOWM0Yi02OWEwLTQ2MDQtYWRkMS02ZTYwM2I4NWVhODUifQ.R73jseNjkkXFGaVXChgLixTvg3zkIaZlMX4uheicoPY")

// 	resp, err := http.DefaultClient.Do(req)
// 	require.NoError(t, err)
// 	require.Equal(t, http.StatusCreated, resp.StatusCode)
// }

// func TestCreateCase_MissingTitle(t *testing.T) {
// 	payload := map[string]interface{}{
// 		"description":         "Should fail with missing title",
// 		"status":              "open",
// 		"priority":            "medium",
// 		"investigation_stage": "initial",
// 		"created_by":          "b5e19c4b-69a0-4604-add1-6e603b85ea85",
// 		"team_name":           "Team Gamma",
// 	}

// 	body, _ := json.Marshal(payload)

// 	req, _ := http.NewRequest("POST", "http://localhost:8080/api/v1/cases", bytes.NewBuffer(body))
// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InRlc3R1c2VyQGV4YW1wbGUuY29tIiwiZXhwIjoxNzUyMDkzNDM0LCJpYXQiOjE3NTIwMDcwMzQsInJvbGUiOiJBZG1pbiIsInRva2VuX3ZlcnNpb24iOjEsInVzZXJfaWQiOiJiNWUxOWM0Yi02OWEwLTQ2MDQtYWRkMS02ZTYwM2I4NWVhODUifQ.R73jseNjkkXFGaVXChgLixTvg3zkIaZlMX4uheicoPY")

// 	resp, err := http.DefaultClient.Do(req)
// 	require.NoError(t, err)
// 	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
// }

// func TestCreateCase_MissingTeamName(t *testing.T) {
// 	payload := map[string]interface{}{
// 		"title":               "Case without team name",
// 		"description":         "Should fail with missing team_name",
// 		"status":              "open",
// 		"priority":            "low",
// 		"investigation_stage": "analysis",
// 		"created_by":          "b5e19c4b-69a0-4604-add1-6e603b85ea85",
// 		// no team_name
// 	}

// 	body, _ := json.Marshal(payload)

// 	req, _ := http.NewRequest("POST", "http://localhost:8080/api/v1/cases", bytes.NewBuffer(body))
// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InRlc3R1c2VyQGV4YW1wbGUuY29tIiwiZXhwIjoxNzUyMDkzNDM0LCJpYXQiOjE3NTIwMDcwMzQsInJvbGUiOiJBZG1pbiIsInRva2VuX3ZlcnNpb24iOjEsInVzZXJfaWQiOiJiNWUxOWM0Yi02OWEwLTQ2MDQtYWRkMS02ZTYwM2I4NWVhODUifQ.R73jseNjkkXFGaVXChgLixTvg3zkIaZlMX4uheicoPY")

// 	resp, err := http.DefaultClient.Do(req)
// 	require.NoError(t, err)
// 	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
// }
