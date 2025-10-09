package unit_tests

import (
	"aegis-api/handlers"
	"aegis-api/services_/auditlog"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockDownloadService is a mock implementation of the download service
type mockDownloadService struct {
	filename     string
	stream       io.ReadCloser
	filetype     string
	err          error
	calledWithID uuid.UUID
	callCount    int
}

func (m *mockDownloadService) DownloadEvidence(evidenceID uuid.UUID) (string, io.ReadCloser, string, error) {
	m.calledWithID = evidenceID
	m.callCount++
	if m.err != nil {
		return "", nil, "", m.err
	}
	return m.filename, m.stream, m.filetype, nil
}

// mockAuditLogger is a mock implementation of the audit logger
type mockAuditLogger struct {
	logs      []auditlog.AuditLog
	callCount int
	err       error
}

func (m *mockAuditLogger) Log(c *gin.Context, log auditlog.AuditLog) error {
	m.logs = append(m.logs, log)
	m.callCount++
	return m.err
}

func (m *mockAuditLogger) getLastLog() auditlog.AuditLog {
	if len(m.logs) == 0 {
		return auditlog.AuditLog{}
	}
	return m.logs[len(m.logs)-1]
}

// mockReadCloser wraps a reader with a Close method
type mockReadCloser struct {
	io.Reader
	closed    bool
	closeErr  error
	readCalls int
}

func (m *mockReadCloser) Read(p []byte) (n int, err error) {
	m.readCalls++
	return m.Reader.Read(p)
}

func (m *mockReadCloser) Close() error {
	m.closed = true
	return m.closeErr
}

// mockGinContext creates a minimal mock Gin context for testing
func mockGinContext(method, path string, params map[string]string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, nil)

	// Add params
	if params != nil {
		for key, val := range params {
			c.Params = append(c.Params, gin.Param{Key: key, Value: val})
		}
	}

	return c, w
}

func TestDownloadHandler_Download_Success(t *testing.T) {
	validUUID := uuid.New()
	fileContent := "test file content"
	mockStream := &mockReadCloser{Reader: strings.NewReader(fileContent)}

	mockService := &mockDownloadService{
		filename: "test-evidence.pdf",
		stream:   mockStream,
		filetype: "application/pdf",
	}
	mockAudit := &mockAuditLogger{}

	handler := handlers.NewDownloadHandlerWithInterfaces(mockService, mockAudit)

	c, w := mockGinContext("GET", "/evidence/download/"+validUUID.String(), map[string]string{
		"id": validUUID.String(),
	})

	// Execute
	handler.Download(c)

	// Assert HTTP Response
	assert.Equal(t, http.StatusOK, w.Code, "Status code should be 200")
	assert.Equal(t, "attachment; filename=test-evidence.pdf", w.Header().Get("Content-Disposition"))
	assert.Equal(t, "application/pdf", w.Header().Get("Content-Type"))
	assert.Equal(t, fileContent, w.Body.String(), "Response body should contain file content")

	// Assert Service Interaction
	assert.Equal(t, 1, mockService.callCount, "Service should be called once")
	assert.Equal(t, validUUID, mockService.calledWithID, "Service should be called with correct UUID")

	// Assert Stream was closed
	assert.True(t, mockStream.closed, "Stream should be closed")
	assert.Greater(t, mockStream.readCalls, 0, "Stream should have been read")

	// Assert Audit Log
	assert.Equal(t, 1, mockAudit.callCount, "Audit logger should be called once")
	auditLog := mockAudit.getLastLog()
	assert.Equal(t, "DOWNLOAD_EVIDENCE", auditLog.Action)
	assert.Equal(t, "SUCCESS", auditLog.Status)
	assert.Equal(t, "evidence", auditLog.Target.Type)
	assert.Equal(t, validUUID.String(), auditLog.Target.ID)
	assert.Contains(t, auditLog.Description, "test-evidence.pdf")
	assert.Contains(t, auditLog.Description, "successfully")
}

