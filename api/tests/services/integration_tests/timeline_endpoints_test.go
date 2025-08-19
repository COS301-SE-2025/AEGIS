package integration_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_CreateTimelineEvent_And_ListByCase(t *testing.T) {
	// Create a case to attach the timeline event to
	title := "Timeline Case " + time.Now().Format(time.RFC3339Nano)
	caseBody := fmt.Sprintf(`{
		"title": %q,
		"description": "timeline test",
		"team_name": "test-team",
		"status": "open"
	}`, title)
	w := doRequest("POST", "/cases", caseBody)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	resp := decodeJSON(t, w.Body.Bytes())
	caseID := resp["id"].(string)
	require.NotEmpty(t, caseID)

	// Create a timeline event
	eventBody := fmt.Sprintf(`{
		"case_id": %q,
		"title": "Event 1",
		"description": "First event",
		"timestamp": %q
	}`, caseID, time.Now().Format(time.RFC3339))
	w = doRequest("POST", "/cases/"+caseID+"/timeline", eventBody)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	eventResp := decodeJSON(t, w.Body.Bytes())
	eventID := eventResp["id"].(string)
	require.NotEmpty(t, eventID)
	require.Equal(t, "First event", eventResp["description"]) // Changed from "title" to "description"

	// List timeline events for the case
	w = doRequest("GET", "/cases/"+caseID+"/timeline", "")
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var events []map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &events)
	require.NoError(t, err, "Failed to unmarshal timeline events list")
	require.NotEmpty(t, events)
}

func Test_TimelineEvent_Validation(t *testing.T) {
	// Try to create a timeline event with missing fields
	body := `{"case_id": "", "title": ""}`
	w := doRequest("POST", "/cases/invalid-id/timeline", body)
	require.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
}

func Test_UpdateTimelineEvent(t *testing.T) {
	// Create a case and event first
	title := "Update Timeline Case " + time.Now().Format(time.RFC3339Nano)
	caseBody := fmt.Sprintf(`{
		"title": %q,
		"description": "timeline test",
		"team_name": "test-team",
		"status": "open"
	}`, title)
	w := doRequest("POST", "/cases", caseBody)
	resp := decodeJSON(t, w.Body.Bytes())
	caseID := resp["id"].(string)

	eventBody := fmt.Sprintf(`{
		"case_id": %q,
		"title": "Event to Update",
		"description": "Update me",
		"timestamp": %q
	}`, caseID, time.Now().Format(time.RFC3339))
	w = doRequest("POST", "/cases/"+caseID+"/timeline", eventBody)
	eventResp := decodeJSON(t, w.Body.Bytes())
	eventID := eventResp["id"].(string)

	// Update the event
	updateBody := `{"title": "Updated Event", "description": "Updated desc"}`
	w = doRequest("PATCH", "/timeline/"+eventID, updateBody)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	updated := decodeJSON(t, w.Body.Bytes())
	require.Equal(t, "Updated desc", updated["description"]) // Changed from "title" to "description"
}

func Test_DeleteTimelineEvent(t *testing.T) {
	// Create a case and event first
	title := "Delete Timeline Case " + time.Now().Format(time.RFC3339Nano)
	caseBody := fmt.Sprintf(`{
		"title": %q,
		"description": "timeline test",
		"team_name": "test-team",
		"status": "open"
	}`, title)
	w := doRequest("POST", "/cases", caseBody)
	resp := decodeJSON(t, w.Body.Bytes())
	caseID := resp["id"].(string)

	eventBody := fmt.Sprintf(`{
		"case_id": %q,
		"title": "Event to Delete",
		"description": "Delete me",
		"timestamp": %q
	}`, caseID, time.Now().Format(time.RFC3339))
	w = doRequest("POST", "/cases/"+caseID+"/timeline", eventBody)
	eventResp := decodeJSON(t, w.Body.Bytes())
	eventID := eventResp["id"].(string)

	// Delete the event
	w = doRequest("DELETE", "/timeline/"+eventID, "")
	require.Equal(t, http.StatusNoContent, w.Code, w.Body.String())
}

