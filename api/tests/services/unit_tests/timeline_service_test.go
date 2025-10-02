package unit_tests

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/datatypes"

	"aegis-api/services_/timeline"
)

// --- Mock definition ---

type MockTimelineRepository struct {
	mock.Mock
}

func (m *MockTimelineRepository) AutoMigrate() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTimelineRepository) Create(event *timeline.TimelineEvent) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockTimelineRepository) GetByID(id string) (*timeline.TimelineEvent, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*timeline.TimelineEvent), args.Error(1)
}

func (m *MockTimelineRepository) ListByCase(caseID string) ([]*timeline.TimelineEvent, error) {
	args := m.Called(caseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*timeline.TimelineEvent), args.Error(1)
}

func (m *MockTimelineRepository) Update(event *timeline.TimelineEvent) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockTimelineRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTimelineRepository) UpdateOrder(caseID string, orderedIDs []string) error {
	args := m.Called(caseID, orderedIDs)
	return args.Error(0)
}

func (m *MockTimelineRepository) FindByID(eventID string) (*timeline.TimelineEvent, error) {
	args := m.Called(eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*timeline.TimelineEvent), args.Error(1)
}

// --- Test helpers ---

func createTestEvent(id string) *timeline.TimelineEvent {
	now := time.Now()
	return &timeline.TimelineEvent{
		ID:          id,
		CaseID:      "case-123",
		Description: "Test event",
		Severity:    "High",
		AnalystName: "John Doe",
		CreatedAt:   now,
		UpdatedAt:   now,
		Evidence:    datatypes.JSON([]byte(`["evidence1","evidence2"]`)),
		Tags:        datatypes.JSON([]byte(`["tag1","tag2"]`)),
	}
}

// --- Unit Tests ---

func TestTimelineService_AddEvent_Success(t *testing.T) {
	mockRepo := new(MockTimelineRepository)
	service := timeline.NewService(mockRepo)

	event := createTestEvent("event-1")

	// Setup mock expectations
	mockRepo.On("Create", event).Return(nil)

	// Call the service method
	result, err := service.AddEvent(event)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, event.ID, result.ID)

	// Verify expectations
	mockRepo.AssertExpectations(t)
}

func TestTimelineService_AddEvent_NilEvent(t *testing.T) {
	mockRepo := new(MockTimelineRepository)
	service := timeline.NewService(mockRepo)

	// Call the service method
	result, err := service.AddEvent(nil)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, timeline.ErrInvalidEvent, err)

	// Verify expectations
	mockRepo.AssertNotCalled(t, "Create")
}

func TestTimelineService_AddEvent_EmptyEvidenceAndTags(t *testing.T) {
	mockRepo := new(MockTimelineRepository)
	service := timeline.NewService(mockRepo)

	event := &timeline.TimelineEvent{
		ID:          "event-1",
		CaseID:      "case-123",
		Description: "Test",
		Evidence:    datatypes.JSON([]byte{}),
		Tags:        datatypes.JSON([]byte{}),
	}

	// Setup mock expectations
	mockRepo.On("Create", mock.MatchedBy(func(e *timeline.TimelineEvent) bool {
		return string(e.Evidence) == "[]" && string(e.Tags) == "[]"
	})).Return(nil)

	// Call the service method
	result, err := service.AddEvent(event)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "[]", string(result.Evidence))
	assert.Equal(t, "[]", string(result.Tags))

	// Verify expectations
	mockRepo.AssertExpectations(t)
}

func TestTimelineService_AddEvent_RepositoryError(t *testing.T) {
	mockRepo := new(MockTimelineRepository)
	service := timeline.NewService(mockRepo)

	event := createTestEvent("event-1")
	repoErr := errors.New("database connection failed")

	// Setup mock expectations
	mockRepo.On("Create", event).Return(repoErr)

	// Call the service method
	result, err := service.AddEvent(event)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, repoErr, err)

	// Verify expectations
	mockRepo.AssertExpectations(t)
}

func TestTimelineService_GetEvent_Success(t *testing.T) {
	mockRepo := new(MockTimelineRepository)
	service := timeline.NewService(mockRepo)

	expectedEvent := createTestEvent("event-1")

	// Setup mock expectations
	mockRepo.On("GetByID", "event-1").Return(expectedEvent, nil)

	// Call the service method
	result, err := service.GetEvent("event-1")

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedEvent.ID, result.ID)

	// Verify expectations
	mockRepo.AssertExpectations(t)
}

