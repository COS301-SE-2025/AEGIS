package metadata_test

import (
	"aegis-api/services_/evidence/metadata"
	"bytes"
	"errors"
	"io"
	"os"
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

type MockIPFS struct {
	mock.Mock
}

func (m *MockIPFS) UploadFile(path string) (string, error) {
	args := m.Called(path)
	return args.String(0), args.Error(1)
}

func (m *MockIPFS) Download(cid string) (io.ReadCloser, error) {
	args := m.Called(cid)
	if r := args.Get(0); r != nil {
		return r.(io.ReadCloser), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestUploadEvidence_Success(t *testing.T) {
	mockRepo := new(MockRepo)
	mockIPFS := new(MockIPFS)
	service := metadata.NewService(mockRepo, mockIPFS)

	fakePath := "temp_testfile.txt"
	fakeCID := "Qm123"
	content := "Hello, AEGIS!"

	// 1. Write temporary file
	err := os.WriteFile(fakePath, []byte(content), 0644)
	assert.NoError(t, err)
	defer os.Remove(fakePath)

	req := metadata.UploadEvidenceRequest{
		CaseID:     uuid.New(),
		UploadedBy: uuid.New(),
		Filename:   "testfile.txt",
		FileType:   "text/plain",
		FilePath:   fakePath,
		FileSize:   int64(len(content)),
		Metadata: map[string]string{
			"source": "user-upload",
		},
	}

	mockIPFS.On("UploadFile", fakePath).Return(fakeCID, nil)

	mockRepo.On("SaveEvidence", mock.AnythingOfType("*metadata.Evidence")).Return(nil).Run(func(args mock.Arguments) {
		e := args.Get(0).(*metadata.Evidence)
		assert.Equal(t, req.Filename, e.Filename)
		assert.Equal(t, req.FileType, e.FileType)
		assert.Equal(t, fakeCID, e.IpfsCID)
		assert.Equal(t, req.Metadata["source"], e.Metadata["source"])
	})

	err = service.UploadEvidence(req)
	assert.NoError(t, err)

	mockIPFS.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestUploadEvidence_IPFSError(t *testing.T) {
	mockRepo := new(MockRepo)
	mockIPFS := new(MockIPFS)
	service := metadata.NewService(mockRepo, mockIPFS)

	req := metadata.UploadEvidenceRequest{
		FilePath: "nonexistent.txt",
	}

	mockIPFS.On("UploadFile", "nonexistent.txt").Return("", errors.New("IPFS error"))

	err := service.UploadEvidence(req)
	assert.Error(t, err)

	mockIPFS.AssertExpectations(t)
}

func TestDownloadEvidence_Success(t *testing.T) {
	mockRepo := new(MockRepo)
	mockIPFS := new(MockIPFS)
	service := metadata.NewService(mockRepo, mockIPFS)

	id := uuid.New()
	ev := &metadata.Evidence{
		ID:       id,
		Filename: "result.txt",
		FileType: "text/plain",
		IpfsCID:  "Qm456",
	}
	mockRepo.On("FindEvidenceByID", id).Return(ev, nil)

	content := "Test download"
	stream := io.NopCloser(bytes.NewBufferString(content))
	mockIPFS.On("Download", "Qm456").Return(stream, nil)

	filename, filetype, reader, err := service.DownloadEvidence(id)

	assert.NoError(t, err)
	assert.Equal(t, ev.Filename, filename)
	assert.Equal(t, ev.FileType, filetype)

	buf := new(bytes.Buffer)
	_, _ = io.Copy(buf, reader)
	assert.Equal(t, content, buf.String())

	mockRepo.AssertExpectations(t)
	mockIPFS.AssertExpectations(t)
}

func TestDownloadEvidence_RepoError(t *testing.T) {
	mockRepo := new(MockRepo)
	mockIPFS := new(MockIPFS)
	service := metadata.NewService(mockRepo, mockIPFS)

	id := uuid.New()
	mockRepo.On("FindEvidenceByID", id).Return(nil, errors.New("not found"))

	_, _, _, err := service.DownloadEvidence(id)
	assert.Error(t, err)

	mockRepo.AssertExpectations(t)
}

func TestDownloadEvidence_IPFSError(t *testing.T) {
	mockRepo := new(MockRepo)
	mockIPFS := new(MockIPFS)
	service := metadata.NewService(mockRepo, mockIPFS)

	id := uuid.New()
	ev := &metadata.Evidence{ID: id, IpfsCID: "QmErr", Filename: "bad", FileType: "text"}
	mockRepo.On("FindEvidenceByID", id).Return(ev, nil)
	mockIPFS.On("Download", "QmErr").Return(nil, errors.New("IPFS down"))

	_, _, _, err := service.DownloadEvidence(id)
	assert.Error(t, err)

	mockRepo.AssertExpectations(t)
	mockIPFS.AssertExpectations(t)
}
