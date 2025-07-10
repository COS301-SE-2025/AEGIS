package unit_tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	

	"aegis-api/services_/evidence/evidence_viewer"
)

func TestServiceGetEvidenceFilesByCaseID(t *testing.T) {
	mockRepo := new(evidence_viewer.MockEvidenceViewer)
	service := &evidence_viewer.EvidenceService{Repo: mockRepo}

	expected := []evidence_viewer.EvidenceFile{
		{ID: "ev123", Data: []byte("file1 bytes")},
		{ID: "ev124", Data: []byte("file2 bytes")},
	}

	mockRepo.On("GetEvidenceFilesByCaseID", "case456").Return(expected, nil)

	files, err := service.GetEvidenceFilesByCaseID("case456")
	assert.NoError(t, err)
	assert.Equal(t, expected, files)
	mockRepo.AssertCalled(t, "GetEvidenceFilesByCaseID", "case456")
}

func TestServiceGetEvidenceFileByID(t *testing.T) {
	mockRepo := new(evidence_viewer.MockEvidenceViewer)
	service := &evidence_viewer.EvidenceService{Repo: mockRepo}

	expected := &evidence_viewer.EvidenceFile{
		ID:   "ev123",
		Data: []byte("actual file bytes"),
	}

	mockRepo.On("GetEvidenceFileByID", "ev123").Return(expected, nil)

	file, err := service.GetEvidenceFileByID("ev123")
	assert.NoError(t, err)
	assert.Equal(t, expected, file)
	mockRepo.AssertCalled(t, "GetEvidenceFileByID", "ev123")
}

func TestServiceSearchEvidenceFiles(t *testing.T) {
	mockRepo := new(evidence_viewer.MockEvidenceViewer)
	service := &evidence_viewer.EvidenceService{Repo: mockRepo}

	expected := []evidence_viewer.EvidenceFile{
		{ID: "ev001", Data: []byte("notes pdf bytes")},
	}

	mockRepo.On("SearchEvidenceFiles", "notes").Return(expected, nil)

	files, err := service.SearchEvidenceFiles("notes")
	assert.NoError(t, err)
	assert.Equal(t, expected, files)
	mockRepo.AssertCalled(t, "SearchEvidenceFiles", "notes")
}

func TestServiceGetFilteredEvidenceFiles(t *testing.T) {
	mockRepo := new(evidence_viewer.MockEvidenceViewer)
	service := &evidence_viewer.EvidenceService{Repo: mockRepo}

	filters := map[string]interface{}{"file_type": "pdf"}
	expected := []evidence_viewer.EvidenceFile{
		{ID: "ev002", Data: []byte("filtered pdf bytes")},
	}

	mockRepo.On("GetFilteredEvidenceFiles", "case789", filters, "uploaded_at", "desc").Return(expected, nil)

	files, err := service.GetFilteredEvidenceFiles("case789", filters, "uploaded_at", "desc")
	assert.NoError(t, err)
	assert.Equal(t, expected, files)
	mockRepo.AssertCalled(t, "GetFilteredEvidenceFiles", "case789", filters, "uploaded_at", "desc")
}
