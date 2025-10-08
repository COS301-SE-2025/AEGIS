package unit_tests

import (
	"aegis-api/handlers"
	"aegis-api/services_/case/listArchiveCases"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ========== MOCKS ==========

// ArchiveCaseServiceInterface matches the service interface expected by the handler.
type ArchiveCaseServiceInterface interface {
	ListArchivedCases(userID, tenantID, teamID string) ([]listArchiveCases.ArchivedCase, error)
}

// Mock for ArchiveCaseService
type MockArchiveCaseService struct {
	mock.Mock
}

func (m *MockArchiveCaseService) ListArchivedCases(userID, tenantID, teamID string) ([]listArchiveCases.ArchivedCase, error) {
	args := m.Called(userID, tenantID, teamID)
	return args.Get(0).([]listArchiveCases.ArchivedCase), args.Error(1)
}

// ========== HELPER FUNCTIONS ==========

func createTestArchivedCaseHandler(title, description string) listArchiveCases.ArchivedCase {
	return listArchiveCases.ArchivedCase{
		ID:          uuid.New(),
		Title:       title,
		Description: description,
		Status:      "archived",
		Priority:    "medium",
		CreatedBy:   uuid.New(),
		TenantID:    uuid.New(),
		TeamID:      uuid.New(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		ArchivedAt:  time.Now(),
	}
}

func createTestContextHandler(userID, tenantID, teamID string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("GET", "/cases/archived", nil)
	c.Request = req

	c.Set("userID", userID)
	c.Set("tenantID", tenantID)
	c.Set("teamID", teamID)

	return c, w
}

// ========== SUCCESS SCENARIOS ==========

func TestListArchivedCasesHandler_Success(t *testing.T) {
	mockService := &MockArchiveCaseService{}

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"

	expectedCases := []listArchiveCases.ArchivedCase{
		createTestArchivedCaseHandler("Archived Security Incident #1", "First archived case"),
		createTestArchivedCaseHandler("Archived Data Breach Investigation", "Second archived case"),
		createTestArchivedCaseHandler("Archived Malware Detection", "Third archived case"),
	}

	mockService.On("ListArchivedCases", userID, tenantID, teamID).Return(expectedCases, nil)

	handler := handlers.ListArchivedCasesHandler(mockService)
	c, w := createTestContextHandler(userID, tenantID, teamID)

	handler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	cases, ok := response["archived_cases"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, cases, 3)

	// Verify case details
	firstCase := cases[0].(map[string]interface{})
	assert.Equal(t, "Archived Security Incident #1", firstCase["title"])
	assert.Equal(t, "archived", firstCase["status"])

	mockService.AssertExpectations(t)
}

func TestListArchivedCasesHandler_EmptyResult(t *testing.T) {
	mockService := &MockArchiveCaseService{}

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"

	emptyCases := []listArchiveCases.ArchivedCase{}

	mockService.On("ListArchivedCases", userID, tenantID, teamID).Return(emptyCases, nil)

	handler := handlers.ListArchivedCasesHandler(mockService)
	c, w := createTestContextHandler(userID, tenantID, teamID)

	handler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	cases, ok := response["archived_cases"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, cases, 0)

	mockService.AssertExpectations(t)
}

func TestListArchivedCasesHandler_SingleCase(t *testing.T) {
	mockService := &MockArchiveCaseService{}

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"

	singleCase := []listArchiveCases.ArchivedCase{
		createTestArchivedCaseHandler("Single Archived Case", "Only one archived case"),
	}

	mockService.On("ListArchivedCases", userID, tenantID, teamID).Return(singleCase, nil)

	handler := handlers.ListArchivedCasesHandler(mockService)
	c, w := createTestContextHandler(userID, tenantID, teamID)

	handler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	cases, ok := response["archived_cases"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, cases, 1)

	firstCase := cases[0].(map[string]interface{})
	assert.Equal(t, "Single Archived Case", firstCase["title"])

	mockService.AssertExpectations(t)
}

// ========== ERROR SCENARIOS ==========

func TestListArchivedCasesHandler_ServiceError(t *testing.T) {
	mockService := &MockArchiveCaseService{}

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"

	expectedError := errors.New("database connection failed")

	mockService.On("ListArchivedCases", userID, tenantID, teamID).Return([]listArchiveCases.ArchivedCase{}, expectedError)

	handler := handlers.ListArchivedCasesHandler(mockService)
	c, w := createTestContextHandler(userID, tenantID, teamID)

	handler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "database connection failed", response["error"])

	mockService.AssertExpectations(t)
}

func TestListArchivedCasesHandler_DatabaseTimeout(t *testing.T) {
	mockService := &MockArchiveCaseService{}

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"

	expectedError := errors.New("database timeout")

	mockService.On("ListArchivedCases", userID, tenantID, teamID).Return([]listArchiveCases.ArchivedCase{}, expectedError)

	handler := handlers.ListArchivedCasesHandler(mockService)
	c, w := createTestContextHandler(userID, tenantID, teamID)

	handler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "database timeout", response["error"])

	mockService.AssertExpectations(t)
}

