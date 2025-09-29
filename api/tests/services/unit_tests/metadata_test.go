package unit_tests

import (
	"aegis-api/services_/evidence/metadata"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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
	metadata.GormRepository // embed so it satisfies type assertion
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

func NewMockGormRepository() metadata.Repository {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		// Disable foreign key constraints for easier testing
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to test database: %v", err))
	}

	// Auto-migrate all models that your evidence service might use
	err = db.AutoMigrate(
		&metadata.Evidence{},
		&metadata.EvidenceLog{},
		// Add other models if needed, e.g.:
		// &tags.Tag{},
		// &timeline.Event{},
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to migrate test database: %v", err))
	}

	// Optionally seed required data
	// For example, if you need a default tag:
	// db.Create(&tags.Tag{Name: "urgent", Description: "Urgent evidence"})

	return metadata.NewGormRepository(db)
}
func TestUploadEvidence_Success(t *testing.T) {
	repo := NewMockGormRepository()
	mockIPFS := new(MockIPFSUploaderMetadata)
	service := metadata.NewService(repo, mockIPFS)

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

	err = service.UploadEvidence(req)
	assert.NoError(t, err)

	mockIPFS.AssertExpectations(t)
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
