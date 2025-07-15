package unit_tests

import (
	"aegis-api/services_/evidence/upload"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockIPFSClient struct {
	mock.Mock
}

func (m *MockIPFSClient) UploadFile(file io.Reader) (string, error) {
	args := m.Called(file)
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
	t.Parallel()

	mockClient := new(MockIPFSClient)
	svc := upload.NewEvidenceService(mockClient)

	fileContent := "hello world"
	reader := strings.NewReader(fileContent)
	expectedCID := "Qm123abc"

	mockClient.On("UploadFile", mock.AnythingOfType("*strings.Reader")).Return(expectedCID, nil)
	defer mockClient.AssertExpectations(t)

	cid, err := svc.UploadFile(reader)

	assert.NoError(t, err)
	assert.Equal(t, expectedCID, cid)
}

func TestUploadFile_Error(t *testing.T) {
	t.Parallel()

	mockClient := new(MockIPFSClient)
	svc := upload.NewEvidenceService(mockClient)

	reader := strings.NewReader("bad file")
	mockClient.On("UploadFile", mock.AnythingOfType("*strings.Reader")).
		Return("", errors.New("upload failed"))
	defer mockClient.AssertExpectations(t)

	cid, err := svc.UploadFile(reader)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "upload failed")
	assert.Empty(t, cid)
}
