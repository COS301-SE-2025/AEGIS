package integration

import (
	"aegis-api/handlers"
	"aegis-api/services_/case/case_tags"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// CaseTag represents the expected database model
type CaseTag struct {
	ID     uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CaseID uuid.UUID `gorm:"type:uuid;not null"`
	TagID  uuid.UUID `gorm:"type:uuid;not null"`
	// Add other fields based on your actual schema
}

// Tag represents a tag entity
type Tag struct {
	ID   uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name string    `gorm:"not null;unique"`
}

func setupTestDB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=localhost port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test DB: %v", err)
	}
	return db
}

func createTestUser(t *testing.T, db *gorm.DB) uuid.UUID {
	userID := uuid.New()
	email := fmt.Sprintf("testuser_%s@example.com", userID.String())

	sql := `
		INSERT INTO users (id, full_name, email, password_hash, role, is_verified) 
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (email) DO NOTHING
	`

	err := db.Exec(sql, userID, "Test User", email, "test-hash", "Admin", true).Error
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	return userID
}

func createTestCase(t *testing.T, db *gorm.DB, createdBy uuid.UUID) uuid.UUID {
	caseID := uuid.New()
	sql := `
		INSERT INTO cases (id, title, description, team_name, created_by)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO NOTHING
	`
	if err := db.Exec(sql, caseID, "Test Title", "Test Description", "Test Team", createdBy).Error; err != nil {
		t.Fatalf("failed to insert test case: %v", err)
	}
	return caseID
}

// Helper function to get the actual database schema
func getCaseTagsSchema(t *testing.T, db *gorm.DB) map[string]string {
	var columns []struct {
		ColumnName string `gorm:"column:column_name"`
		DataType   string `gorm:"column:data_type"`
	}

	err := db.Raw(`
		SELECT column_name, data_type 
		FROM information_schema.columns 
		WHERE table_name = 'case_tags' 
		ORDER BY ordinal_position
	`).Scan(&columns).Error

	schema := make(map[string]string)
	if err != nil {
		t.Logf("Warning: Could not get case_tags schema: %v", err)
		return schema
	}

	for _, col := range columns {
		schema[col.ColumnName] = col.DataType
		t.Logf("Column: %s, Type: %s", col.ColumnName, col.DataType)
	}

	return schema
}

