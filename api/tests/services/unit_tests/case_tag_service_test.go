package unit_tests

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	services "aegis-api/services_/case_tags"
	
)

type MockCaseTagRepository struct {
	mock.Mock
}

func (m *MockCaseTagRepository) AddTagsToCase(ctx context.Context, userID, caseID uuid.UUID, tags []string) error {
	args := m.Called(ctx, userID, caseID, tags)
	return args.Error(0)
}

func (m *MockCaseTagRepository) RemoveTagsFromCase(ctx context.Context, userID, caseID uuid.UUID, tags []string) error {
	args := m.Called(ctx, userID, caseID, tags)
	return args.Error(0)
}

func (m *MockCaseTagRepository) GetTagsForCase(ctx context.Context, caseID uuid.UUID) ([]string, error) {
	args := m.Called(ctx, caseID)
	return args.Get(0).([]string), args.Error(1)
}


func TestTagCase(t *testing.T) {
	mockRepo := new(MockCaseTagRepository)
	svc := services.NewCaseTagService(mockRepo)

	userID := uuid.New()
	caseID := uuid.New()
	tags := []string{"urgent", "legal"}

	mockRepo.
		On("AddTagsToCase", mock.Anything, userID, caseID, tags).
		Return(nil)

	err := svc.TagCase(context.Background(), userID, caseID, tags)
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestUntagCase(t *testing.T) {
	mockRepo := new(MockCaseTagRepository)
	svc := services.NewCaseTagService(mockRepo)

	userID := uuid.New()
	caseID := uuid.New()
	tags := []string{"duplicate"}

	mockRepo.
		On("RemoveTagsFromCase", mock.Anything, userID, caseID, tags).
		Return(nil)

	err := svc.UntagCase(context.Background(), userID, caseID, tags)
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}
