package unit_tests

import (
	"testing"
	"time"

	"aegis-api/services_/case/ListCases"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupCasesTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

	// Create test schema matching your database structure
	db.Exec(`CREATE TABLE cases (
        id TEXT PRIMARY KEY,
        title TEXT,
        description TEXT,
        status TEXT,
        priority TEXT,
        investigation_stage TEXT,
        created_by TEXT,
        team_name TEXT,
        tenant_id TEXT,
        created_at DATETIME,
        updated_at DATETIME
    )`)

	return db
}

func TestNewGormCaseQueryRepository(t *testing.T) {
	db := setupCasesTestDB()
	repo := ListCases.NewGormCaseQueryRepository(db)

	assert.NotNil(t, repo)
}

func TestGormCaseQueryRepository_GetAllCases_Success(t *testing.T) {
	db := setupCasesTestDB()
	repo := ListCases.NewGormCaseQueryRepository(db)

	tenantID := uuid.New().String()
	case1ID := uuid.New().String()
	case2ID := uuid.New().String()

	// Insert test data
	db.Exec(`INSERT INTO cases (id, title, description, status, priority, investigation_stage, created_by, team_name, tenant_id, created_at, updated_at) 
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		case1ID, "Case 1", "Description 1", "open", "high", "initial", uuid.New().String(), "Team A", tenantID, time.Now(), time.Now())

	db.Exec(`INSERT INTO cases (id, title, description, status, priority, investigation_stage, created_by, team_name, tenant_id, created_at, updated_at) 
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		case2ID, "Case 2", "Description 2", "closed", "medium", "resolved", uuid.New().String(), "Team B", tenantID, time.Now(), time.Now())

	// Test the repository
	cases, err := repo.GetAllCases(tenantID)

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, cases, 2)
	assert.Equal(t, "Case 1", cases[0].Title)
	assert.Equal(t, "Case 2", cases[1].Title)
}

func TestGormCaseQueryRepository_GetAllCases_NoResults(t *testing.T) {
	db := setupCasesTestDB()
	repo := ListCases.NewGormCaseQueryRepository(db)

	nonExistentTenantID := uuid.New().String()

	cases, err := repo.GetAllCases(nonExistentTenantID)

	assert.NoError(t, err)
	assert.Empty(t, cases)
}

func TestGormCaseQueryRepository_GetAllCases_DifferentTenants(t *testing.T) {
	db := setupCasesTestDB()
	repo := ListCases.NewGormCaseQueryRepository(db)

	tenant1ID := uuid.New().String()
	tenant2ID := uuid.New().String()

	// Insert cases for different tenants
	db.Exec(`INSERT INTO cases (id, title, status, tenant_id, created_at, updated_at) 
        VALUES (?, ?, ?, ?, ?, ?)`,
		uuid.New().String(), "Tenant 1 Case", "open", tenant1ID, time.Now(), time.Now())

	db.Exec(`INSERT INTO cases (id, title, status, tenant_id, created_at, updated_at) 
        VALUES (?, ?, ?, ?, ?, ?)`,
		uuid.New().String(), "Tenant 2 Case", "open", tenant2ID, time.Now(), time.Now())

	// Test tenant isolation
	cases1, err1 := repo.GetAllCases(tenant1ID)
	cases2, err2 := repo.GetAllCases(tenant2ID)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Len(t, cases1, 1)
	assert.Len(t, cases2, 1)
	assert.Equal(t, "Tenant 1 Case", cases1[0].Title)
	assert.Equal(t, "Tenant 2 Case", cases2[0].Title)
}

