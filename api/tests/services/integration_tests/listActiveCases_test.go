// file: tests/services/integration_tests/case_active_endpoints_test.go
package integration_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// ensure the tenant row exists
func ensureTenant(t *testing.T, id uuid.UUID, name string) {
	t.Helper()
	if name == "" {
		name = "tenant-" + id.String()[:8]
	}
	execSQLT(t, `
		INSERT INTO tenants (id, name, created_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (id) DO NOTHING`,
		id, name,
	)
}

// execSQL is a tiny helper to run parameterized SQL in tests.
func execSQL(t *testing.T, query string, args ...any) {
	t.Helper()
	_, err := pgSQL.ExecContext(tcCtx, query, args...)
	if err != nil {
		t.Fatalf("exec sql: %v\nquery: %s\nargs: %#v", err, query, args)
	}
}

// ensure the team row exists (must reference an existing tenant)
func ensureTeam(t *testing.T, teamID, tenantID uuid.UUID, teamName string) {
	t.Helper()
	if teamName == "" {
		teamName = "team-" + teamID.String()[:8]
	}
	// make sure the tenant exists first
	ensureTenant(t, tenantID, "tenant-"+tenantID.String()[:8])

	execSQLT(t, `
		INSERT INTO teams (id, team_name, tenant_id, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		ON CONFLICT (id) DO NOTHING`,
		teamID, teamName, tenantID,
	)
}

func ensureUser(t *testing.T, id uuid.UUID) {
	t.Helper()
	execSQLT(t, `
        INSERT INTO users
        (id, full_name, email, password_hash, role, is_verified, tenant_id, team_id, created_at)
        VALUES ($1, $2, $3, 'x', 'Admin', true, $4, $5, NOW())
        ON CONFLICT (id) DO NOTHING`,
		id, "user-"+id.String()[:8], id.String()+"@t.local", FixedTenantID, FixedTeamID,
	)
}

func seedCase(t *testing.T, id uuid.UUID, title, status string,
	createdBy, tenantID, teamID uuid.UUID, teamName string,
) {
	t.Helper()
	ensureUser(t, createdBy)
	ensureTenant(t, tenantID, "tenant-"+tenantID.String()[:8])
	ensureTeam(t, teamID, tenantID, teamName)
	execSQLT(t, `
        INSERT INTO cases (id, title, team_name, status, created_by, tenant_id, team_id, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
        ON CONFLICT (id) DO NOTHING`,
		id, title, teamName, status, createdBy, tenantID, teamID,
	)
}

func seedCaseRole(t *testing.T, userID, caseID, tenantID uuid.UUID) {
	t.Helper()
	execSQLT(t, `
        INSERT INTO case_user_roles (user_id, case_id, role, assigned_at, tenant_id)
        VALUES ($1,$2,'Incident Responder',NOW(),$3)
        ON CONFLICT (user_id, case_id) DO NOTHING`,
		userID, caseID, tenantID,
	)
}

// local helper (uses global pgSQL from bootstrap)
func execSQLT(t *testing.T, q string, args ...any) {
	t.Helper()
	if _, err := pgSQL.Exec(q, args...); err != nil {
		t.Fatalf("exec sql: %v", err)
	}
}

func idsFromArray(t *testing.T, body []byte) []string {
	var arr []map[string]any
	require.NoError(t, json.Unmarshal(body, &arr), string(body))
	ids := make([]string, 0, len(arr))
	for _, it := range arr {
		if s, ok := it["id"].(string); ok && s != "" {
			ids = append(ids, s)
		}
	}
	return ids
}

// Keep the old helper if some tests still use it
// func doRequest(method, url, body string) *httptest.ResponseRecorder {
// 	req := httptest.NewRequest(method, url, strings.NewReader(body))
// 	req.Header.Set("Content-Type", "application/json")
// 	w := httptest.NewRecorder()
// 	router.ServeHTTP(w, req)
// 	return w
// }

// NEW: headered variant used by your active-cases tests
func doRequestAuth(method, url, body string, userID, tenantID, teamID uuid.UUID) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-UserID", userID.String())
	req.Header.Set("X-Test-TenantID", tenantID.String())
	req.Header.Set("X-Test-TeamID", teamID.String())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// --- helpers ---

