package unit_tests

import (
	"aegis-api/handlers" // Import the real handler package
	"aegis-api/services_/admin/get_collaborators"
	"aegis-api/services_/auditlog"
	"aegis-api/structs"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockGetCollaboratorsService implements handlers.CollaboratorService
type MockGetCollaboratorsService struct {
	mock.Mock
}

func (m *MockGetCollaboratorsService) GetCollaborators(caseID uuid.UUID) ([]get_collaborators.Collaborator, error) {
	args := m.Called(caseID)
	return args.Get(0).([]get_collaborators.Collaborator), args.Error(1)
}

// MockAuditLogger implements handlers.AuditService
type MockAuditLogger struct {
	mock.Mock
}

func (m *MockAuditLogger) Log(c *gin.Context, log auditlog.AuditLog) error {
	args := m.Called(c, log)
	return args.Error(0) // Return the error from mock expectations
}

// Helper function to create a test Gin context
func createTestContextCollaborators(method, url string, body interface{}) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	var bodyBytes []byte
	if body != nil {
		bodyBytes, _ = json.Marshal(body)
	}

	req, _ := http.NewRequest(method, url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	return c, w
}

// Helper function to create collaborator data
func createTestCollaborator(fullName, email, role string) get_collaborators.Collaborator {
	return get_collaborators.Collaborator{
		ID:         uuid.New(),
		FullName:   fullName,
		Email:      email,
		Role:       role,
		AssignedAt: time.Now(),
	}
}

// ========== SUCCESS SCENARIOS ==========

func TestGetCollaboratorsByCaseID_Success(t *testing.T) {
	mockService := &MockGetCollaboratorsService{}
	mockAuditLogger := &MockAuditLogger{}

	handler := handlers.NewGetCollaboratorsHandler(mockService, mockAuditLogger)

	caseID := uuid.New()
	expectedCollaborators := []get_collaborators.Collaborator{
		createTestCollaborator("John Doe", "john@example.com", "analyst"),
		createTestCollaborator("Jane Smith", "jane@example.com", "incident_responder"),
	}

	mockService.On("GetCollaborators", caseID).Return(expectedCollaborators, nil)
	mockAuditLogger.On("Log", mock.Anything, mock.MatchedBy(func(log auditlog.AuditLog) bool {
		return log.Action == "GET_COLLABORATORS_FOR_CASE" &&
			log.Status == "SUCCESS" &&
			log.Target.ID == caseID.String() &&
			log.Target.Type == "case" &&
			log.Service == "cases"
	})).Return(nil) // ✅ Add this!

	c, w := createTestContextCollaborators("GET", "/cases/"+caseID.String()+"/collaborators", nil)
	c.Params = []gin.Param{{Key: "case_id", Value: caseID.String()}}

	handler.GetCollaboratorsByCaseID(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response structs.SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "Collaborators retrieved successfully", response.Message)
	assert.NotNil(t, response.Data)

	mockService.AssertExpectations(t)
	mockAuditLogger.AssertExpectations(t)
}

func TestGetCollaboratorsByCaseID_EmptyResult(t *testing.T) {
	mockService := &MockGetCollaboratorsService{}
	mockAuditLogger := &MockAuditLogger{}
	handler := handlers.NewGetCollaboratorsHandler(mockService, mockAuditLogger)

	caseID := uuid.New()
	emptyCollaborators := []get_collaborators.Collaborator{}

	mockService.On("GetCollaborators", caseID).Return(emptyCollaborators, nil)
	mockAuditLogger.On("Log", mock.Anything, mock.MatchedBy(func(log auditlog.AuditLog) bool {
		return log.Status == "SUCCESS"
	})).Return(nil) // ✅ Add this!

	c, w := createTestContextCollaborators("GET", "/cases/"+caseID.String()+"/collaborators", nil)
	c.Params = []gin.Param{{Key: "case_id", Value: caseID.String()}}

	handler.GetCollaboratorsByCaseID(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response structs.SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "Collaborators retrieved successfully", response.Message)

	mockService.AssertExpectations(t)
	mockAuditLogger.AssertExpectations(t)
}

func TestGetCollaboratorsByCaseID_SingleCollaborator(t *testing.T) {
	mockService := &MockGetCollaboratorsService{}
	mockAuditLogger := &MockAuditLogger{}
	handler := handlers.NewGetCollaboratorsHandler(mockService, mockAuditLogger)

	caseID := uuid.New()
	expectedCollaborators := []get_collaborators.Collaborator{
		createTestCollaborator("Solo Analyst", "solo@example.com", "forensics_analyst"),
	}

	mockService.On("GetCollaborators", caseID).Return(expectedCollaborators, nil)
	mockAuditLogger.On("Log", mock.Anything, mock.MatchedBy(func(log auditlog.AuditLog) bool {
		return log.Status == "SUCCESS"
	})).Return(nil) // ✅ Add this!

	c, w := createTestContextCollaborators("GET", "/cases/"+caseID.String()+"/collaborators", nil)
	c.Params = []gin.Param{{Key: "case_id", Value: caseID.String()}}

	handler.GetCollaboratorsByCaseID(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response structs.SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	mockService.AssertExpectations(t)
	mockAuditLogger.AssertExpectations(t)
}

func TestGetCollaboratorsByCaseID_LargeDataset(t *testing.T) {
	mockService := &MockGetCollaboratorsService{}
	mockAuditLogger := &MockAuditLogger{}
	handler := handlers.NewGetCollaboratorsHandler(mockService, mockAuditLogger)

	caseID := uuid.New()

	// Generate 100 collaborators
	expectedCollaborators := make([]get_collaborators.Collaborator, 100)
	for i := 0; i < 100; i++ {
		expectedCollaborators[i] = createTestCollaborator(
			fmt.Sprintf("User %d", i+1),
			fmt.Sprintf("user%d@example.com", i+1),
			"analyst",
		)
	}

	mockService.On("GetCollaborators", caseID).Return(expectedCollaborators, nil)
	mockAuditLogger.On("Log", mock.Anything, mock.MatchedBy(func(log auditlog.AuditLog) bool {
		return log.Status == "SUCCESS"
	})).Return(nil) // ✅ Add this!

	c, w := createTestContextCollaborators("GET", "/cases/"+caseID.String()+"/collaborators", nil)
	c.Params = []gin.Param{{Key: "case_id", Value: caseID.String()}}

	handler.GetCollaboratorsByCaseID(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response structs.SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	mockService.AssertExpectations(t)
	mockAuditLogger.AssertExpectations(t)
}

// ========== VALIDATION ERROR TESTS ==========

func TestGetCollaboratorsByCaseID_MissingCaseID(t *testing.T) {
	mockService := &MockGetCollaboratorsService{}
	mockAuditLogger := &MockAuditLogger{}
	handler := handlers.NewGetCollaboratorsHandler(mockService, mockAuditLogger)

	mockAuditLogger.On("Log", mock.Anything, mock.MatchedBy(func(log auditlog.AuditLog) bool {
		return log.Action == "GET_COLLABORATORS_FOR_CASE" &&
			log.Status == "FAILED" &&
			log.Description == "Missing case_id parameter" &&
			log.Target.ID == "" &&
			log.Target.Type == "case"
	})).Return(nil) // ✅ Add this!

	c, w := createTestContextCollaborators("GET", "/cases//collaborators", nil)
	c.Params = []gin.Param{{Key: "case_id", Value: ""}}

	handler.GetCollaboratorsByCaseID(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response structs.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invalid_request", response.Error)
	assert.Equal(t, "case_id parameter is required", response.Message)

	mockAuditLogger.AssertExpectations(t)
	mockService.AssertNotCalled(t, "GetCollaborators")
}

func TestGetCollaboratorsByCaseID_InvalidUUIDFormats(t *testing.T) {
	testCases := []struct {
		name        string
		caseID      string
		description string
	}{
		{
			name:        "Invalid UUID format",
			caseID:      "invalid-uuid",
			description: "Should fail with invalid UUID",
		},
		{
			name:        "Invalid UUID without hyphens (too short)",
			caseID:      "123e4567e89b12d3a456426614174", // Too short - 31 chars instead of 32
			description: "Should fail - incomplete UUID without hyphens",
		},
		{
			name:        "Too short",
			caseID:      "123",
			description: "Should fail with short string",
		},
		{
			name:        "Too long",
			caseID:      "123e4567-e89b-12d3-a456-426614174000-extra",
			description: "Should fail with extra characters",
		},
		{
			name:        "Special characters",
			caseID:      "123e4567-e89b-12d3-a456-42661417400@",
			description: "Should fail with special characters",
		},
		{
			name:        "Invalid hex characters",
			caseID:      "123g4567-e89b-12d3-a456-426614174000", // 'g' is not hex
			description: "Should fail with non-hex characters",
		},
		{
			name:        "Wrong hyphen positions",
			caseID:      "123e456-7e89b-12d3-a456-426614174000",
			description: "Should fail with incorrect hyphen placement",
		},
		{
			name:        "Missing hyphens in middle",
			caseID:      "123e4567e89b-12d3-a456-426614174000",
			description: "Should fail with missing hyphens",
		},
		{
			name:        "Only hyphens",
			caseID:      "--------",
			description: "Should fail with only hyphens",
		},
		{
			name:        "Empty with spaces",
			caseID:      "   ",
			description: "Should fail with only spaces",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := &MockGetCollaboratorsService{}
			mockAuditLogger := &MockAuditLogger{}
			handler := handlers.NewGetCollaboratorsHandler(mockService, mockAuditLogger)

			mockAuditLogger.On("Log", mock.Anything, mock.MatchedBy(func(log auditlog.AuditLog) bool {
				return log.Action == "GET_COLLABORATORS_FOR_CASE" &&
					log.Status == "FAILED" &&
					log.Description == "Invalid case_id format" &&
					log.Target.ID == tc.caseID &&
					log.Target.Type == "case"
			})).Return(nil) // ✅ Add this!

			c, w := createTestContextCollaborators("GET", "/cases/"+tc.caseID+"/collaborators", nil)
			c.Params = []gin.Param{{Key: "case_id", Value: tc.caseID}}

			handler.GetCollaboratorsByCaseID(c)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response structs.ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "invalid_request", response.Error)
			assert.Equal(t, "invalid case_id format", response.Message)

			mockAuditLogger.AssertExpectations(t)
			mockService.AssertNotCalled(t, "GetCollaborators")
		})
	}
}

// ========== ENHANCED VALID UUID TESTS ==========

func TestGetCollaboratorsByCaseID_ValidUUIDFormats(t *testing.T) {
	testCases := []struct {
		name   string
		caseID string
	}{
		{
			name:   "Standard UUID",
			caseID: "123e4567-e89b-12d3-a456-426614174000",
		},
		{
			name:   "Nil UUID",
			caseID: "00000000-0000-0000-0000-000000000000",
		},
		{
			name:   "UUID with uppercase (VALID - Go accepts this)",
			caseID: "123E4567-E89B-12D3-A456-426614174000",
		},
		{
			name:   "UUID without hyphens (VALID - Go accepts this)",
			caseID: "123e4567e89b12d3a456426614174000", // 32 chars exactly
		},
		{
			name:   "Mixed case UUID without hyphens",
			caseID: "123E4567e89B12d3A456426614174000",
		},
		{
			name:   "Random valid UUID",
			caseID: uuid.New().String(),
		},
		{
			name:   "Max values UUID",
			caseID: "ffffffff-ffff-ffff-ffff-ffffffffffff",
		},
		{
			name:   "Version 1 UUID",
			caseID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		},
		{
			name:   "Version 4 UUID",
			caseID: "550e8400-e29b-41d4-a716-446655440000",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := &MockGetCollaboratorsService{}
			mockAuditLogger := &MockAuditLogger{}
			handler := handlers.NewGetCollaboratorsHandler(mockService, mockAuditLogger)

			parsedUUID, err := uuid.Parse(tc.caseID)
			assert.NoError(t, err, "Test case should have valid UUID")

			emptyCollaborators := []get_collaborators.Collaborator{}

			mockService.On("GetCollaborators", parsedUUID).Return(emptyCollaborators, nil)
			mockAuditLogger.On("Log", mock.Anything, mock.MatchedBy(func(log auditlog.AuditLog) bool {
				return log.Status == "SUCCESS" && log.Target.ID == parsedUUID.String()
			})).Return(nil) // ✅ Add this!

			c, w := createTestContextCollaborators("GET", "/cases/"+tc.caseID+"/collaborators", nil)
			c.Params = []gin.Param{{Key: "case_id", Value: tc.caseID}}

			handler.GetCollaboratorsByCaseID(c)

			assert.Equal(t, http.StatusOK, w.Code)

			mockService.AssertExpectations(t)
			mockAuditLogger.AssertExpectations(t)
		})
	}
}

// ========== EDGE CASE UUID BOUNDARY TESTS ==========

func TestGetCollaboratorsByCaseID_UUIDBoundaryTests(t *testing.T) {
	testCases := []struct {
		name        string
		caseID      string
		shouldFail  bool
		description string
	}{
		{
			name:        "31 characters without hyphens",
			caseID:      "123e4567e89b12d3a45642661417400", // 31 chars
			shouldFail:  true,
			description: "One character short",
		},
		{
			name:        "33 characters without hyphens",
			caseID:      "123e4567e89b12d3a456426614174000a", // 33 chars
			shouldFail:  true,
			description: "One character too long",
		},
		{
			name:        "32 characters but invalid hex",
			caseID:      "123g4567e89b12d3a456426614174000", // 'g' invalid
			shouldFail:  true,
			description: "Correct length but invalid hex",
		},
		{
			name:        "Valid 32 character hex",
			caseID:      "123e4567e89b12d3a456426614174000", // Valid
			shouldFail:  false,
			description: "Exactly 32 valid hex characters",
		},
		{
			name:        "UUID with wrong version bits",
			caseID:      "123e4567-e89b-82d3-a456-426614174000", // Invalid version
			shouldFail:  false,                                  // Go still parses this
			description: "Invalid version but Go accepts it",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := &MockGetCollaboratorsService{}
			mockAuditLogger := &MockAuditLogger{}
			handler := handlers.NewGetCollaboratorsHandler(mockService, mockAuditLogger)

			if tc.shouldFail {
				mockAuditLogger.On("Log", mock.Anything, mock.MatchedBy(func(log auditlog.AuditLog) bool {
					return log.Status == "FAILED" && log.Description == "Invalid case_id format"
				})).Return(nil) // ✅ Add this!
			} else {
				parsedUUID, _ := uuid.Parse(tc.caseID)
				mockService.On("GetCollaborators", parsedUUID).Return([]get_collaborators.Collaborator{}, nil)
				mockAuditLogger.On("Log", mock.Anything, mock.MatchedBy(func(log auditlog.AuditLog) bool {
					return log.Status == "SUCCESS"
				})).Return(nil) // ✅ Add this!
			}

			c, w := createTestContextCollaborators("GET", "/cases/"+tc.caseID+"/collaborators", nil)
			c.Params = []gin.Param{{Key: "case_id", Value: tc.caseID}}

			handler.GetCollaboratorsByCaseID(c)

			if tc.shouldFail {
				assert.Equal(t, http.StatusBadRequest, w.Code)
				mockService.AssertNotCalled(t, "GetCollaborators")
			} else {
				assert.Equal(t, http.StatusOK, w.Code)
				mockService.AssertExpectations(t)
			}

			mockAuditLogger.AssertExpectations(t)
		})
	}
}

