// file: services_/report/unit_tests/repo_mongo_recent_test.go
package unit_tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"

	// Import the production package under test.
	report "aegis-api/services_/report"
)

/*
	This suite unit-tests the Mongo aggregation helper that finds the latest section
	update per report: LatestUpdateByReportIDs.

	We use the mongo-driver's mtest mock harness (ClientType: mtest.Mock), which lets us
	inject fake server responses (cursor batches, command errors) without a real MongoDB.

	BASIC FLOW:
	  - Construct the repository using the mocked collection (mt.Coll)
	  - Push "server" responses onto mtest with AddMockResponses(...)
	  - Call the method under test
	  - Assert the decoded results and any errors

	NOTE: This file references ns(mt) to build the "<db>.<collection>" namespace for
	      CreateCursorResponse. If you don't already have it in a shared helper, add:

	      func ns(mt *mtest.T) string {
	          return fmt.Sprintf("%s.%s", mt.DB.Name(), mt.Coll.Name())
	      }
*/

// TestMongoReportRepository_LatestUpdateByReportIDs exercises three scenarios:
//  1. Empty input → returns empty map and performs no DB call
//  2. "Happy path" → two aggregation rows are decoded into map[reportID]time.Time
//  3. Aggregation command error → error is propagated to the caller
func TestMongoReportRepository_LatestUpdateByReportIDs(t *testing.T) {
	// Create an mtest runner that uses a mocked client, DB, and collection.
	mt := mtest.New(t, mtest.NewOptions().
		ClientType(mtest.Mock).
		DatabaseName("testdb").
		CollectionName("report_contents"),
	)

	mt.Run("empty input returns empty map", func(mt *mtest.T) {
		// Build the concrete repo using the mocked collection.
		repo := report.NewReportMongoRepo(mt.Coll)

		// No report IDs supplied → method should short-circuit and return an empty map.
		got, err := repo.LatestUpdateByReportIDs(context.Background(), nil, "tenant-1", "team-1")
		require.NoError(mt, err)
		require.Empty(mt, got)

		// We did not add any mock responses (AddMockResponses), which is fine here because
		// the function should exit before issuing an Aggregate() call.
	})

	mt.Run("happy path aggregates two report IDs", func(mt *mtest.T) {
		repo := report.NewReportMongoRepo(mt.Coll)

		// Create two timestamps to simulate the "lastUpdate" coming back from the $group stage.
		t1 := time.Now().Add(-30 * time.Minute)
		t2 := time.Now()

		// Fabricate a cursor response with two aggregation result documents:
		//   { _id: "rep-1", lastUpdate: t1 }
		//   { _id: "rep-2", lastUpdate: t2 }
		//
		// mtest models cursor batches with "firstBatch" and "nextBatch". Here we use a single
		// "firstBatch" that contains both docs, followed by an empty "nextBatch" to close the cursor.
		first := mtest.CreateCursorResponse(
			1, ns(mt), mtest.FirstBatch,
			bson.D{{Key: "_id", Value: "rep-1"}, {Key: "lastUpdate", Value: primitive.NewDateTimeFromTime(t1)}},
			bson.D{{Key: "_id", Value: "rep-2"}, {Key: "lastUpdate", Value: primitive.NewDateTimeFromTime(t2)}},
		)
		end := mtest.CreateCursorResponse(0, ns(mt), mtest.NextBatch)

		// Push our fake server responses onto the harness. The next Aggregate() call will consume these.
		mt.AddMockResponses(first, end)

		// Execute and assert that both report IDs are present with the right timestamps (±1s).
		got, err := repo.LatestUpdateByReportIDs(context.Background(), []string{"rep-1", "rep-2"}, "tenant-1", "team-1")
		require.NoError(mt, err)
		require.Len(mt, got, 2)
		require.WithinDuration(mt, t1, got["rep-1"], time.Second)
		require.WithinDuration(mt, t2, got["rep-2"], time.Second)
	})

	mt.Run("aggregate error bubbles up", func(mt *mtest.T) {
		repo := report.NewReportMongoRepo(mt.Coll)

		// Instead of returning a cursor, instruct mtest to return a command error for the Aggregate.
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Message: "aggregation failed",
			Code:    1,
			Name:    "CommandError",
		}))

		// Expect the error to be propagated by the repository method.
		got, err := repo.LatestUpdateByReportIDs(context.Background(), []string{"rep-1"}, "tenant", "team")
		require.Error(mt, err)
		require.Nil(mt, got)
	})
}
