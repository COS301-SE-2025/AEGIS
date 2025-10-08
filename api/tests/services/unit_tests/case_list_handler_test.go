package unit_tests

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"aegis-api/handlers"
	"aegis-api/services_/auditlog"
	"aegis-api/services_/case/ListCases"
	listArchiveCases "aegis-api/services_/case/listArchiveCases"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ========== MOCKS ==========

// Mock services
type MockListCasesServiceListCases struct {
	mock.Mock
}

func (m *MockListCasesServiceListCases) GetAllCases(tenantID string) ([]ListCases.Case, error) {
	args := m.Called(tenantID)
	return args.Get(0).([]ListCases.Case), args.Error(1)
}

func (m *MockListCasesServiceListCases) GetCasesByUser(tenantID, userID string) ([]ListCases.Case, error) {
	args := m.Called(tenantID, userID)
	return args.Get(0).([]ListCases.Case), args.Error(1)
}

func (m *MockListCasesServiceListCases) GetFilteredCases(tenantID, status, priority, createdBy, teamName, titleTerm, sortBy, order, userID, teamID string) ([]ListCases.Case, error) {
	args := m.Called(tenantID, status, priority, createdBy, teamName, titleTerm, sortBy, order, userID, teamID)
	return args.Get(0).([]ListCases.Case), args.Error(1)
}

func (m *MockListCasesServiceListCases) GetCaseByID(caseID, tenantID string) (*ListCases.Case, error) {
	args := m.Called(caseID, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ListCases.Case), args.Error(1)
}

type MockListArchivedCasesService struct {
	mock.Mock
}

func (m *MockListArchivedCasesService) ListArchivedCases(userID, tenantID, teamID string) ([]listArchiveCases.ArchivedCase, error) {
	args := m.Called(userID, tenantID, teamID)
	return args.Get(0).([]listArchiveCases.ArchivedCase), args.Error(1)
}

type MockCacheListCases struct {
	mock.Mock
}

func (m *MockCacheListCases) Get(ctx context.Context, key string) (string, bool, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Bool(1), args.Error(2)
}

func (m *MockCacheListCases) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockCacheListCases) Del(ctx context.Context, keys ...string) (int, error) {
	args := m.Called(ctx, keys)
	return args.Int(0), args.Error(1)
}

// Mock dependencies for AuditLogger
type MockMongoLoggerListCases struct {
	mock.Mock
}