// ========== PARAMETER HANDLING EDGE CASES ==========

func TestGetCollaboratorsByCaseID_ParameterEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		setupParams    func() []gin.Param
		expectedStatus int
		expectedError  string
		description    string
	}{
		{
			name: "Duplicate case_id parameters",
			setupParams: func() []gin.Param {
				return []gin.Param{
					{Key: "case_id", Value: "123e4567-e89b-12d3-a456-426614174000"},
					{Key: "case_id", Value: "different-value"},
				}
			},
			expectedStatus: http.StatusOK, // Gin uses first match
			expectedError:  "",
			description:    "Should use first case_id parameter",
		},
		{
			name: "Case sensitive parameter names",
			setupParams: func() []gin.Param {
				return []gin.Param{
					{Key: "Case_ID", Value: "123e4567-e89b-12d3-a456-426614174000"},
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid_request",
			description:    "Parameter names should be case sensitive",
		},
		{
			name: "Parameter with spaces in key",
			setupParams: func() []gin.Param {
				return []gin.Param{
					{Key: "case id", Value: "123e4567-e89b-12d3-a456-426614174000"},
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid_request",
			description:    "Spaces in parameter key should fail",
		},
		{
			name: "Many parameters with case_id last",
			setupParams: func() []gin.Param {
				params := make([]gin.Param, 100)
				for i := 0; i < 99; i++ {
					params[i] = gin.Param{Key: fmt.Sprintf("param_%d", i), Value: "value"}
				}
				params[99] = gin.Param{Key: "case_id", Value: "123e4567-e89b-12d3-a456-426614174000"}
				return params
			},
			expectedStatus: http.StatusOK,
			expectedError:  "",
			description:    "Should find case_id even with many parameters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockGetCollaboratorsService{}
			mockAuditLogger := &MockAuditLogger{}
			handler := handlers.NewGetCollaboratorsHandler(mockService, mockAuditLogger)

			params := tt.setupParams()

			if tt.expectedStatus == http.StatusOK {
				caseID, _ := uuid.Parse("123e4567-e89b-12d3-a456-426614174000")
				mockService.On("GetCollaborators", caseID).Return([]get_collaborators.Collaborator{}, nil)
			}
			mockAuditLogger.On("Log", mock.Anything, mock.Anything).Return(nil) // ✅ Return nil error

			c, w := createTestContextCollaborators("GET", "/cases/test/collaborators", nil)
			c.Params = params

			handler.GetCollaboratorsByCaseID(c)

			assert.Equal(t, tt.expectedStatus, w.Code, tt.description)

			if tt.expectedError != "" {
				var response structs.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response.Error)
			}

			mockAuditLogger.AssertExpectations(t)
			if tt.expectedStatus == http.StatusOK {
				mockService.AssertExpectations(t)
			}
		})
	}
}

// ========== CONCURRENT UUID VALIDATION TESTS ==========

func TestGetCollaboratorsByCaseID_ConcurrentUUIDValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	validUUIDs := []string{
		"123e4567-e89b-12d3-a456-426614174000",
		"550e8400-e29b-41d4-a716-446655440000",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		"00000000-0000-0000-0000-000000000000",
		"ffffffff-ffff-ffff-ffff-ffffffffffff",
	}

	mockService := &MockGetCollaboratorsService{}
	mockAuditLogger := &MockAuditLogger{}
	handler := handlers.NewGetCollaboratorsHandler(mockService, mockAuditLogger)

	// Setup mocks for all UUIDs
	for _, uuidStr := range validUUIDs {
		parsedUUID, _ := uuid.Parse(uuidStr)
		mockService.On("GetCollaborators", parsedUUID).Return([]get_collaborators.Collaborator{}, nil).Times(5)
	}
	mockAuditLogger.On("Log", mock.Anything, mock.Anything).Return(nil).Times(25) // ✅ Add .Return(nil)

	var wg sync.WaitGroup
	results := make(chan int, 25)

	// Test each UUID concurrently with multiple goroutines
	for _, uuidStr := range validUUIDs {
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(uuid string, id int) {
				defer wg.Done()

				c, w := createTestContextCollaborators("GET", "/cases/"+uuid+"/collaborators", nil)
				c.Params = []gin.Param{{Key: "case_id", Value: uuid}}

				handler.GetCollaboratorsByCaseID(c)
				results <- w.Code
			}(uuidStr, i)
		}
	}

	wg.Wait()
	close(results)

	// Verify all requests succeeded
	successCount := 0
	for code := range results {
		if code == http.StatusOK {
			successCount++
		}
	}

	assert.Equal(t, 25, successCount)
	mockService.AssertExpectations(t)
	mockAuditLogger.AssertExpectations(t)
}

