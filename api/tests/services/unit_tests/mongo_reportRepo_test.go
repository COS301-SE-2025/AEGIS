package unit_tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"

	// 👇 Update this to your actual module path
	report "aegis-api/services_/report"
)

// ns builds the "db.collection" namespace string required by CreateCursorResponse.
// mtest needs a single string for the namespace instead of separate db/collection names.
func ns(mt *mtest.T) string {
	return fmt.Sprintf("%s.%s", mt.DB.Name(), mt.Coll.Name())
}

// Covers: SaveReportContent + GetReportContent (found + not found).
// These use mtest's Mock client to simulate MongoDB server replies without a real DB.
func TestReportMongoRepo_SaveAndGet(t *testing.T) {
	// Create a mock test harness bound to "testdb.reports"
	mt := mtest.New(t, mtest.NewOptions().
		ClientType(mtest.Mock).
		DatabaseName("testdb").
		CollectionName("reports"),
	)

	mt.Run("SaveReportContent stamps ID and timestamps", func(mt *mtest.T) {
		repo := report.NewReportMongoRepo(mt.Coll)

		// Simulate a successful InsertOne command reply.
		// The driver only needs ok/n/nModified fields to consider it successful here.
		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "ok", Value: 1},
			bson.E{Key: "n", Value: 1},
		))

		// Missing ID/CreatedAt/UpdatedAt should be filled by the method.
		rc := &report.ReportContentMongo{
			ReportID: uuid.NewString(),
			TenantID: uuid.NewString(),
			TeamID:   uuid.NewString(),
		}
		err := repo.SaveReportContent(context.Background(), rc)
		require.NoError(mt, err)
		assert.False(mt, rc.ID.IsZero(), "ID should be set when inserting")
		assert.False(mt, rc.CreatedAt.IsZero(), "CreatedAt should be set")
		assert.False(mt, rc.UpdatedAt.IsZero(), "UpdatedAt should be set")
	})

	mt.Run("GetReportContent - found", func(mt *mtest.T) {
		repo := report.NewReportMongoRepo(mt.Coll)
		oid := primitive.NewObjectID()
		now := time.Now()

		// A findOne under the hood is modeled as a cursor with a first batch + an empty next batch.
		// FirstBatch has exactly one doc → found.
		first := mtest.CreateCursorResponse(
			1, ns(mt), mtest.FirstBatch,
			bson.D{
				{Key: "_id", Value: oid},
				{Key: "report_id", Value: uuid.NewString()},
				{Key: "tenant_id", Value: "tenant-1"},
				{Key: "team_id", Value: "team-1"},
				{Key: "sections", Value: bson.A{}},
				{Key: "created_at", Value: primitive.NewDateTimeFromTime(now)},
				{Key: "updated_at", Value: primitive.NewDateTimeFromTime(now)},
			},
		)
		// Cursor close for findOne
		end := mtest.CreateCursorResponse(0, ns(mt), mtest.NextBatch)
		mt.AddMockResponses(first, end)

		got, err := repo.GetReportContent(context.Background(), oid, "tenant-1", "team-1")
		require.NoError(mt, err)
		require.NotNil(mt, got, "should return a document")
		assert.Equal(mt, oid, got.ID)
		assert.Equal(mt, "tenant-1", got.TenantID)
		assert.Equal(mt, "team-1", got.TeamID)
	})

	mt.Run("GetReportContent - not found => nil", func(mt *mtest.T) {
		repo := report.NewReportMongoRepo(mt.Coll)

		// First batch with zero docs → not found for findOne semantics.
		first := mtest.CreateCursorResponse(1, ns(mt), mtest.FirstBatch /* no docs */)
		end := mtest.CreateCursorResponse(0, ns(mt), mtest.NextBatch)
		mt.AddMockResponses(first, end)

		got, err := repo.GetReportContent(context.Background(), primitive.NewObjectID(), "t", "tm")
		require.NoError(mt, err)
		assert.Nil(mt, got, "should return nil when no doc matches")
	})
}

