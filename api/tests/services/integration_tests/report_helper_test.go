package integration_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// --- before ---
// func execSQL(t *testing.T, q string, args ...any) { ... t.Fatalf(...) }
// func seedCoreFixtures(t *testing.T) { ... execSQL(t, ...) }

// --- after ---
func execSQLNoTB(q string, args ...any) error {
	_, err := pgSQL.Exec(q, args...)
	return err
}

func seedCoreFixtures() error {
	if err := execSQLNoTB(`INSERT INTO tenants (id, name)
	                       VALUES ($1, $2) ON CONFLICT (id) DO NOTHING`,
		FixedTenantID, "test-tenant",
	); err != nil {
		return fmt.Errorf("seed tenants: %w", err)
	}

	if err := execSQLNoTB(`INSERT INTO teams (id, team_name, tenant_id)
	                       VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING`,
		FixedTeamID, "test-team", FixedTenantID,
	); err != nil {
		return fmt.Errorf("seed teams: %w", err)
	}

	if err := execSQLNoTB(`INSERT INTO users
	      (id, full_name, email, password_hash, role, is_verified, tenant_id, team_id)
	      VALUES ($1,$2,$3,$4,'Admin',true,$5,$6)
	      ON CONFLICT (id) DO NOTHING`,
		FixedUserID, "Test User", "tester@example.com", "x", FixedTenantID, FixedTeamID,
	); err != nil {
		return fmt.Errorf("seed users: %w", err)
	}
	return nil
}

func ensureCaseRow(caseID uuid.UUID) error {
	return execSQLNoTB(`INSERT INTO cases (id, title, team_name, created_by, tenant_id)
	                    VALUES ($1,$2,$3,$4,$5) ON CONFLICT (id) DO NOTHING`,
		caseID, "Case "+caseID.String()[:8], "test-team", FixedUserID, FixedTenantID,
	)
}

// ---------------------------
// JSON helper utilities
// ---------------------------

func parseFlatOrData(b []byte) (map[string]any, error) {
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	// If payload nests under "data", unwrap it
	if data, ok := m["data"].(map[string]any); ok {
		return data, nil
	}
	return m, nil
}

func mustArrayOrItems(t *testing.T, b []byte) []any {
	t.Helper()

	// Try plain array first
	var arr []any
	if err := json.Unmarshal(b, &arr); err == nil {
		return arr
	}

	// Fall back to object with items/data
	var obj map[string]any
	require.NoError(t, json.Unmarshal(b, &obj))
	if v, ok := obj["items"]; ok {
		if a, ok := v.([]any); ok {
			return a
		}
	}
	if v, ok := obj["data"]; ok {
		if a, ok := v.([]any); ok {
			return a
		}
	}
	require.FailNow(t, "expected array body or {items|data: []}")
	return nil
}

// ---------------------------
// Helpers that use real routes
// ---------------------------

// helper: create a report via the production path and return its ID + caseID
func createReport(t *testing.T) (reportID uuid.UUID, caseID uuid.UUID) {
	t.Helper()
	caseID = uuid.New()

	// Ensure the case exists for any JOINs your handlers perform
	ensureCaseRow(caseID)

	w := doRequest("POST", "/reports/cases/"+caseID.String(), "")
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	data, err := parseFlatOrData(w.Body.Bytes())
	require.NoError(t, err)

	idStr := ""
	if v, _ := data["id"].(string); v != "" {
		idStr = v
	} else if v, _ := data["report_id"].(string); v != "" {
		idStr = v
	}
	require.NotEmpty(t, idStr, "create report should return an id")

	reportID = uuid.MustParse(idStr)
	return
}

