package unit_tests

import (
	"testing"
	"time"

	"aegis-api/services/evidence"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

//
// ─── MOCK REPOSITORY ──────────────────────────────────────────────
//

type MockEvidenceRepo struct {
	mock.Mock
}

func (m *MockEvidenceRepo) PreloadMetadata(id uuid.UUID) (*evidence.Evidence, error) {
	args := m.Called(id)
	return args.Get(0).(*evidence.Evidence), args.Error(1)
}

// func (m *MockEvidenceRepo) SaveEvidence(e *evidence.Evidence) error                         { return nil }
// func (m *MockEvidenceRepo) AttachTags(e *evidence.Evidence, tags []string) error            { return nil }
// func (m *MockEvidenceRepo) FindByID(id uuid.UUID) (*evidence.Evidence, error)              { return nil, nil }
func (m *MockEvidenceRepo) DeleteByID(id uuid.UUID) error                                  { return nil }
// func (m *MockEvidenceRepo) FindByCase(caseID uuid.UUID) ([]evidence.Evidence, error)       { return nil, nil }

func (m *MockEvidenceRepo) FindByUser(id uuid.UUID) ([]evidence.Evidence, error) {
    args := m.Called(id)
    return args.Get(0).([]evidence.Evidence), args.Error(1)
}

//
// ─── DUMMY LOGGER & IPFS ──────────────────────────────────────────
//

// type DummyLogger struct{}

// func (d *DummyLogger) Log(userID, evidenceID, filename string) error {
// 	return nil
// }

// type DummyIPFSClient struct{}

// func (d *DummyIPFSClient) Upload(path string) (string, error) {
// 	return "mockCID", nil
// }
//
// ─── TEST ─────────────────────────────────────────────────────────
//
func (m *MockEvidenceRepo) AttachTags(e *evidence.Evidence, tags []string) error {
	args := m.Called(e, tags)
	return args.Error(0)
}
func (m *MockEvidenceRepo) FindByCase(caseID uuid.UUID) ([]evidence.Evidence, error) {
	args := m.Called(caseID)
	return args.Get(0).([]evidence.Evidence), args.Error(1)
}
func (m *MockEvidenceRepo) SaveEvidence(e *evidence.Evidence) error {
	args := m.Called(e)
	return args.Error(0)
}

func (m *MockEvidenceRepo) FindByID(id uuid.UUID) (*evidence.Evidence, error) {
	args := m.Called(id)

	// 👇 Check if value is nil before casting
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*evidence.Evidence), args.Error(1)
}



func TestGetEvidenceMetadata(t *testing.T) {
	mockRepo := new(MockEvidenceRepo)
	mockLogger := new(DummyLogger)
	mockIPFS := new(DummyIPFSClient)

	service := evidence.NewEvidenceService(mockIPFS, mockRepo, mockLogger)

	evidenceID := uuid.New()
	expectedEvidence := &evidence.Evidence{
		ID:         evidenceID,
		Filename:   "sample.txt",
		FileType:   "text/plain",
		IpfsCID:    "QmExampleCID",
		FileSize:   123,
		Checksum:   "abc123",
		Metadata:   map[string]interface{}{"source": "mock"},
		Tags:       []evidence.Tag{{Name: "tag1"}, {Name: "tag2"}},
		CaseID:     uuid.New(),
		UploadedBy: uuid.New(),
		UploadedAt: time.Now(),
	}

	mockRepo.On("PreloadMetadata", evidenceID).Return(expectedEvidence, nil)

	meta, err := service.GetEvidenceMetadata(evidenceID.String())

	assert.NoError(t, err)
	assert.Equal(t, expectedEvidence.Filename, meta.Filename)
	assert.Equal(t, expectedEvidence.IpfsCID, meta.IpfsCID)
	assert.Equal(t, expectedEvidence.FileSize, meta.FileSize)
	assert.ElementsMatch(t, []string{"tag1", "tag2"}, meta.Tags)

	mockRepo.AssertExpectations(t)
}