// Covers: UpdateSection, AddSection, DeleteSection
// Each subtest sets up one UpdateOne mock response to simulate success or failure.
func TestReportMongoRepo_SectionMutations(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().
		ClientType(mtest.Mock).
		DatabaseName("testdb").
		CollectionName("reports"),
	)

	mt.Run("UpdateSection - success", func(mt *mtest.T) {
		repo := report.NewReportMongoRepo(mt.Coll)

		// Simulate UpdateOne matched and modified.
		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "ok", Value: 1}, bson.E{Key: "n", Value: 1}, bson.E{Key: "nModified", Value: 1},
		))

		err := repo.UpdateSection(context.Background(), primitive.NewObjectID(), primitive.NewObjectID(), "new", "t", "tm")
		require.NoError(mt, err)
	})

	mt.Run("UpdateSection - section not found", func(mt *mtest.T) {
		repo := report.NewReportMongoRepo(mt.Coll)

		// Simulate UpdateOne with no matching docs (n = 0).
		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "ok", Value: 1}, bson.E{Key: "n", Value: 0}, bson.E{Key: "nModified", Value: 0},
		))

		err := repo.UpdateSection(context.Background(), primitive.NewObjectID(), primitive.NewObjectID(), "new", "t", "tm")
		require.Error(mt, err)
		assert.Contains(mt, err.Error(), "section not found")
	})

	mt.Run("AddSection - success", func(mt *mtest.T) {
		repo := report.NewReportMongoRepo(mt.Coll)

		// UpdateOne (with $push) succeeded → section added.
		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "ok", Value: 1}, bson.E{Key: "n", Value: 1}, bson.E{Key: "nModified", Value: 1},
		))

		err := repo.AddSection(context.Background(), primitive.NewObjectID(), report.ReportSection{Title: "T", Content: "C", Order: 1}, "t", "tm")
		require.NoError(mt, err)
	})

	mt.Run("AddSection - report not found", func(mt *mtest.T) {
		repo := report.NewReportMongoRepo(mt.Coll)

		// No matched docs → "report not found" path.
		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "ok", Value: 1}, bson.E{Key: "n", Value: 0}, bson.E{Key: "nModified", Value: 0},
		))

		err := repo.AddSection(context.Background(), primitive.NewObjectID(), report.ReportSection{Title: "x"}, "t", "tm")
		require.Error(mt, err)
		assert.Contains(mt, err.Error(), "report not found")
	})

	mt.Run("DeleteSection - success", func(mt *mtest.T) {
		repo := report.NewReportMongoRepo(mt.Coll)

		// $pull matched → section deleted.
		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "ok", Value: 1}, bson.E{Key: "n", Value: 1}, bson.E{Key: "nModified", Value: 1},
		))

		err := repo.DeleteSection(context.Background(), primitive.NewObjectID(), primitive.NewObjectID(), "t", "tm")
		require.NoError(mt, err)
	})

	mt.Run("DeleteSection - not found", func(mt *mtest.T) {
		repo := report.NewReportMongoRepo(mt.Coll)

		// No matched docs → ErrSectionNotFound.
		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "ok", Value: 1}, bson.E{Key: "n", Value: 0}, bson.E{Key: "nModified", Value: 0},
		))

		err := repo.DeleteSection(context.Background(), primitive.NewObjectID(), primitive.NewObjectID(), "t", "tm")
		require.Error(mt, err)
		assert.ErrorIs(mt, err, report.ErrSectionNotFound)
	})
}

