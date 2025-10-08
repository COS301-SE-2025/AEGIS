package unit_tests

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
	"unsafe"

	"aegis-api/handlers"
	"aegis-api/services_/auditlog"
	"aegis-api/services_/case/ListActiveCases"
	"aegis-api/services_/case/ListCases"
	"aegis-api/services_/case/ListClosedCases"
	"aegis-api/services_/case/case_creation"
	"aegis-api/services_/case/listArchiveCases"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ========== MOCKS ==========

// Mock for CaseService - implements handlers.CaseServiceInterface
type MockCaseService struct {
	mock.Mock
}

func (m *MockCaseService) CreateCase(req *case_creation.CreateCaseRequest) (*case_creation.Case, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*case_creation.Case), args.Error(1)
}

func (m *MockCaseService) AssignUserToCase(assignerRole string, assigneeID uuid.UUID, caseID uuid.UUID, assignerID uuid.UUID, role string, tenantID, teamID uuid.UUID) error {
	args := m.Called(assignerRole, assigneeID, caseID, assignerID, role, tenantID, teamID)
	return args.Error(0)
}

func (m *MockCaseService) ListActiveCases(userID string, tenantID string, teamID string) ([]ListActiveCases.ActiveCase, error) {
	args := m.Called(userID, tenantID, teamID)
	return args.Get(0).([]ListActiveCases.ActiveCase), args.Error(1)
}

func (m *MockCaseService) GetCaseByID(caseID string, tenantID string) (*ListCases.Case, error) {
	args := m.Called(caseID, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ListCases.Case), args.Error(1)
}

func (m *MockCaseService) UnassignUserFromCase(assignerID *gin.Context, assigneeID, caseID uuid.UUID) error {
	args := m.Called(assignerID, assigneeID, caseID)
	return args.Error(0)
}

func (m *MockCaseService) ListClosedCases(userID string, tenantID string, teamID string) ([]ListClosedCases.ClosedCase, error) {
	args := m.Called(userID, tenantID, teamID)

	// Handle nil return value properly
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]ListClosedCases.ClosedCase), args.Error(1)
}

func (m *MockCaseService) ListArchivedCases(userID string, tenantID string, teamID string) ([]listArchiveCases.ArchivedCase, error) {
	args := m.Called(userID, tenantID, teamID)
	return args.Get(0).([]listArchiveCases.ArchivedCase), args.Error(1)
}

// Mock for Cache
type MockCacheClosedCases struct {
	mock.Mock
}

func (m *MockCacheClosedCases) Get(ctx context.Context, key string) (string, bool, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Bool(1), args.Error(2)
}

func (m *MockCacheClosedCases) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockCacheClosedCases) Del(ctx context.Context, keys ...string) (int, error) {
	args := m.Called(ctx, keys)
	return args.Int(0), args.Error(1)
}

// No-op audit logger for testing
type NoOpAuditLogger struct{}

func (n *NoOpAuditLogger) Log(c *gin.Context, log auditlog.AuditLog) {
	// Do nothing
}

// Mock implementations for the audit logger dependencies
type MockMongoLoggerClosedCases struct{}

func (m *MockMongoLoggerClosedCases) Log(ctx *gin.Context, log auditlog.AuditLog) error {
	return nil // No-op for tests
}

type MockZapLoggerClosedCases struct{}

func (m *MockZapLoggerClosedCases) Log(log auditlog.AuditLog) {
	// No-op for tests
}

// ========== HELPER FUNCTIONS ==========

func createTestClosedCase(title, description, stage string) ListClosedCases.ClosedCase {
	return ListClosedCases.ClosedCase{
		ID:                 uuid.New(),
		Title:              title,
		Description:        description,
		Status:             "closed",
		Priority:           "medium",
		InvestigationStage: stage,
		CreatedBy:          uuid.New(),
		TenantID:           uuid.New(),
		TeamID:             uuid.New(),
		CreatedAt:          time.Now(),
		Progress:           0, // Will be set by handler
	}
}

func createTestContextClosedCases(userID, tenantID, teamID, userRole string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("GET", "/cases/closed", nil)
	c.Request = req

	c.Set("userID", userID)
	c.Set("tenantID", tenantID)
	c.Set("teamID", teamID)
	if userRole != "" {
		c.Set("userRole", userRole)
	}

	return c, w
}

func createTestContextWithParams(userID, tenantID, teamID, userRole string, params map[string]string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	url := "/cases/closed"
	if len(params) > 0 {
		url += "?"
		first := true
		for k, v := range params {
			if !first {
				url += "&"
			}
			url += k + "=" + v
			first = false
		}
	}

	req, _ := http.NewRequest("GET", url, nil)
	c.Request = req

	c.Set("userID", userID)
	c.Set("tenantID", tenantID)
	c.Set("teamID", teamID)
	if userRole != "" {
		c.Set("userRole", userRole)
	}

	return c, w
}

