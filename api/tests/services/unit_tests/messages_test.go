package unit_tests

import (
	"aegis-api/pkg/websocket"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// Model definitions
type ThreadMessage struct {
	ID              uuid.UUID         `json:"id"`
	ThreadID        uuid.UUID         `json:"thread_id"`
	ParentMessageID *uuid.UUID        `json:"parent_message_id,omitempty"`
	UserID          uuid.UUID         `json:"user_id"`
	Message         string            `json:"message"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	IsApproved      bool              `json:"is_approved"`
	ApprovedBy      *uuid.UUID        `json:"approved_by,omitempty"`
	ApprovedAt      *time.Time        `json:"approved_at,omitempty"`
	Mentions        []MessageMention  `json:"mentions,omitempty"`
	Reactions       []MessageReaction `json:"reactions,omitempty"`
}

type MessageMention struct {
	MessageID       uuid.UUID `json:"message_id"`
	MentionedUserID uuid.UUID `json:"mentioned_user_id"`
	CreatedAt       time.Time `json:"created_at"`
}

type MessageReaction struct {
	ID        uuid.UUID `json:"id"`
	MessageID uuid.UUID `json:"message_id"`
	UserID    uuid.UUID `json:"user_id"`
	Reaction  string    `json:"reaction"`
	CreatedAt time.Time `json:"created_at"`
}

// Repository interface
type MessageRepository interface {
	CreateMessage(msg *ThreadMessage) error
	GetMessagesByThread(threadID uuid.UUID) ([]ThreadMessage, error)
	ApproveMessage(messageID, approverID uuid.UUID) error
	AddReaction(messageID, userID uuid.UUID, reaction string) error
	RemoveReaction(messageID, userID uuid.UUID) error
	GetReplies(parentMessageID uuid.UUID) ([]ThreadMessage, error)
	GetMessageByID(messageID uuid.UUID) (*ThreadMessage, error)
	AddMentions(messageID uuid.UUID, mentions []uuid.UUID) error
}

// Service interface
type MessageService interface {
	SendMessage(threadID, userID uuid.UUID, message string, parentMessageID *uuid.UUID, mentions []uuid.UUID) (*ThreadMessage, error)
	GetMessagesByThread(threadID uuid.UUID) ([]ThreadMessage, error)
	ApproveMessage(messageID, approverID uuid.UUID) error
	AddReaction(messageID, userID uuid.UUID, reaction string) error
	RemoveReaction(messageID, userID uuid.UUID) error
	GetReplies(parentMessageID uuid.UUID) ([]ThreadMessage, error)
	AddMentions(messageID uuid.UUID, mentions []uuid.UUID) error
	GetMessageByID(messageID uuid.UUID) (*ThreadMessage, error)
}

// MockMessageRepository is a mock implementation of MessageRepository
type MockMessageRepository struct {
	mock.Mock
	DB *gorm.DB
}

func (m *MockMessageRepository) CreateMessage(msg *ThreadMessage) error {
	args := m.Called(msg)
	return args.Error(0)
}

func (m *MockMessageRepository) GetMessagesByThread(threadID uuid.UUID) ([]ThreadMessage, error) {
	args := m.Called(threadID)
	return args.Get(0).([]ThreadMessage), args.Error(1)
}

func (m *MockMessageRepository) ApproveMessage(messageID, approverID uuid.UUID) error {
	args := m.Called(messageID, approverID)
	return args.Error(0)
}

func (m *MockMessageRepository) AddReaction(messageID, userID uuid.UUID, reaction string) error {
	args := m.Called(messageID, userID, reaction)
	return args.Error(0)
}

func (m *MockMessageRepository) RemoveReaction(messageID, userID uuid.UUID) error {
	args := m.Called(messageID, userID)
	return args.Error(0)
}

func (m *MockMessageRepository) GetReplies(parentMessageID uuid.UUID) ([]ThreadMessage, error) {
	args := m.Called(parentMessageID)
	return args.Get(0).([]ThreadMessage), args.Error(1)
}

func (m *MockMessageRepository) GetMessageByID(messageID uuid.UUID) (*ThreadMessage, error) {
	args := m.Called(messageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ThreadMessage), args.Error(1)
}

func (m *MockMessageRepository) AddMentions(messageID uuid.UUID, mentions []uuid.UUID) error {
	args := m.Called(messageID, mentions)
	return args.Error(0)
}

// MockDB is a mock implementation of gorm.DB for direct database operations

// func (m *MockDB) Model(value interface{}) *MockDB {
// 	m.Called(value)
// 	return m
// }

// func (m *MockDB) Where(query interface{}, args ...interface{}) *MockDB {
// 	m.Called(query, args)
// 	return m
// }

// func (m *MockDB) Update(column string, value interface{}) *MockDB {
// 	m.Called(column, value)
// 	return &MockDB{Mock: mock.Mock{}}
// }

// MessageServiceImpl is a concrete implementation of MessageService for testing
type MessageServiceImpl struct {
	repo MessageRepository
	hub  *websocket.Hub
}

func NewMessageService(repo MessageRepository, hub *websocket.Hub) MessageService {
	return &MessageServiceImpl{
		repo: repo,
		hub:  hub,
	}
}

func (s *MessageServiceImpl) SendMessage(threadID, userID uuid.UUID, message string, parentMessageID *uuid.UUID, mentions []uuid.UUID) (*ThreadMessage, error) {
	// Validate inputs
	if message == "" {
		return nil, errors.New("message cannot be empty")
	}

	// Create the message model
	msg := &ThreadMessage{
		ID:              uuid.New(),
		ThreadID:        threadID,
		ParentMessageID: parentMessageID,
		UserID:          userID,
		Message:         message,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Save the message to the database
	err := s.repo.CreateMessage(msg)
	if err != nil {
		return nil, err
	}

	// Add mentions separately
	if len(mentions) > 0 {
		err = s.repo.AddMentions(msg.ID, mentions)
		if err != nil {
			return nil, err
		}
	}

	return msg, nil
}

func (s *MessageServiceImpl) GetMessagesByThread(threadID uuid.UUID) ([]ThreadMessage, error) {
	return s.repo.GetMessagesByThread(threadID)
}

func (s *MessageServiceImpl) ApproveMessage(messageID, approverID uuid.UUID) error {
	return s.repo.ApproveMessage(messageID, approverID)
}

func (s *MessageServiceImpl) AddReaction(messageID, userID uuid.UUID, reaction string) error {
	return s.repo.AddReaction(messageID, userID, reaction)
}

func (s *MessageServiceImpl) RemoveReaction(messageID, userID uuid.UUID) error {
	return s.repo.RemoveReaction(messageID, userID)
}

func (s *MessageServiceImpl) GetReplies(parentMessageID uuid.UUID) ([]ThreadMessage, error) {
	return s.repo.GetReplies(parentMessageID)
}

func (s *MessageServiceImpl) AddMentions(messageID uuid.UUID, mentions []uuid.UUID) error {
	if len(mentions) == 0 {
		return nil // No mentions to add
	}

	return s.repo.AddMentions(messageID, mentions)
}

func (s *MessageServiceImpl) GetMessageByID(messageID uuid.UUID) (*ThreadMessage, error) {
	return s.repo.GetMessageByID(messageID)
}

// MessageServiceTestSuite defines our test suite
type MessageServiceTestSuite struct {
	suite.Suite
	service  MessageService
	mockRepo *MockMessageRepository
	mockHub  *websocket.Hub
	//mockDB       *MockDB
	testUserID   uuid.UUID
	testThreadID uuid.UUID
}

// SetupTest is called before each test
func (suite *MessageServiceTestSuite) SetupTest() {
	suite.mockRepo = new(MockMessageRepository)
	suite.mockHub = &websocket.Hub{}
	//suite.mockDB = new(MockDB)
	suite.mockRepo.DB = &gorm.DB{} // Mock DB for direct operations
	suite.service = NewMessageService(suite.mockRepo, suite.mockHub)

	// Generate test UUIDs
	suite.testUserID = uuid.New()
	suite.testThreadID = uuid.New()
}

// Test SendMessage success case
func (suite *MessageServiceTestSuite) TestSendMessage_Success() {
	message := "Hello, World!"
	mentions := []uuid.UUID{uuid.New(), uuid.New()}

	suite.mockRepo.On("CreateMessage", mock.MatchedBy(func(msg *ThreadMessage) bool {
		return msg.Message == message && msg.ThreadID == suite.testThreadID && msg.UserID == suite.testUserID
	})).Return(nil)
	suite.mockRepo.On("AddMentions", mock.AnythingOfType("uuid.UUID"), mentions).Return(nil)

	result, err := suite.service.SendMessage(suite.testThreadID, suite.testUserID, message, nil, mentions)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), message, result.Message)
	assert.Equal(suite.T(), suite.testThreadID, result.ThreadID)
	assert.Equal(suite.T(), suite.testUserID, result.UserID)
	assert.NotEqual(suite.T(), uuid.Nil, result.ID)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test SendMessage empty message failure
func (suite *MessageServiceTestSuite) TestSendMessage_EmptyMessage_Failure() {
	result, err := suite.service.SendMessage(suite.testThreadID, suite.testUserID, "", nil, nil)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), "message cannot be empty", err.Error())

	suite.mockRepo.AssertNotCalled(suite.T(), "CreateMessage")
}

// Test SendMessage repository error
func (suite *MessageServiceTestSuite) TestSendMessage_CreateMessage_Failure() {
	message := "Test message"
	expectedError := errors.New("database error")

	suite.mockRepo.On("CreateMessage", mock.MatchedBy(func(msg *ThreadMessage) bool {
		return msg.Message == message && msg.ThreadID == suite.testThreadID
	})).Return(expectedError)

	result, err := suite.service.SendMessage(suite.testThreadID, suite.testUserID, message, nil, nil)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), expectedError, err)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test SendMessage mentions error
func (suite *MessageServiceTestSuite) TestSendMessage_AddMentions_Failure() {
	message := "Test message"
	mentions := []uuid.UUID{uuid.New()}
	expectedError := errors.New("mentions error")

	suite.mockRepo.On("CreateMessage", mock.AnythingOfType("*unit_tests.ThreadMessage")).Return(nil)
	suite.mockRepo.On("AddMentions", mock.AnythingOfType("uuid.UUID"), mentions).Return(expectedError)

	result, err := suite.service.SendMessage(suite.testThreadID, suite.testUserID, message, nil, mentions)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), expectedError, err)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test GetMessagesByThread success
func (suite *MessageServiceTestSuite) TestGetMessagesByThread_Success() {
	expectedMessages := []ThreadMessage{
		{ID: uuid.New(), ThreadID: suite.testThreadID, Message: "Message 1"},
		{ID: uuid.New(), ThreadID: suite.testThreadID, Message: "Message 2"},
	}

	suite.mockRepo.On("GetMessagesByThread", suite.testThreadID).Return(expectedMessages, nil)

	messages, err := suite.service.GetMessagesByThread(suite.testThreadID)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedMessages, messages)
	assert.Len(suite.T(), messages, 2)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test GetMessagesByThread error
func (suite *MessageServiceTestSuite) TestGetMessagesByThread_Error() {
	expectedError := errors.New("database error")
	suite.mockRepo.On("GetMessagesByThread", suite.testThreadID).Return([]ThreadMessage{}, expectedError)

	messages, err := suite.service.GetMessagesByThread(suite.testThreadID)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedError, err)
	assert.Empty(suite.T(), messages)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test ApproveMessage success
func (suite *MessageServiceTestSuite) TestApproveMessage_Success() {
	messageID := uuid.New()
	approverID := uuid.New()
	// testMessage := &ThreadMessage{
	// 	ID:       messageID,
	// 	ThreadID: suite.testThreadID,
	// 	Message:  "Test message",
	// }

	suite.mockRepo.On("ApproveMessage", messageID, approverID).Return(nil)
	//suite.mockRepo.On("GetMessageByID", messageID).Return(testMessage, nil)

	err := suite.service.ApproveMessage(messageID, approverID)

	assert.NoError(suite.T(), err)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test ApproveMessage failure
func (suite *MessageServiceTestSuite) TestApproveMessage_Failure() {
	messageID := uuid.New()
	approverID := uuid.New()
	expectedError := errors.New("approval failed")

	suite.mockRepo.On("ApproveMessage", messageID, approverID).Return(expectedError)

	err := suite.service.ApproveMessage(messageID, approverID)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedError, err)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test AddReaction success
func (suite *MessageServiceTestSuite) TestAddReaction_Success() {
	messageID := uuid.New()
	reaction := "👍"
	// testMessage := &ThreadMessage{
	// 	ID:       messageID,
	// 	ThreadID: suite.testThreadID,
	// 	Message:  "Test message",
	// }

	suite.mockRepo.On("AddReaction", messageID, suite.testUserID, reaction).Return(nil)
	//suite.mockRepo.On("GetMessageByID", messageID).Return(testMessage, nil)

	err := suite.service.AddReaction(messageID, suite.testUserID, reaction)

	assert.NoError(suite.T(), err)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test AddReaction failure
func (suite *MessageServiceTestSuite) TestAddReaction_Failure() {
	messageID := uuid.New()
	reaction := "👍"
	expectedError := errors.New("reaction failed")

	suite.mockRepo.On("AddReaction", messageID, suite.testUserID, reaction).Return(expectedError)

	err := suite.service.AddReaction(messageID, suite.testUserID, reaction)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedError, err)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test RemoveReaction success
func (suite *MessageServiceTestSuite) TestRemoveReaction_Success() {
	messageID := uuid.New()
	// testMessage := &ThreadMessage{
	// 	ID:       messageID,
	// 	ThreadID: suite.testThreadID,
	// 	Message:  "Test message",
	// }

	suite.mockRepo.On("RemoveReaction", messageID, suite.testUserID).Return(nil)
	//suite.mockRepo.On("GetMessageByID", messageID).Return(testMessage, nil)

	err := suite.service.RemoveReaction(messageID, suite.testUserID)

	assert.NoError(suite.T(), err)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test RemoveReaction failure
func (suite *MessageServiceTestSuite) TestRemoveReaction_Failure() {
	messageID := uuid.New()
	expectedError := errors.New("remove reaction failed")

	suite.mockRepo.On("RemoveReaction", messageID, suite.testUserID).Return(expectedError)

	err := suite.service.RemoveReaction(messageID, suite.testUserID)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedError, err)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test GetReplies success
func (suite *MessageServiceTestSuite) TestGetReplies_Success() {
	parentMessageID := uuid.New()
	expectedReplies := []ThreadMessage{
		{ID: uuid.New(), ParentMessageID: &parentMessageID, Message: "Reply 1"},
		{ID: uuid.New(), ParentMessageID: &parentMessageID, Message: "Reply 2"},
	}

	suite.mockRepo.On("GetReplies", parentMessageID).Return(expectedReplies, nil)

	replies, err := suite.service.GetReplies(parentMessageID)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedReplies, replies)
	assert.Len(suite.T(), replies, 2)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test GetReplies failure
func (suite *MessageServiceTestSuite) TestGetReplies_Failure() {
	parentMessageID := uuid.New()
	expectedError := errors.New("get replies failed")

	suite.mockRepo.On("GetReplies", parentMessageID).Return([]ThreadMessage{}, expectedError)

	replies, err := suite.service.GetReplies(parentMessageID)

	assert.Error(suite.T(), err)
	assert.Empty(suite.T(), replies)
	assert.Equal(suite.T(), expectedError, err)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test AddMentions success
func (suite *MessageServiceTestSuite) TestAddMentions_Success() {
	messageID := uuid.New()
	mentions := []uuid.UUID{uuid.New(), uuid.New()}

	suite.mockRepo.On("AddMentions", messageID, mentions).Return(nil)

	err := suite.service.AddMentions(messageID, mentions)

	assert.NoError(suite.T(), err)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test AddMentions empty mentions
func (suite *MessageServiceTestSuite) TestAddMentions_EmptyMentions_Success() {
	messageID := uuid.New()
	mentions := []uuid.UUID{}

	err := suite.service.AddMentions(messageID, mentions)

	assert.NoError(suite.T(), err)

	suite.mockRepo.AssertNotCalled(suite.T(), "AddMentions")
}

// Test AddMentions failure
func (suite *MessageServiceTestSuite) TestAddMentions_Failure() {
	messageID := uuid.New()
	mentions := []uuid.UUID{uuid.New()}
	expectedError := errors.New("add mentions failed")

	suite.mockRepo.On("AddMentions", messageID, mentions).Return(expectedError)

	err := suite.service.AddMentions(messageID, mentions)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedError, err)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test GetMessageByID success
func (suite *MessageServiceTestSuite) TestGetMessageByID_Success() {
	messageID := uuid.New()
	expectedMessage := &ThreadMessage{
		ID:      messageID,
		Message: "Test message",
	}

	suite.mockRepo.On("GetMessageByID", messageID).Return(expectedMessage, nil)

	message, err := suite.service.GetMessageByID(messageID)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedMessage, message)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test GetMessageByID failure
func (suite *MessageServiceTestSuite) TestGetMessageByID_Failure() {
	messageID := uuid.New()
	expectedError := errors.New("message not found")

	suite.mockRepo.On("GetMessageByID", messageID).Return(nil, expectedError)

	message, err := suite.service.GetMessageByID(messageID)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), message)
	assert.Equal(suite.T(), expectedError, err)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Run the test suite
func TestMessageServiceTestSuite(t *testing.T) {
	suite.Run(t, new(MessageServiceTestSuite))
}

// Test NewMessageService
func TestNewMessageService(t *testing.T) {
	mockRepo := new(MockMessageRepository)
	mockHub := &websocket.Hub{}

	service := NewMessageService(mockRepo, mockHub)

	assert.NotNil(t, service)
	// Type assertion to access private fields for testing
	serviceImpl := service.(*MessageServiceImpl)
	assert.Equal(t, mockRepo, serviceImpl.repo)
	assert.Equal(t, mockHub, serviceImpl.hub)
}
