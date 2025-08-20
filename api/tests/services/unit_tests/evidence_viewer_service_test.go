package unit_tests

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"aegis-api/services_/evidence/evidence_viewer"
)

type MockEvidenceViewer struct {
	mock.Mock
}

func (m *MockEvidenceViewer) GetEvidenceFilesByCaseID(caseID string) ([]evidence_viewer.EvidenceFile, error) {
	args := m.Called(caseID)
	return args.Get(0).([]evidence_viewer.EvidenceFile), args.Error(1)
}

func (m *MockEvidenceViewer) GetEvidenceFileByID(fileID string) (*evidence_viewer.EvidenceFile, error) {
	args := m.Called(fileID)
	return args.Get(0).(*evidence_viewer.EvidenceFile), args.Error(1)
}

func (m *MockEvidenceViewer) SearchEvidenceFiles(term string) ([]evidence_viewer.EvidenceFile, error) {
	args := m.Called(term)
	return args.Get(0).([]evidence_viewer.EvidenceFile), args.Error(1)
}

func (m *MockEvidenceViewer) GetFilteredEvidenceFiles(caseID string, filters map[string]interface{}, sortBy string, order string) ([]evidence_viewer.EvidenceFile, error) {
	args := m.Called(caseID, filters, sortBy, order)
	return args.Get(0).([]evidence_viewer.EvidenceFile), args.Error(1)
}

// ────────────────────────────────
// ✅ POSITIVE TESTS
// ────────────────────────────────

func TestService_GetEvidenceFilesByCaseID_Success(t *testing.T) {
	t.Parallel()
	mockRepo := new(MockEvidenceViewer)
	service := &evidence_viewer.EvidenceService{Repo: mockRepo}

	expected := []evidence_viewer.EvidenceFile{
		{ID: "ev123", Data: []byte("file1 bytes")},
		{ID: "ev124", Data: []byte("file2 bytes")},
	}

	mockRepo.On("GetEvidenceFilesByCaseID", "case456").Return(expected, nil)

	files, err := service.GetEvidenceFilesByCaseID("case456")
	assert.NoError(t, err)
	assert.Equal(t, expected, files)

	mockRepo.AssertExpectations(t)
}

func TestService_GetEvidenceFileByID_Success(t *testing.T) {
	t.Parallel()
	mockRepo := new(MockEvidenceViewer)
	service := &evidence_viewer.EvidenceService{Repo: mockRepo}

	expected := &evidence_viewer.EvidenceFile{
		ID:   "ev123",
		Data: []byte("actual file bytes"),
	}

	mockRepo.On("GetEvidenceFileByID", "ev123").Return(expected, nil)

	file, err := service.GetEvidenceFileByID("ev123")
	assert.NoError(t, err)
	assert.Equal(t, expected, file)

	mockRepo.AssertExpectations(t)
}

func TestService_SearchEvidenceFiles_Success(t *testing.T) {
	t.Parallel()
	mockRepo := new(MockEvidenceViewer)
	service := &evidence_viewer.EvidenceService{Repo: mockRepo}

	expected := []evidence_viewer.EvidenceFile{
		{ID: "ev001", Data: []byte("notes pdf bytes")},
	}

	mockRepo.On("SearchEvidenceFiles", "notes").Return(expected, nil)

	files, err := service.SearchEvidenceFiles("notes")
	assert.NoError(t, err)
	assert.Equal(t, expected, files)

	mockRepo.AssertExpectations(t)
}

func TestService_GetFilteredEvidenceFiles_Success(t *testing.T) {
	t.Parallel()
	mockRepo := new(MockEvidenceViewer)
	service := &evidence_viewer.EvidenceService{Repo: mockRepo}

	filters := map[string]interface{}{"file_type": "pdf"}
	expected := []evidence_viewer.EvidenceFile{
		{ID: "ev002", Data: []byte("filtered pdf bytes")},
	}

	mockRepo.On("GetFilteredEvidenceFiles", "case789", filters, "uploaded_at", "desc").Return(expected, nil)

	files, err := service.GetFilteredEvidenceFiles("case789", filters, "uploaded_at", "desc")
	assert.NoError(t, err)
	assert.Equal(t, expected, files)

	mockRepo.AssertExpectations(t)
}

// ────────────────────────────────
// ❌ NEGATIVE TESTS
// ────────────────────────────────

func TestService_GetEvidenceFileByID_Error(t *testing.T) {
	t.Parallel()
	mockRepo := new(MockEvidenceViewer)
	service := &evidence_viewer.EvidenceService{Repo: mockRepo}

	mockRepo.On("GetEvidenceFileByID", "missing123").Return((*evidence_viewer.EvidenceFile)(nil), errors.New("not found"))

	file, err := service.GetEvidenceFileByID("missing123")
	assert.Error(t, err)
	assert.Nil(t, file)
	assert.EqualError(t, err, "not found")

	mockRepo.AssertExpectations(t)
}

func TestService_SearchEvidenceFiles_Error(t *testing.T) {
	t.Parallel()
	mockRepo := new(MockEvidenceViewer)
	service := &evidence_viewer.EvidenceService{Repo: mockRepo}

	mockRepo.On("SearchEvidenceFiles", "nonexistent").
		Return([]evidence_viewer.EvidenceFile(nil), errors.New("search failed"))

	files, err := service.SearchEvidenceFiles("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, files)
	assert.EqualError(t, err, "search failed")

	mockRepo.AssertExpectations(t)
}