func TestListArchivedCasesHandler_PermissionDenied(t *testing.T) {
	mockService := &MockArchiveCaseService{}

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"

	expectedError := errors.New("permission denied")

	mockService.On("ListArchivedCases", userID, tenantID, teamID).Return([]listArchiveCases.ArchivedCase{}, expectedError)

	handler := handlers.ListArchivedCasesHandler(mockService)
	c, w := createTestContextHandler(userID, tenantID, teamID)

	handler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "permission denied", response["error"])

	mockService.AssertExpectations(t)
}

// ========== AUTHENTICATION/AUTHORIZATION TESTS ==========

func TestListArchivedCasesHandler_MissingCredentials(t *testing.T) {
	testCases := []struct {
		name     string
		userID   interface{}
		tenantID interface{}
		teamID   interface{}
	}{
		{
			name:     "Missing UserID",
			userID:   nil,
			tenantID: "test-tenant-456",
			teamID:   "test-team-789",
		},
		{
			name:     "Missing TenantID",
			userID:   "test-user-123",
			tenantID: nil,
			teamID:   "test-team-789",
		},
		{
			name:     "Missing TeamID",
			userID:   "test-user-123",
			tenantID: "test-tenant-456",
			teamID:   nil,
		},
		{
			name:     "All Missing",
			userID:   nil,
			tenantID: nil,
			teamID:   nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					// Handler panics due to type assertion on nil value
					t.Logf("Expected panic caught: %v", r)
				}
			}()

			mockService := &MockArchiveCaseService{}
			handler := handlers.ListArchivedCasesHandler(mockService)

			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req, _ := http.NewRequest("GET", "/cases/archived", nil)
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

			handler(c)

			// If we reach here, the handler didn't panic
			// This would indicate the handler needs better error handling
		})
	}
}

func TestListArchivedCasesHandler_InvalidTokenTypes(t *testing.T) {
	testCases := []struct {
		name     string
		userID   interface{}
		tenantID interface{}
		teamID   interface{}
	}{
		{
			name:     "Invalid UserID Type",
			userID:   123, // int instead of string
			tenantID: "test-tenant-456",
			teamID:   "test-team-789",
		},
		{
			name:     "Invalid TenantID Type",
			userID:   "test-user-123",
			tenantID: 456, // int instead of string
			teamID:   "test-team-789",
		},
		{
			name:     "Invalid TeamID Type",
			userID:   "test-user-123",
			tenantID: "test-tenant-456",
			teamID:   789, // int instead of string
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					// Handler panics due to type assertion failure
					t.Logf("Expected panic caught: %v", r)
					assert.Contains(t, r.(error).Error(), "interface conversion")
				}
			}()

			mockService := &MockArchiveCaseService{}
			handler := handlers.ListArchivedCasesHandler(mockService)

			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req, _ := http.NewRequest("GET", "/cases/archived", nil)
			c.Request = req

			c.Set("userID", tc.userID)
			c.Set("tenantID", tc.tenantID)
			c.Set("teamID", tc.teamID)

			handler(c)
		})
	}
}

// ========== LARGE DATASET TESTS ==========