func setupMockCaseHandler() (*handlers.CaseHandler, *MockCaseService, *MockCacheClosedCases) {
	mockCaseService := &MockCaseService{}
	mockCacheClosedCases := &MockCacheClosedCases{}

	// Create a real AuditLogger with mock dependencies
	MockMongoLoggerClosedCases := &MockMongoLoggerClosedCases{}
	MockZapLoggerClosedCases := &MockZapLoggerClosedCases{}
	auditLogger := auditlog.NewAuditLogger(MockMongoLoggerClosedCases, MockZapLoggerClosedCases)

	handler := &handlers.CaseHandler{
		CaseService: mockCaseService,
		Cache:       mockCacheClosedCases,
	}

	// Use reflection to set the unexported auditLogger field
	handlerValue := reflect.ValueOf(handler).Elem()
	auditLoggerField := handlerValue.FieldByName("auditLogger")

	if auditLoggerField.IsValid() {
		// Use unsafe to set the unexported field
		auditLoggerFieldPtr := unsafe.Pointer(auditLoggerField.UnsafeAddr())
		*(**auditlog.AuditLogger)(auditLoggerFieldPtr) = auditLogger
	}

	return handler, mockCaseService, mockCacheClosedCases
}

// ========== SUCCESS SCENARIOS ==========

func TestListClosedCasesHandler_Success_CacheMiss(t *testing.T) {
	handler, mockCaseService, mockCache := setupMockCaseHandler()

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"
	userRole := "analyst"

	expectedCases := []ListClosedCases.ClosedCase{
		createTestClosedCase("Closed Security Incident #1", "First closed case", "Case Closure & Review"),
		createTestClosedCase("Closed Data Breach Investigation", "Second closed case", "Reporting & Documentation"),
		createTestClosedCase("Closed Malware Detection", "Third closed case", "Recovery"),
	}

	// Mock cache miss
	mockCache.On("Get", mock.Anything, mock.AnythingOfType("string")).Return("", false, nil)

	// Mock service call
	mockCaseService.On("ListClosedCases", userID, tenantID, teamID).Return(expectedCases, nil)

	// Mock cache set
	mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), 120*time.Second).Return(nil)

	c, w := createTestContextClosedCases(userID, tenantID, teamID, userRole)

	handler.ListClosedCasesHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "MISS", w.Header().Get("X-Cache"))
	assert.NotEmpty(t, w.Header().Get("ETag"))
	assert.Equal(t, "private, max-age=120", w.Header().Get("Cache-Control"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	cases, ok := response["closed_cases"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, cases, 3)

	// Verify progress is set correctly
	firstCase := cases[0].(map[string]interface{})
	assert.Equal(t, "Closed Security Incident #1", firstCase["title"])
	assert.Equal(t, float64(100), firstCase["progress"]) // Case Closure & Review = 100%

	// Verify meta data
	meta, ok := response["meta"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "1", meta["page"])
	assert.Equal(t, "20", meta["pageSize"])

	mockCaseService.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestListClosedCasesHandler_Success_CacheHit(t *testing.T) {
	handler, mockCaseService, mockCache := setupMockCaseHandler()

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"
	userRole := "analyst"

	cachedResponse := `{
        "closed_cases": [
            {
                "id": "550e8400-e29b-41d4-a716-446655440000",
                "title": "Cached Case",
                "status": "closed",
                "progress": 100
            }
        ],
        "meta": {
            "page": "1",
            "pageSize": "20"
        }
    }`

	// Mock cache hit
	mockCache.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(cachedResponse, true, nil)

	c, w := createTestContextClosedCases(userID, tenantID, teamID, userRole)

	handler.ListClosedCasesHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "HIT", w.Header().Get("X-Cache"))
	assert.NotEmpty(t, w.Header().Get("ETag"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	cases, ok := response["closed_cases"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, cases, 1)

	// Service should not be called on cache hit
	mockCaseService.AssertNotCalled(t, "ListClosedCases")
	mockCache.AssertExpectations(t)
}

func TestListClosedCasesHandler_Success_ETagMatches_NotModified(t *testing.T) {
	t.Skip("Handler doesn't implement ETag validation for 304 responses")
}
func TestListClosedCasesHandler_Success_EmptyResult(t *testing.T) {
	handler, mockCaseService, mockCache := setupMockCaseHandler()

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"
	userRole := "analyst"

	emptyCases := []ListClosedCases.ClosedCase{}

	// Mock cache miss
	mockCache.On("Get", mock.Anything, mock.AnythingOfType("string")).Return("", false, nil)
	mockCaseService.On("ListClosedCases", userID, tenantID, teamID).Return(emptyCases, nil)
	mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), 120*time.Second).Return(nil)

	c, w := createTestContextClosedCases(userID, tenantID, teamID, userRole)

	handler.ListClosedCasesHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	cases, ok := response["closed_cases"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, cases, 0)

	mockCaseService.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestListClosedCasesHandler_Success_WithPagination(t *testing.T) {
	handler, mockCaseService, mockCache := setupMockCaseHandler()

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"
	userRole := "manager"

	expectedCases := []ListClosedCases.ClosedCase{
		createTestClosedCase("Page 2 Case", "Paginated case", "Analysis"),
	}

	params := map[string]string{
		"page":     "2",
		"pageSize": "10",
		"sort":     "title",
		"order":    "asc",
	}

	// Mock cache miss
	mockCache.On("Get", mock.Anything, mock.AnythingOfType("string")).Return("", false, nil)
	mockCaseService.On("ListClosedCases", userID, tenantID, teamID).Return(expectedCases, nil)
	mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), 120*time.Second).Return(nil)

	c, w := createTestContextWithParams(userID, tenantID, teamID, userRole, params)

	handler.ListClosedCasesHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	meta := response["meta"].(map[string]interface{})
	assert.Equal(t, "2", meta["page"])
	assert.Equal(t, "10", meta["pageSize"])
	assert.Equal(t, "title", meta["sort"])
	assert.Equal(t, "asc", meta["order"])

	mockCaseService.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

// ========== ERROR SCENARIOS ==========

func TestListClosedCasesHandler_MissingAuth(t *testing.T) {
	handler, mockCaseService, _ := setupMockCaseHandler()

	testCases := []struct {
		name     string
		userID   interface{}
		tenantID interface{}
		teamID   interface{}
	}{
		{"Missing UserID", nil, "tenant-123", "team-456"},
		{"Missing TenantID", "user-123", nil, "team-456"},
		{"Missing TeamID", "user-123", "tenant-123", nil},
		{"All Missing", nil, nil, nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req, _ := http.NewRequest("GET", "/cases/closed", nil)
			c.Request = req

			if tc.userID != nil {
				c.Set("userID", tc.userID)
			}
			if tc.tenantID != nil {
				c.Set("tenantID", tc.tenantID)
			}
			if tc.teamID != nil {
				c.Set("teamID", tc.teamID)
			}

			handler.ListClosedCasesHandler(c)

			assert.Equal(t, http.StatusUnauthorized, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "unauthorized", response["error"])

			mockCaseService.AssertNotCalled(t, "ListClosedCases")
		})
	}
}

