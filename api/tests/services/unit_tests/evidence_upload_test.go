package unit_tests

import (
	"testing"
	"aegis-api/services/evidence"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ─────────────────────────────────────────────
// MOCKS
// ─────────────────────────────────────────────

type MockIPFSClient struct {
	mock.Mock
}


func (m *MockIPFSClient) Upload(path string) (string, error) {
	args := m.Called(path)
	return args.String(0), args.Error(1)
}

// ✅ ADD THIS METHOD
func (m *MockIPFSClient) Download(cid string) ([]byte, error) {
	args := m.Called(cid)
	return args.Get(0).([]byte), args.Error(1)
}

type MockEvidenceRepository struct {
	mock.Mock
}

func (m *MockEvidenceRepository) SaveEvidence(e *evidence.Evidence) error {
	args := m.Called(e)
	return args.Error(0)
}

func (m *MockEvidenceRepository) AttachTags(e *evidence.Evidence, tags []string) error {
	args := m.Called(e, tags)
	return args.Error(0)
}

func (m *MockEvidenceRepository) FindByID(id uuid.UUID) (*evidence.Evidence, error) {
	args := m.Called(id)

	// 👇 Check if value is nil before casting
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*evidence.Evidence), args.Error(1)
}


func (m *MockEvidenceRepository) DeleteByID(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}


func (m *MockEvidenceRepository) FindByCase(caseID uuid.UUID) ([]evidence.Evidence, error) {
	args := m.Called(caseID)
	return args.Get(0).([]evidence.Evidence), args.Error(1)
}

func (m *MockEvidenceRepository) FindByUser(userID uuid.UUID) ([]evidence.Evidence, error) {
	args := m.Called(userID)
	return args.Get(0).([]evidence.Evidence), args.Error(1)
}

func (m *MockEvidenceRepository) PreloadMetadata(id uuid.UUID) (*evidence.Evidence, error) {
	args := m.Called(id)
	return args.Get(0).(*evidence.Evidence), args.Error(1)
}

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Log(userID, evidenceID, filename string) error {
	args := m.Called(userID, evidenceID, filename)
	return args.Error(0)
}

// type DummyLogger struct{}
// func (d *DummyLogger) Log(userID, evidenceID, filename string) error { return nil }

// type DummyIPFSClient struct{}
// func (d *DummyIPFSClient) Upload(path string) (string, error) { return "mockCID", nil }



// ─────────────────────────────────────────────
// TEST CASE
// ─────────────────────────────────────────────

func TestUploadEvidenceService_Mocked(t *testing.T) {
	mockIPFS := new(MockIPFSClient)
	mockRepo := new(MockEvidenceRepository)
	mockLogger := new(MockLogger)

	service := evidence.NewEvidenceService(mockIPFS, mockRepo, mockLogger)

	req := evidence.UploadEvidenceRequest{
		CaseID:     uuid.New().String(),
		UploadedBy: uuid.New().String(),
		Filename:   "sample.txt",
		FileType:   "text/plain",
		IpfsCID:    "QmFakeCID123",
		FileSize:   42,
		Checksum:   "mockchecksum",
		Metadata:   map[string]interface{}{"source": "test"},
		Tags:       []string{"test", "evidence"},
	}

	// Expectations
	mockRepo.On("SaveEvidence", mock.AnythingOfType("*evidence.Evidence")).Return(nil)
	mockRepo.On("AttachTags", mock.AnythingOfType("*evidence.Evidence"), req.Tags).Return(nil)
	mockLogger.On("Log", mock.Anything, mock.Anything, req.Filename).Return(nil)

	// Execute
	result, err := service.UploadEvidence(req)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, req.Filename, result.Filename)
	assert.Equal(t, req.FileType, result.FileType)

	mockRepo.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}
