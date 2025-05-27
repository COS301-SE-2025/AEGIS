package unit_tests

import (
	"testing"
	"time"

	"aegis-api/services/evidence"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	//"github.com/stretchr/testify/mock"
)

// ─── MOCK REPO ───────────────────────────────────────────────

// type MockEvidenceRepo struct {
// 	mock.Mock
// }

// func (m *MockEvidenceRepo) FindByID(id uuid.UUID) (*evidence.Evidence, error) {
// 	args := m.Called(id)
// 	return args.Get(0).(*evidence.Evidence), args.Error(1)
// }

// func (m *MockEvidenceRepo) SaveEvidence(e *evidence.Evidence) error                         { return nil }
// func (m *MockEvidenceRepo) AttachTags(e *evidence.Evidence, tags []string) error            { return nil }
// func (m *MockEvidenceRepo) DeleteByID(id uuid.UUID) error                                  { return nil }
// func (m *MockEvidenceRepo) FindByCase(caseID uuid.UUID) ([]evidence.Evidence, error)       { return nil, nil }
// func (m *MockEvidenceRepo) FindByUser(userID uuid.UUID) ([]evidence.Evidence, error)       { return nil, nil }
// func (m *MockEvidenceRepo) PreloadMetadata(id uuid.UUID) (*evidence.Evidence, error)       { return nil, nil }

// ─── DUMMY LOGGER & IPFS ─────────────────────────────────────

// type DummyLogger struct{}
// func (d *DummyLogger) Log(userID, evidenceID, filename string) error { return nil }

// type DummyIPFSClient struct{}
// func (d *DummyIPFSClient) Upload(path string) (string, error) { return "dummyCID", nil }

// ─── TESTS ───────────────────────────────────────────────────

func TestGetEvidenceByID_Success(t *testing.T) {
	mockRepo := new(MockEvidenceRepo)
	mockLogger := new(DummyLogger)
	mockIPFS := new(DummyIPFSClient)

	service := evidence.NewEvidenceService(mockIPFS, mockRepo, mockLogger)

	evidenceID := uuid.New()
	expected := &evidence.Evidence{
		ID:       evidenceID,
		Filename: "found.txt",
		IpfsCID:  "QmFoundCID",
		UploadedAt: time.Now(),
	}

	mockRepo.On("FindByID", evidenceID).Return(expected, nil)

	result, err := service.GetEvidenceByID(evidenceID.String())

	assert.NoError(t, err)
	assert.Equal(t, expected.Filename, result.Filename)
	mockRepo.AssertExpectations(t)
}

func TestGetEvidenceByID_NotFound(t *testing.T) {
	mockRepo := new(MockEvidenceRepo)
	mockLogger := new(DummyLogger)
	mockIPFS := new(DummyIPFSClient)

	service := evidence.NewEvidenceService(mockIPFS, mockRepo, mockLogger)

	nonexistentID := uuid.New()
	mockRepo.On("FindByID", nonexistentID).Return(nil, assert.AnError)

	_, err := service.GetEvidenceByID(nonexistentID.String())
	assert.Error(t, err)
}

func TestGetEvidenceByID_InvalidUUID(t *testing.T) {
	mockRepo := new(MockEvidenceRepo)
	mockLogger := new(DummyLogger)
	mockIPFS := new(DummyIPFSClient)

	service := evidence.NewEvidenceService(mockIPFS, mockRepo, mockLogger)

	_, err := service.GetEvidenceByID("not-a-uuid")
	assert.Error(t, err)
}
