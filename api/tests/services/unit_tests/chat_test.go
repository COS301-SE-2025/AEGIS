package unit_tests

import (
	//"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http"

	"testing"
	"time"

	chat "aegis-api/services_/chat"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// --- Mock Collection ---

type MockChatCollection struct {
	mock.Mock
}

func (m *MockChatCollection) InsertOne(ctx context.Context, doc interface{}, opts ...interface{}) (*mongo.InsertOneResult, error) {
	args := m.Called(ctx, doc)
	return args.Get(0).(*mongo.InsertOneResult), args.Error(1)
}
func (m *MockChatCollection) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...interface{}) (*mongo.UpdateResult, error) {
	args := m.Called(ctx, filter, update)
	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}
func (m *MockChatCollection) CountDocuments(ctx context.Context, filter interface{}, opts ...interface{}) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// --- Mock Repository (if you have a repository interface) ---

type MockChatRepository struct {
	mock.Mock
}

// DeleteMessage implements chat.ChatRepository.
func (m *MockChatRepository) DeleteMessage(ctx context.Context, messageID primitive.ObjectID) error {
	panic("unimplemented")
}

// GetGroupByID implements chat.ChatRepository.
func (m *MockChatRepository) GetGroupByID(ctx context.Context, groupID primitive.ObjectID) (*chat.ChatGroup, error) {
	panic("unimplemented")
}

// GetGroupMembers implements chat.ChatRepository.
func (m *MockChatRepository) GetGroupMembers(ctx context.Context, groupID primitive.ObjectID) ([]*chat.Member, error) {
	panic("unimplemented")
}

// GetMessageByID implements chat.ChatRepository.
func (m *MockChatRepository) GetMessageByID(ctx context.Context, messageID primitive.ObjectID) (*chat.Message, error) {
	panic("unimplemented")
}

// GetMessages implements chat.ChatRepository.
func (m *MockChatRepository) GetMessages(ctx context.Context, groupID primitive.ObjectID, limit int, before *primitive.ObjectID) ([]*chat.Message, error) {
	panic("unimplemented")
}

// GetUnreadCount implements chat.ChatRepository.
func (m *MockChatRepository) GetUnreadCount(ctx context.Context, groupID primitive.ObjectID, userEmail string) (int, error) {
	panic("unimplemented")
}

// GetUserGroups implements chat.ChatRepository.
func (m *MockChatRepository) GetUserGroups(ctx context.Context, userEmail string) ([]*chat.ChatGroup, error) {
	panic("unimplemented")
}

// IsGroupAdmin implements chat.ChatRepository.
func (m *MockChatRepository) IsGroupAdmin(ctx context.Context, groupID primitive.ObjectID, userEmail string) (bool, error) {
	panic("unimplemented")
}

// MarkMessagesAsRead implements chat.ChatRepository.
func (m *MockChatRepository) MarkMessagesAsRead(ctx context.Context, groupID primitive.ObjectID, messageIDs []primitive.ObjectID, userEmail string) error {
	panic("unimplemented")
}

// RemoveMemberFromGroup implements chat.ChatRepository.
func (m *MockChatRepository) RemoveMemberFromGroup(ctx context.Context, groupID primitive.ObjectID, userEmail string) error {
	panic("unimplemented")
}

// SearchMessages implements chat.ChatRepository.
func (m *MockChatRepository) SearchMessages(ctx context.Context, groupID primitive.ObjectID, query string, limit int, skip int) ([]*chat.Message, error) {
	panic("unimplemented")
}

// UpdateGroup implements chat.ChatRepository.
func (m *MockChatRepository) UpdateGroup(ctx context.Context, group *chat.ChatGroup) error {
	panic("unimplemented")
}

// UpdateLastMessage implements chat.ChatRepository.
func (m *MockChatRepository) UpdateLastMessage(ctx context.Context, groupID primitive.ObjectID, lastMessage *chat.LastMessage) error {
	panic("unimplemented")
}

