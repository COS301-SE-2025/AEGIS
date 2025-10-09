// unit_tests/websocket_test_helpers.go
package unit_tests

import (
	wshub "aegis-api/pkg/websocket"
	"aegis-api/services_/notification"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/mock"
)

// MockNotificationService implements notification.NotificationServiceInterface
type MockWsNotificationService struct {
	mock.Mock
}

func (m *MockWsNotificationService) SaveNotification(notif *notification.Notification) error {
	args := m.Called(notif)
	return args.Error(0)
}

func (m *MockWsNotificationService) GetNotificationsForUser(tenantID, teamID, userID string) ([]notification.Notification, error) {
	args := m.Called(tenantID, teamID, userID)
	return args.Get(0).([]notification.Notification), args.Error(1)
}

func (m *MockWsNotificationService) MarkAsRead(notificationIDs []string) error {
	args := m.Called(notificationIDs)
	return args.Error(0)
}

func (m *MockWsNotificationService) ArchiveNotifications(notificationIDs []string) error {
	args := m.Called(notificationIDs)
	return args.Error(0)
}

func (m *MockWsNotificationService) DeleteNotifications(notificationIDs []string) error {
	args := m.Called(notificationIDs)
	return args.Error(0)
}

// TestHubWrapper wraps Hub for testing
type TestHubWrapper struct {
	*wshub.Hub
}

func NewTestHub(notificationService *MockWsNotificationService) *TestHubWrapper {
	mockService := &notification.NotificationService{}
	if notificationService != nil {
		// You might need to adapt this based on your actual NotificationService structure
	}

	hub := wshub.NewHub(mockService)
	return &TestHubWrapper{hub}
}

// Helper function to create mock WebSocket connections
func CreateMockWebSocket() (*httptest.Server, *websocket.Conn) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		// Handle connection with timeout
		go func() {
			defer conn.Close()
			conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			for {
				if _, _, err := conn.NextReader(); err != nil {
					break
				}
				conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			}
		}()
	}))

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		server.Close()
		return nil, nil
	}

	return server, conn
}

// TestClientWrapper for testing client methods
type TestClientWrapper struct {
	*wshub.Client
	Hub *TestHubWrapper
}

func NewTestClient(hub *TestHubWrapper, userID, caseID string, conn *websocket.Conn) *TestClientWrapper {
	client := &wshub.Client{
		// Minimal client for tests: embed the shared client struct if needed in future tests
		Hub:  hub.Hub,
		Send: make(chan []byte, 256),
	}
	return &TestClientWrapper{client, hub}
}
