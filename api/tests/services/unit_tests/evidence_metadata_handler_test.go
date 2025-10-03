package unit_tests

// import (
// 	"aegis-api/handlers"
// 	"aegis-api/services_/evidence/metadata"
// 	"bytes"
// 	"fmt"
// 	"io"
// 	"mime/multipart"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/gin-gonic/gin"
// 	"github.com/google/uuid"
// )

// // testContext simulates a minimal gin.Context
// type testContext struct {
// 	*gin.Context
// 	userID     string
// 	userRole   string
// 	tenantID   string
// 	teamID     string
// 	email      string
// 	clientIP   string
// 	userAgent  string
// 	params     map[string]string
// 	postForm   map[string]string
// 	formFiles  map[string][]*multipart.FileHeader
// 	jsonStatus int
// 	jsonOutput *bytes.Buffer
// 	getCalled  map[string]bool
// }

// func newTestContext() *testContext {
// 	gin.SetMode(gin.TestMode)
// 	w := httptest.NewRecorder()
// 	c, _ := gin.CreateTestContext(w)
// 	// Initialize request with a valid RemoteAddr to avoid ClientIP panic
// 	req := httptest.NewRequest("POST", "/", nil)
// 	req.RemoteAddr = "127.0.0.1:12345"
// 	c.Request = req
// 	return &testContext{
// 		Context:    c,
// 		getCalled:  make(map[string]bool),
// 		params:     make(map[string]string),
// 		postForm:   make(map[string]string),
// 		jsonOutput: new(bytes.Buffer),
// 	}
// }

// func (c *testContext) Get(key string) (interface{}, bool) {
// 	c.getCalled[key] = true
// 	switch key {
// 	case "userID":
// 		return c.userID, c.userID != ""
// 	case "userRole":
// 		return c.userRole, c.userRole != ""
// 	case "tenantID":
// 		return c.tenantID, c.tenantID != ""
// 	case "teamID":
// 		return c.teamID, c.teamID != ""
// 	case "email":
// 		return c.email, c.email != ""
// 	}
// 	return nil, false
// }

// func (c *testContext) Param(key string) string {
// 	return c.params[key]
// }

// func (c *testContext) ClientIP() string {
// 	return c.clientIP
// }

// func (c *testContext) Request() *http.Request {
// 	req := c.Context.Request
// 	if req == nil {
// 		req = httptest.NewRequest("POST", "/", nil)
// 		req.RemoteAddr = c.clientIP + ":12345"
// 	}
// 	req.Header.Set("User-Agent", c.userAgent)
// 	if len(c.postForm) > 0 || len(c.formFiles) > 0 {
// 		body := &bytes.Buffer{}
// 		writer := multipart.NewWriter(body)
// 		for key, value := range c.postForm {
// 			writer.WriteField(key, value)
// 		}
// 		for _, files := range c.formFiles {
// 			for _, file := range files {
// 				part, err := writer.CreateFormFile("files", file.Filename)
// 				if err != nil {
// 					return req
// 				}
// 				part.Write([]byte("dummy content"))
// 			}
// 		}
// 		writer.Close()
// 		req.Header.Set("Content-Type", writer.FormDataContentType())
// 		req.Body = http.NoBody
// 		req.ContentLength = int64(body.Len())
// 		req.Body = io.NopCloser(body)
// 	}
// 	return req
// }

// func (c *testContext) PostForm(key string) string {
// 	return c.postForm[key]
// }

// func (c *testContext) MultipartForm() (*multipart.Form, error) {
// 	if len(c.formFiles) == 0 && len(c.postForm) == 0 {
// 		return nil, fmt.Errorf("no multipart form")
// 	}
// 	return &multipart.Form{
// 		Value: map[string][]string{},
// 		File:  c.formFiles,
// 	}, nil
// }