// UpdateMessage implements chat.ChatRepository.
func (m *MockChatRepository) UpdateMessage(ctx context.Context, message *chat.Message) error {
	panic("unimplemented")
}

// MarkMessagesAsDelivered implements chat.ChatRepository.
func (m *MockChatRepository) MarkMessagesAsDelivered(ctx context.Context, groupID primitive.ObjectID, messageIDs []primitive.ObjectID, userEmail string) error {
	args := m.Called(ctx, groupID, messageIDs, userEmail)
	return args.Error(0)
}

// GetUndeliveredMessages implements chat.ChatRepository.
func (m *MockChatRepository) GetUndeliveredMessages(ctx context.Context, userEmail string, limit int, before *primitive.ObjectID) ([]*chat.Message, error) {
	args := m.Called(ctx, userEmail, limit, before)
	return args.Get(0).([]*chat.Message), args.Error(1)
}

func (m *MockChatRepository) CreateGroup(ctx context.Context, group *chat.ChatGroup) error {
	args := m.Called(ctx, group)
	return args.Error(0)
}
func (m *MockChatRepository) AddMemberToGroup(ctx context.Context, groupID primitive.ObjectID, member *chat.Member) error {
	args := m.Called(ctx, groupID, member)
	return args.Error(0)
}
func (m *MockChatRepository) CreateMessage(ctx context.Context, msg *chat.Message) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}
func (m *MockChatRepository) IsUserInGroup(ctx context.Context, groupID primitive.ObjectID, userEmail string) (bool, error) {
	args := m.Called(ctx, groupID, userEmail)
	return args.Bool(0), args.Error(1)
}

// func (m *MockChatRepository) GetUndeliveredMessages(ctx context.Context, groupID string, limit int, before *primitive.ObjectID) ([]*chat.Message, error) {
// 	args := m.Called(ctx, groupID, limit, before)
// 	return args.Get(0).([]*chat.Message), args.Error(1)
// }

type MockIPFSUploader struct {
	mock.Mock
}

func (m *MockIPFSUploader) UploadFile(ctx context.Context, file multipart.File, fileName string) (*chat.IPFSUploadResult, error) {
	args := m.Called(ctx, file, fileName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*chat.IPFSUploadResult), args.Error(1)
}

func (m *MockIPFSUploader) UploadBytes(ctx context.Context, data []byte, fileName string) (*chat.IPFSUploadResult, error) {
	args := m.Called(ctx, data, fileName)
	return args.Get(0).(*chat.IPFSUploadResult), args.Error(1)
}

func (m *MockIPFSUploader) GetFileURL(hash string) string {
	args := m.Called(hash)
	return args.String(0)
}

func (m *MockIPFSUploader) DeleteFile(ctx context.Context, hash string) error {
	args := m.Called(ctx, hash)
	return args.Error(0)
}

type MockWebSocketManager struct {
	mock.Mock
}

// GetActiveUsers implements chat.WebSocketManager.
func (m *MockWebSocketManager) GetActiveUsers(groupID string) []string {
	panic("unimplemented")
}

// HandleConnection implements chat.WebSocketManager.
func (m *MockWebSocketManager) HandleConnection(userEmail string, w http.ResponseWriter, r *http.Request) error {
	panic("unimplemented")
}

// RemoveUserFromGroup implements chat.WebSocketManager.
func (m *MockWebSocketManager) RemoveUserFromGroup(userEmail string, groupID string) error {
	panic("unimplemented")
}

func (m *MockWebSocketManager) BroadcastToGroup(groupID string, message interface{}) error {
	args := m.Called(groupID, message)
	return args.Error(0)
}

func (m *MockWebSocketManager) SendToUser(email string, message interface{}) error {
	args := m.Called(email, message)
	return args.Error(0)
}
func (m *MockChatRepository) DeleteGroup(ctx context.Context, groupID primitive.ObjectID) error {
	args := m.Called(ctx, groupID)
	return args.Error(0)
}
func (m *MockWebSocketManager) AddUserToGroup(email, groupID string) error {
	args := m.Called(email, groupID)
	return args.Error(0)
}