func TestTimelineService_GetEvent_NotFound(t *testing.T) {
	mockRepo := new(MockTimelineRepository)
	service := timeline.NewService(mockRepo)

	notFoundErr := errors.New("event not found")

	// Setup mock expectations
	mockRepo.On("GetByID", "nonexistent").Return(nil, notFoundErr)

	// Call the service method
	result, err := service.GetEvent("nonexistent")

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, notFoundErr, err)

	// Verify expectations
	mockRepo.AssertExpectations(t)
}

func TestTimelineService_ListEvents_Success(t *testing.T) {
	mockRepo := new(MockTimelineRepository)
	service := timeline.NewService(mockRepo)

	now := time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)
	events := []*timeline.TimelineEvent{
		{
			ID:          "event-1",
			CaseID:      "case-123",
			Description: "First event",
			Severity:    "High",
			AnalystName: "Alice",
			CreatedAt:   now,
			Evidence:    datatypes.JSON([]byte(`["evidence1"]`)),
			Tags:        datatypes.JSON([]byte(`["tag1"]`)),
		},
		{
			ID:          "event-2",
			CaseID:      "case-123",
			Description: "Second event",
			Severity:    "Medium",
			AnalystName: "Bob",
			CreatedAt:   now.Add(1 * time.Hour),
			Evidence:    datatypes.JSON([]byte(`["evidence2"]`)),
			Tags:        datatypes.JSON([]byte(`["tag2"]`)),
		},
	}

	// Setup mock expectations
	mockRepo.On("ListByCase", "case-123").Return(events, nil)

	// Call the service method
	result, err := service.ListEvents("case-123")

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.Equal(t, "event-1", result[0].ID)
	assert.Equal(t, "2024-01-15", result[0].Date)
	assert.Equal(t, "14:30", result[0].Time)
	assert.Equal(t, "High", result[0].Severity)

	// Verify expectations
	mockRepo.AssertExpectations(t)
}

// func TestTimelineService_ListEvents_EmptyList(t *testing.T) {
// 	mockRepo := new(MockTimelineRepository)
// 	service := timeline.NewService(mockRepo)

// 	// Setup mock expectations
// 	mockRepo.On("ListByCase", "case-456").Return([]*timeline.TimelineEvent{}, nil)

// 	// Call the service method
// 	result, err := service.ListEvents("case-456")

// 	// Assertions
// 	assert.NoError(t, err)
// 	assert.NotNil(t, result)
// 	assert.Empty(t, result)

// 	// Verify expectations
// 	mockRepo.AssertExpectations(t)
// }

func TestTimelineService_ListEvents_RepositoryError(t *testing.T) {
	mockRepo := new(MockTimelineRepository)
	service := timeline.NewService(mockRepo)

	repoErr := errors.New("database error")

	// Setup mock expectations
	mockRepo.On("ListByCase", "case-123").Return(nil, repoErr)

	// Call the service method
	result, err := service.ListEvents("case-123")

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, repoErr, err)

	// Verify expectations
	mockRepo.AssertExpectations(t)
}

func TestTimelineService_UpdateEvent_Success(t *testing.T) {
	mockRepo := new(MockTimelineRepository)
	service := timeline.NewService(mockRepo)

	event := createTestEvent("event-1")
	event.Description = "Updated description"

	// Setup mock expectations
	mockRepo.On("Update", event).Return(nil)
	mockRepo.On("GetByID", "event-1").Return(event, nil)

	// Call the service method
	result, err := service.UpdateEvent(event)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Updated description", result.Description)

	// Verify expectations
	mockRepo.AssertExpectations(t)
}

func TestTimelineService_UpdateEvent_NilEvent(t *testing.T) {
	mockRepo := new(MockTimelineRepository)
	service := timeline.NewService(mockRepo)

	// Call the service method
	result, err := service.UpdateEvent(nil)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, timeline.ErrInvalidEvent, err)

	// Verify expectations
	mockRepo.AssertNotCalled(t, "Update")
}

func TestTimelineService_UpdateEvent_EmptyID(t *testing.T) {
	mockRepo := new(MockTimelineRepository)
	service := timeline.NewService(mockRepo)

	event := createTestEvent("")

	// Call the service method
	result, err := service.UpdateEvent(event)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, timeline.ErrInvalidEvent, err)

	// Verify expectations
	mockRepo.AssertNotCalled(t, "Update")
}

