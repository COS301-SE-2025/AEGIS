package unit_tests

import (
	"aegis-api/services_/evidence/metadata"
	"encoding/json"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Minimal IPFS mock for error test
type MockIPFS_Struct struct {
	mock.Mock
}

func (m *MockIPFS_Struct) UploadFile(r io.Reader) (string, error) {
	args := m.Called(r)
	return args.String(0), args.Error(1)
}

// Implement the Download method to satisfy upload.IPFSClientImp interface.
func (m *MockIPFS_Struct) Download(cid string) (io.ReadCloser, error) {
	args := m.Called(cid)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

type MockEvidenceRepo struct {
	mock.Mock
}

func (m *MockEvidenceRepo) SaveEvidence(e *metadata.Evidence) error {
	args := m.Called(e)
	return args.Error(0)
}

// Implement FindEvidenceByCaseID to satisfy metadata.Repository interface.
func (m *MockEvidenceRepo) FindEvidenceByCaseID(caseID uuid.UUID) ([]metadata.Evidence, error) {
	args := m.Called(caseID)
	return args.Get(0).([]metadata.Evidence), args.Error(1)
}

func (m *MockEvidenceRepo) FindEvidenceByID(id uuid.UUID) (*metadata.Evidence, error) {
	args := m.Called(id)
	return args.Get(0).(*metadata.Evidence), args.Error(1)
}

// Implement AppendEvidenceLog to satisfy metadata.Repository interface.
func (m *MockEvidenceRepo) AppendEvidenceLog(log *metadata.EvidenceLog) error {
	args := m.Called(log)
	return args.Error(0)
}

// Use the MockIPFS defined in evidence_download_test.go, or rename this struct if both are needed.
// For example, rename to MockIPFSUploader if you need both mocks:

type MockIPFSUploaderMetadata struct {
	mock.Mock
}

func (m *MockIPFSUploaderMetadata) UploadFile(r io.Reader) (string, error) {
	args := m.Called(r)
	return args.String(0), args.Error(1)
}

// Implement the Download method to satisfy upload.IPFSClientImp interface.
func (m *MockIPFSUploaderMetadata) Download(cid string) (io.ReadCloser, error) {
	args := m.Called(cid)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}
func TestUploadEvidence_Success(t *testing.T) {
	mockRepo := new(MockEvidenceRepo)
	mockIPFS := new(MockIPFSUploaderMetadata)
	service := metadata.NewService(mockRepo, mockIPFS)

	content := "Hello, AEGIS!"
	tmpFile, err := os.CreateTemp("", "testfile-*.txt")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	_, err = tmpFile.Write([]byte(content))
	assert.NoError(t, err)

	// Reopen the file for reading
	fileReader, err := os.Open(tmpFile.Name())
	assert.NoError(t, err)
	defer fileReader.Close()

	req := metadata.UploadEvidenceRequest{
		CaseID:     uuid.New(),
		UploadedBy: uuid.New(),
		Filename:   "testfile.txt",
		FileType:   "text/plain",
		FileSize:   int64(len(content)),
		FileData:   fileReader,
		Metadata: map[string]string{
			"source": "user-upload",
		},
	}

	mockIPFS.On("UploadFile", mock.MatchedBy(func(r interface{}) bool {
		_, ok := r.(io.Reader)
		return ok
	})).Return("Qm123", nil)

	// Add expectation for AppendEvidenceLog to prevent panic
	mockRepo.On("AppendEvidenceLog", mock.AnythingOfType("*metadata.EvidenceLog")).Return(nil)

	mockRepo.On("SaveEvidence", mock.AnythingOfType("*metadata.Evidence")).Return(nil).Run(func(args mock.Arguments) {
		e := args.Get(0).(*metadata.Evidence)
		assert.Equal(t, req.Filename, e.Filename)
		assert.Equal(t, req.FileType, e.FileType)
		assert.Equal(t, "Qm123", e.IpfsCID)

		var meta map[string]string
		err := json.Unmarshal([]byte(e.Metadata), &meta)
		assert.NoError(t, err)
		assert.Equal(t, req.Metadata["source"], meta["source"])
	})

	err = service.UploadEvidence(req)
	assert.NoError(t, err)

	mockIPFS.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestUploadEvidence_IPFSError(t *testing.T) {
	mockRepo := new(MockEvidenceRepo)
	mockIPFS := new(MockIPFS_Struct)
	service := metadata.NewService(mockRepo, mockIPFS)

	fileReader := os.NewFile(0, os.DevNull) // dummy reader (non-readable for IPFS)
	defer fileReader.Close()

	req := metadata.UploadEvidenceRequest{
		CaseID:     uuid.New(),
		UploadedBy: uuid.New(),
		Filename:   "fail.txt",
		FileType:   "text/plain",
		FileSize:   100,
		FileData:   fileReader,
		Metadata:   map[string]string{},
	}

	mockIPFS.On("UploadFile", mock.MatchedBy(func(r io.Reader) bool {
		return r != nil
	})).Return("", errors.New("IPFS error"))

	err := service.UploadEvidence(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "IPFS upload failed")

	mockIPFS.AssertExpectations(t)
}