func TestCreateGroup_Success(t *testing.T) {
	repo := new(MockChatRepository)
	group := &chat.ChatGroup{
		ID:        primitive.NewObjectID(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsActive:  true,
	}

	repo.On("CreateGroup", mock.Anything, group).Return(nil)

	err := repo.CreateGroup(context.Background(), group)
	assert.NoError(t, err)
}

func TestCreateGroup_DBError(t *testing.T) {
	repo := new(MockChatRepository)
	group := &chat.ChatGroup{
		ID: primitive.NewObjectID(),
	}

	repo.On("CreateGroup", mock.Anything, group).Return(errors.New("db error"))

	err := repo.CreateGroup(context.Background(), group)
	assert.Error(t, err)
	assert.EqualError(t, err, "db error")
}

func TestAddMemberToGroup_Success(t *testing.T) {
	repo := new(MockChatRepository)
	groupID := primitive.NewObjectID()
	member := &chat.Member{UserEmail: "test@example.com"}

	repo.On("AddMemberToGroup", mock.Anything, groupID, member).Return(nil)

	err := repo.AddMemberToGroup(context.Background(), groupID, member)
	assert.NoError(t, err)
}

func TestAddMemberToGroup_DBError(t *testing.T) {
	repo := new(MockChatRepository)
	groupID := primitive.NewObjectID()
	member := &chat.Member{UserEmail: "test@example.com"}

	repo.On("AddMemberToGroup", mock.Anything, groupID, member).Return(errors.New("update failed"))

	err := repo.AddMemberToGroup(context.Background(), groupID, member)
	assert.Error(t, err)
	assert.EqualError(t, err, "update failed")
}

func TestCreateMessage_Success(t *testing.T) {
	repo := new(MockChatRepository)
	msg := &chat.Message{
		ID:      primitive.NewObjectID(),
		Content: "Hello",
	}

	repo.On("CreateMessage", mock.Anything, msg).Return(nil)

	err := repo.CreateMessage(context.Background(), msg)
	assert.NoError(t, err)
}

func TestCreateMessage_DBError(t *testing.T) {
	repo := new(MockChatRepository)
	msg := &chat.Message{
		ID:      primitive.NewObjectID(),
		Content: "Hello",
	}

	repo.On("CreateMessage", mock.Anything, msg).Return(errors.New("insert failed"))

	err := repo.CreateMessage(context.Background(), msg)
	assert.Error(t, err)
	assert.EqualError(t, err, "insert failed")
}

func TestIsUserInGroup_True(t *testing.T) {
	repo := new(MockChatRepository)
	groupID := primitive.NewObjectID()
	email := "user@example.com"

	repo.On("IsUserInGroup", mock.Anything, groupID, email).Return(true, nil)

	ok, err := repo.IsUserInGroup(context.Background(), groupID, email)
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestIsUserInGroup_False(t *testing.T) {
	repo := new(MockChatRepository)
	groupID := primitive.NewObjectID()
	email := "user@example.com"

	repo.On("IsUserInGroup", mock.Anything, groupID, email).Return(false, nil)

	ok, err := repo.IsUserInGroup(context.Background(), groupID, email)
	assert.NoError(t, err)
	assert.False(t, ok)
}

// --- Additional Tests ---

func TestCreateGroup_NilGroup(t *testing.T) {
	repo := new(MockChatRepository)
	repo.On("CreateGroup", mock.Anything, (*chat.ChatGroup)(nil)).Return(errors.New("nil group"))

	err := repo.CreateGroup(context.Background(), nil)
	assert.Error(t, err)
	assert.EqualError(t, err, "nil group")
}

func TestAddMemberToGroup_NilMember(t *testing.T) {
	repo := new(MockChatRepository)
	groupID := primitive.NewObjectID()
	repo.On("AddMemberToGroup", mock.Anything, groupID, (*chat.Member)(nil)).Return(errors.New("nil member"))

	err := repo.AddMemberToGroup(context.Background(), groupID, nil)
	assert.Error(t, err)
	assert.EqualError(t, err, "nil member")
}

func TestCreateMessage_NilMessage(t *testing.T) {
	repo := new(MockChatRepository)
	repo.On("CreateMessage", mock.Anything, (*chat.Message)(nil)).Return(errors.New("nil message"))

	err := repo.CreateMessage(context.Background(), nil)
	assert.Error(t, err)
	assert.EqualError(t, err, "nil message")
}

func TestIsUserInGroup_EmptyEmail(t *testing.T) {
	repo := new(MockChatRepository)
	groupID := primitive.NewObjectID()
	email := ""

	repo.On("IsUserInGroup", mock.Anything, groupID, email).Return(false, errors.New("empty email"))

	ok, err := repo.IsUserInGroup(context.Background(), groupID, email)
	assert.Error(t, err)
	assert.False(t, ok)
	assert.EqualError(t, err, "empty email")
}

func TestAddMemberToGroup_ContextCancelled(t *testing.T) {
	repo := new(MockChatRepository)
	groupID := primitive.NewObjectID()
	member := &chat.Member{UserEmail: "cancel@example.com"}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repo.On("AddMemberToGroup", mock.Anything, groupID, member).Return(context.Canceled)

	err := repo.AddMemberToGroup(ctx, groupID, member)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestCreateGroup_ContextDeadlineExceeded(t *testing.T) {
	repo := new(MockChatRepository)
	group := &chat.ChatGroup{ID: primitive.NewObjectID()}
	ctx, cancel := context.WithTimeout(context.Background(), 0)
	cancel()

	repo.On("CreateGroup", mock.Anything, group).Return(context.DeadlineExceeded)

	err := repo.CreateGroup(ctx, group)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestIsUserInGroup_DBError(t *testing.T) {
	repo := new(MockChatRepository)
	groupID := primitive.NewObjectID()
	email := "user@example.com"

	repo.On("IsUserInGroup", mock.Anything, groupID, email).Return(false, errors.New("db error"))

	ok, err := repo.IsUserInGroup(context.Background(), groupID, email)
	assert.Error(t, err)
	assert.False(t, ok)
	assert.EqualError(t, err, "db error")
}

func TestCreateGroup_DuplicateGroup(t *testing.T) {
	repo := new(MockChatRepository)
	group := &chat.ChatGroup{ID: primitive.NewObjectID()}

	repo.On("CreateGroup", mock.Anything, group).Return(errors.New("duplicate group"))

	err := repo.CreateGroup(context.Background(), group)
	assert.Error(t, err)
	assert.EqualError(t, err, "duplicate group")
}

func TestAddMemberToGroup_DuplicateMember(t *testing.T) {
	repo := new(MockChatRepository)
	groupID := primitive.NewObjectID()
	member := &chat.Member{UserEmail: "duplicate@example.com"}

	repo.On("AddMemberToGroup", mock.Anything, groupID, member).Return(errors.New("member already exists"))

	err := repo.AddMemberToGroup(context.Background(), groupID, member)
	assert.Error(t, err)
	assert.EqualError(t, err, "member already exists")
}

func TestCreateMessage_EmptyContent(t *testing.T) {
	repo := new(MockChatRepository)
	msg := &chat.Message{
		ID:      primitive.NewObjectID(),
		Content: "",
	}

	repo.On("CreateMessage", mock.Anything, msg).Return(errors.New("empty content"))

	err := repo.CreateMessage(context.Background(), msg)
	assert.Error(t, err)
	assert.EqualError(t, err, "empty content")
}

func TestIsUserInGroup_NilContext(t *testing.T) {
	repo := new(MockChatRepository)
	groupID := primitive.NewObjectID()
	email := "user@example.com"

	// Simulate nil context (should not panic, but context.Context cannot be nil in practice)
	repo.On("IsUserInGroup", nil, groupID, email).Return(false, errors.New("nil context"))

	ok, err := repo.IsUserInGroup(nil, groupID, email)
	assert.Error(t, err)
	assert.False(t, ok)
	assert.EqualError(t, err, "nil context")
}