func TestDownloadHandler_Download_InvalidUUID(t *testing.T) {
	mockService := &mockDownloadService{}
	mockAudit := &mockAuditLogger{}

	handler := handlers.NewDownloadHandlerWithInterfaces(mockService, mockAudit)

	invalidID := "not-a-uuid"
	c, w := mockGinContext("GET", "/evidence/download/"+invalidID, map[string]string{
		"id": invalidID,
	})

	// Execute
	handler.Download(c)

	// Assert HTTP Response
	assert.Equal(t, http.StatusBadRequest, w.Code, "Status code should be 400")
	assert.Contains(t, w.Body.String(), "Invalid evidence ID")

	// Assert Service was NOT called
	assert.Equal(t, 0, mockService.callCount, "Service should not be called with invalid UUID")

	// Assert Audit Log
	assert.Equal(t, 1, mockAudit.callCount, "Audit logger should be called once")
	auditLog := mockAudit.getLastLog()
	assert.Equal(t, "DOWNLOAD_EVIDENCE", auditLog.Action)
	assert.Equal(t, "FAILED", auditLog.Status)
	assert.Equal(t, invalidID, auditLog.Target.ID)
	assert.Contains(t, auditLog.Description, "Invalid UUID format")
}

func TestDownloadHandler_Download_ServiceError(t *testing.T) {
	testCases := []struct {
		name         string
		serviceError error
		expectedMsg  string
	}{
		{
			name:         "NotFound",
			serviceError: errors.New("evidence not found"),
			expectedMsg:  "evidence not found",
		},
		{
			name:         "DatabaseError",
			serviceError: errors.New("database connection failed"),
			expectedMsg:  "database connection failed",
		},
		{
			name:         "PermissionError",
			serviceError: errors.New("access denied"),
			expectedMsg:  "access denied",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validUUID := uuid.New()
			mockService := &mockDownloadService{
				err: tc.serviceError,
			}
			mockAudit := &mockAuditLogger{}

			handler := handlers.NewDownloadHandlerWithInterfaces(mockService, mockAudit)

			c, w := mockGinContext("GET", "/evidence/download/"+validUUID.String(), map[string]string{
				"id": validUUID.String(),
			})

			// Execute
			handler.Download(c)

			// Assert HTTP Response
			assert.Equal(t, http.StatusInternalServerError, w.Code)
			assert.Contains(t, w.Body.String(), "Failed to download evidence")
			assert.Contains(t, w.Body.String(), tc.expectedMsg)

			// Assert Service was called
			assert.Equal(t, 1, mockService.callCount, "Service should be called once")
			assert.Equal(t, validUUID, mockService.calledWithID)

			// Assert Audit Log
			assert.Equal(t, 1, mockAudit.callCount)
			auditLog := mockAudit.getLastLog()
			assert.Equal(t, "FAILED", auditLog.Status)
			assert.Contains(t, auditLog.Description, tc.expectedMsg)
		})
	}
}

func TestDownloadHandler_Download_DifferentFileTypes(t *testing.T) {
	testCases := []struct {
		name     string
		filename string
		filetype string
		content  string
	}{
		{"PDF", "document.pdf", "application/pdf", "PDF content"},
		{"Image", "photo.jpg", "image/jpeg", "JPEG binary"},
		{"Video", "video.mp4", "video/mp4", "MP4 binary"},
		{"Text", "log.txt", "text/plain", "Log file content"},
		{"ZIP", "archive.zip", "application/zip", "ZIP binary"},
		{"JSON", "data.json", "application/json", `{"key":"value"}`},
		{"XML", "config.xml", "application/xml", "<root></root>"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validUUID := uuid.New()
			mockStream := &mockReadCloser{Reader: strings.NewReader(tc.content)}

			mockService := &mockDownloadService{
				filename: tc.filename,
				stream:   mockStream,
				filetype: tc.filetype,
			}
			mockAudit := &mockAuditLogger{}

			handler := handlers.NewDownloadHandlerWithInterfaces(mockService, mockAudit)

			c, w := mockGinContext("GET", "/evidence/download/"+validUUID.String(), map[string]string{
				"id": validUUID.String(),
			})

			handler.Download(c)

			// Assert
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "attachment; filename="+tc.filename, w.Header().Get("Content-Disposition"))
			assert.Equal(t, tc.filetype, w.Header().Get("Content-Type"))
			assert.Equal(t, tc.content, w.Body.String())
			assert.True(t, mockStream.closed, "Stream should be closed")
			assert.Equal(t, 1, mockService.callCount)
			assert.Equal(t, 1, mockAudit.callCount)
		})
	}
}

