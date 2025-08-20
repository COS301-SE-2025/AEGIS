package unit_tests

// import (
// 	"bytes"

// 	"encoding/json"
// 	"fmt"
// 	"net/http"
// 	"net/http/httptest"
// 	"os"
// 	"testing"
// 	"time"

// // Your database package
// 	"aegis-api/handlers"
// // Your ORM models
// 	"aegis-api/services_/report/update_status"

// 	"github.com/gin-gonic/gin"
// 	"github.com/google/uuid"
// 	"github.com/stretchr/testify/assert"
// 	"gorm.io/driver/postgres"
// 	"gorm.io/gorm"
// )

// // setupTestDB connects to a real PostgreSQL test database.
// // This is the core of an integration test.
// func setupTestDB(t *testing.T) *gorm.DB {
// 	dsn := fmt.Sprintf(
// 		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
// 		os.Getenv("DB_HOST_TEST"),
// 		os.Getenv("DB_PORT_TEST"),
// 		os.Getenv("DB_USER_TEST"),
// 		os.Getenv("DB_PASSWORD_TEST"),
// 		os.Getenv("DB_NAME_TEST"),
// 	)

// 	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
// 	if err != nil {
// 		t.Fatalf("failed to connect to test DB: %v", err)
// 	}
// 	return db
// }

// func TestUpdateReportStatusIntegration(t *testing.T) {
// 	gin.SetMode(gin.TestMode)

// 	// Step 1: Set up the test database and start a transaction.
// 	testDB := setupTestDB(t)
// 	tx := testDB.Begin()
// 	defer tx.Rollback() // Rollback everything to clean up

// 	// Step 2: Insert required data (fixtures) into the database.
// 	// You will need to insert a tenant, team, and examiner/user first
// 	// as these are likely foreign key dependencies for a report.
// 	testTenantID := uuid.New()
// 	tx.Exec(`INSERT INTO tenants (id, name, created_at, updated_at) VALUES (?, ?, ?, ?)`,
// 		testTenantID, "TestTenant", time.Now(), time.Now())

// 	testTeamID := uuid.New()
// 	tx.Exec(`INSERT INTO teams (id, name, tenant_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
// 		testTeamID, "TestTeam", testTenantID, time.Now(), time.Now())

// 	testExaminerID := uuid.New()
// 	tx.Exec(`INSERT INTO users (id, full_name, email, password_hash, role, team_id, tenant_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
// 		testExaminerID, "Test Examiner", "examiner@test.com", "hashedpassword", "Examiner", testTeamID, testTenantID, time.Now(), time.Now())

// 	testCaseID := uuid.New()
// 	tx.Exec(`INSERT INTO cases (id, title, created_by, tenant_id) VALUES (?, ?, ?, ?)`,
// 		testCaseID, "Test Case", testExaminerID, testTenantID)

// 	// Insert the report that we will update.
// 	reportID := uuid.New()
// 	now := time.Now()
// 	reportToInsert := update_status.Report{
// 		ID:           reportID,
// 		CaseID:       testCaseID,
// 		ExaminerID:   testExaminerID,
// 		Name:         "Integration Test Report",
// 		Status:       "draft",
// 		CreatedAt:    time.Now(),
// 		UpdatedAt:    time.Now(),
// 		TenantID:     testTenantID,
// 		MongoID:      uuid.New().String(), // Example required field
// 		ReportNumber: "123-IT",
// 		Version:      1,
// 		DateExamined: &now,
// 		FilePath:     "/test/path",
// 	}

// 	result := tx.Create(&reportToInsert)
// 	if result.Error != nil {
// 		t.Fatalf("Failed to insert test report fixture: %v", result.Error)
// 	}

// 	// Step 3: Set up the real dependencies (repository, service, handler, and router).
// 	reportRepo := update_status.NewReportStatusRepository(tx) // Pass the transaction
// 	reportService := update_status.NewReportStatusService(reportRepo)
// 	reportHandler := handlers.NewReportStatusHandler(reportService)

// 	r := gin.Default()
// 	reportHandler.RegisterRoutes(r)

// 	// Step 4: Define and send the HTTP request.
// 	w := httptest.NewRecorder()
// 	reqBody := map[string]string{"status": "review"}
// 	bodyBytes, _ := json.Marshal(reqBody)

// 	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/reports/%s/status", reportID.String()), bytes.NewBuffer(bodyBytes))
// 	req.Header.Set("Content-Type", "application/json")

// 	r.ServeHTTP(w, req)

// 	// Step 5: Assert the HTTP response is as expected.
// 	assert.Equal(t, http.StatusOK, w.Code)

// 	var responseBody map[string]interface{}
// 	err := json.Unmarshal(w.Body.Bytes(), &responseBody)
// 	assert.NoError(t, err)
// 	assert.Equal(t, "review", responseBody["status"])

// 	// Step 6: Verify the state change in the database.
// 	var updatedReport update_status.Report
// 	result = tx.First(&updatedReport, "id = ?", reportID)
// 	assert.NoError(t, result.Error)
// 	assert.Equal(t, "review", updatedReport.Status)
// }
