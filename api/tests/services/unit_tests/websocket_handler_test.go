// unit_tests/handler_test.go
package unit_tests

import (
	"aegis-api/pkg/chatModels"
	wshub "aegis-api/pkg/websocket"
	"aegis-api/services_/notification"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockWebSocketManager implements chatModels.WebSocketManager interface
type MockWSocketManager struct {
	mock.Mock
}

// AddConnection implements chatModels.WebSocketManager.
func (m *MockWSocketManager) AddConnection(userID string, caseID string, conn *websocket.Conn) {
	panic("unimplemented")
}

// BroadcastToCase implements chatModels.WebSocketManager.
func (m *MockWSocketManager) BroadcastToCase(caseID string, message chatModels.WebSocketMessage) error {
	panic("unimplemented")
}

func (m *MockWSocketManager) HandleConnection(w http.ResponseWriter, r *http.Request) error {
	args := m.Called(w, r)
	return args.Error(0)
}

func (m *MockWSocketManager) BroadcastToGroup(groupID string, message chatModels.WebSocketMessage) error {
	args := m.Called(groupID, message)
	return args.Error(0)
}

func (m *MockWSocketManager) AddUserToGroup(userID, userEmail, groupID string, conn *websocket.Conn) error {
	args := m.Called(userID, userEmail, groupID, conn)
	return args.Error(0)
}

func (m *MockWSocketManager) RemoveUserFromGroup(userID, groupID string) error {
	args := m.Called(userID, groupID)
	return args.Error(0)
}

func (m *MockWSocketManager) SendToUser(userID string, message interface{}) error {
	args := m.Called(userID, message)
	return args.Error(0)
}

func (m *MockWSocketManager) BroadcastTypingStart(groupID, userEmail string) error {
	args := m.Called(groupID, userEmail)
	return args.Error(0)
}

func (m *MockWSocketManager) BroadcastTypingStop(groupID, userEmail string) error {
	args := m.Called(groupID, userEmail)
	return args.Error(0)
}

func (m *MockWSocketManager) GetActiveUsers(groupID string) []string {
	args := m.Called(groupID)
	return args.Get(0).([]string)
}

func TestRegisterWebSocketRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	rg := router.Group("/ws")

	mockManager := new(MockWSocketManager)

	wshub.RegisterWebSocketRoutes(rg, mockManager)

	// Verify route is registered
	routes := router.Routes()
	found := false
	for _, route := range routes {
		if strings.Contains(route.Path, "/ws/cases/:caseId") {
			found = true
			break
		}
	}

	assert.True(t, found, "WebSocket route should be registered")
}

func TestWebSocketRoute_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	rg := router.Group("/ws")

	mockManager := new(MockWSocketManager)
	mockManager.On("HandleConnection", mock.Anything, mock.Anything).Return(nil)

	wshub.RegisterWebSocketRoutes(rg, mockManager)

	req := httptest.NewRequest("GET", "/ws/cases/case123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// The route should be hit (status may vary based on HandleConnection implementation)
	assert.NotEqual(t, 404, w.Code, "Route should exist")
}

func TestWebSocketRoute_HandleConnectionError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	rg := router.Group("/ws")

	mockManager := new(MockWSocketManager)
	expectedErr := fmt.Errorf("connection failed")
	mockManager.On("HandleConnection", mock.Anything, mock.Anything).Return(expectedErr)

	wshub.RegisterWebSocketRoutes(rg, mockManager)

	req := httptest.NewRequest("GET", "/ws/cases/case123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should handle error gracefully
	mockManager.AssertExpectations(t)
}

