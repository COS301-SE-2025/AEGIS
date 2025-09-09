package integration_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// helper: insert a minimal case row directly (avoids depending on other handlers)
func insertCaseRow(t *testing.T, id uuid.UUID) {
	t.Helper()
	_, err := pgSQL.Exec(`
		INSERT INTO cases (id, title, team_name, created_by, tenant_id, team_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (id) DO NOTHING
	`, id, "AssignCase "+id.String()[:8], "test-team", FixedUserID, FixedTenantID, FixedTeamID)
	require.NoError(t, err)
}

// helper: count mapping rows
func countMapping(t *testing.T, userID, caseID uuid.UUID) int {
	t.Helper()
	var n int
	err := pgSQL.QueryRow(`SELECT COUNT(*) FROM case_user_roles WHERE user_id=$1 AND case_id=$2`, userID, caseID).Scan(&n)
	require.NoError(t, err)
	return n
}

// helper: fetch the persisted team_id for a mapping
func getMappingTeamID(t *testing.T, userID, caseID uuid.UUID) uuid.UUID {
	t.Helper()
	var teamIDStr string
	err := pgSQL.QueryRow(
		`SELECT team_id::text FROM case_user_roles WHERE user_id=$1 AND case_id=$2`,
		userID, caseID,
	).Scan(&teamIDStr)
	require.NoError(t, err)
	got, err := uuid.Parse(teamIDStr)
	require.NoError(t, err)
	return got
}

func Test_CaseAssign_Then_Unassign(t *testing.T) {
	caseID := uuid.New()
	insertCaseRow(t, caseID)

	role := "Forensic Analyst" // must exist in user_role enum

	// --- Assign
	body := fmt.Sprintf(`{"user_id":%q,"case_id":%q,"role":%q}`, FixedUserID.String(), caseID.String(), role)
	w := doRequest("POST", "/cases/assign", body)
	require.True(t, w.Code == http.StatusCreated || w.Code == http.StatusOK, w.Body.String())

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, FixedUserID.String(), resp["user_id"])
	require.Equal(t, caseID.String(), resp["case_id"])
	require.Equal(t, role, resp["role"])

	// verify in DB
	require.Equal(t, 1, countMapping(t, FixedUserID, caseID))

	// NEW: verify team_id persisted correctly
	gotTeam := getMappingTeamID(t, FixedUserID, caseID)
	require.Equal(t, FixedTeamID, gotTeam, "team_id should match FixedTeamID")

	// --- Duplicate assign -> 409
	w = doRequest("POST", "/cases/assign", body)
	require.Equal(t, http.StatusConflict, w.Code, w.Body.String())

	// still only one row
	require.Equal(t, 1, countMapping(t, FixedUserID, caseID))

	// --- Unassign
	w = doRequest("POST", "/cases/unassign", fmt.Sprintf(`{"user_id":%q,"case_id":%q}`, FixedUserID.String(), caseID.String()))
	require.True(t, w.Code == http.StatusNoContent || w.Code == http.StatusOK, w.Body.String())

	// verify gone
	require.Equal(t, 0, countMapping(t, FixedUserID, caseID))
}

func Test_CaseAssign_InvalidIDs(t *testing.T) {
	// bad UUIDs
	w := doRequest("POST", "/cases/assign", `{"user_id":"nope","case_id":"also-bad","role":"Forensic Analyst"}`)
	require.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
}

func Test_Unassign_NotPresent(t *testing.T) {
	caseID := uuid.New()
	insertCaseRow(t, caseID)

	// ensure no existing mapping
	n := countMapping(t, FixedUserID, caseID)
	require.Equal(t, 0, n)

	// unassign should be idempotent in our test handler: 204 if nothing there after op
	w := doRequest("POST", "/cases/unassign", fmt.Sprintf(`{"user_id":%q,"case_id":%q}`, FixedUserID.String(), caseID.String()))
	require.True(t, w.Code == http.StatusNoContent || w.Code == http.StatusOK, w.Body.String())

	// still none
	require.Equal(t, 0, countMapping(t, FixedUserID, caseID))
}

// optional: sanity to ensure FK exists so insert fails for non-existent case
func Test_Assign_Fails_When_Case_Missing(t *testing.T) {
	missingCase := uuid.New() // not inserted
	w := doRequest("POST", "/cases/assign", fmt.Sprintf(`{"user_id":%q,"case_id":%q,"role":"Forensic Analyst"}`, FixedUserID.String(), missingCase.String()))
	// Depending on FK settings, this may be 400 with pq error text; accept 4xx/5xx but assert it failed.
	require.True(t, w.Code >= 400, w.Body.String())
}