func (m *MockMongoLoggerListCases) Log(ctx *gin.Context, log auditlog.AuditLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

type MockZapLoggerListCases struct {
	mock.Mock
}

func (m *MockZapLoggerListCases) Log(log auditlog.AuditLog) {
	m.Called(log)
}

// ========== HELPER FUNCTIONS ==========

// Helper function to create test case data
func createTestCases() []ListCases.Case {
	return []ListCases.Case{
		{
			ID:          uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
			Title:       "Test Case 1",
			Description: "Test Description 1",
			Status:      "open",
			Priority:    "high",
			CreatedBy:   uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
			TenantID:    uuid.MustParse("550e8400-e29b-41d4-a716-446655440003"),
		},
		{
			ID:          uuid.MustParse("550e8400-e29b-41d4-a716-446655440004"),
			Title:       "Test Case 2",
			Description: "Test Description 2",
			Status:      "closed",
			Priority:    "medium",
			CreatedBy:   uuid.MustParse("550e8400-e29b-41d4-a716-446655440005"),
			TenantID:    uuid.MustParse("550e8400-e29b-41d4-a716-446655440003"),
		},
	}
}

func createTestArchivedCases() []listArchiveCases.ArchivedCase {
	return []listArchiveCases.ArchivedCase{
		{
			ID:          uuid.MustParse("550e8400-e29b-41d4-a716-446655440006"),
			Title:       "Archived Case 1",
			Description: "Archived Description 1",
			Status:      "archived",
			Priority:    "high",
			CreatedBy:   uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
			TenantID:    uuid.MustParse("550e8400-e29b-41d4-a716-446655440003"),
		},
		{
			ID:          uuid.MustParse("550e8400-e29b-41d4-a716-446655440007"),
			Title:       "Archived Case 2",
			Description: "Archived Description 2",
			Status:      "archived",
			Priority:    "medium",
			CreatedBy:   uuid.MustParse("550e8400-e29b-41d4-a716-446655440005"),
			TenantID:    uuid.MustParse("550e8400-e29b-41d4-a716-446655440003"),
		},
	}
}

// Create test context helper (similar to list_active_cases_handler_test.go)
func createTestContextListCases(userID, tenantID, teamID, userRole string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("GET", "/cases/all", nil)
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	if userID != "" {
		c.Set("userID", userID)
	}
	if tenantID != "" {
		c.Set("tenantID", tenantID)
	}
	if teamID != "" {
		c.Set("teamID", teamID)
	}
	if userRole != "" {
		c.Set("userRole", userRole)
		c.Set("role", userRole) // Some handlers might use "role" instead of "userRole"
	}
	if userID != "" {
		c.Set("email", "test@example.com")
	}

	return c, w
}

// Create test handler using constructor (like in list_active_cases_handler_test.go)
func createTestHandlerListCases(mockListCasesService *MockListCasesServiceListCases, mockListArchived *MockListArchivedCasesService, mockCache *MockCacheListCases) *handlers.CaseHandler {
	// Create real AuditLogger with mock dependencies
	mockMongo := &MockMongoLoggerListCases{}
	mockZap := &MockZapLoggerListCases{}

	// Set up the mocks to do nothing (or capture calls if needed)
	mockMongo.On("Log", mock.Anything, mock.Anything).Return(nil)
	mockZap.On("Log", mock.Anything).Return()

	realAuditLogger := auditlog.NewAuditLogger(mockMongo, mockZap)

	return handlers.NewCaseHandler(
		nil,                  // caseService
		mockListCasesService, // listCasesService
		nil,                  // listActiveCasesService
		nil,                  // listClosedCasesService
		mockListArchived,     // listArchivedCasesService
		realAuditLogger,      // auditLogger
		nil,                  // userRepo
		nil,                  // updateCaseService
		mockCache,            // cacheClient
	)
}

// ========== SUCCESS SCENARIOS ==========

func TestGetAllCasesHandler_Success(t *testing.T) {
	mockListCases := &MockListCasesServiceListCases{}
	mockListArchived := &MockListArchivedCasesService{}
	mockCache := &MockCacheListCases{}

	handler := createTestHandlerListCases(mockListCases, mockListArchived, mockCache)

	testCases := createTestCases()

	// Mock cache miss
	mockCache.On("Get", mock.Anything, mock.AnythingOfType("string")).Return("", false, nil)
	mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil)

	// Mock service success
	mockListCases.On("GetAllCases", "tenant-1").Return(testCases, nil)

	c, w := createTestContextListCases("test-user", "tenant-1", "team-1", "admin")

	handler.GetAllCasesHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "cases")
	assert.Contains(t, response, "meta")

	cases := response["cases"].([]interface{})
	assert.Len(t, cases, 2)

	mockListCases.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestGetAllCasesHandler_CacheHit(t *testing.T) {
	mockListCases := &MockListCasesServiceListCases{}
	mockListArchived := &MockListArchivedCasesService{}
	mockCache := &MockCacheListCases{}

	handler := createTestHandlerListCases(mockListCases, mockListArchived, mockCache)

	cachedResponse := `{"cases":[{"id":"550e8400-e29b-41d4-a716-446655440001","title":"Cached Case"}],"meta":{"page":"1"}}`

	// Mock cache hit
	mockCache.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(cachedResponse, true, nil)

	c, w := createTestContextListCases("test-user", "tenant-1", "team-1", "admin")

	handler.GetAllCasesHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("ETag"), "")
	assert.Equal(t, "private, max-age=120", w.Header().Get("Cache-Control"))

	// Service should not be called on cache hit
	mockListCases.AssertNotCalled(t, "GetAllCases")
	mockCache.AssertExpectations(t)
}