// func (c *testContext) JSON(status int, obj interface{}) {
// 	c.jsonStatus = status
// 	c.jsonOutput = new(bytes.Buffer)
// 	switch v := obj.(type) {
// 	case map[string]interface{}:
// 		fmt.Fprintf(c.jsonOutput, "{")
// 		first := true
// 		for k, val := range v {
// 			if !first {
// 				c.jsonOutput.WriteString(",")
// 			}
// 			switch v := val.(type) {
// 			case string:
// 				fmt.Fprintf(c.jsonOutput, `"%s":"%s"`, k, v)
// 			case bool:
// 				fmt.Fprintf(c.jsonOutput, `"%s":%t`, k, v)
// 			case int:
// 				fmt.Fprintf(c.jsonOutput, `"%s":%d`, k, v)
// 			default:
// 				fmt.Fprintf(c.jsonOutput, `"%s":"%v"`, k, v)
// 			}
// 			first = false
// 		}
// 		c.jsonOutput.WriteString("}")
// 	case []interface{}:
// 		fmt.Fprintf(c.jsonOutput, "[")
// 		for i, item := range v {
// 			if m, ok := item.(map[string]interface{}); ok {
// 				fmt.Fprintf(c.jsonOutput, "{")
// 				first := true
// 				for k, val := range m {
// 					if !first {
// 						c.jsonOutput.WriteString(",")
// 					}
// 					switch v := val.(type) {
// 					case string:
// 						fmt.Fprintf(c.jsonOutput, `"%s":"%s"`, k, v)
// 					case uuid.UUID:
// 						fmt.Fprintf(c.jsonOutput, `"%s":"%s"`, k, v.String())
// 					default:
// 						fmt.Fprintf(c.jsonOutput, `"%s":"%v"`, k, v)
// 					}
// 					first = false
// 				}
// 				fmt.Fprintf(c.jsonOutput, "}")
// 			} else {
// 				fmt.Fprintf(c.jsonOutput, `"%v"`, item)
// 			}
// 			if i < len(v)-1 {
// 				c.jsonOutput.WriteString(",")
// 			}
// 		}
// 		c.jsonOutput.WriteString("]")
// 	case metadata.Evidence:
// 		fmt.Fprintf(c.jsonOutput, `{"id":"%s","filename":"%s"}`, v.ID.String(), v.Filename)
// 	case *metadata.Evidence:
// 		if v != nil {
// 			fmt.Fprintf(c.jsonOutput, `{"id":"%s","filename":"%s"}`, v.ID.String(), v.Filename)
// 		} else {
// 			c.jsonOutput.WriteString("null")
// 		}
// 	default:
// 		fmt.Fprintf(c.jsonOutput, "%v", obj)
// 	}
// }

// // mockMetadataService, mockAuditLogger, mockCacheClient definitions remain unchanged
// // ... (include the existing mockMetadataService, mockAuditLogger, and mockCacheClient from your code)

// func TestUploadEvidence(t *testing.T) {
// 	validCaseID := uuid.New().String()
// 	validUserID := uuid.New().String()

