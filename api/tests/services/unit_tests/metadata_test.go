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

func TestUploadEvidence_Success(t *testing.T) {
	mockRepo := new(MockRepo)
	mockIPFS := new(MockIPFS)
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
	mockRepo := new(MockRepo)
	mockIPFS := new(MockIPFS)
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