// Covers: UpdateSections (success + not found) and ReorderSection (success + section missing).
// UpdateSections normalizes IDs/timestamps and enforces ordering. ReorderSection does a read-then-write.
func TestReportMongoRepo_UpdateAndReorder(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().
		ClientType(mtest.Mock).
		DatabaseName("testdb").
		CollectionName("reports"),
	)

	mt.Run("UpdateSections - success and normalizes", func(mt *mtest.T) {
		repo := report.NewReportMongoRepo(mt.Coll)

		// Simulate UpdateOne success for replacing 'sections' array.
		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "ok", Value: 1}, bson.E{Key: "n", Value: 1}, bson.E{Key: "nModified", Value: 1},
		))

		// Missing IDs/timestamps should be generated; order should remain ascending.
		sections := []report.ReportSection{
			{Title: "B", Content: "b", Order: 2},
			{Title: "A", Content: "a", Order: 1},
		}
		err := repo.UpdateSections(context.Background(), primitive.NewObjectID(), sections, "tenant", "team")
		require.NoError(mt, err)
		for _, s := range sections {
			assert.False(mt, s.ID.IsZero(), "ID should be set")
			assert.False(mt, s.CreatedAt.IsZero(), "CreatedAt should be set")
			assert.False(mt, s.UpdatedAt.IsZero(), "UpdatedAt should be set")
		}
		assert.LessOrEqual(mt, sections[0].Order, sections[1].Order, "sections should be sorted by Order")
	})

	mt.Run("UpdateSections - doc not found", func(mt *mtest.T) {
		repo := report.NewReportMongoRepo(mt.Coll)

		// No matched doc → ErrMongoReportNotFound.
		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "ok", Value: 1}, bson.E{Key: "n", Value: 0}, bson.E{Key: "nModified", Value: 0},
		))

		err := repo.UpdateSections(context.Background(), primitive.NewObjectID(), []report.ReportSection{{Title: "x", Order: 1}}, "t", "tm")
		require.Error(mt, err)
		assert.ErrorIs(mt, err, report.ErrMongoReportNotFound)
	})

	mt.Run("ReorderSection - success", func(mt *mtest.T) {
		repo := report.NewReportMongoRepo(mt.Coll)
		rid := primitive.NewObjectID()
		s1, s2, s3 := primitive.NewObjectID(), primitive.NewObjectID(), primitive.NewObjectID()
		now := time.Now()

		// ReorderSection first does a FindOne to load the doc (cursor first + end),
		// then calls UpdateSections to persist reordered array (UpdateOne success).
		first := mtest.CreateCursorResponse(
			1, ns(mt), mtest.FirstBatch,
			bson.D{
				{Key: "_id", Value: rid},
				{Key: "tenant_id", Value: "t"},
				{Key: "team_id", Value: "tm"},
				{Key: "report_id", Value: uuid.NewString()},
				{Key: "sections", Value: bson.A{
					bson.D{{Key: "_id", Value: s1}, {Key: "title", Value: "A"}, {Key: "content", Value: "a"}, {Key: "order", Value: 1}, {Key: "created_at", Value: primitive.NewDateTimeFromTime(now)}, {Key: "updated_at", Value: primitive.NewDateTimeFromTime(now)}},
					bson.D{{Key: "_id", Value: s2}, {Key: "title", Value: "B"}, {Key: "content", Value: "b"}, {Key: "order", Value: 2}, {Key: "created_at", Value: primitive.NewDateTimeFromTime(now)}, {Key: "updated_at", Value: primitive.NewDateTimeFromTime(now)}},
					bson.D{{Key: "_id", Value: s3}, {Key: "title", Value: "C"}, {Key: "content", Value: "c"}, {Key: "order", Value: 3}, {Key: "created_at", Value: primitive.NewDateTimeFromTime(now)}, {Key: "updated_at", Value: primitive.NewDateTimeFromTime(now)}},
				}},
				{Key: "created_at", Value: primitive.NewDateTimeFromTime(now)},
				{Key: "updated_at", Value: primitive.NewDateTimeFromTime(now)},
			},
		)
		end := mtest.CreateCursorResponse(0, ns(mt), mtest.NextBatch)
		updateOK := mtest.CreateSuccessResponse(bson.E{Key: "ok", Value: 1}, bson.E{Key: "n", Value: 1}, bson.E{Key: "nModified", Value: 1})
		mt.AddMockResponses(first, end, updateOK)

		err := repo.ReorderSection(context.Background(), rid, s3, 1, "t", "tm")
		require.NoError(mt, err)
	})

	mt.Run("ReorderSection - section not found", func(mt *mtest.T) {
		repo := report.NewReportMongoRepo(mt.Coll)
		rid := primitive.NewObjectID()
		now := time.Now()

		// FindOne returns a doc with zero sections → cannot find the section to move.
		first := mtest.CreateCursorResponse(
			1, ns(mt), mtest.FirstBatch,
			bson.D{
				{Key: "_id", Value: rid},
				{Key: "tenant_id", Value: "t"},
				{Key: "team_id", Value: "tm"},
				{Key: "report_id", Value: uuid.NewString()},
				{Key: "sections", Value: bson.A{}},
				{Key: "created_at", Value: primitive.NewDateTimeFromTime(now)},
				{Key: "updated_at", Value: primitive.NewDateTimeFromTime(now)},
			},
		)
		end := mtest.CreateCursorResponse(0, ns(mt), mtest.NextBatch)
		mt.AddMockResponses(first, end)

		err := repo.ReorderSection(context.Background(), rid, primitive.NewObjectID(), 2, "t", "tm")
		require.Error(mt, err)
		assert.Contains(mt, err.Error(), "section not found")
	})
}