func TestTimelineService_UpdateEvent_RepositoryError(t *testing.T) {
	mockRepo := new(MockTimelineRepository)
	service := timeline.NewService(mockRepo)

	event := createTestEvent("event-1")
	repoErr := errors.New("update failed")

	// Setup mock expectations
	mockRepo.On("Update", event).Return(repoErr)

	// Call the service method
	result, err := service.UpdateEvent(event)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, repoErr, err)

	// Verify expectations
	mockRepo.AssertExpectations(t)
}

func TestTimelineService_DeleteEvent_Success(t *testing.T) {
	mockRepo := new(MockTimelineRepository)
	service := timeline.NewService(mockRepo)

	// Setup mock expectations
	mockRepo.On("Delete", "event-1").Return(nil)

	// Call the service method
	err := service.DeleteEvent("event-1")

	// Assertions
	assert.NoError(t, err)

	// Verify expectations
	mockRepo.AssertExpectations(t)
}

func TestTimelineService_DeleteEvent_RepositoryError(t *testing.T) {
	mockRepo := new(MockTimelineRepository)
	service := timeline.NewService(mockRepo)

	repoErr := errors.New("delete failed")

	// Setup mock expectations
	mockRepo.On("Delete", "event-1").Return(repoErr)

	// Call the service method
	err := service.DeleteEvent("event-1")

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, repoErr, err)

	// Verify expectations
	mockRepo.AssertExpectations(t)
}

func TestTimelineService_ReorderEvents_Success(t *testing.T) {
	mockRepo := new(MockTimelineRepository)
	service := timeline.NewService(mockRepo)

	orderedIDs := []string{"event-3", "event-1", "event-2"}

	// Setup mock expectations
	mockRepo.On("UpdateOrder", "case-123", orderedIDs).Return(nil)

	// Call the service method
	err := service.ReorderEvents("case-123", orderedIDs)

	// Assertions
	assert.NoError(t, err)

	// Verify expectations
	mockRepo.AssertExpectations(t)
}

func TestTimelineService_ReorderEvents_EmptyList(t *testing.T) {
	mockRepo := new(MockTimelineRepository)
	service := timeline.NewService(mockRepo)

	orderedIDs := []string{}

	// Setup mock expectations
	mockRepo.On("UpdateOrder", "case-123", orderedIDs).Return(nil)

	// Call the service method
	err := service.ReorderEvents("case-123", orderedIDs)

	// Assertions
	assert.NoError(t, err)

	// Verify expectations
	mockRepo.AssertExpectations(t)
}

func TestTimelineService_ReorderEvents_RepositoryError(t *testing.T) {
	mockRepo := new(MockTimelineRepository)
	service := timeline.NewService(mockRepo)

	orderedIDs := []string{"event-1", "event-2"}
	repoErr := errors.New("reorder failed")

	// Setup mock expectations
	mockRepo.On("UpdateOrder", "case-123", orderedIDs).Return(repoErr)

	// Call the service method
	err := service.ReorderEvents("case-123", orderedIDs)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, repoErr, err)

	// Verify expectations
	mockRepo.AssertExpectations(t)
}

func TestTimelineService_GetEventByID_Success(t *testing.T) {
	mockRepo := new(MockTimelineRepository)
	service := timeline.NewService(mockRepo)

	expectedEvent := createTestEvent("event-1")

	// Setup mock expectations
	mockRepo.On("FindByID", "event-1").Return(expectedEvent, nil)

	// Call the service method
	result, err := service.GetEventByID("event-1")

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedEvent.ID, result.ID)

	// Verify expectations
	mockRepo.AssertExpectations(t)
}

func TestTimelineService_GetEventByID_NotFound(t *testing.T) {
	mockRepo := new(MockTimelineRepository)
	service := timeline.NewService(mockRepo)

	notFoundErr := errors.New("event not found")

	// Setup mock expectations
	mockRepo.On("FindByID", "nonexistent").Return(nil, notFoundErr)

	// Call the service method
	result, err := service.GetEventByID("nonexistent")

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, notFoundErr, err)

	// Verify expectations
	mockRepo.AssertExpectations(t)
}