func TestGetAllCasesHandler_ServiceError(t *testing.T) {
	mockListCases := &MockListCasesServiceListCases{}
	mockListArchived := &MockListArchivedCasesService{}
	mockCache := &MockCacheListCases{}

	handler := createTestHandlerListCases(mockListCases, mockListArchived, mockCache)

	// Mock cache miss
	mockCache.On("Get", mock.Anything, mock.AnythingOfType("string")).Return("", false, nil)

	// Mock service error
	mockListCases.On("GetAllCases", "tenant-1").Return([]ListCases.Case{}, errors.New("service error"))

	c, w := createTestContextListCases("test-user", "tenant-1", "team-1", "admin")

	handler.GetAllCasesHandler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "could not retrieve cases", response["error"])

	mockListCases.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestGetCasesByUserHandler_Success(t *testing.T) {
	mockListCases := &MockListCasesServiceListCases{}
	mockListArchived := &MockListArchivedCasesService{}
	mockCache := &MockCacheListCases{}

	handler := createTestHandlerListCases(mockListCases, mockListArchived, mockCache)

	testCases := createTestCases()

	// Mock service success
	mockListCases.On("GetCasesByUser", "tenant-1", "user-123").Return(testCases, nil)

	c, w := createTestContextListCases("test-user", "tenant-1", "team-1", "admin")
	// Update the request URL to include the user_id parameter
	req, _ := http.NewRequest("GET", "/cases/user/user-123", nil)
	c.Request = req
	c.Params = []gin.Param{{Key: "user_id", Value: "user-123"}}

	handler.GetCasesByUserHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "cases")

	cases := response["cases"].([]interface{})
	assert.Len(t, cases, 2)

	mockListCases.AssertExpectations(t)
}

func TestGetFilteredCasesHandler_Success(t *testing.T) {
	mockListCases := &MockListCasesServiceListCases{}
	mockListArchived := &MockListArchivedCasesService{}
	mockCache := &MockCacheListCases{}

	handler := createTestHandlerListCases(mockListCases, mockListArchived, mockCache)

	testCases := createTestCases()

	// Mock service success
	mockListCases.On("GetFilteredCases",
		"tenant-1", "open", "high", "", "", "", "", "", "test-user", "team-1").Return(testCases, nil)

	c, w := createTestContextListCases("test-user", "tenant-1", "team-1", "admin")
	// Update the request URL to include query parameters
	req, _ := http.NewRequest("GET", "/cases/filter?status=open&priority=high", nil)
	c.Request = req

	handler.GetFilteredCasesHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "cases")

	mockListCases.AssertExpectations(t)
}