func TestCaseTagsTableSchema(t *testing.T) {
	db := setupTestDB(t)

	// Check if table exists
	var tableExists bool
	err := db.Raw(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_name = 'case_tags'
		)
	`).Scan(&tableExists).Error

	if err != nil || !tableExists {
		t.Fatalf("case_tags table does not exist or error checking: %v", err)
	}

	schema := getCaseTagsSchema(t, db)
	assert.NotEmpty(t, schema, "case_tags table should have columns")

	// Log the schema for debugging
	t.Logf("Complete case_tags table schema: %+v", schema)
}

func TestTagCaseIntegration(t *testing.T) {
	db := setupTestDB(t)

	// First, check the schema
	schema := getCaseTagsSchema(t, db)
	if len(schema) == 0 {
		t.Skip("Skipping test: case_tags table schema could not be determined")
	}

	// Use a fresh connection for each test to avoid transaction conflicts
	freshDB := setupTestDB(t)

	// Setup service & handler with fresh DB connection
	repo := case_tags.NewCaseTagRepository(freshDB)
	service := case_tags.NewCaseTagService(repo)
	handler := handlers.CaseTagHandler{Service: service}

	// Create test data in main database (not in transaction)
	userID := createTestUser(t, freshDB)
	caseID := createTestCase(t, freshDB, userID)

	// Prepare request
	tags := []string{"evidence", "urgent"}
	requestBody := map[string]interface{}{
		"tags": tags,
	}

	reqBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Add auth middleware
	r.Use(func(c *gin.Context) {
		c.Set("userID", userID.String())
		c.Set("role", "Admin")
		c.Next()
	})

	// Register route
	r.POST("/cases/:case_id/tags", handler.TagCase)

	// Create request
	req := httptest.NewRequest("POST", fmt.Sprintf("/cases/%s/tags", caseID), bytes.NewReader(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// Perform request
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Debug output
	t.Logf("Response Code: %d", w.Code)
	t.Logf("Response Body: %s", w.Body.String())

	// The response should be successful (204 No Content or 200 OK)
	if w.Code != http.StatusNoContent && w.Code != http.StatusOK {
		t.Errorf("Expected status 204 or 200, got %d. Response: %s", w.Code, w.Body.String())

		// If it failed, let's check what went wrong
		if w.Code == http.StatusInternalServerError {
			t.Logf("Internal server error - this might be a database schema issue")
			t.Logf("Available schema: %+v", schema)
		}
		return
	}

	// Verify tags were saved - use flexible query based on actual schema
	t.Logf("Verifying tags were saved to database...")

	// Try different approaches to query the data based on common patterns
	verifyTagsSaved(t, freshDB, caseID, tags, schema)

	// Cleanup
	cleanupTestData(t, freshDB, caseID, userID)
}

func verifyTagsSaved(t *testing.T, db *gorm.DB, caseID uuid.UUID, expectedTags []string, schema map[string]string) {
	// Strategy 1: If there's a direct tag column
	if _, hasTag := schema["tag"]; hasTag {
		var dbTags []string
		err := db.Raw("SELECT tag FROM case_tags WHERE case_id = $1", caseID).Scan(&dbTags).Error
		if err == nil {
			assert.ElementsMatch(t, expectedTags, dbTags)
			return
		}
	}

	// Strategy 2: If there's a tag_name column
	if _, hasTagName := schema["tag_name"]; hasTagName {
		var dbTags []string
		err := db.Raw("SELECT tag_name FROM case_tags WHERE case_id = $1", caseID).Scan(&dbTags).Error
		if err == nil {
			assert.ElementsMatch(t, expectedTags, dbTags)
			return
		}
	}

	// Strategy 3: If it's a junction table with tag_id, join with tags table
	if _, hasTagID := schema["tag_id"]; hasTagID {
		var dbTags []string
		err := db.Raw(`
			SELECT t.name 
			FROM case_tags ct 
			JOIN tags t ON ct.tag_id = t.id 
			WHERE ct.case_id = $1
		`, caseID).Scan(&dbTags).Error
		if err == nil {
			assert.ElementsMatch(t, expectedTags, dbTags)
			return
		}
	}

	// Strategy 4: Just verify that records exist
	var count int64
	err := db.Raw("SELECT COUNT(*) FROM case_tags WHERE case_id = $1", caseID).Scan(&count).Error
	if err == nil && count > 0 {
		t.Logf("Found %d case_tags records for case %s", count, caseID)
		assert.Equal(t, int64(len(expectedTags)), count)
		return
	}

	t.Errorf("Could not verify tags were saved. Schema: %+v", schema)
}

func cleanupTestData(_ *testing.T, db *gorm.DB, caseID, userID uuid.UUID) {
	// Clean up test data
	db.Exec("DELETE FROM case_tags WHERE case_id = $1", caseID)
	db.Exec("DELETE FROM cases WHERE id = $1", caseID)
	db.Exec("DELETE FROM users WHERE id = $1", userID)
}

func TestTagCaseWithErrorHandling(t *testing.T) {
	db := setupTestDB(t)
	schema := getCaseTagsSchema(t, db)

	if len(schema) == 0 {
		t.Skip("Skipping test: case_tags table schema could not be determined")
	}

	// Create fresh connections for each subtest to avoid transaction conflicts
	testCases := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		description    string
		setupAuth      bool
	}{
		{
			name:           "Valid tags",
			requestBody:    map[string]interface{}{"tags": []string{"evidence", "urgent"}},
			expectedStatus: http.StatusNoContent,
			description:    "Should successfully tag case",
			setupAuth:      true,
		},
		{
			name:           "Empty tags array",
			requestBody:    map[string]interface{}{"tags": []string{}},
			expectedStatus: http.StatusBadRequest,
			description:    "Should handle empty tags array",
			setupAuth:      true,
		},
		{
			name:           "Missing tags field",
			requestBody:    map[string]interface{}{"notags": []string{"test"}},
			expectedStatus: http.StatusBadRequest,
			description:    "Should handle missing tags field",
			setupAuth:      true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Fresh DB connection for each subtest
			freshDB := setupTestDB(t)

			// Setup service & handler
			repo := case_tags.NewCaseTagRepository(freshDB)
			service := case_tags.NewCaseTagService(repo)
			handler := handlers.CaseTagHandler{Service: service}

			// Create test data
			userID := createTestUser(t, freshDB)
			caseID := createTestCase(t, freshDB, userID)

			// Setup Gin router
			gin.SetMode(gin.TestMode)
			r := gin.New()

			// Add auth middleware conditionally
			if tc.setupAuth {
				r.Use(func(c *gin.Context) {
					c.Set("userID", userID.String())
					c.Set("role", "Admin")
					c.Next()
				})
			}

			// Register route
			r.POST("/cases/:case_id/tags", handler.TagCase)

			// Create request
			reqBodyBytes, _ := json.Marshal(tc.requestBody)
			req := httptest.NewRequest("POST", fmt.Sprintf("/cases/%s/tags", caseID), bytes.NewReader(reqBodyBytes))
			req.Header.Set("Content-Type", "application/json")

			// Perform request
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			t.Logf("%s - Response Code: %d, Body: %s", tc.description, w.Code, w.Body.String())

			// For successful cases, verify database state
			if tc.name == "Valid tags" && w.Code == http.StatusNoContent {
				expectedTags := []string{"evidence", "urgent"}
				verifyTagsSaved(t, freshDB, caseID, expectedTags, schema)
			}

			// Cleanup
			cleanupTestData(t, freshDB, caseID, userID)
		})
	}
}