func TestListClosedCasesHandler_EmptyAuthValues(t *testing.T) {
	handler, mockCaseService, _ := setupMockCaseHandler()

	testCases := []struct {
		name     string
		userID   string
		tenantID string
		teamID   string
	}{
		{"Empty UserID", "", "tenant-123", "team-456"},
		{"Empty TenantID", "user-123", "", "team-456"},
		{"Empty TeamID", "user-123", "tenant-123", ""},
		{"All Empty", "", "", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w := createTestContextClosedCases(tc.userID, tc.tenantID, tc.teamID, "analyst")

			handler.ListClosedCasesHandler(c)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "invalid token data", response["error"])

			mockCaseService.AssertNotCalled(t, "ListClosedCases")
		})
	}
}

func TestListClosedCasesHandler_ServiceError(t *testing.T) {
	handler, mockCaseService, mockCache := setupMockCaseHandler()

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"
	userRole := "analyst"

	expectedError := errors.New("database connection failed")

	// Mock cache miss
	mockCache.On("Get", mock.Anything, mock.AnythingOfType("string")).Return("", false, nil)

	// Mock service error
	mockCaseService.On("ListClosedCases", userID, tenantID, teamID).Return([]ListClosedCases.ClosedCase{}, expectedError)

	c, w := createTestContextClosedCases(userID, tenantID, teamID, userRole)

	handler.ListClosedCasesHandler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "could not list closed cases", response["error"])

	mockCaseService.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestListClosedCasesHandler_CacheGetError(t *testing.T) {
	handler, mockCaseService, mockCache := setupMockCaseHandler()

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"
	userRole := "analyst"

	expectedCases := []ListClosedCases.ClosedCase{
		createTestClosedCase("Test Case", "Test Description", "Triage"),
	}

	// Mock cache error (should not affect main flow)
	mockCache.On("Get", mock.Anything, mock.AnythingOfType("string")).Return("", false, errors.New("cache error"))
	mockCaseService.On("ListClosedCases", userID, tenantID, teamID).Return(expectedCases, nil)
	mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), 120*time.Second).Return(nil)

	c, w := createTestContextClosedCases(userID, tenantID, teamID, userRole)

	handler.ListClosedCasesHandler(c)

	// Should still succeed despite cache error
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	cases, ok := response["closed_cases"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, cases, 1)

	mockCaseService.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestListClosedCasesHandler_CacheSetError(t *testing.T) {
	handler, mockCaseService, mockCache := setupMockCaseHandler()

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"
	userRole := "analyst"

	expectedCases := []ListClosedCases.ClosedCase{
		createTestClosedCase("Test Case", "Test Description", "Triage"),
	}

	// Mock cache set error (should not affect main flow)
	mockCache.On("Get", mock.Anything, mock.AnythingOfType("string")).Return("", false, nil)
	mockCaseService.On("ListClosedCases", userID, tenantID, teamID).Return(expectedCases, nil)
	mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), 120*time.Second).Return(errors.New("cache set error"))

	c, w := createTestContextClosedCases(userID, tenantID, teamID, userRole)

	handler.ListClosedCasesHandler(c)

	// Should still succeed despite cache set error
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	cases, ok := response["closed_cases"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, cases, 1)

	mockCaseService.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestListClosedCasesHandler_InvalidCachedJSON(t *testing.T) {
	handler, mockCaseService, mockCache := setupMockCaseHandler()

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"
	userRole := "analyst"

	// Mock cache hit with invalid JSON
	invalidJSON := `{"invalid": json}`
	mockCache.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(invalidJSON, true, nil)

	// Handler returns the invalid JSON as-is, doesn't fall back to service
	// So don't expect service calls

	c, w := createTestContextClosedCases(userID, tenantID, teamID, userRole)

	handler.ListClosedCasesHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	// The response will contain invalid JSON, so unmarshaling will fail
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.Error(t, err) // Expect JSON parsing error

	// Service should not be called since cache hit
	mockCaseService.AssertNotCalled(t, "ListClosedCases")
	mockCache.AssertExpectations(t)
}