// 	tests := []struct {
// 		name           string
// 		context        *testContext
// 		service        *mockMetadataService
// 		logger         *mockAuditLogger
// 		cacheClient    *mockCacheClient
// 		expectedStatus int
// 		expectedBody   string
// 	}{
// 		{
// 			name: "Missing Tenant or Team Context",
// 			context: &testContext{
// 				Context:   newTestContext().Context,
// 				userID:    "user123",
// 				userRole:  "admin",
// 				clientIP:  "127.0.0.1",
// 				userAgent: "test-agent",
// 			},
// 			service:        &mockMetadataService{},
// 			logger:         nil, // Match handler call
// 			cacheClient:    &mockCacheClient{},
// 			expectedStatus: http.StatusUnauthorized,
// 			expectedBody:   `{"error":"Tenant or Team context missing"}`,
// 		},
// 		{
// 			name: "Invalid Case ID",
// 			context: &testContext{
// 				Context:   newTestContext().Context,
// 				userID:    "user123",
// 				userRole:  "admin",
// 				tenantID:  "tenant123",
// 				teamID:    "team123",
// 				clientIP:  "127.0.0.1",
// 				userAgent: "test-agent",
// 				postForm:  map[string]string{"caseId": "invalid", "uploadedBy": validUserID},
// 			},
// 			service:        &mockMetadataService{},
// 			logger:         nil,
// 			cacheClient:    &mockCacheClient{},
// 			expectedStatus: http.StatusBadRequest,
// 			expectedBody:   `{"error":"Invalid caseId format"}`,
// 		},
// 		{
// 			name: "Invalid UploadedBy ID",
// 			context: &testContext{
// 				Context:   newTestContext().Context,
// 				userID:    "user123",
// 				userRole:  "admin",
// 				tenantID:  "tenant123",
// 				teamID:    "team123",
// 				clientIP:  "127.0.0.1",
// 				userAgent: "test-agent",
// 				postForm:  map[string]string{"caseId": validCaseID, "uploadedBy": "invalid"},
// 			},
// 			service:        &mockMetadataService{},
// 			logger:         nil,
// 			cacheClient:    &mockCacheClient{},
// 			expectedStatus: http.StatusBadRequest,
// 			expectedBody:   `{"error":"Invalid uploadedBy format"}`,
// 		},
// 		{
// 			name: "No Files Uploaded",
// 			context: &testContext{
// 				Context:   newTestContext().Context,
// 				userID:    "user123",
// 				userRole:  "admin",
// 				tenantID:  "tenant123",
// 				teamID:    "team123",
// 				clientIP:  "127.0.0.1",
// 				userAgent: "test-agent",
// 				postForm:  map[string]string{"caseId": validCaseID, "uploadedBy": validUserID},
// 			},
// 			service:        &mockMetadataService{},
// 			logger:         nil,
// 			cacheClient:    &mockCacheClient{},
// 			expectedStatus: http.StatusBadRequest,
// 			expectedBody:   `{"error":"No files uploaded"}`,
// 		},
// 		{
// 			name: "Service Error",
// 			context: &testContext{
// 				Context:   newTestContext().Context,
// 				userID:    "user123",
// 				userRole:  "admin",
// 				tenantID:  "tenant123",
// 				teamID:    "team123",
// 				clientIP:  "127.0.0.1",
// 				userAgent: "test-agent",
// 				postForm:  map[string]string{"caseId": validCaseID, "uploadedBy": validUserID, "fileType": "text/plain"},
// 				formFiles: map[string][]*multipart.FileHeader{
// 					"files": {{Filename: "test.txt", Size: 10}},
// 				},
// 			},
// 			service: &mockMetadataService{
// 				uploadEvidenceError: fmt.Errorf("service error"),
// 			},
// 			logger:         nil,
// 			cacheClient:    &mockCacheClient{},
// 			expectedStatus: http.StatusInternalServerError,
// 			expectedBody:   `{"error":"service error"}`,
// 		},
// 		{
// 			name: "Success",
// 			context: &testContext{
// 				Context:   newTestContext().Context,
// 				userID:    "user123",
// 				userRole:  "admin",
// 				tenantID:  "tenant123",
// 				teamID:    "team123",
// 				clientIP:  "127.0.0.1",
// 				userAgent: "test-agent",
// 				postForm:  map[string]string{"caseId": validCaseID, "uploadedBy": validUserID, "fileType": "text/plain"},
// 				formFiles: map[string][]*multipart.FileHeader{
// 					"files": {{Filename: "test.txt", Size: 10}},
// 				},
// 			},
// 			service:        &mockMetadataService{},
// 			logger:         &mockAuditLogger{},
// 			cacheClient:    &mockCacheClient{delResult: 1},
// 			expectedStatus: http.StatusOK,
// 			expectedBody:   `{"message":"Evidence uploaded successfully"}`,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Set context values
// 			tt.context.Set("userID", tt.context.userID)
// 			tt.context.Set("userRole", tt.context.userRole)
// 			tt.context.Set("tenantID", tt.context.tenantID)
// 			tt.context.Set("teamID", tt.context.teamID)
// 			// Update request with form data
// 			tt.context.Context.Request = tt.context.Request()

// 			handler := handlers.NewMetadataHandler(tt.service, tt.logger, tt.cacheClient)
// 			handler.UploadEvidence(tt.context.Context)
// 			if tt.context.jsonStatus != tt.expectedStatus {
// 				t.Errorf("expected status %d, got %d", tt.expectedStatus, tt.context.jsonStatus)
// 			}
// 			if tt.context.jsonOutput == nil || tt.context.jsonOutput.String() != tt.expectedBody {
// 				t.Errorf("expected body %q, got %q", tt.expectedBody, tt.context.jsonOutput.String())
// 			}
// 			if tt.name == "Success" {
// 				if !tt.service.uploadEvidenceCalled {
// 					t.Error("expected UploadEvidence to be called")
// 				}
// 				if tt.logger != nil && tt.logger.callCount == 0 {
// 					t.Error("expected auditLogger.Log to be called")
// 				}
// 				if !tt.cacheClient.delCalled {
// 					t.Error("expected cacheClient.Del to be called")
// 				}
// 			}
// 		})
// 	}
// }