func seedTenantTeamUser(t *testing.T) (tenantID, teamID, userID uuid.UUID) {
	t.Helper()
	tenantID = uuid.New()
	teamID = uuid.New()
	userID = uuid.New()

	execSQL(t, `INSERT INTO tenants (id, name) VALUES ($1, $2)`,
		tenantID, "t-"+tenantID.String()[:8],
	)
	execSQL(t, `INSERT INTO teams (id, team_name, tenant_id) VALUES ($1, $2, $3)`,
		teamID, "team-"+teamID.String()[:8], tenantID,
	)
	// role must be a valid enum; 'Admin' exists in your schema
	execSQL(t, `INSERT INTO users (id, full_name, email, password_hash, role, is_verified, tenant_id, team_id)
	            VALUES ($1,$2,$3,$4,'Admin',true,$5,$6)`,
		userID, "User "+userID.String()[:8], userID.String()[:8]+"@ex.com", "x", tenantID, teamID,
	)
	return
}

func insertCase(t *testing.T, id uuid.UUID, title string, createdBy, tenantID, teamID uuid.UUID, status string) {
	t.Helper()
	execSQL(t, `INSERT INTO cases
	            (id, title, description, status, investigation_stage, priority, team_name, created_by, tenant_id, team_id)
	            VALUES ($1,$2,'', $3, 'Triage', 'medium', 'test-team', $4, $5, $6)`,
		id, title, status, createdBy, tenantID, teamID,
	)
}

func mustIDsFromArray(t *testing.T, body []byte) []uuid.UUID {
	t.Helper()
	var arr []map[string]any
	require.NoError(t, json.Unmarshal(body, &arr), string(body))
	out := make([]uuid.UUID, 0, len(arr))
	for _, m := range arr {
		if s, _ := m["id"].(string); s != "" {
			if id, err := uuid.Parse(s); err == nil {
				out = append(out, id)
			}
		}
	}
	return out
}

func idsString(ids []uuid.UUID) string {
	ss := make([]string, len(ids))
	for i, id := range ids {
		ss[i] = id.String()
	}
	return fmt.Sprintf("%v", ss)
}

// --- tests ---

func Test_ListActiveCases_Basic(t *testing.T) {
	tenantID, teamID, userID := seedTenantTeamUser(t)
	err := pgDB.Exec(`DO $$ 
        BEGIN
            IF NOT EXISTS (
                SELECT 1 FROM pg_enum 
                WHERE enumlabel = 'archived' 
                AND enumtypid = (SELECT oid FROM pg_type WHERE typname = 'case_status')
            ) THEN
                ALTER TYPE case_status ADD VALUE 'archived';
            END IF;
        END $$`).Error
	require.NoError(t, err)

	// Seed exactly 3 open cases for this tenant/team/user
	want := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	insertCase(t, want[0], "alpha", userID, tenantID, teamID, "open")
	insertCase(t, want[1], "beta", userID, tenantID, teamID, "open")
	insertCase(t, want[2], "gamma", userID, tenantID, teamID, "open")

	// Control noise: a closed case in same tenant/team should be excluded
	insertCase(t, uuid.New(), "closed-one", userID, tenantID, teamID, "closed")

	// More noise: cases in a different tenant/team should not appear
	otherTenant, otherTeam, otherUser := seedTenantTeamUser(t)
	insertCase(t, uuid.New(), "foreign-1", otherUser, otherTenant, otherTeam, "open")
	insertCase(t, uuid.New(), "foreign-2", otherUser, otherTenant, otherTeam, "open")

	// Call endpoint scoped to our unique tenant/team/user
	w := doRequestAuth("GET", "/cases/active", "", userID, tenantID, teamID)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	got := mustIDsFromArray(t, w.Body.Bytes())
	require.Len(t, got, len(want), "unexpected number of cases: %s", idsString(got))

	// Verify the exact set (order-insensitive)
	wantSet := map[uuid.UUID]struct{}{want[0]: {}, want[1]: {}, want[2]: {}}
	for _, id := range got {
		if _, ok := wantSet[id]; !ok {
			t.Fatalf("unexpected case id in results: %s (got: %s)", id, idsString(got))
		}
	}
}

func Test_ListActiveCases_NoMatches(t *testing.T) {
	// Fresh, isolated tenant/team/user with no cases seeded
	tenantID, teamID, userID := seedTenantTeamUser(t)

	w := doRequestAuth("GET", "/cases/active", "", userID, tenantID, teamID)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var arr []any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &arr), w.Body.String())
	require.Len(t, arr, 0, "expected no cases for this user/tenant/team; got: %s", w.Body.String())
}
