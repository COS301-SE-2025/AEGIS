package integration

// import (
// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"net/http"
// 	"net/http/httptest"
// 	"os"
// 	"testing"

// 	"github.com/gin-gonic/gin"
// 	"github.com/google/uuid"
// 	"github.com/stretchr/testify/assert"
// 	"gorm.io/driver/postgres"
// 	"gorm.io/gorm"
	
// 	"aegis-api/handlers"
	
// 	"aegis-api/services_/evidence/evidence_tag"
// )

// func setupTestDB(t *testing.T) *gorm.DB {
// 	dsn := fmt.Sprintf(
// 		"host=localhost port=%s user=%s password=%s dbname=%s sslmode=disable",
// 		os.Getenv("DB_PORT"),
// 		os.Getenv("DB_USER"),
// 		os.Getenv("DB_PASSWORD"),
// 		os.Getenv("DB_NAME"),
// 	)

// 	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
// 	if err != nil {
// 		t.Fatalf("failed to connect to test DB: %v", err)
// 	}
// 	return db
// }


// func TestEvidenceTagIntegration(t *testing.T) {
// 	db := setupTestDB(t)

// 	// Start transaction
// 	tx := db.Begin()
// 	defer tx.Rollback()

// 	// --- Step 1: Insert test user (with all required fields) ---
// 	testUserID := uuid.New()
// 	// Use unique email for each test run
// 	testEmail := fmt.Sprintf("testuser_%s@example.com", testUserID.String())
// 	err := tx.Exec(`
// 		INSERT INTO users (id, full_name, email, password_hash, role, is_verified) 
// 		VALUES (?, ?, ?, ?, ?, ?)
// 		ON CONFLICT (email) DO NOTHING`,
// 		testUserID, 
// 		"Test User", 
// 		testEmail, 
// 		"hashedpassword", 
// 		"Admin", 
// 		true,
// 	).Error
// 	if err != nil {
// 		t.Fatalf("failed to insert test user: %v", err)
// 	}

// 	// --- Step 2: Insert test case (required for case_id) ---
// 	testCaseID := uuid.New()
// 	err = tx.Exec(`
// 		INSERT INTO cases (
// 			id, title, description, TeamName, created_by
// 		) VALUES (?, ?, ?, ?, ?)`,
// 		testCaseID, "Test Case Title", "Short test description", "Test Team", testUserID,
// 	).Error
// 	if err != nil {
// 		t.Fatalf("failed to insert test case: %v", err)
// 	}

// 	// --- Step 3: Insert test evidence ---
// 	testEvidenceID := uuid.New()

	

// 	err = tx.Exec(`
// 		INSERT INTO evidence (
// 			id, case_id, uploaded_by, filename, file_type, ipfs_cid, file_size, checksum, metadata
// 		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
// 		testEvidenceID, testCaseID, testUserID,
// 		"testfile.pdf", "application/pdf", "QmFakeCID", 12345, "fakechecksum", `{"test": "data"}`,
// 	).Error
// 	if err != nil {
// 		t.Fatalf("failed to insert test evidence: %v", err)
// 	}

	

// 	// --- Step 4: Set up repository, service, handler ---
// 	repo := evidence_tag.NewEvidenceTagRepository(tx)
// 	service := evidence_tag.NewEvidenceTagService(repo)
// 	handler := &handlers.EvidenceTagHandler{Service: service}

// 	// --- Step 5: Setup Gin router with test-specific routes (no auth middleware) ---
// 	r := gin.Default()
// 	r.Use(func(c *gin.Context) {
// 		// Set context values that handlers might need
// 		c.Set("userID", testUserID.String())
// 		c.Set("role", "Admin")
// 		c.Next()
// 	})

// 	// Register test-specific routes without authentication middleware
// 	apiGroup := r.Group("/api")
// 	apiGroup.POST("/evidence-tags/tag", handler.TagEvidence)
// 	apiGroup.POST("/evidence-tags/untag", handler.UntagEvidence)
// 	apiGroup.GET("/evidence-tags/:evidence_id", handler.GetEvidenceTags)

// 	// --- Step 6: Test Tag Evidence ---
// 	body := map[string]interface{}{
// 		"evidence_id": testEvidenceID.String(),
// 		"tags":        []string{"Sensitive", "Urgent"},
// 	}
// 	bodyBytes, _ := json.Marshal(body)

// 	req, _ := http.NewRequest("POST", "/api/evidence-tags/tag", bytes.NewBuffer(bodyBytes))
// 	req.Header.Set("Content-Type", "application/json")

// 	w := httptest.NewRecorder()
// 	r.ServeHTTP(w, req)

// 	assert.Equal(t, http.StatusOK, w.Code)
// 	assert.Contains(t, w.Body.String(), "Evidence tagged successfully")

// 	// --- Step 7: Verify Get Evidence Tags ---
// 	getReq, _ := http.NewRequest("GET",
// 		fmt.Sprintf("/api/evidence-tags/%s", testEvidenceID.String()), nil)
// 	getResp := httptest.NewRecorder()
// 	r.ServeHTTP(getResp, getReq)

// 	assert.Equal(t, http.StatusOK, getResp.Code)
// 	assert.Contains(t, getResp.Body.String(), "\"sensitive\"")
// 	//assert.Contains(t, getResp.Body.String(), "Sensitive")
// 	assert.Contains(t, getResp.Body.String(), "\"urgent\"")

// 	// --- Step 8: Untag Evidence ---
// 	untagBody := map[string]interface{}{
// 		"evidence_id": testEvidenceID.String(),
// 		"tags":        []string{"Sensitive"},
// 	}
// 	untagBytes, _ := json.Marshal(untagBody)

// 	untagReq, _ := http.NewRequest("POST", "/api/evidence-tags/untag", bytes.NewBuffer(untagBytes))
// 	untagReq.Header.Set("Content-Type", "application/json")

// 	untagResp := httptest.NewRecorder()
// 	r.ServeHTTP(untagResp, untagReq)

// 	assert.Equal(t, http.StatusOK, untagResp.Code)
// 	assert.Contains(t, untagResp.Body.String(), "Evidence untagged successfully")

// 	// --- Step 9: Verify Tags After Untag ---
// 	getReq2, _ := http.NewRequest("GET",
// 		fmt.Sprintf("/api/evidence-tags/%s", testEvidenceID.String()), nil)
// 	getResp2 := httptest.NewRecorder()
// 	r.ServeHTTP(getResp2, getReq2)

// 	assert.Equal(t, http.StatusOK, getResp2.Code)
// 	assert.NotContains(t, getResp2.Body.String(), "Sensitive")
// 	assert.Contains(t, getResp2.Body.String(), "urgent")
// }