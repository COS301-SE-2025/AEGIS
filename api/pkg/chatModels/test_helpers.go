// chatModels/websocket_test_helpers.go
package chatModels

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockUserService implements UserService for testing
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserService) GetUsersByEmails(ctx context.Context, emails []string) ([]*User, error) {
	args := m.Called(ctx, emails)
	return args.Get(0).([]*User), args.Error(1)
}

func (m *MockUserService) ValidateUserExists(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

// MockChatRepository implements ChatRepository for testing
type MockChatRepository struct {
	mock.Mock
}

func (m *MockChatRepository) CreateGroup(ctx context.Context, group *ChatGroup) error {
	args := m.Called(ctx, group)
	return args.Error(0)
}

func (m *MockChatRepository) GetGroupByID(ctx context.Context, groupID primitive.ObjectID) (*ChatGroup, error) {
	args := m.Called(ctx, groupID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ChatGroup), args.Error(1)
}

func (m *MockChatRepository) GetUserGroups(ctx context.Context, userEmail string) ([]*ChatGroup, error) {
	args := m.Called(ctx, userEmail)
	return args.Get(0).([]*ChatGroup), args.Error(1)
}

func (m *MockChatRepository) UpdateGroup(ctx context.Context, group *ChatGroup) error {
	args := m.Called(ctx, group)
	return args.Error(0)
}

func (m *MockChatRepository) DeleteGroup(ctx context.Context, groupID primitive.ObjectID) error {
	args := m.Called(ctx, groupID)
	return args.Error(0)
}

func (m *MockChatRepository) AddMemberToGroup(ctx context.Context, groupID primitive.ObjectID, member *Member) error {
	args := m.Called(ctx, groupID, member)
	return args.Error(0)
}

func (m *MockChatRepository) RemoveMemberFromGroup(ctx context.Context, groupID primitive.ObjectID, userEmail string) error {
	args := m.Called(ctx, groupID, userEmail)
	return args.Error(0)
}

func (m *MockChatRepository) IsUserInGroup(ctx context.Context, groupID primitive.ObjectID, userEmail string) (bool, error) {
	args := m.Called(ctx, groupID, userEmail)
	return args.Bool(0), args.Error(1)
}

func (m *MockChatRepository) UpdateLastMessage(ctx context.Context, groupID primitive.ObjectID, lastMessage *LastMessage) error {
	args := m.Called(ctx, groupID, lastMessage)
	return args.Error(0)
}

func (m *MockChatRepository) GetGroupsByCaseID(ctx context.Context, caseID primitive.ObjectID) ([]*ChatGroup, error) {
	args := m.Called(ctx, caseID)
	return args.Get(0).([]*ChatGroup), args.Error(1)
}

func (m *MockChatRepository) UpdateGroupImage(ctx context.Context, groupID primitive.ObjectID, imageURL string) error {
	args := m.Called(ctx, groupID, imageURL)
	return args.Error(0)
}

func (m *MockChatRepository) CreateMessage(ctx context.Context, message *Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockChatRepository) GetMessageByID(ctx context.Context, messageID primitive.ObjectID) (*Message, error) {
	args := m.Called(ctx, messageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Message), args.Error(1)
}

func (m *MockChatRepository) GetMessages(ctx context.Context, groupID primitive.ObjectID, limit int, before *primitive.ObjectID) ([]*Message, error) {
	args := m.Called(ctx, groupID, limit, before)
	return args.Get(0).([]*Message), args.Error(1)
}

func (m *MockChatRepository) SearchMessages(ctx context.Context, groupID primitive.ObjectID, query string, limit int, skip int) ([]*Message, error) {
	args := m.Called(ctx, groupID, query, limit, skip)
	return args.Get(0).([]*Message), args.Error(1)
}

func (m *MockChatRepository) UpdateMessage(ctx context.Context, message *Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockChatRepository) DeleteMessage(ctx context.Context, messageID primitive.ObjectID) error {
	args := m.Called(ctx, messageID)
	return args.Error(0)
}

func (m *MockChatRepository) MarkMessagesAsRead(ctx context.Context, groupID primitive.ObjectID, messageIDs []primitive.ObjectID, userEmail string) error {
	args := m.Called(ctx, groupID, messageIDs, userEmail)
	return args.Error(0)
}

func (m *MockChatRepository) GetUnreadCount(ctx context.Context, groupID primitive.ObjectID, userEmail string) (int, error) {
	args := m.Called(ctx, groupID, userEmail)
	return args.Int(0), args.Error(1)
}

func (m *MockChatRepository) GetGroupMembers(ctx context.Context, groupID primitive.ObjectID) ([]*Member, error) {
	args := m.Called(ctx, groupID)
	return args.Get(0).([]*Member), args.Error(1)
}

func (m *MockChatRepository) IsGroupAdmin(ctx context.Context, groupID primitive.ObjectID, userEmail string) (bool, error) {
	args := m.Called(ctx, groupID, userEmail)
	return args.Bool(0), args.Error(1)
}

func (m *MockChatRepository) GetUndeliveredMessages(ctx context.Context, userEmail string, limit int, before *primitive.ObjectID) ([]*Message, error) {
	args := m.Called(ctx, userEmail, limit, before)
	return args.Get(0).([]*Message), args.Error(1)
}

func (m *MockChatRepository) MarkMessagesAsDelivered(ctx context.Context, groupID primitive.ObjectID, messageIDs []primitive.ObjectID, userEmail string) error {
	args := m.Called(ctx, groupID, messageIDs, userEmail)
	return args.Error(0)
}

// TestWebSocketManager is an exported wrapper for testing
type TestWebSocketManager struct {
	*webSocketManager
}

// NewTestWebSocketManager creates a new WebSocket manager for testing
func NewTestWebSocketManager(userService UserService, repo ChatRepository) *TestWebSocketManager {
	return &TestWebSocketManager{
		webSocketManager: NewWebSocketManager(userService, repo).(*webSocketManager),
	}
}

// Export some internal methods for testing
func (w *TestWebSocketManager) HandleTypingIndicator(userEmail, groupID string, isTyping bool) {
	w.handleTypingIndicator(userEmail, groupID, isTyping)
}

func (w *TestWebSocketManager) CleanupTypingIndicators(stopChan chan struct{}) {
	w.cleanupTypingIndicators(stopChan)
}

// GetInternalState returns internal state for testing (use carefully)
func (w *TestWebSocketManager) GetInternalState() (map[string][]string, map[string][]string, map[string]map[string]time.Time) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	// Return copies to avoid race conditions
	groupUsers := make(map[string][]string)
	for k, v := range w.groupUsers {
		groupUsers[k] = append([]string{}, v...)
	}

	userGroups := make(map[string][]string)
	for k, v := range w.userGroups {
		userGroups[k] = append([]string{}, v...)
	}

	typingUsers := make(map[string]map[string]time.Time)
	for k, v := range w.typingUsers {
		typingUsers[k] = make(map[string]time.Time)
		for k2, v2 := range v {
			typingUsers[k][k2] = v2
		}
	}

	return groupUsers, userGroups, typingUsers
}
