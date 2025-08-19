package integration

import (
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

	"aegis-api/handlers"
	"aegis-api/services_/case/case_evidence_totals"
)

func setupCaseEvidenceTestDB(t *testing.T) *gorm.DB {
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

func TestGetDashboardTotals(t *testing.T) {
	db := setupCaseEvidenceTestDB(t)
	tx := db.Begin()
	defer tx.Rollback()

	userID := uuid.New()
	caseID := uuid.New()
	evidenceID := uuid.New()

	// Insert test user
	err := tx.Exec(`
		INSERT INTO users (id, full_name, email, password_hash, role, is_verified) 
		VALUES (?, 'Tester', 'dashboard@example.com', 'hashedpwd', 'Admin', true)
		ON CONFLICT (email) DO NOTHING
	`, userID).Error
	assert.NoError(t, err)

	// Insert one valid case (status = open)
	err = tx.Exec(`
		INSERT INTO cases (id, title, team_name, created_by,status)
		VALUES (?, 'Included Case', 'TeamX', ?, 'open')
	`, caseID, userID).Error
	assert.NoError(t, err)

	// Insert one excluded case (status = archived)
	err = tx.Exec(`
		INSERT INTO cases (id, title, team_name, created_by,status)
		VALUES (?, 'Excluded Case', 'TeamX', ?, 'archived')
	`, uuid.New(), userID).Error
	assert.NoError(t, err)

	// Insert evidence attached to included case
	err = tx.Exec(`
		INSERT INTO evidence (
			id, case_id, uploaded_by, filename, file_type, ipfs_cid, file_size, checksum, metadata
		) VALUES (?, ?, ?, 'dash.pdf', 'application/pdf', 'fakeCID', 1024, 'checksum123', '{}')
	`, evidenceID, caseID, userID).Error
	assert.NoError(t, err)

	// Setup everything else
	repo := case_evidence_totals.NewCaseEviRepository(tx)
	service := case_evidence_totals.NewDashboardService(repo)
	handler := handlers.NewCaseEvidenceTotalsHandler(service)

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Set("userID", userID.String()) // <- context-based user
		c.Set("role", "Admin")
		c.Next()
	})
	r.GET("/api/dashboard/totals", handler.GetDashboardTotals)

	// Test default filtering (should include "open", "ongoing", "closed")
	req, _ := http.NewRequest("GET", "/api/dashboard/totals", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, 1, int(response["case_count"].(float64)))      // should NOT include "archived"
	assert.Equal(t, 1, int(response["evidence_count"].(float64)))  // filtered by uploaded_by

	// Test with query param override (e.g., exclude "open")
	req2, _ := http.NewRequest("GET", "/api/dashboard/totals?statuses=archived", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)

	var response2 map[string]interface{}
	err = json.Unmarshal(w2.Body.Bytes(), &response2)
	assert.NoError(t, err)

	assert.Equal(t, 1, int(response2["case_count"].(float64))) // Only archived now
	assert.Equal(t, 1, int(response2["evidence_count"].(float64))) // Evidence only linked to "open"
}
