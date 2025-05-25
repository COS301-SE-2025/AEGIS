package services

import (
	"testing"
	"time"
	"aegis-api/db"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"aegis-api/services/ListCases"
)

func init() {
	if err := db.Connect(); err != nil {
		panic("❌ Failed to connect to DB: " + err.Error())
	}
}

func seedTestCases(t *testing.T) []ListCases.Case {
	cases := []ListCases.Case{
		{
			ID:                  uuid.New(),
			Title:               "Unauthorized Access Detected",
			Status:              "open",
			Priority:            "high",
			InvestigationStage:  "analysis",
			CreatedBy:           uuid.MustParse("8fb89568-3c52-4535-af33-d2f1266def52"),
			CreatedAt:           time.Now(),
		},
		{
			ID:                  uuid.New(),
			Title:               "Suspicious Network Activity",
			Status:              "under_review",
			Priority:            "medium",
			InvestigationStage:  "research",
			CreatedBy:           uuid.MustParse("8fb89568-3c52-4535-af33-d2f1266def52"),
			CreatedAt:           time.Now(),
		},
	}

	for _, c := range cases {
		if err := db.DB.Create(&c).Error; err != nil {
			t.Fatalf("❌ Failed to insert test case: %v", err)
		}
	}

	return cases
}

func TestGetFilteredCases_ByStatus(t *testing.T) {
	service := ListCases.NewListCasesService()
	seedTestCases(t)

	// Filter for "open" cases only
	results, err := service.GetFilteredCases("open", "", "", "", "created_at", "desc")
	assert.NoError(t, err)
	assert.NotEmpty(t, results)

	for _, c := range results {
		assert.Equal(t, "open", c.Status)
	}
}

func TestGetFilteredCases_ByPriority(t *testing.T) {
	service := ListCases.NewListCasesService()
	seedTestCases(t)

	results, err := service.GetFilteredCases("", "medium", "", "", "created_at", "desc")
	assert.NoError(t, err)
	assert.NotEmpty(t, results)

	for _, c := range results {
		assert.Equal(t, "medium", c.Priority)
	}
}
func TestGetFilteredCases_InvalidSortAndOrder(t *testing.T) {
	service := ListCases.NewListCasesService()

	// Should fall back to default (created_at desc)
	results, err := service.GetFilteredCases("", "", "", "", "invalid_field", "invalid_order")

	assert.NoError(t, err)
	assert.NotNil(t, results)
	t.Logf("✅ Handled invalid sort/order gracefully. Returned %d results.", len(results))
}
func TestGetFilteredCases_NoFilters(t *testing.T) {
	service := ListCases.NewListCasesService()

	results, err := service.GetFilteredCases("", "", "", "",  "","")
	assert.NoError(t, err)

	t.Logf("✅ Returned %d cases with no filters", len(results))
	assert.GreaterOrEqual(t, len(results), 1)
}

func TestGetFilteredCases_CombinedFilters(t *testing.T) {
	service := ListCases.NewListCasesService()

	// Make sure test data includes one open/high/createdBy case
	results, err := service.GetFilteredCases("open", "high", "ded0a1b3-4712-46b5-8d01-fafbaf3f8236", "", "created_at", "desc")

	assert.NoError(t, err)

	for _, c := range results {
		assert.Equal(t, "open", c.Status)
		assert.Equal(t, "high", c.Priority)
		assert.Equal(t, "ded0a1b3-4712-46b5-8d01-fafbaf3f8236", c.CreatedBy.String())
	}
	t.Logf("✅ Passed combined filter logic, returned %d case(s)", len(results))
}
func TestGetFilteredCases_TitleSearchMatch(t *testing.T) {
	service := ListCases.NewListCasesService()
	seedTestCases(t)

	titles := []string{"unauthorized", "suspicious", "", "notfound"}
	expectedMin := []int{1, 1, 2, 0}

	for i, term := range titles {
		results, err := service.GetFilteredCases("", "", "", term, "created_at", "desc")
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), expectedMin[i], "Failed for term: %s", term)
		t.Logf("✅ Title search '%s' returned %d result(s)", term, len(results))
	}
}

func tearDownCases(t *testing.T) {
	if err := db.DB.Exec("DELETE FROM cases").Error; err != nil {
		t.Fatalf("❌ Failed to clean up test cases: %v", err)
	}
}