func TestDownloadHandler_Download_LargeFile(t *testing.T) {
	validUUID := uuid.New()
	largeContent := strings.Repeat("A", 1024*1024) // 1MB
	mockStream := &mockReadCloser{Reader: strings.NewReader(largeContent)}

	mockService := &mockDownloadService{
		filename: "large-file.bin",
		stream:   mockStream,
		filetype: "application/octet-stream",
	}
	mockAudit := &mockAuditLogger{}

	handler := handlers.NewDownloadHandlerWithInterfaces(mockService, mockAudit)

	c, w := mockGinContext("GET", "/evidence/download/"+validUUID.String(), map[string]string{
		"id": validUUID.String(),
	})

	handler.Download(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, len(largeContent), w.Body.Len(), "Full content should be written")
	assert.True(t, mockStream.closed, "Stream should be closed")
	assert.Greater(t, mockStream.readCalls, 1, "Large file should require multiple reads")
}

func TestDownloadHandler_Download_EmptyFile(t *testing.T) {
	validUUID := uuid.New()
	mockStream := &mockReadCloser{Reader: strings.NewReader("")}

	mockService := &mockDownloadService{
		filename: "empty.txt",
		stream:   mockStream,
		filetype: "text/plain",
	}
	mockAudit := &mockAuditLogger{}

	handler := handlers.NewDownloadHandlerWithInterfaces(mockService, mockAudit)

	c, w := mockGinContext("GET", "/evidence/download/"+validUUID.String(), map[string]string{
		"id": validUUID.String(),
	})

	handler.Download(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Body.String(), "Body should be empty")
	assert.True(t, mockStream.closed, "Stream should be closed even for empty files")
}

func TestDownloadHandler_Download_SpecialCharactersInFilename(t *testing.T) {
	testCases := []struct {
		name     string
		filename string
	}{
		{"Spaces", "test file.pdf"},
		{"Parentheses", "file (2024).pdf"},
		{"Dashes", "file-name-with-dashes.txt"},
		{"Underscores", "file_name_with_underscores.txt"},
		{"Numbers", "file123.pdf"},
		{"Mixed", "Test_File-123 (final).pdf"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validUUID := uuid.New()
			mockStream := &mockReadCloser{Reader: strings.NewReader("content")}

			mockService := &mockDownloadService{
				filename: tc.filename,
				stream:   mockStream,
				filetype: "application/pdf",
			}
			mockAudit := &mockAuditLogger{}

			handler := handlers.NewDownloadHandlerWithInterfaces(mockService, mockAudit)

			c, w := mockGinContext("GET", "/evidence/download/"+validUUID.String(), map[string]string{
				"id": validUUID.String(),
			})

			handler.Download(c)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Header().Get("Content-Disposition"), tc.filename)
		})
	}
}

func TestDownloadHandler_Download_AuditLogError(t *testing.T) {
	validUUID := uuid.New()
	mockStream := &mockReadCloser{Reader: strings.NewReader("content")}

	mockService := &mockDownloadService{
		filename: "test.pdf",
		stream:   mockStream,
		filetype: "application/pdf",
	}
	mockAudit := &mockAuditLogger{
		err: errors.New("audit log service unavailable"),
	}

	handler := handlers.NewDownloadHandlerWithInterfaces(mockService, mockAudit)

	c, w := mockGinContext("GET", "/evidence/download/"+validUUID.String(), map[string]string{
		"id": validUUID.String(),
	})

	handler.Download(c)

	// Download should still succeed even if audit logging fails
	assert.Equal(t, http.StatusOK, w.Code, "Download should succeed despite audit log failure")
	assert.Equal(t, "content", w.Body.String())
	assert.True(t, mockStream.closed)

	// Verify audit logger was attempted
	assert.Equal(t, 1, mockAudit.callCount, "Audit logger should have been called")
}

