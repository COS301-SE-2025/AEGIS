package upload_test

import (
	"aegis-api/services_/evidence/upload"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockIPFSClient struct {
	mock.Mock
}

func (m *MockIPFSClient) UploadFile(path string) (string, error) {
	args := m.Called(path)
	return args.String(0), args.Error(1)
}

func (m *MockIPFSClient) Download(cid string) (io.ReadCloser, error) {
	args := m.Called(cid)
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	return result.(io.ReadCloser), args.Error(1)
}

func TestUploadFile_Success(t *testing.T) {
	mockClient := new(MockIPFSClient)
	svc := upload.NewEvidenceService(mockClient)

	path := "test.txt"
	expectedCID := "Qm123abc"

	mockClient.On("UploadFile", path).Return(expectedCID, nil)

	cid, err := svc.UploadFile(path)

	assert.NoError(t, err)
	assert.Equal(t, expectedCID, cid)
	mockClient.AssertExpectations(t)
}

func TestUploadFile_Error(t *testing.T) {
	mockClient := new(MockIPFSClient)
	svc := upload.NewEvidenceService(mockClient)

	path := "missing.txt"
	mockClient.On("UploadFile", path).Return("", errors.New("file not found"))

	cid, err := svc.UploadFile(path)

	assert.Error(t, err)
	assert.Empty(t, cid)
	mockClient.AssertExpectations(t)
}