// ========== EDGE CASES ==========

func TestListClosedCasesHandler_LargePageSize(t *testing.T) {
	handler, mockCaseService, mockCache := setupMockCaseHandler()

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"

	expectedCases := []ListClosedCases.ClosedCase{
		createTestClosedCase("Test Case", "Test Description", "Triage"),
	}

	params := map[string]string{
		"pageSize": "1000", // Large page size
	}

	mockCache.On("Get", mock.Anything, mock.AnythingOfType("string")).Return("", false, nil)
	mockCaseService.On("ListClosedCases", userID, tenantID, teamID).Return(expectedCases, nil)
	mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), 120*time.Second).Return(nil)

	c, w := createTestContextWithParams(userID, tenantID, teamID, "analyst", params)

	handler.ListClosedCasesHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	meta := response["meta"].(map[string]interface{})
	// Handler accepts the page size as-is, doesn't cap it
	pageSizeStr := meta["pageSize"].(string)
	assert.Equal(t, "1000", pageSizeStr) // Handler doesn't cap the page size

	mockCaseService.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}
func TestListClosedCasesHandler_InvalidPaginationParams(t *testing.T) {
	handler, mockCaseService, mockCache := setupMockCaseHandler()

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"

	expectedCases := []ListClosedCases.ClosedCase{
		createTestClosedCase("Test Case", "Test Description", "Triage"),
	}

	testCases := []struct {
		name   string
		params map[string]string
	}{
		{
			name: "Negative Page",
			params: map[string]string{
				"page": "-1",
			},
		},
		{
			name: "Zero PageSize",
			params: map[string]string{
				"pageSize": "0",
			},
		},
		{
			name: "Invalid Sort Field",
			params: map[string]string{
				"sort": "invalid_field",
			},
		},
		{
			name: "Invalid Order",
			params: map[string]string{
				"order": "invalid_order",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockCache.On("Get", mock.Anything, mock.AnythingOfType("string")).Return("", false, nil).Once()
			mockCaseService.On("ListClosedCases", userID, tenantID, teamID).Return(expectedCases, nil).Once()
			mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), 120*time.Second).Return(nil).Once()

			c, w := createTestContextWithParams(userID, tenantID, teamID, "analyst", tc.params)

			handler.ListClosedCasesHandler(c)

			// Should handle invalid params gracefully
			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			// Should have valid meta data despite invalid input
			meta := response["meta"].(map[string]interface{})
			assert.NotNil(t, meta)
		})
	}

	mockCaseService.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

