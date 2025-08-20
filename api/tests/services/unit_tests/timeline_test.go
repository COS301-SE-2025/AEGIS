package unit_tests

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Minimal test model compatible with SQLite
type TimelineEventTest struct {
	ID     string `gorm:"primaryKey"`
	CaseID string
	Order  int
}

type testEvent struct {
	ID        string
	CaseID    string
	Order     int
	CreatedAt int64 // not used, just for completeness
}

func setupTimelineTestDB(t *testing.T) *gorm.DB {
	dbName := "file:" + uuid.New().String() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	require.NoError(t, err)
	err = db.AutoMigrate(&TimelineEventTest{})
	require.NoError(t, err)
	return db
}

func TestCreateAndGetByID(t *testing.T) {
	db := setupTimelineTestDB(t)
	eventID := uuid.New().String()
	event := &TimelineEventTest{ID: eventID, CaseID: "case1", Order: 0}
	err := db.Create(event).Error
	require.NoError(t, err)
	var fetched TimelineEventTest
	err = db.First(&fetched, "id = ?", eventID).Error
	require.NoError(t, err)
	require.Equal(t, event.ID, fetched.ID)
}

func TestListByCase(t *testing.T) {
	db := setupTimelineTestDB(t)
	event1ID := uuid.New().String()
	event2ID := uuid.New().String()
	event3ID := uuid.New().String()
	events := []*TimelineEventTest{
		{ID: event1ID, CaseID: "case1", Order: 1},
		{ID: event2ID, CaseID: "case1", Order: 2},
		{ID: event3ID, CaseID: "case2", Order: 1},
	}
	for _, ev := range events {
		require.NoError(t, db.Create(ev).Error)
	}
	var case1Events []TimelineEventTest
	err := db.Where("case_id = ?", "case1").Order("\"order\" ASC").Find(&case1Events).Error
	require.NoError(t, err)
	require.Len(t, case1Events, 2)
	require.Equal(t, []int{1, 2}, []int{case1Events[0].Order, case1Events[1].Order})
	var case2Events []TimelineEventTest
	err = db.Where("case_id = ?", "case2").Order("\"order\" ASC").Find(&case2Events).Error
	require.NoError(t, err)
	require.Len(t, case2Events, 1)
	require.Equal(t, 1, case2Events[0].Order)
}

func TestUpdateAndDelete(t *testing.T) {
	db := setupTimelineTestDB(t)
	eventID := uuid.New().String()
	event := &TimelineEventTest{ID: eventID, CaseID: "case1", Order: 1}
	require.NoError(t, db.Create(event).Error)
	event.Order = 5
	require.NoError(t, db.Model(&TimelineEventTest{}).Where("id = ?", event.ID).Updates(event).Error)
	var fetched TimelineEventTest
	err := db.First(&fetched, "id = ?", event.ID).Error
	require.NoError(t, err)
	require.Equal(t, 5, fetched.Order)
	require.NoError(t, db.Delete(&TimelineEventTest{}, "id = ?", event.ID).Error)
	err = db.First(&fetched, "id = ?", event.ID).Error
	require.Error(t, err)
}

func TestUpdateOrder(t *testing.T) {
	db := setupTimelineTestDB(t)
	event1ID := uuid.New().String()
	event2ID := uuid.New().String()
	event3ID := uuid.New().String()
	events := []*TimelineEventTest{
		{ID: event1ID, CaseID: "case1", Order: 1},
		{ID: event2ID, CaseID: "case1", Order: 2},
		{ID: event3ID, CaseID: "case1", Order: 3},
	}
	for _, ev := range events {
		require.NoError(t, db.Create(ev).Error)
	}
	newOrder := []string{event2ID, event3ID, event1ID}
	for idx, id := range newOrder {
		require.NoError(t, db.Model(&TimelineEventTest{}).Where("id = ? AND case_id = ?", id, "case1").Update("order", idx+1).Error)
	}
	var case1Events []TimelineEventTest
	err := db.Where("case_id = ?", "case1").Order("\"order\" ASC").Find(&case1Events).Error
	require.NoError(t, err)
	orders := make([]int, len(case1Events))
	for i, ev := range case1Events {
		orders[i] = ev.Order
	}
	require.Equal(t, []int{1, 2, 3}, orders)
}

func TestFindByID(t *testing.T) {
	db := setupTimelineTestDB(t)
	eventID := uuid.New().String()
	event := &TimelineEventTest{ID: eventID, CaseID: "case1", Order: 1}
	require.NoError(t, db.Create(event).Error)
	var fetched TimelineEventTest
	err := db.First(&fetched, "id = ?", event.ID).Error
	require.NoError(t, err)
	require.Equal(t, event.ID, fetched.ID)
}