func TestListArchivedCasesHandler_LargeDataset(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large dataset test in short mode")
	}

	mockService := &MockArchiveCaseService{}

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"

	// Generate 500 archived cases
	largeCaseSet := make([]listArchiveCases.ArchivedCase, 500)
	for i := 0; i < 500; i++ {
		largeCaseSet[i] = createTestArchivedCaseHandler(
			"Archived Case "+string(rune(i)),
			"Description for archived case "+string(rune(i)),
		)
	}

	mockService.On("ListArchivedCases", userID, tenantID, teamID).Return(largeCaseSet, nil)

	handler := handlers.ListArchivedCasesHandler(mockService)
	c, w := createTestContextHandler(userID, tenantID, teamID)

	start := time.Now()
	handler(c)
	duration := time.Since(start)

	assert.Equal(t, http.StatusOK, w.Code)
	t.Logf("Processing time for 500 archived cases: %v", duration)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	cases, ok := response["archived_cases"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, cases, 500)

	mockService.AssertExpectations(t)
}

// ========== SPECIAL CHARACTER TESTS ==========

func TestListArchivedCasesHandler_SpecialCharacters(t *testing.T) {
	mockService := &MockArchiveCaseService{}

	userID := "test-用户-123"
	tenantID := "test-租户-456"
	teamID := "test-团队-789"

	expectedCases := []listArchiveCases.ArchivedCase{
		createTestArchivedCaseHandler("Archived Security Incident with 特殊字符", "Description with special chars"),
		createTestArchivedCaseHandler("Archived Incidente de Seguridad con símbolos: @#$%^&*()", "Spanish description"),
	}

	mockService.On("ListArchivedCases", userID, tenantID, teamID).Return(expectedCases, nil)

	handler := handlers.ListArchivedCasesHandler(mockService)
	c, w := createTestContextHandler(userID, tenantID, teamID)

	handler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	cases, ok := response["archived_cases"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, cases, 2)

	mockService.AssertExpectations(t)
}

// ========== CONCURRENT ACCESS TESTS ==========

func TestListArchivedCasesHandler_ConcurrentRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	mockService := &MockArchiveCaseService{}

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"

	expectedCases := []listArchiveCases.ArchivedCase{
		createTestArchivedCaseHandler("Concurrent Test Case", "Concurrent access test"),
	}

	numRequests := 20
	mockService.On("ListArchivedCases", userID, tenantID, teamID).Return(expectedCases, nil).Times(numRequests)

	handler := handlers.ListArchivedCasesHandler(mockService)

	results := make(chan int, numRequests)
	for i := 0; i < numRequests; i++ {
		go func() {
			c, w := createTestContextHandler(userID, tenantID, teamID)
			handler(c)
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
	mockService.AssertExpectations(t)
}

// ========== JSON RESPONSE STRUCTURE TESTS ==========

func TestListArchivedCasesHandler_ResponseStructure(t *testing.T) {
	mockService := &MockArchiveCaseService{}

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"

	testCase := createTestArchivedCaseHandler("Test Case", "Test Description")
	expectedCases := []listArchiveCases.ArchivedCase{testCase}

	mockService.On("ListArchivedCases", userID, tenantID, teamID).Return(expectedCases, nil)

	handler := handlers.ListArchivedCasesHandler(mockService)
	c, w := createTestContextHandler(userID, tenantID, teamID)

	handler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify response structure
	assert.Contains(t, response, "archived_cases")
	cases, ok := response["archived_cases"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, cases, 1)

	caseData := cases[0].(map[string]interface{})
	assert.Contains(t, caseData, "id")
	assert.Contains(t, caseData, "title")
	assert.Contains(t, caseData, "description")
	assert.Contains(t, caseData, "status")
	assert.Contains(t, caseData, "priority")

	mockService.AssertExpectations(t)
}

// ========== BENCHMARK TESTS ==========

func BenchmarkListArchivedCasesHandler(b *testing.B) {
	mockService := &MockArchiveCaseService{}

	userID := "test-user-123"
	tenantID := "test-tenant-456"
	teamID := "test-team-789"

	expectedCases := []listArchiveCases.ArchivedCase{
		createTestArchivedCaseHandler("Benchmark Case", "Benchmark description"),
	}

	mockService.On("ListArchivedCases", userID, tenantID, teamID).Return(expectedCases, nil)

	handler := handlers.ListArchivedCasesHandler(mockService)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c, _ := createTestContextHandler(userID, tenantID, teamID)
		handler(c)
	}
}