// helper: add a section and return its Mongo ObjectID by querying Mongo
func addSectionFindOID(t *testing.T, reportID uuid.UUID, title, content string, order int) primitive.ObjectID {
	t.Helper()

	body := fmt.Sprintf(`{"title":%q,"content":%q,"order":%d}`, title, content, order)
	w := doRequest("POST", "/reports/"+reportID.String()+"/sections", body)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	// Find the document containing a section with this exact title
	filter := bson.M{
		"report_id": reportID.String(),
		"tenant_id": FixedTenantID.String(),
		"team_id":   FixedTeamID.String(),
		"sections": bson.M{
			"$elemMatch": bson.M{"title": title},
		},
	}

	var doc bson.M
	require.NoError(t, mongoColl.FindOne(tcCtx, filter).Decode(&doc), "section doc must be present")

	// Extract the matched section's _id; be tolerant of types
	sections, _ := doc["sections"].(primitive.A)
	require.NotEmpty(t, sections, "expected sections array")

	for _, raw := range sections {
		if m, ok := raw.(bson.M); ok && m["title"] == title {
			if oid, ok := m["_id"].(primitive.ObjectID); ok {
				return oid
			}
		}
	}
	require.FailNow(t, "section %q not found / missing _id", title)
	return primitive.NilObjectID
}

// decode JSON into any
func decodeJSON(t *testing.T, b []byte) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal(b, &m), "body=%s", string(b))
	return m
}

// Recursively search for a string value under any of the given keys.
func findString(v any, keys ...string) (string, bool) {
	switch x := v.(type) {
	case map[string]any:
		// direct hit
		for _, k := range keys {
			if s, ok := x[k].(string); ok && s != "" {
				return s, true
			}
		}
		// unwrap some common containers eagerly
		for _, k := range []string{"data", "report", "payload", "result"} {
			if m, ok := x[k].(map[string]any); ok {
				if s, ok2 := findString(m, keys...); ok2 {
					return s, true
				}
			}
		}
		// otherwise, walk all values
		for _, vv := range x {
			if s, ok := findString(vv, keys...); ok {
				return s, true
			}
		}
	case []any:
		for _, it := range x {
			if s, ok := findString(it, keys...); ok {
				return s, true
			}
		}
	}
	return "", false
}

func mustFindString(t *testing.T, v any, keys ...string) string {
	t.Helper()
	if s, ok := findString(v, keys...); ok {
		return s
	}
	require.FailNowf(t, "missing string field", "looked for keys: %v; v=%#v", keys, v)
	return ""
}

// Recursively find an array; prefer given keys if present (items, data, reports, ...).
func findArray(v any, preferredKeys ...string) ([]any, bool) {
	switch x := v.(type) {
	case []any:
		return x, true
	case map[string]any:
		// key-preferred lookups
		for _, k := range preferredKeys {
			if arr, ok := x[k].([]any); ok {
				return arr, true
			}
			// sometimes arrays sit inside one nesting level
			if m, ok := x[k].(map[string]any); ok {
				if arr, ok2 := findArray(m, preferredKeys...); ok2 {
					return arr, true
				}
			}
		}
		// common wrappers
		for _, k := range []string{"data", "payload", "result"} {
			if m, ok := x[k].(map[string]any); ok {
				if arr, ok2 := findArray(m, preferredKeys...); ok2 {
					return arr, true
				}
			}
		}
		// fallback: first array we can find
		for _, vv := range x {
			if arr, ok := findArray(vv, preferredKeys...); ok {
				return arr, true
			}
		}
	}
	return nil, false
}

func mustFindArray(t *testing.T, b []byte) []any {
	t.Helper()
	root := decodeJSON(t, b)
	if arr, ok := findArray(root, "items", "data", "reports", "rows", "list", "content"); ok && len(arr) > 0 {
		return arr
	}
	require.FailNowf(t, "expected array/list", "body=%s", string(b))
	return nil
}

func findStringDeep(v any, keys ...string) (string, bool) {
	switch x := v.(type) {
	case map[string]any:
		for _, k := range keys {
			if s, ok := x[k].(string); ok && s != "" {
				return s, true
			}
		}
		for _, vv := range x {
			if s, ok := findStringDeep(vv, keys...); ok {
				return s, true
			}
		}
	case []any:
		for _, it := range x {
			if s, ok := findStringDeep(it, keys...); ok {
				return s, true
			}
		}
	}
	return "", false
}
