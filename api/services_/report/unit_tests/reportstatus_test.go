package unit_tests

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"aegis-api/handlers"
	

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"context"
	"strings"
	"github.com/google/uuid"
	"aegis-api/services_/report/update_status"
	
)



type MockReportStatusService struct {
	UpdateStatusFunc func(ctx context.Context, reportID uuid.UUID, status update_status.ReportStatus) (*update_status.Report, error)
}

func (m *MockReportStatusService) UpdateStatus(ctx context.Context, reportID uuid.UUID, status update_status.ReportStatus) (*update_status.Report, error) {
	if m.UpdateStatusFunc != nil {
		return m.UpdateStatusFunc(ctx, reportID, status)
	}
	return nil, nil
}





func TestUpdateStatus_Success(t *testing.T) {
    gin.SetMode(gin.TestMode)

    reportID := uuid.New()
    mockService := &MockReportStatusService{
        UpdateStatusFunc: func(ctx context.Context, id uuid.UUID, status update_status.ReportStatus) (*update_status.Report, error) {
            return &update_status.Report{
                ID:     id,
                Status: status,
            }, nil
        },
    }

    handler := handlers.NewReportStatusHandler(mockService)

    // Setup router for testing
    r := gin.Default()
    r.PUT("/reports/:id/status", handler.UpdateStatus)

    w := httptest.NewRecorder()
    reqBody := strings.NewReader(`{"status":"review"}`)
    req, _ := http.NewRequest(http.MethodPut, "/reports/"+reportID.String()+"/status", reqBody)
    req.Header.Set("Content-Type", "application/json")

    r.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
    assert.Contains(t, w.Body.String(), `"status":"review"`)
}


func TestUpdateStatus_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewReportStatusHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := bytes.NewBufferString(`{"wrong":"field"}`)
	req, _ := http.NewRequest(http.MethodPut, "/reports/123/status", body)
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.UpdateStatus(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
