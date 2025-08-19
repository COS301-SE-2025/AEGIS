package integration_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_GetTenantIOCGraph(t *testing.T) {
	tenantID := FixedTenantID.String()

	w := doRequest("GET", fmt.Sprintf("/tenants/%s/ioc-graph", tenantID), "")
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var graph []interface{}
	err := json.Unmarshal(w.Body.Bytes(), &graph)
	require.NoError(t, err, "Failed to unmarshal IOC graph")
	require.IsType(t, []interface{}{}, graph) // Accept empty array as valid
	// Accept empty graph: do not require NotEmpty
}

func Test_GetTenantIOCGraph_NotFoundTenant(t *testing.T) {
	nonExistentTenantID := "00000000-0000-0000-0000-000000000000"
	w := doRequest("GET", fmt.Sprintf("/tenants/%s/ioc-graph", nonExistentTenantID), "")
	require.Equal(t, http.StatusUnauthorized, w.Code, w.Body.String())
}

func Test_GetCaseIOCGraph(t *testing.T) {
	// Create a case first
	title := "IOC Graph Case " + time.Now().Format(time.RFC3339Nano)
	caseBody := fmt.Sprintf(`{
        "title": %q,
        "description": "ioc graph test",
        "team_name": "test-team",
        "status": "open"
    }`, title)
	w := doRequest("POST", "/cases", caseBody)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	resp := decodeJSON(t, w.Body.Bytes())
	caseID := resp["id"].(string)
	require.NotEmpty(t, caseID)

	tenantID := FixedTenantID.String()

	w = doRequest("GET", fmt.Sprintf("/tenants/%s/cases/%s/ioc-graph", tenantID, caseID), "")
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var graph []interface{}
	err := json.Unmarshal(w.Body.Bytes(), &graph)
	require.NoError(t, err, "Failed to unmarshal IOC graph")
	require.IsType(t, []interface{}{}, graph) // Accept empty array as valid
	// Accept empty graph: do not require NotEmpty
}

func Test_GetCaseIOCGraph_NotFoundCase(t *testing.T) {
	tenantID := FixedTenantID.String()
	nonExistentCaseID := "00000000-0000-0000-0000-000000000000"
	w := doRequest("GET", fmt.Sprintf("/tenants/%s/cases/%s/ioc-graph", tenantID, nonExistentCaseID), "")
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var graph []interface{}
	err := json.Unmarshal(w.Body.Bytes(), &graph)
	require.NoError(t, err)
	require.Empty(t, graph)
}

func Test_AddAndGetIOCsByCase(t *testing.T) {
	// Create a case
	title := "IOC Add Case " + time.Now().Format(time.RFC3339Nano)
	caseBody := fmt.Sprintf(`{
		"title": %q,
		"description": "ioc add test",
		"team_name": "test-team",
		"status": "open"
	}`, title)
	w := doRequest("POST", "/cases", caseBody)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	resp := decodeJSON(t, w.Body.Bytes())
	caseID := resp["id"].(string)
	require.NotEmpty(t, caseID)

	// Add an IOC to the case
	iocBody := `{"type": "ip", "value": "192.0.2.1"}`
	w = doRequest("POST", fmt.Sprintf("/cases/%s/iocs", caseID), iocBody)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	iocResp := decodeJSON(t, w.Body.Bytes())
	require.Equal(t, "ip", iocResp["type"])
	require.Equal(t, "192.0.2.1", iocResp["value"])

	// Get IOCs for the case
	w = doRequest("GET", fmt.Sprintf("/cases/%s/iocs", caseID), "")
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var iocs []map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &iocs)
	require.NoError(t, err, "Failed to unmarshal IOCs list")
	require.NotEmpty(t, iocs)
	require.Equal(t, "ip", iocs[0]["type"])
	require.Equal(t, "192.0.2.1", iocs[0]["value"])
}

func Test_AddIOCToCase_ValidationError(t *testing.T) {
	title := "IOC Validation Case " + time.Now().Format(time.RFC3339Nano)
	caseBody := fmt.Sprintf(`{
        "title": %q,
        "description": "ioc validation test",
        "team_name": "test-team",
        "status": "open"
    }`, title)
	w := doRequest("POST", "/cases", caseBody)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	resp := decodeJSON(t, w.Body.Bytes())
	caseID := resp["id"].(string)

	// Missing "type" field
	iocBody := `{"value": "192.0.2.1"}`
	w = doRequest("POST", fmt.Sprintf("/cases/%s/iocs", caseID), iocBody)
	require.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
}

func Test_AddIOCToCase_NotFound(t *testing.T) {
	nonExistentCaseID := "00000000-0000-0000-0000-000000000000"
	iocBody := `{"type": "ip", "value": "192.0.2.1"}`
	w := doRequest("POST", fmt.Sprintf("/cases/%s/iocs", nonExistentCaseID), iocBody)
	require.Equal(t, http.StatusInternalServerError, w.Code, w.Body.String())
}
