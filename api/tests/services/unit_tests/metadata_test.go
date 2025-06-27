package unit_tests

import (
	"aegis-api/services_/evidence/metadata"
	"encoding/json"
	"errors"
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

		// Unmarshal JSON metadata
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

	req := metadata.UploadEvidenceRequest{
		FilePath: "nonexistent.txt",
	}

	mockIPFS.On("UploadFile", "nonexistent.txt").Return("", errors.New("IPFS error"))

	err := service.UploadEvidence(req)
	assert.Error(t, err)

	mockIPFS.AssertExpectations(t)
}