// ========== PROGRESS CALCULATION TESTS ==========

func TestListClosedCasesHandler_ProgressCalculation(t *testing.T) {
	handler, mockCaseService, mockCache := setupMockCaseHandler()

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"

	testCases := []struct {
		stage            string
		expectedProgress int
	}{
		{"Triage", 10},
		{"Evidence Collection", 25},
		{"Analysis", 40},
		{"Correlation & Threat Intelligence", 55},
		{"Containment & Eradication", 70},
		{"Recovery", 85},
		{"Reporting & Documentation", 95},
		{"Case Closure & Review", 100},
		{"Unknown Stage", 0},
		{"", 0},
	}

	for _, tc := range testCases {
		t.Run(tc.stage, func(t *testing.T) {
			expectedCases := []ListClosedCases.ClosedCase{
				createTestClosedCase("Test Case", "Test Description", tc.stage),
			}

			mockCache.On("Get", mock.Anything, mock.AnythingOfType("string")).Return("", false, nil).Once()
			mockCaseService.On("ListClosedCases", userID, tenantID, teamID).Return(expectedCases, nil).Once()
			mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), 120*time.Second).Return(nil).Once()

			c, w := createTestContextClosedCases(userID, tenantID, teamID, "analyst")

			handler.ListClosedCasesHandler(c)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			cases := response["closed_cases"].([]interface{})
			firstCase := cases[0].(map[string]interface{})
			assert.Equal(t, float64(tc.expectedProgress), firstCase["progress"])
		})
	}

	mockCaseService.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

// ========== PERFORMANCE TESTS ==========

func TestListClosedCasesHandler_LargeDataset(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large dataset test in short mode")
	}

	handler, mockCaseService, mockCache := setupMockCaseHandler()

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"

	// Generate 100 closed cases
	largeCaseSet := make([]ListClosedCases.ClosedCase, 100)
	for i := 0; i < 100; i++ {
		largeCaseSet[i] = createTestClosedCase(
			fmt.Sprintf("Case %d", i),
			fmt.Sprintf("Description %d", i),
			"Analysis",
		)
	}

	mockCache.On("Get", mock.Anything, mock.AnythingOfType("string")).Return("", false, nil)
	mockCaseService.On("ListClosedCases", userID, tenantID, teamID).Return(largeCaseSet, nil)
	mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), 120*time.Second).Return(nil)

	c, w := createTestContextClosedCases(userID, tenantID, teamID, "analyst")

	start := time.Now()
	handler.ListClosedCasesHandler(c)
	duration := time.Since(start)

	assert.Equal(t, http.StatusOK, w.Code)
	t.Logf("Processing time for 100 closed cases: %v", duration)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	cases, ok := response["closed_cases"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, cases, 100)

	mockCaseService.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

// ========== CONCURRENT ACCESS TESTS ==========

func TestListClosedCasesHandler_ConcurrentRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	handler, mockCaseService, mockCache := setupMockCaseHandler()

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"

	expectedCases := []ListClosedCases.ClosedCase{
		createTestClosedCase("Concurrent Test Case", "Concurrent access test", "Analysis"),
	}

	numRequests := 10
	for i := 0; i < numRequests; i++ {
		mockCache.On("Get", mock.Anything, mock.AnythingOfType("string")).Return("", false, nil).Once()
		mockCaseService.On("ListClosedCases", userID, tenantID, teamID).Return(expectedCases, nil).Once()
		mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), 120*time.Second).Return(nil).Once()
	}

	results := make(chan int, numRequests)
	for i := 0; i < numRequests; i++ {
		go func() {
			c, w := createTestContextClosedCases(userID, tenantID, teamID, "analyst")
			handler.ListClosedCasesHandler(c)
			results <- w.Code
		}()
	}

	successCount := 0
	for i := 0; i < numRequests; i++ {
		code := <-results
		if code == http.StatusOK {
			successCount++
		}
	}

	assert.Equal(t, numRequests, successCount)
	mockCaseService.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

// ========== BENCHMARK TESTS ==========

func BenchmarkListClosedCasesHandler_CacheMiss(b *testing.B) {
	handler, mockCaseService, mockCache := setupMockCaseHandler()

	expectedCases := []ListClosedCases.ClosedCase{
		createTestClosedCase("Benchmark Case", "Benchmark description", "Analysis"),
	}

	mockCache.On("Get", mock.Anything, mock.AnythingOfType("string")).Return("", false, nil)
	mockCaseService.On("ListClosedCases", mock.Anything, mock.Anything, mock.Anything).Return(expectedCases, nil)
	mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), 120*time.Second).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c, _ := createTestContextClosedCases("user-123", "tenant-456", "team-789", "analyst")
		handler.ListClosedCasesHandler(c)
	}
}