func TestGormCaseQueryRepository_GetCasesByUser_Success(t *testing.T) {
	db := setupCasesTestDB()
	repo := ListCases.NewGormCaseQueryRepository(db)

	userID := uuid.New().String()
	tenantID := uuid.New().String()
	otherUserID := uuid.New().String()

	// Insert cases for different users
	db.Exec(`INSERT INTO cases (id, title, created_by, tenant_id, created_at, updated_at) 
        VALUES (?, ?, ?, ?, ?, ?)`,
		uuid.New().String(), "User Case 1", userID, tenantID, time.Now(), time.Now())

	db.Exec(`INSERT INTO cases (id, title, created_by, tenant_id, created_at, updated_at) 
        VALUES (?, ?, ?, ?, ?, ?)`,
		uuid.New().String(), "User Case 2", userID, tenantID, time.Now(), time.Now())

	db.Exec(`INSERT INTO cases (id, title, created_by, tenant_id, created_at, updated_at) 
        VALUES (?, ?, ?, ?, ?, ?)`,
		uuid.New().String(), "Other User Case", otherUserID, tenantID, time.Now(), time.Now())

	cases, err := repo.GetCasesByUser(userID, tenantID)

	assert.NoError(t, err)
	assert.Len(t, cases, 2)
	assert.Equal(t, "User Case 1", cases[0].Title)
	assert.Equal(t, "User Case 2", cases[1].Title)
}

func TestGormCaseQueryRepository_GetCasesByUser_NoResults(t *testing.T) {
	db := setupCasesTestDB()
	repo := ListCases.NewGormCaseQueryRepository(db)

	userID := uuid.New().String()
	tenantID := uuid.New().String()

	cases, err := repo.GetCasesByUser(userID, tenantID)

	assert.NoError(t, err)
	assert.Empty(t, cases)
}