// ========== LOCALE AND INTERNATIONALIZATION TESTS ==========

func TestGetCollaboratorsByCaseID_InternationalizationHandling(t *testing.T) {
	testCases := []struct {
		name        string
		caseID      string
		description string
	}{
		{
			name:        "Arabic numerals",
			caseID:      "١٢٣e٤٥٦٧-e٨٩b-١٢d٣-a٤٥٦-٤٢٦٦١٤١٧٤٠٠",
			description: "Arabic-Indic digits should be rejected",
		},
		{
			name:        "Chinese characters",
			caseID:      "一二三e四五六七-e八九b-一二d三-a四五六-四二六六一四一七四零零零",
			description: "Chinese numerals should be rejected",
		},
		{
			name:        "Cyrillic characters",
			caseID:      "123е4567-е89б-12д3-а456-426614174000",
			description: "Cyrillic letters that look like Latin should be rejected",
		},
		{
			name:        "Greek letters",
			caseID:      "123ε4567-ε89β-12δ3-α456-426614174000",
			description: "Greek letters should be rejected",
		},
		{
			name:        "Mathematical symbols",
			caseID:      "123∈4567-∈89∆-12∂3-∀456-426614174000",
			description: "Mathematical symbols should be rejected",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := &MockGetCollaboratorsService{}
			mockAuditLogger := &MockAuditLogger{}
			handler := handlers.NewGetCollaboratorsHandler(mockService, mockAuditLogger)

			mockAuditLogger.On("Log", mock.Anything, mock.MatchedBy(func(log auditlog.AuditLog) bool {
				return log.Status == "FAILED" && log.Description == "Invalid case_id format"
			})).Return(nil) // ✅ Add this!

			c, w := createTestContextCollaborators("GET", "/cases/"+tc.caseID+"/collaborators", nil)
			c.Params = []gin.Param{{Key: "case_id", Value: tc.caseID}}

			handler.GetCollaboratorsByCaseID(c)

			assert.Equal(t, http.StatusBadRequest, w.Code, tc.description)

			var response structs.ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "invalid_request", response.Error)
			assert.Equal(t, "invalid case_id format", response.Message)

			mockAuditLogger.AssertExpectations(t)
			mockService.AssertNotCalled(t, "GetCollaborators")
		})
	}
}