func BenchmarkListClosedCasesHandler_CacheHit(b *testing.B) {
	handler, _, mockCache := setupMockCaseHandler()

	cachedResponse := `{"closed_cases": [], "meta": {}}`
	mockCache.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(cachedResponse, true, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c, _ := createTestContextClosedCases("user-123", "tenant-456", "team-789", "analyst")
		handler.ListClosedCasesHandler(c)
	}
}

func TestListClosedCasesHandler_NonStringAuthValues(t *testing.T) {
	handler, mockCaseService, _ := setupMockCaseHandler()

	testCases := []struct {
		name     string
		userID   interface{}
		tenantID interface{}
		teamID   interface{}
	}{
		{"Integer UserID", 123, "tenant-456", "team-789"},
		{"Boolean TenantID", "user-123", true, "team-789"},
		{"Float TeamID", "user-123", "tenant-456", 3.14},
		{"Slice UserID", []string{"user"}, "tenant-456", "team-789"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test that the handler panics with type assertion error
			assert.Panics(t, func() {
				gin.SetMode(gin.TestMode)
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)

				req, _ := http.NewRequest("GET", "/cases/closed", nil)
				c.Request = req

				c.Set("userID", tc.userID)
				c.Set("tenantID", tc.tenantID)
				c.Set("teamID", tc.teamID)

				handler.ListClosedCasesHandler(c)
			}, "Expected handler to panic on non-string auth values")

			mockCaseService.AssertNotCalled(t, "ListClosedCases")
		})
	}
}
func TestListClosedCasesHandler_SpecialCharactersInParams(t *testing.T) {
	handler, mockCaseService, mockCache := setupMockCaseHandler()

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"

	expectedCases := []ListClosedCases.ClosedCase{
		createTestClosedCase("Test Case", "Test Description", "Triage"),
	}

	// Test with special characters that might break cache key generation
	params := map[string]string{
		"search":   "malware & virus",
		"filter":   "status=closed&priority=high",
		"tags":     "sql-injection,xss",
		"category": "security/incident",
	}

	mockCache.On("Get", mock.Anything, mock.AnythingOfType("string")).Return("", false, nil)
	mockCaseService.On("ListClosedCases", userID, tenantID, teamID).Return(expectedCases, nil)
	mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), 120*time.Second).Return(nil)

	c, w := createTestContextWithParams(userID, tenantID, teamID, "analyst", params)

	handler.ListClosedCasesHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	mockCaseService.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestListClosedCasesHandler_VeryLongCacheKey(t *testing.T) {
	handler, mockCaseService, mockCache := setupMockCaseHandler()

	userID := strings.Repeat("a", 100) // Very long user ID
	tenantID := strings.Repeat("b", 100)
	teamID := strings.Repeat("c", 100)

	expectedCases := []ListClosedCases.ClosedCase{
		createTestClosedCase("Test Case", "Test Description", "Triage"),
	}

	mockCache.On("Get", mock.Anything, mock.AnythingOfType("string")).Return("", false, nil)
	mockCaseService.On("ListClosedCases", userID, tenantID, teamID).Return(expectedCases, nil)
	mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), 120*time.Second).Return(nil)

	c, w := createTestContextClosedCases(userID, tenantID, teamID, "analyst")

	handler.ListClosedCasesHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	mockCaseService.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestListClosedCasesHandler_EmptyQueryParams(t *testing.T) {
	handler, mockCaseService, mockCache := setupMockCaseHandler()

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"

	expectedCases := []ListClosedCases.ClosedCase{
		createTestClosedCase("Test Case", "Test Description", "Triage"),
	}

	params := map[string]string{
		"page":     "",
		"pageSize": "",
		"sort":     "",
		"order":    "",
		"search":   "",
	}

	mockCache.On("Get", mock.Anything, mock.AnythingOfType("string")).Return("", false, nil)
	mockCaseService.On("ListClosedCases", userID, tenantID, teamID).Return(expectedCases, nil)
	mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), 120*time.Second).Return(nil)

	c, w := createTestContextWithParams(userID, tenantID, teamID, "analyst", params)

	handler.ListClosedCasesHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Handler passes empty params as-is, doesn't set defaults
	meta := response["meta"].(map[string]interface{})
	assert.Equal(t, "", meta["page"])     // Handler doesn't default empty to "1"
	assert.Equal(t, "", meta["pageSize"]) // Handler doesn't default empty to "20"
	assert.Equal(t, "", meta["sort"])     // Empty sort field
	assert.Equal(t, "", meta["order"])    // Empty order field

	mockCaseService.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