func TestDownloadHandler_Download_StreamCloseError(t *testing.T) {
	validUUID := uuid.New()
	mockStream := &mockReadCloser{
		Reader:   strings.NewReader("content"),
		closeErr: errors.New("failed to close stream"),
	}

	mockService := &mockDownloadService{
		filename: "test.pdf",
		stream:   mockStream,
		filetype: "application/pdf",
	}
	mockAudit := &mockAuditLogger{}

	handler := handlers.NewDownloadHandlerWithInterfaces(mockService, mockAudit)

	c, w := mockGinContext("GET", "/evidence/download/"+validUUID.String(), map[string]string{
		"id": validUUID.String(),
	})

	handler.Download(c)

	// Should still succeed and return content even if close fails
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "content", w.Body.String())
	assert.True(t, mockStream.closed, "Close should have been attempted")
}

func TestNewDownloadHandler(t *testing.T) {
	t.Run("CreateHandlerWithInterfaces", func(t *testing.T) {
		mockService := &mockDownloadService{}
		mockAudit := &mockAuditLogger{}

		handler := handlers.NewDownloadHandlerWithInterfaces(mockService, mockAudit)

		require.NotNil(t, handler, "Handler should not be nil")
	})

	t.Run("HandlerWithNilService", func(t *testing.T) {
		mockAudit := &mockAuditLogger{}
		handler := handlers.NewDownloadHandlerWithInterfaces(nil, mockAudit)

		require.NotNil(t, handler, "Handler should not be nil even with nil service")

		// It will panic when Download is called, but that's expected
	})

	t.Run("HandlerWithNilAuditLogger", func(t *testing.T) {
		mockService := &mockDownloadService{}
		handler := handlers.NewDownloadHandlerWithInterfaces(mockService, nil)

		require.NotNil(t, handler, "Handler should not be nil even with nil audit logger")
	})
}

func TestDownloadHandler_EdgeCases(t *testing.T) {
	t.Run("EmptyUUIDParam", func(t *testing.T) {
		mockService := &mockDownloadService{}
		mockAudit := &mockAuditLogger{}
		handler := handlers.NewDownloadHandlerWithInterfaces(mockService, mockAudit)

		c, w := mockGinContext("GET", "/evidence/download/", map[string]string{
			"id": "",
		})

		handler.Download(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, 0, mockService.callCount, "Service should not be called")
	})

	t.Run("NilUUIDParam", func(t *testing.T) {
		mockService := &mockDownloadService{}
		mockAudit := &mockAuditLogger{}
		handler := handlers.NewDownloadHandlerWithInterfaces(mockService, mockAudit)

		c, w := mockGinContext("GET", "/evidence/download/", nil)

		handler.Download(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, 0, mockService.callCount)
	})

	t.Run("ZeroUUID", func(t *testing.T) {
		mockStream := &mockReadCloser{Reader: strings.NewReader("content")}
		mockService := &mockDownloadService{
			filename: "test.pdf",
			stream:   mockStream,
			filetype: "application/pdf",
		}
		mockAudit := &mockAuditLogger{}
		handler := handlers.NewDownloadHandlerWithInterfaces(mockService, mockAudit)

		zeroUUID := uuid.UUID{} // 00000000-0000-0000-0000-000000000000
		c, w := mockGinContext("GET", "/evidence/download/"+zeroUUID.String(), map[string]string{
			"id": zeroUUID.String(),
		})

		handler.Download(c)

		// Zero UUID is valid, so request should succeed
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, 1, mockService.callCount)
		assert.Equal(t, zeroUUID, mockService.calledWithID)
	})
}