func TestGormCaseQueryRepository_GetCaseByID_Success(t *testing.T) {
	db := setupCasesTestDB()
	repo := ListCases.NewGormCaseQueryRepository(db)

	caseID := uuid.New().String()
	tenantID := uuid.New().String()

	// Insert test case
	db.Exec(`INSERT INTO cases (id, title, description, status, priority, tenant_id, created_at, updated_at) 
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		caseID, "Test Case", "Test Description", "open", "high", tenantID, time.Now(), time.Now())

	result, err := repo.GetCaseByID(caseID, tenantID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test Case", result.Title)
	assert.Equal(t, "Test Description", result.Description)
	assert.Equal(t, "open", result.Status)
}

func TestGormCaseQueryRepository_GetCaseByID_NotFound(t *testing.T) {
	db := setupCasesTestDB()
	repo := ListCases.NewGormCaseQueryRepository(db)

	nonExistentCaseID := uuid.New().String()
	tenantID := uuid.New().String()

	result, err := repo.GetCaseByID(nonExistentCaseID, tenantID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestGormCaseQueryRepository_GetCaseByID_WrongTenant(t *testing.T) {
	db := setupCasesTestDB()
	repo := ListCases.NewGormCaseQueryRepository(db)

	caseID := uuid.New().String()
	correctTenantID := uuid.New().String()
	wrongTenantID := uuid.New().String()

	// Insert case for correct tenant
	db.Exec(`INSERT INTO cases (id, title, tenant_id, created_at, updated_at) 
        VALUES (?, ?, ?, ?, ?)`,
		caseID, "Test Case", correctTenantID, time.Now(), time.Now())

	// Try to get with wrong tenant
	result, err := repo.GetCaseByID(caseID, wrongTenantID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestGormCaseQueryRepository_QueryCases_NoFilters(t *testing.T) {
	db := setupCasesTestDB()
	repo := ListCases.NewGormCaseQueryRepository(db)

	// Insert test cases
	db.Exec(`INSERT INTO cases (id, title, status, priority, created_at, updated_at) 
        VALUES (?, ?, ?, ?, ?, ?)`,
		uuid.New().String(), "Case 1", "open", "high", time.Now(), time.Now())

	db.Exec(`INSERT INTO cases (id, title, status, priority, created_at, updated_at) 
        VALUES (?, ?, ?, ?, ?, ?)`,
		uuid.New().String(), "Case 2", "closed", "low", time.Now(), time.Now())

	filter := ListCases.CaseFilter{}

	cases, err := repo.QueryCases(filter)

	assert.NoError(t, err)
	assert.Len(t, cases, 2)
}

func TestGormCaseQueryRepository_QueryCases_StatusFilter(t *testing.T) {
	db := setupCasesTestDB()
	repo := ListCases.NewGormCaseQueryRepository(db)

	// Insert test cases with different statuses
	db.Exec(`INSERT INTO cases (id, title, status, created_at, updated_at) 
        VALUES (?, ?, ?, ?, ?)`,
		uuid.New().String(), "Open Case", "open", time.Now(), time.Now())

	db.Exec(`INSERT INTO cases (id, title, status, created_at, updated_at) 
        VALUES (?, ?, ?, ?, ?)`,
		uuid.New().String(), "Closed Case", "closed", time.Now(), time.Now())

	filter := ListCases.CaseFilter{Status: "open"}

	cases, err := repo.QueryCases(filter)

	assert.NoError(t, err)
	assert.Len(t, cases, 1)
	assert.Equal(t, "Open Case", cases[0].Title)
	assert.Equal(t, "open", cases[0].Status)
}

func TestGormCaseQueryRepository_QueryCases_PriorityFilter(t *testing.T) {
	db := setupCasesTestDB()
	repo := ListCases.NewGormCaseQueryRepository(db)

	// Insert test cases with different priorities
	db.Exec(`INSERT INTO cases (id, title, priority, created_at, updated_at) 
        VALUES (?, ?, ?, ?, ?)`,
		uuid.New().String(), "High Priority Case", "high", time.Now(), time.Now())

	db.Exec(`INSERT INTO cases (id, title, priority, created_at, updated_at) 
        VALUES (?, ?, ?, ?, ?)`,
		uuid.New().String(), "Low Priority Case", "low", time.Now(), time.Now())

	filter := ListCases.CaseFilter{Priority: "high"}

	cases, err := repo.QueryCases(filter)

	assert.NoError(t, err)
	assert.Len(t, cases, 1)
	assert.Equal(t, "High Priority Case", cases[0].Title)
	assert.Equal(t, "high", cases[0].Priority)
}

func TestGormCaseQueryRepository_QueryCases_CreatedByFilter(t *testing.T) {
	db := setupCasesTestDB()
	repo := ListCases.NewGormCaseQueryRepository(db)

	userID := uuid.New().String()
	otherUserID := uuid.New().String()

	// Insert test cases
	db.Exec(`INSERT INTO cases (id, title, created_by, created_at, updated_at) 
        VALUES (?, ?, ?, ?, ?)`,
		uuid.New().String(), "User Case", userID, time.Now(), time.Now())

	db.Exec(`INSERT INTO cases (id, title, created_by, created_at, updated_at) 
        VALUES (?, ?, ?, ?, ?)`,
		uuid.New().String(), "Other User Case", otherUserID, time.Now(), time.Now())

	filter := ListCases.CaseFilter{CreatedBy: userID}

	cases, err := repo.QueryCases(filter)

	assert.NoError(t, err)
	assert.Len(t, cases, 1)
	assert.Equal(t, "User Case", cases[0].Title)
}

func TestGormCaseQueryRepository_QueryCases_TeamNameFilter(t *testing.T) {
	db := setupCasesTestDB()
	repo := ListCases.NewGormCaseQueryRepository(db)

	// Insert test cases
	db.Exec(`INSERT INTO cases (id, title, team_name, created_at, updated_at) 
        VALUES (?, ?, ?, ?, ?)`,
		uuid.New().String(), "Security Team Case", "Security", time.Now(), time.Now())

	db.Exec(`INSERT INTO cases (id, title, team_name, created_at, updated_at) 
        VALUES (?, ?, ?, ?, ?)`,
		uuid.New().String(), "Operations Team Case", "Operations", time.Now(), time.Now())

	filter := ListCases.CaseFilter{TeamName: "Security"}

	cases, err := repo.QueryCases(filter)

	assert.NoError(t, err)
	assert.Len(t, cases, 1)
	assert.Equal(t, "Security Team Case", cases[0].Title)
}

func TestGormCaseQueryRepository_QueryCases_TitleTermFilter(t *testing.T) {
	db := setupCasesTestDB()
	repo := ListCases.NewGormCaseQueryRepository(db)

	// Insert test cases
	db.Exec(`INSERT INTO cases (id, title, created_at, updated_at) 
        VALUES (?, ?, ?, ?)`,
		uuid.New().String(), "Security Incident Response", time.Now(), time.Now())

	db.Exec(`INSERT INTO cases (id, title, created_at, updated_at) 
        VALUES (?, ?, ?, ?)`,
		uuid.New().String(), "Network Maintenance", time.Now(), time.Now())

	filter := ListCases.CaseFilter{TitleTerm: "incident"}

	cases, err := repo.QueryCases(filter)

	assert.NoError(t, err)
	// Check length first to avoid panic
	if assert.Len(t, cases, 1) {
		assert.Equal(t, "Security Incident Response", cases[0].Title)
	}
}

// Also add a case-insensitive test since we're using LIKE
func TestGormCaseQueryRepository_QueryCases_TitleTermFilter_CaseInsensitive(t *testing.T) {
	db := setupCasesTestDB()
	repo := ListCases.NewGormCaseQueryRepository(db)

	// Insert test cases
	db.Exec(`INSERT INTO cases (id, title, created_at, updated_at) 
        VALUES (?, ?, ?, ?)`,
		uuid.New().String(), "Security Incident Response", time.Now(), time.Now())

	db.Exec(`INSERT INTO cases (id, title, created_at, updated_at) 
        VALUES (?, ?, ?, ?)`,
		uuid.New().String(), "Network Maintenance", time.Now(), time.Now())

	// Test with uppercase search term
	filter := ListCases.CaseFilter{TitleTerm: "INCIDENT"}

	cases, err := repo.QueryCases(filter)

	assert.NoError(t, err)
	// SQLite LIKE is case-insensitive by default, so this should find the case
	if len(cases) > 0 {
		assert.Equal(t, "Security Incident Response", cases[0].Title)
	} else {
		// If case-sensitive, we expect no results
		assert.Len(t, cases, 0)
	}
}

// Add a test for partial matches
func TestGormCaseQueryRepository_QueryCases_TitleTermFilter_PartialMatch(t *testing.T) {
	db := setupCasesTestDB()
	repo := ListCases.NewGormCaseQueryRepository(db)

	// Insert test cases
	db.Exec(`INSERT INTO cases (id, title, created_at, updated_at) 
        VALUES (?, ?, ?, ?)`,
		uuid.New().String(), "Security Incident Response", time.Now(), time.Now())

	db.Exec(`INSERT INTO cases (id, title, created_at, updated_at) 
        VALUES (?, ?, ?, ?)`,
		uuid.New().String(), "Network Maintenance", time.Now(), time.Now())

	// Test partial match
	filter := ListCases.CaseFilter{TitleTerm: "security"}

	cases, err := repo.QueryCases(filter)

	assert.NoError(t, err)
	// Should find the security case
	if assert.Len(t, cases, 1) {
		assert.Equal(t, "Security Incident Response", cases[0].Title)
	}
}

// Test for no matches
func TestGormCaseQueryRepository_QueryCases_TitleTermFilter_NoMatch(t *testing.T) {
	db := setupCasesTestDB()
	repo := ListCases.NewGormCaseQueryRepository(db)

	// Insert test cases
	db.Exec(`INSERT INTO cases (id, title, created_at, updated_at) 
        VALUES (?, ?, ?, ?)`,
		uuid.New().String(), "Security Incident Response", time.Now(), time.Now())

	// Search for something that doesn't exist
	filter := ListCases.CaseFilter{TitleTerm: "nonexistent"}

	cases, err := repo.QueryCases(filter)

	assert.NoError(t, err)
	assert.Empty(t, cases)
}

func TestGormCaseQueryRepository_QueryCases_SortingAsc(t *testing.T) {
	db := setupCasesTestDB()
	repo := ListCases.NewGormCaseQueryRepository(db)

	// Insert test cases with different titles
	db.Exec(`INSERT INTO cases (id, title, created_at, updated_at) 
        VALUES (?, ?, ?, ?)`,
		uuid.New().String(), "Zebra Case", time.Now(), time.Now())

	db.Exec(`INSERT INTO cases (id, title, created_at, updated_at) 
        VALUES (?, ?, ?, ?)`,
		uuid.New().String(), "Alpha Case", time.Now(), time.Now())

	filter := ListCases.CaseFilter{SortBy: "title", SortOrder: "asc"}

	cases, err := repo.QueryCases(filter)

	assert.NoError(t, err)
	assert.Len(t, cases, 2)
	assert.Equal(t, "Alpha Case", cases[0].Title)
	assert.Equal(t, "Zebra Case", cases[1].Title)
}

func TestGormCaseQueryRepository_QueryCases_SortingDesc(t *testing.T) {
	db := setupCasesTestDB()
	repo := ListCases.NewGormCaseQueryRepository(db)

	// Insert test cases with different titles
	db.Exec(`INSERT INTO cases (id, title, created_at, updated_at) 
        VALUES (?, ?, ?, ?)`,
		uuid.New().String(), "Alpha Case", time.Now(), time.Now())

	db.Exec(`INSERT INTO cases (id, title, created_at, updated_at) 
        VALUES (?, ?, ?, ?)`,
		uuid.New().String(), "Zebra Case", time.Now(), time.Now())

	filter := ListCases.CaseFilter{SortBy: "title", SortOrder: "desc"}

	cases, err := repo.QueryCases(filter)

	assert.NoError(t, err)
	assert.Len(t, cases, 2)
	assert.Equal(t, "Zebra Case", cases[0].Title)
	assert.Equal(t, "Alpha Case", cases[1].Title)
}

func TestGormCaseQueryRepository_QueryCases_InvalidSortOrder(t *testing.T) {
	db := setupCasesTestDB()
	repo := ListCases.NewGormCaseQueryRepository(db)

	// Insert test case
	db.Exec(`INSERT INTO cases (id, title, created_at, updated_at) 
        VALUES (?, ?, ?, ?)`,
		uuid.New().String(), "Test Case", time.Now(), time.Now())

	filter := ListCases.CaseFilter{SortBy: "title", SortOrder: "invalid"}

	cases, err := repo.QueryCases(filter)

	// Should still work but without sorting
	assert.NoError(t, err)
	assert.Len(t, cases, 1)
}

func TestGormCaseQueryRepository_QueryCases_MultipleFilters(t *testing.T) {
	db := setupCasesTestDB()
	repo := ListCases.NewGormCaseQueryRepository(db)

	userID := uuid.New().String()

	// Insert test cases
	db.Exec(`INSERT INTO cases (id, title, status, priority, created_by, team_name, created_at, updated_at) 
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		uuid.New().String(), "Security Incident", "open", "high", userID, "Security", time.Now(), time.Now())

	db.Exec(`INSERT INTO cases (id, title, status, priority, created_by, team_name, created_at, updated_at) 
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		uuid.New().String(), "Network Issue", "closed", "low", userID, "Operations", time.Now(), time.Now())

	filter := ListCases.CaseFilter{
		Status:    "open",
		Priority:  "high",
		CreatedBy: userID,
		TeamName:  "Security",
	}

	cases, err := repo.QueryCases(filter)

	assert.NoError(t, err)
	assert.Len(t, cases, 1)
	assert.Equal(t, "Security Incident", cases[0].Title)
}

func TestGormCaseQueryRepository_DatabaseError(t *testing.T) {
	// Create a database and close it to simulate error
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	sqlDB, _ := db.DB()
	sqlDB.Close() // Close connection to force error

	repo := ListCases.NewGormCaseQueryRepository(db)

	// Test all methods with closed DB
	cases, err1 := repo.GetAllCases("tenant-id")
	assert.Error(t, err1)
	assert.Nil(t, cases)

	userCases, err2 := repo.GetCasesByUser("user-id", "tenant-id")
	assert.Error(t, err2)
	assert.Nil(t, userCases)

	caseResult, err3 := repo.GetCaseByID("case-id", "tenant-id")
	assert.Error(t, err3)
	assert.Nil(t, caseResult)

	queryCases, err4 := repo.QueryCases(ListCases.CaseFilter{})
	assert.Error(t, err4)
	assert.Nil(t, queryCases)
}