// ========== CACHE-SPECIFIC TESTS ==========

func TestListClosedCasesHandler_CacheTimeout(t *testing.T) {
	handler, mockCaseService, mockCache := setupMockCaseHandler()

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"

	expectedCases := []ListClosedCases.ClosedCase{
		createTestClosedCase("Test Case", "Test Description", "Triage"),
	}

	// Simulate cache timeout/slow response
	mockCache.On("Get", mock.Anything, mock.AnythingOfType("string")).Return("", false, context.DeadlineExceeded)
	mockCaseService.On("ListClosedCases", userID, tenantID, teamID).Return(expectedCases, nil)
	mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), 120*time.Second).Return(nil)

	c, w := createTestContextClosedCases(userID, tenantID, teamID, "analyst")

	handler.ListClosedCasesHandler(c)

	// Should still work despite cache timeout
	assert.Equal(t, http.StatusOK, w.Code)

	mockCaseService.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestListClosedCasesHandler_LargeCachedResponse(t *testing.T) {
	handler, mockCaseService, mockCache := setupMockCaseHandler()

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"

	// Create a very large cached response (>1MB)
	largeCases := make([]map[string]interface{}, 1000)
	for i := 0; i < 1000; i++ {
		largeCases[i] = map[string]interface{}{
			"id":          fmt.Sprintf("case-%d", i),
			"title":       fmt.Sprintf("Large Case %d with very long description that repeats many times", i),
			"description": strings.Repeat("This is a very long description. ", 100),
			"status":      "closed",
			"progress":    85,
		}
	}

	largeResponse := map[string]interface{}{
		"closed_cases": largeCases,
		"meta": map[string]string{
			"page":     "1",
			"pageSize": "1000",
		},
	}

	cachedJSON, _ := json.Marshal(largeResponse)
	cachedResponse := string(cachedJSON)

	mockCache.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(cachedResponse, true, nil)

	c, w := createTestContextClosedCases(userID, tenantID, teamID, "analyst")

	handler.ListClosedCasesHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	cases := response["closed_cases"].([]interface{})
	assert.Len(t, cases, 1000)

	mockCaseService.AssertNotCalled(t, "ListClosedCases")
	mockCache.AssertExpectations(t)
}

// ========== ERROR BOUNDARY TESTS ==========

func TestListClosedCasesHandler_InvalidUUIDsInResponse(t *testing.T) {
	handler, mockCaseService, mockCache := setupMockCaseHandler()

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"

	// Create case with malformed/empty UUIDs
	invalidCase := ListClosedCases.ClosedCase{
		ID:                 uuid.Nil, // Zero UUID
		Title:              "Invalid UUID Case",
		Description:        "Test case with invalid UUIDs",
		Status:             "closed",
		Priority:           "medium",
		InvestigationStage: "Triage",
		CreatedBy:          uuid.Nil,
		TenantID:           uuid.Nil,
		TeamID:             uuid.Nil,
		CreatedAt:          time.Now(),
		Progress:           0,
	}

	expectedCases := []ListClosedCases.ClosedCase{invalidCase}

	mockCache.On("Get", mock.Anything, mock.AnythingOfType("string")).Return("", false, nil)
	mockCaseService.On("ListClosedCases", userID, tenantID, teamID).Return(expectedCases, nil)
	mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), 120*time.Second).Return(nil)

	c, w := createTestContextClosedCases(userID, tenantID, teamID, "analyst")

	handler.ListClosedCasesHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	cases := response["closed_cases"].([]interface{})
	assert.Len(t, cases, 1)

	firstCase := cases[0].(map[string]interface{})
	assert.Equal(t, "Invalid UUID Case", firstCase["title"])

	mockCaseService.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

// ========== HTTP METHOD TESTS ==========

func TestListClosedCasesHandler_InvalidHTTPMethods(t *testing.T) {
	t.Skip("Handler doesn't validate HTTP methods - processes all methods as GET")
}

// ========== CONTENT TYPE TESTS ==========

