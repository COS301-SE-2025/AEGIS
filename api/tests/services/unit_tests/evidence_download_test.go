// evidence_download_test.go
package unit_tests

import (
	"aegis-api/services_/evidence/evidence_download"
	"aegis-api/services_/evidence/metadata"
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) SaveEvidence(e *metadata.Evidence) error {
	args := m.Called(e)
	return args.Error(0)
}

func (m *MockRepo) FindEvidenceByID(id uuid.UUID) (*metadata.Evidence, error) {
	args := m.Called(id)
	if ev := args.Get(0); ev != nil {
		return ev.(*metadata.Evidence), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRepo) FindEvidenceByCaseID(caseID uuid.UUID) ([]metadata.Evidence, error) {
	args := m.Called(caseID)
	if evs := args.Get(0); evs != nil {
		return evs.([]metadata.Evidence), args.Error(1)
	}
	return nil, args.Error(1)
}

type MockIPFS struct {
	mock.Mock
}

func (m *MockIPFS) UploadFile(reader io.Reader) (string, error) {
	args := m.Called(reader)
	return args.String(0), args.Error(1)
}

func (m *MockIPFS) Download(cid string) (io.ReadCloser, error) {
	args := m.Called(cid)
	if reader := args.Get(0); reader != nil {
		return reader.(io.ReadCloser), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestDownloadEvidence_Success(t *testing.T) {
	mockRepo := new(MockRepo)
	mockIPFS := new(MockIPFS)
	service := evidence_download.NewService(mockRepo, mockIPFS)

	evidenceID := uuid.New()
	evidence := &metadata.Evidence{
		ID:       evidenceID,
		Filename: "sample.txt",
		FileType: "text/plain",
		IpfsCID:  "Qm123abc",
	}

	mockRepo.On("FindEvidenceByID", evidenceID).Return(evidence, nil)

	expectedContent := "file content"
	mockStream := io.NopCloser(bytes.NewBufferString(expectedContent))
	mockIPFS.On("Download", "Qm123abc").Return(mockStream, nil)

	filename, reader, filetype, err := service.DownloadEvidence(evidenceID)

	assert.NoError(t, err)
	assert.Equal(t, "sample.txt", filename)
	assert.Equal(t, "text/plain", filetype)

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, reader)
	assert.NoError(t, err)
	assert.Equal(t, expectedContent, buf.String())

	mockRepo.AssertExpectations(t)
	mockIPFS.AssertExpectations(t)
}

func TestDownloadEvidence_RepoError(t *testing.T) {
	mockRepo := new(MockRepo)
	mockIPFS := new(MockIPFS)
	service := evidence_download.NewService(mockRepo, mockIPFS)

	evidenceID := uuid.New()
	mockRepo.On("FindEvidenceByID", evidenceID).Return(nil, errors.New("not found"))

	filename, reader, filetype, err := service.DownloadEvidence(evidenceID)

	assert.Error(t, err)
	assert.Empty(t, filename)
	assert.Nil(t, reader)
	assert.Empty(t, filetype)
	mockRepo.AssertExpectations(t)
}

func TestDownloadEvidence_IPFSError(t *testing.T) {
	mockRepo := new(MockRepo)
	mockIPFS := new(MockIPFS)
	service := evidence_download.NewService(mockRepo, mockIPFS)

	evidenceID := uuid.New()
	evidence := &metadata.Evidence{
		ID:       evidenceID,
		Filename: "sample.txt",
		FileType: "text/plain",
		IpfsCID:  "Qm123abc",
	}
	mockRepo.On("FindEvidenceByID", evidenceID).Return(evidence, nil)
	mockIPFS.On("Download", "Qm123abc").Return(nil, errors.New("IPFS failure"))

	filename, reader, filetype, err := service.DownloadEvidence(evidenceID)

	assert.Error(t, err)
	assert.Empty(t, filename)
	assert.Nil(t, reader)
	assert.Empty(t, filetype)

	mockRepo.AssertExpectations(t)
	mockIPFS.AssertExpectations(t)
}
