// file: services_/report/unit_tests/service_recent_test.go
package unit_tests

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	// SUT package
	report "aegis-api/services_/report"
)

/*
	ListRecentReports — behavior to verify:

	1) Calculates the "window" (candidateLimit*3, min 30; default candidateLimit=10 when Limit<=0)
	   and passes that as the 3rd arg to repo.ListRecentCandidates.

	2) When the repo returns no candidates → returns empty slice and does NOT call Mongo.

	3) Calls Mongo.LatestUpdateByReportIDs with:
	   - list of candidate IDs (as strings)
	   - tenantID as string
	   - teamID as string, or "" when opts.TeamID == nil

	4) Merges timestamps: LastModified = max(pg.updated_at, mongo.lastUpdate[reportID])

	5) Sorts by LastModified DESC, truncates to Limit.

	6) Converts LastModified to Africa/Johannesburg timezone.
*/

// Test: No candidates from repo => empty result and no Mongo call.
func TestListRecentReports_EmptyCandidates(t *testing.T) {
	ctx := context.Background()
	repoMock := new(MockRepo)
	mongoMock := new(MockMongo)
	svc := newSvc(repoMock, mongoMock)

	opts := report.RecentReportsOptions{
		TenantID: uuid.New(),
		Limit:    5,
	}
	// window = max(Limit*3, 30) => max(15, 30) = 30
	repoMock.
		On("ListRecentCandidates", ctx, opts, 30).
		Return([]report.Report{}, nil).
		Once()

	out, err := svc.ListRecentReports(ctx, opts)
	require.NoError(t, err)
	require.Empty(t, out)

	// Ensure Mongo isn't called at all.
	mongoMock.AssertNotCalled(t, "LatestUpdateByReportIDs", mock.Anything, mock.Anything, mock.Anything, mock.Anything)

	repoMock.AssertExpectations(t)
}

// Test: Default limit & teamID empty-string scoping to Mongo; time conversion to Africa/Johannesburg.
func TestListRecentReports_DefaultLimit_TeamNil_TimezoneConversion(t *testing.T) {
	ctx := context.Background()
	repoMock := new(MockRepo)
	mongoMock := new(MockMongo)
	svc := newSvc(repoMock, mongoMock)

	tenant := uuid.New()
	// Limit <= 0 → candidateLimit=10, window = max(10*3,30)=30
	opts := report.RecentReportsOptions{
		TenantID: tenant,
		TeamID:   nil, // team scope should become "" when calling Mongo
		Limit:    0,
	}

	// Single candidate from Postgres with a known timestamp.
	id := uuid.New()
	pgUpdated := time.Date(2025, 8, 19, 6, 0, 0, 0, time.UTC)
	repoMock.
		On("ListRecentCandidates", ctx, opts, 30).
		Return([]report.Report{
			{ID: id, Name: "R1", Status: "draft", UpdatedAt: pgUpdated},
		}, nil).
		Once()

	// Mongo returns no newer updates (empty map).
	mongoMock.
		On("LatestUpdateByReportIDs",
			ctx,
			mock.MatchedBy(func(ids []string) bool { return len(ids) == 1 && ids[0] == id.String() }),
			tenant.String(),
			"", // teamID should be empty string when opts.TeamID == nil
		).
		Return(map[string]time.Time{}, nil).
		Once()

	out, err := svc.ListRecentReports(ctx, opts)
	require.NoError(t, err)
	require.Len(t, out, 1)

	// LastModified should be equal to pgUpdated but presented in Africa/Johannesburg.
	loc, _ := time.LoadLocation("Africa/Johannesburg")
	require.Equal(t, pgUpdated.In(loc), out[0].LastModified)
	require.Equal(t, "R1", out[0].Title)
	require.Equal(t, "draft", out[0].Status)

	repoMock.AssertExpectations(t)
	mongoMock.AssertExpectations(t)
}

// Test: Merge, sort, and truncate to Limit, with non-nil TeamID scoping to Mongo.
func TestListRecentReports_MergeSortTruncate_WithTeamID(t *testing.T) {
	ctx := context.Background()
	repoMock := new(MockRepo)
	mongoMock := new(MockMongo)
	svc := newSvc(repoMock, mongoMock)

	tenant := uuid.New()
	team := uuid.New()

	// Limit=2 → candidateLimit=2, window=max(2*3,30)=30
	opts := report.RecentReportsOptions{
		TenantID: tenant,
		TeamID:   &team,
		Limit:    2,
	}

	// Build 3 candidates with varying PG updated_at values.
	idA := uuid.New() // will be updated by Mongo to be newer than PG
	idB := uuid.New() // PG is already newer than Mongo
	idC := uuid.New() // no Mongo entry

	pgA := time.Date(2025, 8, 19, 10, 0, 0, 0, time.UTC)
	pgB := time.Date(2025, 8, 19, 13, 0, 0, 0, time.UTC)
	pgC := time.Date(2025, 8, 19, 9, 0, 0, 0, time.UTC)

	repoMock.
		On("ListRecentCandidates", ctx, opts, 30).
		Return([]report.Report{
			{ID: idA, Name: "A", Status: "s1", UpdatedAt: pgA},
			{ID: idB, Name: "B", Status: "s2", UpdatedAt: pgB},
			{ID: idC, Name: "C", Status: "s3", UpdatedAt: pgC},
		}, nil).
		Once()

	// Mongo returns a newer time for A and an older time for B; no entry for C.
	mA := time.Date(2025, 8, 19, 12, 0, 0, 0, time.UTC) // newer than pgA (10:00)
	mB := time.Date(2025, 8, 19, 11, 0, 0, 0, time.UTC) // older than pgB (13:00)

	mongoMock.
		On("LatestUpdateByReportIDs",
			ctx,
			mock.MatchedBy(func(ids []string) bool {
				// Order doesn't matter, just ensure all three IDs are present.
				set := map[string]bool{}
				for _, s := range ids {
					set[s] = true
				}
				return set[idA.String()] && set[idB.String()] && set[idC.String()] && len(ids) == 3
			}),
			tenant.String(),
			team.String(), // teamID should be the string value when pointer is non-nil
		).
		Return(map[string]time.Time{
			idA.String(): mA,
			idB.String(): mB,
			// idC absent
		}, nil).
		Once()

	out, err := svc.ListRecentReports(ctx, opts)
	require.NoError(t, err)

	// After merge:
	//   A => max(pgA=10:00, mA=12:00) = 12:00
	//   B => max(pgB=13:00, mB=11:00) = 13:00
	//   C => pgC=09:00 (no mongo)
	// Sort desc by LastModified => B(13:00), A(12:00), C(09:00)
	// Truncate to Limit=2 => [B, A]

	require.Len(t, out, 2)
	require.Equal(t, "B", out[0].Title)
	require.Equal(t, "A", out[1].Title)

	// And both times should be presented in Africa/Johannesburg (UTC+2).
	loc, _ := time.LoadLocation("Africa/Johannesburg")
	require.Equal(t, pgB.In(loc), out[0].LastModified) // B used PG (13:00Z)
	require.Equal(t, mA.In(loc), out[1].LastModified)  // A used Mongo (12:00Z)

	repoMock.AssertExpectations(t)
	mongoMock.AssertExpectations(t)
}