func Test_ReorderTimelineEvents(t *testing.T) {
	// Create a case and two events
	title := "Reorder Timeline Case " + time.Now().Format(time.RFC3339Nano)
	caseBody := fmt.Sprintf(`{
		"title": %q,
		"description": "timeline test",
		"team_name": "test-team",
		"status": "open"
	}`, title)
	w := doRequest("POST", "/cases", caseBody)
	resp := decodeJSON(t, w.Body.Bytes())
	caseID := resp["id"].(string)

	eventBody1 := fmt.Sprintf(`{
		"case_id": %q,
		"title": "Event 1",
		"description": "First event",
		"timestamp": %q
	}`, caseID, time.Now().Format(time.RFC3339))
	w = doRequest("POST", "/cases/"+caseID+"/timeline", eventBody1)
	eventResp1 := decodeJSON(t, w.Body.Bytes())
	eventID1 := eventResp1["id"].(string)

	eventBody2 := fmt.Sprintf(`{
		"case_id": %q,
		"title": "Event 2",
		"description": "Second event",
		"timestamp": %q
	}`, caseID, time.Now().Add(time.Minute).Format(time.RFC3339))
	w = doRequest("POST", "/cases/"+caseID+"/timeline", eventBody2)
	eventResp2 := decodeJSON(t, w.Body.Bytes())
	eventID2 := eventResp2["id"].(string)

	// Reorder events
	reorderBody := fmt.Sprintf(`{"ordered_ids": [%q, %q]}`, eventID2, eventID1) // Fix field name
	w = doRequest("POST", "/cases/"+caseID+"/timeline/reorder", reorderBody)
	require.Equal(t, http.StatusNoContent, w.Code, w.Body.String()) //
}

func Test_CreateTimelineEvent_MissingDescription(t *testing.T) {
	// Create a case
	title := "Timeline Case " + time.Now().Format(time.RFC3339Nano)
	caseBody := fmt.Sprintf(`{
		"title": %q,
		"description": "timeline test",
		"team_name": "test-team",
		"status": "open"
	}`, title)
	w := doRequest("POST", "/cases", caseBody)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	resp := decodeJSON(t, w.Body.Bytes())
	caseID := resp["id"].(string)

	// Try to create event with missing description
	eventBody := fmt.Sprintf(`{
		"case_id": %q,
		"title": "Event Missing Description"
	}`, caseID)
	w = doRequest("POST", "/cases/"+caseID+"/timeline", eventBody)
	require.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
}

func Test_GetTimelineEvents_NonExistentCase(t *testing.T) {
	nonExistentCaseID := "00000000-0000-0000-0000-000000000000"
	w := doRequest("GET", "/cases/"+nonExistentCaseID+"/timeline", "")
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var events []map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &events)
	require.NoError(t, err)
	require.Empty(t, events)
}

func Test_UpdateTimelineEvent_NotFound(t *testing.T) {
	nonExistentEventID := "00000000-0000-0000-0000-000000000000"
	updateBody := `{"description": "Should not update"}`
	w := doRequest("PATCH", "/timeline/"+nonExistentEventID, updateBody)
	require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
}

func Test_DeleteTimelineEvent_NotFound(t *testing.T) {
	nonExistentEventID := "00000000-0000-0000-0000-000000000000"
	w := doRequest("DELETE", "/timeline/"+nonExistentEventID, "")
	require.Equal(t, http.StatusNoContent, w.Code, w.Body.String())
}

func Test_ReorderTimelineEvents_InvalidBody(t *testing.T) {
	// Create a case
	title := "Reorder Invalid Body Case " + time.Now().Format(time.RFC3339Nano)
	caseBody := fmt.Sprintf(`{
		"title": %q,
		"description": "timeline test",
		"team_name": "test-team",
		"status": "open"
	}`, title)
	w := doRequest("POST", "/cases", caseBody)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	resp := decodeJSON(t, w.Body.Bytes())
	caseID := resp["id"].(string)

	// Send invalid reorder body
	reorderBody := `{"bad_field": ["id1", "id2"]}`
	w = doRequest("POST", "/cases/"+caseID+"/timeline/reorder", reorderBody)
	require.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
}
