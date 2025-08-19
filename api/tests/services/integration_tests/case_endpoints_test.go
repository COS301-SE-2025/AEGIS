package integration_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_CreateCase_And_GetByID(t *testing.T) {
	title := "Integration Case " + time.Now().Format(time.RFC3339Nano)
	body := fmt.Sprintf(`{
  "title": %q,
  "description": "from tests",
  "team_name": "test-team",
  "status": "open"
}`, title)

	w := doRequest("POST", "/cases", body)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())

	resp := decodeJSON(t, w.Body.Bytes())
	id, _ := resp["id"].(string)
	require.NotEmpty(t, id)
	require.Equal(t, title, resp["title"])

	// GET by id (optional)
	w = doRequest("GET", "/cases/"+id, "")
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
}

func Test_CreateCase_Validation(t *testing.T) {
	w := doRequest("POST", "/cases", `{"title": "", "teamName":"test-team"}`)
	require.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
}
