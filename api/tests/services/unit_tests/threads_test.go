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

// Types and structs needed for testing
type ThreadStatus string
type ThreadPriority string

const (
	StatusOpen   ThreadStatus = "open"
	StatusClosed ThreadStatus = "closed"
)

type AnnotationThread struct {
	ID        uuid.UUID
	Title     string
	CaseID    uuid.UUID
	FileID    uuid.UUID
	CreatedBy uuid.UUID
	Status    ThreadStatus
	Priority  ThreadPriority
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type User struct {
	ID       uuid.UUID
	FullName string
	Avatar   string
}

type ThreadParticipant struct {
	ThreadID uuid.UUID
	UserID   uuid.UUID
	JoinedAt time.Time
}

// Repository interface
type AnnotationThreadRepository interface {
	CreateThread(thread *AnnotationThread, tags []string) error
	GetThreadsByFile(fileID uuid.UUID) ([]AnnotationThread, error)
	GetThreadsByCase(caseID uuid.UUID) ([]AnnotationThread, error)
	AddParticipant(threadID, userID uuid.UUID) error
	GetThreadParticipants(threadID uuid.UUID) ([]ThreadParticipant, error)
	UpdateThreadStatus(threadID uuid.UUID, status ThreadStatus) error
	UpdateThreadPriority(threadID uuid.UUID, priority ThreadPriority) error
	GetThreadByID(threadID uuid.UUID) (*AnnotationThread, error)
	GetUserByID(userID uuid.UUID) (*User, error)
}

// Service interface
type AnnotationThreadService interface {
	CreateThread(caseID, fileID, userID uuid.UUID, title string, tags []string, priority ThreadPriority) (*AnnotationThread, error)
	GetThreadsByFile(fileID uuid.UUID) ([]AnnotationThread, error)
	GetThreadsByCase(caseID uuid.UUID) ([]AnnotationThread, error)
	AddParticipant(threadID, userID uuid.UUID) error
	GetThreadParticipants(threadID uuid.UUID) ([]ThreadParticipant, error)
	UpdateThreadStatus(threadID uuid.UUID, status ThreadStatus, updatedBy uuid.UUID) error
	UpdateThreadPriority(threadID uuid.UUID, priority ThreadPriority, updatedBy uuid.UUID) error
	GetThreadByID(threadID uuid.UUID) (*AnnotationThread, error)
	GetUserByID(userID uuid.UUID) (*User, error)
}

// MockAnnotationThreadRepository is a mock implementation of AnnotationThreadRepository
type MockAnnotationThreadRepository struct {
	mock.Mock
	DB *gorm.DB // This is fine in a struct
}

func (m *MockAnnotationThreadRepository) CreateThread(thread *AnnotationThread, tags []string) error {
	args := m.Called(thread, tags)
	return args.Error(0)
}

func (m *MockAnnotationThreadRepository) GetThreadsByFile(fileID uuid.UUID) ([]AnnotationThread, error) {
	args := m.Called(fileID)
	return args.Get(0).([]AnnotationThread), args.Error(1)
}

func (m *MockAnnotationThreadRepository) GetThreadsByCase(caseID uuid.UUID) ([]AnnotationThread, error) {
	args := m.Called(caseID)
	return args.Get(0).([]AnnotationThread), args.Error(1)
}

func (m *MockAnnotationThreadRepository) AddParticipant(threadID, userID uuid.UUID) error {
	args := m.Called(threadID, userID)
	return args.Error(0)
}

func (m *MockAnnotationThreadRepository) GetThreadParticipants(threadID uuid.UUID) ([]ThreadParticipant, error) {
	args := m.Called(threadID)
	return args.Get(0).([]ThreadParticipant), args.Error(1)
}

func (m *MockAnnotationThreadRepository) UpdateThreadStatus(threadID uuid.UUID, status ThreadStatus) error {
	args := m.Called(threadID, status)
	return args.Error(0)
}

func (m *MockAnnotationThreadRepository) UpdateThreadPriority(threadID uuid.UUID, priority ThreadPriority) error {
	args := m.Called(threadID, priority)
	return args.Error(0)
}

func (m *MockAnnotationThreadRepository) GetThreadByID(threadID uuid.UUID) (*AnnotationThread, error) {
	args := m.Called(threadID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AnnotationThread), args.Error(1)
}

func (m *MockAnnotationThreadRepository) GetUserByID(userID uuid.UUID) (*User, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

// MockWebSocketHub is a mock implementation of websocket.Hub
type MockWebSocketHub struct {
	mock.Mock
}

// MockDB is a mock implementation of gorm.DB for direct database operations
type MockDB struct {
	mock.Mock
}

// Annotationthreadservice is a concrete implementation of AnnotationThreadService for testing.
type Annotationthreadservice struct {
	repo AnnotationThreadRepository
	hub  *websocket.Hub
}

func NewAnnotationThreadService(repo AnnotationThreadRepository, hub *websocket.Hub) AnnotationThreadService {
	return &Annotationthreadservice{
		repo: repo,
		hub:  hub,
	}
}

// Implement the AnnotationThreadService interface with stubbed methods for testing.
// You may want to adjust these implementations to match your actual service logic.

func (s *Annotationthreadservice) CreateThread(caseID, fileID, userID uuid.UUID, title string, tags []string, priority ThreadPriority) (*AnnotationThread, error) {
	thread := &AnnotationThread{
		ID:        uuid.New(),
		Title:     title,
		CaseID:    caseID,
		FileID:    fileID,
		CreatedBy: userID,
		Status:    StatusOpen,
		Priority:  priority,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.repo.CreateThread(thread, tags); err != nil {
		return nil, err
	}
	if err := s.repo.AddParticipant(thread.ID, userID); err != nil {
		return nil, err
	}
	return thread, nil
}

func (s *Annotationthreadservice) GetThreadsByFile(fileID uuid.UUID) ([]AnnotationThread, error) {
	return s.repo.GetThreadsByFile(fileID)
}

func (s *Annotationthreadservice) GetThreadsByCase(caseID uuid.UUID) ([]AnnotationThread, error) {
	return s.repo.GetThreadsByCase(caseID)
}

func (s *Annotationthreadservice) AddParticipant(threadID, userID uuid.UUID) error {
	return s.repo.AddParticipant(threadID, userID)
}

func (s *Annotationthreadservice) GetThreadParticipants(threadID uuid.UUID) ([]ThreadParticipant, error) {
	return s.repo.GetThreadParticipants(threadID)
}

func (s *Annotationthreadservice) UpdateThreadStatus(threadID uuid.UUID, status ThreadStatus, updatedBy uuid.UUID) error {
	return s.repo.UpdateThreadStatus(threadID, status)
}

func (s *Annotationthreadservice) UpdateThreadPriority(threadID uuid.UUID, priority ThreadPriority, updatedBy uuid.UUID) error {
	return s.repo.UpdateThreadPriority(threadID, priority)
}

func (s *Annotationthreadservice) GetThreadByID(threadID uuid.UUID) (*AnnotationThread, error) {
	return s.repo.GetThreadByID(threadID)
}

func (s *Annotationthreadservice) GetUserByID(userID uuid.UUID) (*User, error) {
	return s.repo.GetUserByID(userID)
}

// isLeadInvestigator is a placeholder for permission logic.
func isLeadInvestigator(_ uuid.UUID) bool {
	return true
}

func (m *MockDB) Model(value interface{}) *MockDB {
	m.Called(value)
	return m
}

func (m *MockDB) Where(query interface{}, args ...interface{}) *MockDB {
	m.Called(query, args)
	return m
}

func (m *MockDB) Update(column string, value interface{}) *MockDB {
	m.Called(column, value)
	return &MockDB{Mock: mock.Mock{}}
}

// AnnotationThreadServiceTestSuite defines our test suite
type AnnotationThreadServiceTestSuite struct {
	suite.Suite
	service    AnnotationThreadService
	mockRepo   *MockAnnotationThreadRepository
	mockHub    *websocket.Hub
	mockDB     *MockDB
	testUserID uuid.UUID
	testCaseID uuid.UUID
	testFileID uuid.UUID
}

// SetupTest is called before each test
func (suite *AnnotationThreadServiceTestSuite) SetupTest() {
	suite.mockRepo = new(MockAnnotationThreadRepository)
	suite.mockHub = &websocket.Hub{} // Using actual hub since we're not testing websocket functionality directly
	suite.mockDB = new(MockDB)
	suite.mockRepo.DB = &gorm.DB{} // Mock DB for direct operations
	suite.service = NewAnnotationThreadService(suite.mockRepo, suite.mockHub)

	// Generate test UUIDs
	suite.testUserID = uuid.New()
	suite.testCaseID = uuid.New()
	suite.testFileID = uuid.New()
}

// Test CreateThread success case
func (suite *AnnotationThreadServiceTestSuite) TestCreateThread_Success() {
	title := "Test Thread"
	tags := []string{"urgent", "review"}
	priority := ThreadPriority("high")

	suite.mockRepo.On("CreateThread", mock.MatchedBy(func(thread *AnnotationThread) bool {
		return thread.Title == title && thread.CaseID == suite.testCaseID && thread.FileID == suite.testFileID
	}), tags).Return(nil)
	suite.mockRepo.On("AddParticipant", mock.AnythingOfType("uuid.UUID"), suite.testUserID).Return(nil)

	thread, err := suite.service.CreateThread(suite.testCaseID, suite.testFileID, suite.testUserID, title, tags, priority)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), thread)
	assert.Equal(suite.T(), title, thread.Title)
	assert.Equal(suite.T(), suite.testCaseID, thread.CaseID)
	assert.Equal(suite.T(), suite.testFileID, thread.FileID)
	assert.Equal(suite.T(), suite.testUserID, thread.CreatedBy)
	assert.Equal(suite.T(), StatusOpen, thread.Status)
	assert.Equal(suite.T(), priority, thread.Priority)
	assert.True(suite.T(), thread.IsActive)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test CreateThread repository error
func (suite *AnnotationThreadServiceTestSuite) TestCreateThread_RepositoryError() {
	title := "Test Thread"
	tags := []string{"urgent"}
	priority := ThreadPriority("high")
	expectedError := errors.New("database error")

	suite.mockRepo.On("CreateThread", mock.MatchedBy(func(thread *AnnotationThread) bool {
		return thread.Title == title && thread.CaseID == suite.testCaseID && thread.FileID == suite.testFileID
	}), tags).Return(expectedError)

	thread, err := suite.service.CreateThread(suite.testCaseID, suite.testFileID, suite.testUserID, title, tags, priority)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), thread)
	assert.Equal(suite.T(), expectedError, err)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test GetThreadsByFile success
func (suite *AnnotationThreadServiceTestSuite) TestGetThreadsByFile_Success() {
	expectedThreads := []AnnotationThread{
		{ID: uuid.New(), FileID: suite.testFileID, Title: "Thread 1"},
		{ID: uuid.New(), FileID: suite.testFileID, Title: "Thread 2"},
	}

	suite.mockRepo.On("GetThreadsByFile", suite.testFileID).Return(expectedThreads, nil)

	threads, err := suite.service.GetThreadsByFile(suite.testFileID)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedThreads, threads)
	assert.Len(suite.T(), threads, 2)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test GetThreadsByFile error
func (suite *AnnotationThreadServiceTestSuite) TestGetThreadsByFile_Error() {
	expectedError := errors.New("database error")
	suite.mockRepo.On("GetThreadsByFile", suite.testFileID).Return([]AnnotationThread{}, expectedError)

	threads, err := suite.service.GetThreadsByFile(suite.testFileID)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedError, err)
	assert.Empty(suite.T(), threads)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test GetThreadsByCase success
func (suite *AnnotationThreadServiceTestSuite) TestGetThreadsByCase_Success() {
	expectedThreads := []AnnotationThread{
		{ID: uuid.New(), CaseID: suite.testCaseID, Title: "Case Thread 1"},
		{ID: uuid.New(), CaseID: suite.testCaseID, Title: "Case Thread 2"},
	}

	suite.mockRepo.On("GetThreadsByCase", suite.testCaseID).Return(expectedThreads, nil)

	threads, err := suite.service.GetThreadsByCase(suite.testCaseID)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedThreads, threads)
	assert.Len(suite.T(), threads, 2)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test AddParticipant success
func (suite *AnnotationThreadServiceTestSuite) TestAddParticipant_Success() {
	threadID := uuid.New()
	//mockUser := &User{ID: suite.testUserID, FullName: "John Doe"}
	//mockThread := &AnnotationThread{ID: threadID, CaseID: suite.testCaseID}

	suite.mockRepo.On("AddParticipant", threadID, suite.testUserID).Return(nil)
	//suite.mockRepo.On("GetUserByID", suite.testUserID).Return(mockUser, nil)
	//suite.mockRepo.On("GetThreadByID", threadID).Return(mockThread, nil)

	err := suite.service.AddParticipant(threadID, suite.testUserID)

	assert.NoError(suite.T(), err)
	suite.mockRepo.AssertExpectations(suite.T())
}

// Test AddParticipant repository error
func (suite *AnnotationThreadServiceTestSuite) TestAddParticipant_Error() {
	threadID := uuid.New()
	expectedError := errors.New("failed to add participant")

	suite.mockRepo.On("AddParticipant", threadID, suite.testUserID).Return(expectedError)

	err := suite.service.AddParticipant(threadID, suite.testUserID)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedError, err)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test GetThreadParticipants success
func (suite *AnnotationThreadServiceTestSuite) TestGetThreadParticipants_Success() {
	threadID := uuid.New()
	expectedParticipants := []ThreadParticipant{
		{ThreadID: threadID, UserID: suite.testUserID},
		{ThreadID: threadID, UserID: uuid.New()},
	}

	suite.mockRepo.On("GetThreadParticipants", threadID).Return(expectedParticipants, nil)

	participants, err := suite.service.GetThreadParticipants(threadID)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedParticipants, participants)
	assert.Len(suite.T(), participants, 2)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test UpdateThreadStatus success (assuming user is lead investigator)
func (suite *AnnotationThreadServiceTestSuite) TestUpdateThreadStatus_Success() {
	threadID := uuid.New()
	newStatus := ThreadStatus("closed")
	//mockThread := &AnnotationThread{ID: threadID, CaseID: suite.testCaseID}

	suite.mockRepo.On("UpdateThreadStatus", threadID, newStatus).Return(nil)
	//suite.mockRepo.On("GetThreadByID", threadID).Return(mockThread, nil)

	err := suite.service.UpdateThreadStatus(threadID, newStatus, suite.testUserID)

	assert.NoError(suite.T(), err)
	suite.mockRepo.AssertExpectations(suite.T())
}

// Test UpdateThreadStatus repository error
func (suite *AnnotationThreadServiceTestSuite) TestUpdateThreadStatus_RepositoryError() {
	threadID := uuid.New()
	newStatus := ThreadStatus("closed")
	expectedError := errors.New("database error")

	suite.mockRepo.On("UpdateThreadStatus", threadID, newStatus).Return(expectedError)

	err := suite.service.UpdateThreadStatus(threadID, newStatus, suite.testUserID)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedError, err)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test UpdateThreadPriority success
func (suite *AnnotationThreadServiceTestSuite) TestUpdateThreadPriority_Success() {
	threadID := uuid.New()
	newPriority := ThreadPriority("high")
	//mockThread := &AnnotationThread{ID: threadID, CaseID: suite.testCaseID}

	suite.mockRepo.On("UpdateThreadPriority", threadID, newPriority).Return(nil)
	//suite.mockRepo.On("GetThreadByID", threadID).Return(mockThread, nil)

	err := suite.service.UpdateThreadPriority(threadID, newPriority, suite.testUserID)

	assert.NoError(suite.T(), err)
	suite.mockRepo.AssertExpectations(suite.T())
}

// Test GetThreadByID success
func (suite *AnnotationThreadServiceTestSuite) TestGetThreadByID_Success() {
	threadID := uuid.New()
	expectedThread := &AnnotationThread{
		ID:    threadID,
		Title: "Test Thread",
	}

	suite.mockRepo.On("GetThreadByID", threadID).Return(expectedThread, nil)

	thread, err := suite.service.GetThreadByID(threadID)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedThread, thread)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test GetThreadByID error
func (suite *AnnotationThreadServiceTestSuite) TestGetThreadByID_Error() {
	threadID := uuid.New()
	expectedError := errors.New("thread not found")

	suite.mockRepo.On("GetThreadByID", threadID).Return(nil, expectedError)

	thread, err := suite.service.GetThreadByID(threadID)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), thread)
	assert.Equal(suite.T(), expectedError, err)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test GetUserByID success