// Add this test to your existing file
func TestGetCollaboratorsByCaseID_ServiceError(t *testing.T) {
	mockService := &MockGetCollaboratorsService{}
	mockAuditLogger := &MockAuditLogger{}
	handler := handlers.NewGetCollaboratorsHandler(mockService, mockAuditLogger)

	caseID := uuid.New()
	serviceError := fmt.Errorf("database connection failed")

	mockService.On("GetCollaborators", caseID).Return([]get_collaborators.Collaborator{}, serviceError)
	mockAuditLogger.On("Log", mock.Anything, mock.MatchedBy(func(log auditlog.AuditLog) bool {
		return log.Action == "GET_COLLABORATORS_FOR_CASE" &&
			log.Status == "FAILED" &&
			log.Description == "Could not retrieve collaborators: database connection failed" &&
			log.Target.ID == caseID.String() &&
			log.Target.Type == "case"
	})).Return(nil)

	c, w := createTestContextCollaborators("GET", "/cases/"+caseID.String()+"/collaborators", nil)
	c.Params = []gin.Param{{Key: "case_id", Value: caseID.String()}}

	handler.GetCollaboratorsByCaseID(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response structs.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "internal_error", response.Error)
	assert.Equal(t, "could not retrieve collaborators: database connection failed", response.Message)

	mockService.AssertExpectations(t)
	mockAuditLogger.AssertExpectations(t)
}
func TestGetCollaboratorsByCaseID_AuditLoggerError(t *testing.T) {
	mockService := &MockGetCollaboratorsService{}
	mockAuditLogger := &MockAuditLogger{}
	handler := handlers.NewGetCollaboratorsHandler(mockService, mockAuditLogger)

	caseID := uuid.New()
	expectedCollaborators := []get_collaborators.Collaborator{
		createTestCollaborator("John Doe", "john@example.com", "analyst"),
	}

	mockService.On("GetCollaborators", caseID).Return(expectedCollaborators, nil)
	// Simulate audit logger error
	mockAuditLogger.On("Log", mock.Anything, mock.Anything).Return(fmt.Errorf("audit log failed"))

	c, w := createTestContextCollaborators("GET", "/cases/"+caseID.String()+"/collaborators", nil)
	c.Params = []gin.Param{{Key: "case_id", Value: caseID.String()}}

	handler.GetCollaboratorsByCaseID(c)

	// Should still return success since handler doesn't check audit log errors
	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
	mockAuditLogger.AssertExpectations(t)
}
func TestGetCollaboratorsByCaseID_ServiceErrors(t *testing.T) {
	testCases := []struct {
		name          string
		serviceError  error
		expectedError string
	}{
		{
			name:          "Network timeout",
			serviceError:  fmt.Errorf("network timeout"),
			expectedError: "Could not retrieve collaborators: network timeout", // ✅ Fix capitalization
		},
		{
			name:          "Permission denied",
			serviceError:  fmt.Errorf("permission denied"),
			expectedError: "Could not retrieve collaborators: permission denied", // ✅ Fix capitalization
		},
		{
			name:          "Case not found",
			serviceError:  fmt.Errorf("case not found"),
			expectedError: "Could not retrieve collaborators: case not found", // ✅ Fix capitalization
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := &MockGetCollaboratorsService{}
			mockAuditLogger := &MockAuditLogger{}
			handler := handlers.NewGetCollaboratorsHandler(mockService, mockAuditLogger)

			caseID := uuid.New()

			mockService.On("GetCollaborators", caseID).Return([]get_collaborators.Collaborator{}, tc.serviceError)
			mockAuditLogger.On("Log", mock.Anything, mock.MatchedBy(func(log auditlog.AuditLog) bool {
				return log.Action == "GET_COLLABORATORS_FOR_CASE" &&
					log.Status == "FAILED" &&
					log.Description == tc.expectedError && // Now this will match
					log.Target.ID == caseID.String() &&
					log.Target.Type == "case" &&
					log.Service == "cases"
			})).Return(nil)

			c, w := createTestContextCollaborators("GET", "/cases/"+caseID.String()+"/collaborators", nil)
			c.Params = []gin.Param{{Key: "case_id", Value: caseID.String()}}

			handler.GetCollaboratorsByCaseID(c)

			assert.Equal(t, http.StatusInternalServerError, w.Code)

			var response structs.ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "internal_error", response.Error)
			// Fix the expected message case as well
			expectedMessage := strings.ToLower(string(tc.expectedError[0])) + tc.expectedError[1:]
			assert.Equal(t, expectedMessage, response.Message)

			mockService.AssertExpectations(t)
			mockAuditLogger.AssertExpectations(t)
		})
	}
}
func TestGetCollaboratorsByCaseID_DataVariations(t *testing.T) {
	testCases := []struct {
		name          string
		collaborators []get_collaborators.Collaborator
		description   string
	}{
		{
			name:          "Collaborators with nil values",
			collaborators: []get_collaborators.Collaborator{{ID: uuid.New(), FullName: "", Email: "", Role: ""}},
			description:   "Should handle empty string fields",
		},
		{
			name: "Collaborators with special characters",
			collaborators: []get_collaborators.Collaborator{
				createTestCollaborator("José María", "josé@example.com", "analista"),
			},
			description: "Should handle international characters",
		},
		{
			name: "Very long collaborator data",
			collaborators: []get_collaborators.Collaborator{
				createTestCollaborator(
					strings.Repeat("A", 1000),
					strings.Repeat("a", 100)+"@example.com",
					"super_long_role_name_that_exceeds_normal_limits",
				),
			},
			description: "Should handle very long data",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := &MockGetCollaboratorsService{}
			mockAuditLogger := &MockAuditLogger{}
			handler := handlers.NewGetCollaboratorsHandler(mockService, mockAuditLogger)

			caseID := uuid.New()

			mockService.On("GetCollaborators", caseID).Return(tc.collaborators, nil)
			mockAuditLogger.On("Log", mock.Anything, mock.MatchedBy(func(log auditlog.AuditLog) bool {
				return log.Status == "SUCCESS"
			})).Return(nil)

			c, w := createTestContextCollaborators("GET", "/cases/"+caseID.String()+"/collaborators", nil)
			c.Params = []gin.Param{{Key: "case_id", Value: caseID.String()}}

			handler.GetCollaboratorsByCaseID(c)

			assert.Equal(t, http.StatusOK, w.Code)

			var response structs.SuccessResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.True(t, response.Success)

			mockService.AssertExpectations(t)
			mockAuditLogger.AssertExpectations(t)
		})
	}
}
