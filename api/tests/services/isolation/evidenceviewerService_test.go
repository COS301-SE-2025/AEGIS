package isolation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	

	
	"aegis-api/services_/evidence/evidence_viewer"
)

func TestServiceGetEvidenceByCase(t *testing.T) {
	mockRepo := new(evidence_viewer.MockEvidenceViewer)
	service := &evidence_viewer.EvidenceService{Repo: mockRepo}

	expected := []evidence_viewer.EvidenceResponse{
		{ID: "ev123", CaseID: "case456", Filename: "photo.jpg"},
	}

	mockRepo.On("GetEvidenceByCase", "case456").Return(expected, nil)

	evidence, err := service.GetEvidenceByCase("case456")
	assert.NoError(t, err)
	assert.Equal(t, expected, evidence)
	mockRepo.AssertCalled(t, "GetEvidenceByCase", "case456")
}

func TestServiceGetEvidenceByID(t *testing.T) {
	mockRepo := new(evidence_viewer.MockEvidenceViewer)
	service := &evidence_viewer.EvidenceService{Repo: mockRepo}

	expected := &evidence_viewer.EvidenceResponse{ID: "ev123", Filename: "file.jpg"}
	mockRepo.On("GetEvidenceByID", "ev123").Return(expected, nil)

	evidence, err := service.GetEvidenceByID("ev123")
	assert.NoError(t, err)
	assert.Equal(t, expected, evidence)
	mockRepo.AssertCalled(t, "GetEvidenceByID", "ev123")
}

func TestServiceSearchEvidence(t *testing.T) {
	mockRepo := new(evidence_viewer.MockEvidenceViewer)
	service := &evidence_viewer.EvidenceService{Repo: mockRepo}

	expected := []evidence_viewer.EvidenceResponse{
		{ID: "ev001", Filename: "notes.pdf"},
	}
	mockRepo.On("SearchEvidence", "notes").Return(expected, nil)

	results, err := service.SearchEvidence("notes")
	assert.NoError(t, err)
	assert.Equal(t, expected, results)
	mockRepo.AssertCalled(t, "SearchEvidence", "notes")
}

func TestServiceGetFilteredEvidence(t *testing.T) {
	mockRepo := new(evidence_viewer.MockEvidenceViewer)
	service := &evidence_viewer.EvidenceService{Repo: mockRepo}

	filters := map[string]interface{}{"file_type": "pdf"}
	expected := []evidence_viewer.EvidenceResponse{
		{ID: "ev002", Filename: "document.pdf"},
	}

	mockRepo.On("GetFilteredEvidence", "case789", filters, "uploaded_at", "desc").Return(expected, nil)

	results, err := service.GetFilteredEvidence("case789", filters, "uploaded_at", "desc")
	assert.NoError(t, err)
	assert.Equal(t, expected, results)
	mockRepo.AssertCalled(t, "GetFilteredEvidence", "case789", filters, "uploaded_at", "desc")
}