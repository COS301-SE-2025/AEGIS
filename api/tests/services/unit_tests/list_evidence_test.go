package unit_tests

import (
	"testing"
	"aegis-api/services/evidence"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"time"
	//"fmt"
	
)
// func (d *DummyIPFSClient) Upload(path string) (string, error) {
// 	return "mockCID", nil
// }

func (d *DummyIPFSClient) Download(cid string) ([]byte, error) {
	return []byte("dummy content"), nil
}


func TestListEvidenceByCase(t *testing.T) {
	mockRepo := new(MockEvidenceRepo)
	mockLogger := new(DummyLogger)
	mockIPFS := new(DummyIPFSClient)

	caseID := uuid.New()
	expected := []evidence.Evidence{
		{Filename: "file1.txt", IpfsCID: "cid1", UploadedAt: time.Now()},
		{Filename: "file2.log", IpfsCID: "cid2", UploadedAt: time.Now()},
	}
	mockRepo.On("FindByCase", caseID).Return(expected, nil)

	service := evidence.NewEvidenceService(mockIPFS, mockRepo, mockLogger)
	results, err := service.ListEvidenceByCase(caseID.String())

	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "file1.txt", results[0].Filename)

	mockRepo.AssertExpectations(t)
}
func TestDownloadEvidenceByUser(t *testing.T) {
    mockRepo := new(MockEvidenceRepo)
    mockIPFS := new(MockIPFSClient)
    mockLogger := new(DummyLogger)
    svc := evidence.NewEvidenceService(mockIPFS, mockRepo, mockLogger)

    uid := uuid.New()
    cid := "Qm123abc"
    payload := []byte("mock file content")

    // 1) repo returns one record
    mockRepo.
        On("FindByUser", mock.AnythingOfType("uuid.UUID")).
        Return([]evidence.Evidence{
            {Filename: "file1.txt", FileType: "text/plain", IpfsCID: cid},
        }, nil)

    // 2) ipfs.Download is called with that CID
    mockIPFS.
        On("Download", cid).
        Return(payload, nil)

    files, err := svc.DownloadEvidenceByUser(uid.String())
    assert.NoError(t, err)

    // 3) guard against panic and assert contents
    assert.Len(t, files, 1)
    if len(files) > 0 {
        assert.Equal(t, "file1.txt", files[0].Filename)
        assert.Equal(t, payload, files[0].Content)
    }

    mockRepo.AssertExpectations(t)
    mockIPFS.AssertExpectations(t)
}





func TestListEvidenceByUser_NoResults(t *testing.T) {
	mockRepo := new(MockEvidenceRepo)
	mockLogger := new(DummyLogger)
	mockIPFS := new(DummyIPFSClient)
	service := evidence.NewEvidenceService(mockIPFS, mockRepo, mockLogger)

	userID := uuid.New()

	// ✅ Exact same string comparison logic
	mockRepo.On("FindByUser", mock.MatchedBy(func(id uuid.UUID) bool {
		return id.String() == userID.String()
	})).Return([]evidence.Evidence{}, nil)

	results, err := service.ListEvidenceByUser(userID.String())

	assert.NoError(t, err)
	assert.Empty(t, results)

	mockRepo.AssertExpectations(t)
}
