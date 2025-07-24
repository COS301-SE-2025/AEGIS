package unit_tests

import (
	"aegis-api/handlers"
	"aegis-api/services_/auditlog"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuditLogService struct {
	mock.Mock
}

func (m *MockAuditLogService) GetRecentUserActivities(ctx context.Context, userID string) ([]auditlog.AuditLog, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]auditlog.AuditLog), args.Error(1)
}

func TestGetRecentActivities_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockAuditLogService)
	handler := handlers.NewRecentActivityHandler(mockService)

	expectedLogs := []auditlog.AuditLog{
		{
			ID:          "log-1",
			Timestamp:   time.Now().UTC(),
			Action:      "UPLOAD_EVIDENCE",
			Actor:       auditlog.Actor{ID: "user-123", Role: "Investigator"},
			Service:     "evidence",
			Status:      "SUCCESS",
			Description: "Uploaded file",
		},
	}

	mockService.On("GetRecentUserActivities", mock.Anything, "user-123").Return(expectedLogs, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/auditlogs/recent/user-123", nil) // ✅ FIX
	c.Params = []gin.Param{{Key: "userId", Value: "user-123"}}

	handler.GetRecentActivities(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var jsonResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &jsonResp)
	assert.NoError(t, err)

	assert.True(t, jsonResp["success"].(bool))
	assert.Equal(t, "Recent activities retrieved", jsonResp["message"])
	assert.NotEmpty(t, jsonResp["data"])
}

func TestGetRecentActivities_MissingUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockAuditLogService)
	handler := handlers.NewRecentActivityHandler(mockService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// Intentionally leave c.Params empty

	handler.GetRecentActivities(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "user ID is required")
}

func TestGetRecentActivities_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockAuditLogService)
	handler := handlers.NewRecentActivityHandler(mockService)

	mockService.On("GetRecentUserActivities", mock.Anything, "user-123").
		Return([]auditlog.AuditLog{}, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/auditlogs/recent/user-123", nil) // ✅ FIX
	c.Params = []gin.Param{{Key: "userId", Value: "user-123"}}

	handler.GetRecentActivities(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to retrieve logs")
}