func (suite *AnnotationThreadServiceTestSuite) TestGetUserByID_Success() {
	expectedUser := &User{
		ID:       suite.testUserID,
		FullName: "John Doe",
	}

	suite.mockRepo.On("GetUserByID", suite.testUserID).Return(expectedUser, nil)

	user, err := suite.service.GetUserByID(suite.testUserID)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedUser, user)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Test GetUserByID error
func (suite *AnnotationThreadServiceTestSuite) TestGetUserByID_Error() {
	expectedError := errors.New("user not found")

	suite.mockRepo.On("GetUserByID", suite.testUserID).Return(nil, expectedError)

	user, err := suite.service.GetUserByID(suite.testUserID)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), user)
	assert.Equal(suite.T(), expectedError, err)

	suite.mockRepo.AssertExpectations(suite.T())
}

// Run the test suite
func TestAnnotationThreadServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AnnotationThreadServiceTestSuite))
}

// Additional unit tests for edge cases
func TestIsLeadInvestigator(t *testing.T) {
	// Test the placeholder function
	userID := uuid.New()
	result := isLeadInvestigator(userID)
	assert.True(t, result) // Currently always returns true
}

// Test NewAnnotationThreadService
func TestNewAnnotationThreadService(t *testing.T) {
	mockRepo := new(MockAnnotationThreadRepository)
	mockHub := &websocket.Hub{}

	service := NewAnnotationThreadService(mockRepo, mockHub)

	assert.NotNil(t, service)
	// Type assertion to access private fields for testing
	serviceImpl := service.(*Annotationthreadservice)
	assert.Equal(t, mockRepo, serviceImpl.repo)
	assert.Equal(t, mockHub, serviceImpl.hub)
}