func TestGetCaseByIDHandler_Success(t *testing.T) {
	mockListCases := &MockListCasesServiceListCases{}
	mockListArchived := &MockListArchivedCasesService{}
	mockCache := &MockCacheListCases{}

	handler := createTestHandlerListCases(mockListCases, mockListArchived, mockCache)

	testCase := &ListCases.Case{
		ID:          uuid.MustParse("550e8400-e29b-41d4-a716-446655440008"),
		Title:       "Test Case",
		Description: "Test Description",
		Status:      "open",
		Priority:    "high",
		CreatedBy:   uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
		TenantID:    uuid.MustParse("550e8400-e29b-41d4-a716-446655440003"),
	}

	// Mock cache miss
	mockCache.On("Get", mock.Anything, mock.AnythingOfType("string")).Return("", false, nil)
	mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil)

	// Mock service success
	mockListCases.On("GetCaseByID", "550e8400-e29b-41d4-a716-446655440008", "tenant-1").Return(testCase, nil)

	c, w := createTestContextListCases("test-user", "tenant-1", "team-1", "admin")
	req, _ := http.NewRequest("GET", "/cases/550e8400-e29b-41d4-a716-446655440008", nil)
	c.Request = req
	c.Params = []gin.Param{{Key: "case_id", Value: "550e8400-e29b-41d4-a716-446655440008"}}

	handler.GetCaseByIDHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "case")

	mockListCases.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestListArchivedCases_Success(t *testing.T) {
	mockListCases := &MockListCasesServiceListCases{}
	mockListArchived := &MockListArchivedCasesService{}
	mockCache := &MockCacheListCases{}

	handler := createTestHandlerListCases(mockListCases, mockListArchived, mockCache)

	testArchivedCases := createTestArchivedCases()

	// Mock service success
	mockListArchived.On("ListArchivedCases", "test-user", "tenant-1", "team-1").Return(testArchivedCases, nil)

	c, w := createTestContextListCases("test-user", "tenant-1", "team-1", "admin")
	req, _ := http.NewRequest("GET", "/cases/archived", nil)
	c.Request = req

	handler.ListArchivedCasesHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "archived_cases")

	cases := response["archived_cases"].([]interface{})
	assert.Len(t, cases, 2)

	mockListArchived.AssertExpectations(t)
}

// ========== ERROR SCENARIOS ==========

func TestListArchivedCases_MissingCredentials(t *testing.T) {
	testCases := []struct {
		name     string
		userID   string
		tenantID string
		teamID   string
		expected int
	}{
		{"Missing UserID", "", "tenant", "team", http.StatusBadRequest},
		{"Missing TenantID", "user", "", "team", http.StatusBadRequest},
		{"Missing TeamID", "user", "tenant", "", http.StatusBadRequest},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockListCases := &MockListCasesServiceListCases{}
			mockListArchived := &MockListArchivedCasesService{}
			mockCache := &MockCacheListCases{}

			handler := createTestHandlerListCases(mockListCases, mockListArchived, mockCache)

			c, w := createTestContextListCases(tc.userID, tc.tenantID, tc.teamID, "admin")
			req, _ := http.NewRequest("GET", "/cases/archived", nil)
			c.Request = req

			handler.ListArchivedCasesHandler(c)

			assert.Equal(t, tc.expected, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Contains(t, response, "error")
		})
	}
}

func TestListArchivedCases_ServiceError(t *testing.T) {
	mockListCases := &MockListCasesServiceListCases{}
	mockListArchived := &MockListArchivedCasesService{}
	mockCache := &MockCacheListCases{}

	handler := createTestHandlerListCases(mockListCases, mockListArchived, mockCache)

	// Mock service error
	mockListArchived.On("ListArchivedCases", "test-user", "tenant-1", "team-1").Return([]listArchiveCases.ArchivedCase{}, errors.New("service error"))

	c, w := createTestContextListCases("test-user", "tenant-1", "team-1", "admin")
	req, _ := http.NewRequest("GET", "/cases/archived", nil)
	c.Request = req

	handler.ListArchivedCasesHandler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "could not retrieve archived cases", response["error"])

	mockListArchived.AssertExpectations(t)
}

// ========== EMPTY RESPONSE TESTS ==========

func TestGetAllCasesHandler_EmptyResponse(t *testing.T) {
	mockListCases := &MockListCasesServiceListCases{}
	mockListArchived := &MockListArchivedCasesService{}
	mockCache := &MockCacheListCases{}

	handler := createTestHandlerListCases(mockListCases, mockListArchived, mockCache)

	emptyCases := []ListCases.Case{}

	mockCache.On("Get", mock.Anything, mock.AnythingOfType("string")).Return("", false, nil)
	mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil)
	mockListCases.On("GetAllCases", "tenant-1").Return(emptyCases, nil)

	c, w := createTestContextListCases("test-user", "tenant-1", "team-1", "admin")

	handler.GetAllCasesHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	cases := response["cases"].([]interface{})
	assert.Len(t, cases, 0)

	mockListCases.AssertExpectations(t)
}