// Covers: FindByReportUUID (found + not found) and LatestUpdateByReportIDs aggregate.
// For aggregate/find, mtest again models replies as a cursor: first batch then empty batch.
func TestReportMongoRepo_FindAndAggregate(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().
		ClientType(mtest.Mock).
		DatabaseName("testdb").
		CollectionName("reports"),
	)

	mt.Run("FindByReportUUID - found", func(mt *mtest.T) {
		repo := report.NewReportMongoRepo(mt.Coll)
		rUUID := uuid.New()

		// findOne → one doc in FirstBatch then cursor end.
		first := mtest.CreateCursorResponse(
			1, ns(mt), mtest.FirstBatch,
			bson.D{
				{Key: "_id", Value: primitive.NewObjectID()},
				{Key: "report_id", Value: rUUID.String()},
				{Key: "tenant_id", Value: "t"},
				{Key: "team_id", Value: "tm"},
				{Key: "sections", Value: bson.A{}},
				{Key: "created_at", Value: primitive.NewDateTimeFromTime(time.Now())},
				{Key: "updated_at", Value: primitive.NewDateTimeFromTime(time.Now())},
			},
		)
		end := mtest.CreateCursorResponse(0, ns(mt), mtest.NextBatch)
		mt.AddMockResponses(first, end)

		got, err := repo.FindByReportUUID(context.Background(), rUUID)
		require.NoError(mt, err)
		require.NotNil(mt, got, "document should be found")
		assert.Equal(mt, rUUID.String(), got.ReportID)
	})

	mt.Run("FindByReportUUID - not found", func(mt *mtest.T) {
		repo := report.NewReportMongoRepo(mt.Coll)

		// FirstBatch has no docs → not found.
		first := mtest.CreateCursorResponse(1, ns(mt), mtest.FirstBatch /* no docs */)
		end := mtest.CreateCursorResponse(0, ns(mt), mtest.NextBatch)
		mt.AddMockResponses(first, end)

		got, err := repo.FindByReportUUID(context.Background(), uuid.New())
		require.NoError(mt, err)
		assert.Nil(mt, got, "nil expected when not found")
	})

	mt.Run("LatestUpdateByReportIDs - happy path", func(mt *mtest.T) {
		repo := report.NewReportMongoRepo(mt.Coll)
		t1 := time.Now().Add(-1 * time.Hour)
		t2 := time.Now()

		// Aggregate returns two rows in FirstBatch, then end.
		first := mtest.CreateCursorResponse(
			1, ns(mt), mtest.FirstBatch,
			bson.D{{Key: "_id", Value: "rep-1"}, {Key: "lastUpdate", Value: primitive.NewDateTimeFromTime(t1)}},
			bson.D{{Key: "_id", Value: "rep-2"}, {Key: "lastUpdate", Value: primitive.NewDateTimeFromTime(t2)}},
		)
		end := mtest.CreateCursorResponse(0, ns(mt), mtest.NextBatch)
		mt.AddMockResponses(first, end)

		got, err := repo.LatestUpdateByReportIDs(context.Background(), []string{"rep-1", "rep-2"}, "t", "tm")
		require.NoError(mt, err)
		require.Len(mt, got, 2, "should return both report IDs")
		assert.WithinDuration(mt, t1, got["rep-1"], time.Second)
		assert.WithinDuration(mt, t2, got["rep-2"], time.Second)
	})
}
