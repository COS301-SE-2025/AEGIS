package unit_tests

import (
	"testing"
	"aegis-api/services/evidence"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ──────────────────────────────
// MockEvidenceRepository (partial)
// ──────────────────────────────

type MockEvidenceRepoDelete struct {
	mock.Mock
}

func (m *MockEvidenceRepoDelete) DeleteByID(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

// Stub required to satisfy interface
func (m *MockEvidenceRepoDelete) SaveEvidence(e *evidence.Evidence) error                         { return nil }
func (m *MockEvidenceRepoDelete) AttachTags(e *evidence.Evidence, tags []string) error            { return nil }
func (m *MockEvidenceRepoDelete) FindByID(id uuid.UUID) (*evidence.Evidence, error)              { return nil, nil }
func (m *MockEvidenceRepoDelete) FindByCase(caseID uuid.UUID) ([]evidence.Evidence, error)       { return nil, nil }
func (m *MockEvidenceRepoDelete) FindByUser(userID uuid.UUID) ([]evidence.Evidence, error)       { return nil, nil }
func (m *MockEvidenceRepoDelete) PreloadMetadata(id uuid.UUID) (*evidence.Evidence, error)       { return nil, nil }

type DummyLogger struct{}
func (d *DummyLogger) Log(userID, evidenceID, filename string) error { return nil }

type DummyIPFSClient struct{}
func (d *DummyIPFSClient) Upload(path string) (string, error) { return "mockCID", nil }

func TestDeleteEvidenceByID_Success(t *testing.T) {
	mockRepo := new(MockEvidenceRepoDelete)
	logger := new(DummyLogger)
	ipfs := new(DummyIPFSClient)
	service := evidence.NewEvidenceService(ipfs, mockRepo, logger)

	id := uuid.New()
	mockRepo.On("DeleteByID", id).Return(nil)

	err := service.DeleteEvidenceByID(id.String())
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteEvidenceByID_NotFound(t *testing.T) {
	mockRepo := new(MockEvidenceRepoDelete)
	logger := new(DummyLogger)
	ipfs := new(DummyIPFSClient)
	service := evidence.NewEvidenceService(ipfs, mockRepo, logger)

	nonexistentID := uuid.New()
	mockRepo.On("DeleteByID", nonexistentID).Return(assert.AnError)

	err := service.DeleteEvidenceByID(nonexistentID.String())
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteEvidenceByID_InvalidUUID(t *testing.T) {
	mockRepo := new(MockEvidenceRepoDelete)
	logger := new(DummyLogger)
	ipfs := new(DummyIPFSClient)
	service := evidence.NewEvidenceService(ipfs, mockRepo, logger)

	err := service.DeleteEvidenceByID("invalid-uuid")
	assert.Error(t, err)
}