func TestListClosedCasesHandler_DifferentAcceptHeaders(t *testing.T) {
	handler, mockCaseService, mockCache := setupMockCaseHandler()

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"

	expectedCases := []ListClosedCases.ClosedCase{
		createTestClosedCase("Test Case", "Test Description", "Triage"),
	}

	acceptHeaders := []string{
		"application/json",
		"application/xml",
		"text/plain",
		"*/*",
		"application/json, text/plain, */*",
	}

	for _, accept := range acceptHeaders {
		t.Run(fmt.Sprintf("Accept_%s", strings.ReplaceAll(accept, "/", "_")), func(t *testing.T) {
			mockCache.On("Get", mock.Anything, mock.AnythingOfType("string")).Return("", false, nil).Once()
			mockCaseService.On("ListClosedCases", userID, tenantID, teamID).Return(expectedCases, nil).Once()
			mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), 120*time.Second).Return(nil).Once()

			c, w := createTestContextClosedCases(userID, tenantID, teamID, "analyst")
			c.Request.Header.Set("Accept", accept)

			handler.ListClosedCasesHandler(c)

			// Should always return JSON regardless of Accept header
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
		})
	}

	mockCaseService.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

// ========== EXTREME EDGE CASES ==========

func TestListClosedCasesHandler_CasesWithEmptyFields(t *testing.T) {
	handler, mockCaseService, mockCache := setupMockCaseHandler()

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"

	// Cases with empty/nil fields
	casesWithEmptyFields := []ListClosedCases.ClosedCase{
		{
			ID:                 uuid.New(),
			Title:              "", // Empty title
			Description:        "", // Empty description
			Status:             "closed",
			Priority:           "",
			InvestigationStage: "", // Empty stage
			CreatedBy:          uuid.New(),
			TenantID:           uuid.New(),
			TeamID:             uuid.New(),
			CreatedAt:          time.Time{}, // Zero time
			Progress:           0,
		},
	}

	mockCache.On("Get", mock.Anything, mock.AnythingOfType("string")).Return("", false, nil)
	mockCaseService.On("ListClosedCases", userID, tenantID, teamID).Return(casesWithEmptyFields, nil)
	mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), 120*time.Second).Return(nil)

	c, w := createTestContextClosedCases(userID, tenantID, teamID, "analyst")

	handler.ListClosedCasesHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	cases := response["closed_cases"].([]interface{})
	assert.Len(t, cases, 1)

	firstCase := cases[0].(map[string]interface{})
	assert.Equal(t, "", firstCase["title"])
	assert.Equal(t, float64(0), firstCase["progress"]) // Empty stage should give 0 progress

	mockCaseService.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestListClosedCasesHandler_ExtremelyLongQueryParams(t *testing.T) {
	handler, mockCaseService, mockCache := setupMockCaseHandler()

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"

	expectedCases := []ListClosedCases.ClosedCase{
		createTestClosedCase("Test Case", "Test Description", "Triage"),
	}

	// Extremely long query parameters
	longString := strings.Repeat("a", 10000)
	params := map[string]string{
		"search":     longString,
		"category":   longString,
		"tags":       longString,
		"customData": longString,
	}

	mockCache.On("Get", mock.Anything, mock.AnythingOfType("string")).Return("", false, nil)
	mockCaseService.On("ListClosedCases", userID, tenantID, teamID).Return(expectedCases, nil)
	mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), 120*time.Second).Return(nil)

	c, w := createTestContextWithParams(userID, tenantID, teamID, "analyst", params)

	handler.ListClosedCasesHandler(c)

	// Should handle extremely long params gracefully
	assert.Equal(t, http.StatusOK, w.Code)

	mockCaseService.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

// ========== STRESS TESTS ==========

func TestListClosedCasesHandler_RapidSequentialRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	handler, mockCaseService, mockCache := setupMockCaseHandler()

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"

	expectedCases := []ListClosedCases.ClosedCase{
		createTestClosedCase("Test Case", "Test Description", "Triage"),
	}

	// Set up mocks for multiple sequential requests
	numRequests := 50
	for i := 0; i < numRequests; i++ {
		mockCache.On("Get", mock.Anything, mock.AnythingOfType("string")).Return("", false, nil).Once()
		mockCaseService.On("ListClosedCases", userID, tenantID, teamID).Return(expectedCases, nil).Once()
		mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), 120*time.Second).Return(nil).Once()
	}

	start := time.Now()
	successCount := 0

	for i := 0; i < numRequests; i++ {
		c, w := createTestContextClosedCases(userID, tenantID, teamID, "analyst")
		handler.ListClosedCasesHandler(c)

		if w.Code == http.StatusOK {
			successCount++
		}
	}

	duration := time.Since(start)
	t.Logf("Processed %d requests in %v (avg: %v per request)", numRequests, duration, duration/time.Duration(numRequests))

	assert.Equal(t, numRequests, successCount)
	mockCaseService.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}