func TestWebSocketRoute_WithRealHub(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	rg := router.Group("/ws")

	hub := wshub.NewHub(&notification.NotificationService{})
	go hub.Run()

	wshub.RegisterWebSocketRoutes(rg, hub)

	// Set JWT secret for auth
	os.Setenv("JWT_SECRET_KEY", "test-secret")
	defer os.Unsetenv("JWT_SECRET_KEY")

	// Create valid token
	claims := &wshub.Claims{
		Email:    "test@example.com",
		UserID:   "user123",
		TenantID: "tenant123",
		TeamID:   "team123",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("test-secret"))

	req := httptest.NewRequest("GET", "/ws/cases/case123?token="+tokenString, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should attempt to upgrade connection (will fail in test but route works)
	assert.NotEqual(t, 404, w.Code)
}

func TestWebSocketRoute_MissingCaseID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	rg := router.Group("/ws")

	mockManager := new(MockWSocketManager)

	wshub.RegisterWebSocketRoutes(rg, mockManager)

	req := httptest.NewRequest("GET", "/ws/cases/", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 404 for invalid route
	assert.Equal(t, 404, w.Code)
}

func TestWebSocketRoute_GETMethodOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	rg := router.Group("/ws")

	mockManager := new(MockWSocketManager)

	wshub.RegisterWebSocketRoutes(rg, mockManager)

	// Try POST method (should not work)
	req := httptest.NewRequest("POST", "/ws/cases/case123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code, "Only GET method should be registered")
}

func TestWebSocketRoute_MultipleRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	rg := router.Group("/ws")

	mockManager := new(MockWSocketManager)
	mockManager.On("HandleConnection", mock.Anything, mock.Anything).Return(nil)

	wshub.RegisterWebSocketRoutes(rg, mockManager)

	// Make multiple requests
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", fmt.Sprintf("/ws/cases/case%d", i), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.NotEqual(t, 404, w.Code)
	}

	mockManager.AssertNumberOfCalls(t, "HandleConnection", 3)
}

func TestWebSocketRoute_WithQueryParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	rg := router.Group("/ws")

	mockManager := new(MockWSocketManager)
	mockManager.On("HandleConnection", mock.Anything, mock.Anything).Return(nil)

	wshub.RegisterWebSocketRoutes(rg, mockManager)

	req := httptest.NewRequest("GET", "/ws/cases/case123?token=test-token&other=param", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.NotEqual(t, 404, w.Code)
	mockManager.AssertExpectations(t)
}

func TestWebSocketRoute_DifferentCaseIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	rg := router.Group("/ws")

	mockManager := new(MockWSocketManager)
	mockManager.On("HandleConnection", mock.Anything, mock.Anything).Return(nil)

	wshub.RegisterWebSocketRoutes(rg, mockManager)

	caseIDs := []string{"case123", "case-abc-def", "12345", "test_case"}

	for _, caseID := range caseIDs {
		req := httptest.NewRequest("GET", "/ws/cases/"+caseID, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.NotEqual(t, 404, w.Code, "Should accept case ID: "+caseID)
	}
}

func TestWebSocketRoute_WithHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	rg := router.Group("/ws")

	mockManager := new(MockWSocketManager)
	mockManager.On("HandleConnection", mock.Anything, mock.Anything).Return(nil)

	wshub.RegisterWebSocketRoutes(rg, mockManager)

	req := httptest.NewRequest("GET", "/ws/cases/case123", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("User-Agent", "Test Client")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.NotEqual(t, 404, w.Code)
	mockManager.AssertExpectations(t)
}

func TestWebSocketRoute_Concurrent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	rg := router.Group("/ws")

	mockManager := new(MockWSocketManager)
	mockManager.On("HandleConnection", mock.Anything, mock.Anything).Return(nil)

	wshub.RegisterWebSocketRoutes(rg, mockManager)

	done := make(chan bool, 5)

	for i := 0; i < 5; i++ {
		go func(index int) {
			req := httptest.NewRequest("GET", fmt.Sprintf("/ws/cases/case%d", index), nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.NotEqual(t, 404, w.Code)
			done <- true
		}(i)
	}

	// Wait for all requests to complete
	for i := 0; i < 5; i++ {
		<-done
	}

	mockManager.AssertNumberOfCalls(t, "HandleConnection", 5)
}

func TestWebSocketRoute_HandlerLogging(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	rg := router.Group("/ws")

	mockManager := new(MockWSocketManager)
	mockManager.On("HandleConnection", mock.Anything, mock.Anything).Return(nil)

	wshub.RegisterWebSocketRoutes(rg, mockManager)

	req := httptest.NewRequest("GET", "/ws/cases/case123", nil)
	w := httptest.NewRecorder()

	// This test verifies the route logs properly (manual verification in actual logs)
	router.ServeHTTP(w, req)

	mockManager.AssertCalled(t, "HandleConnection", mock.Anything, mock.Anything)
}
