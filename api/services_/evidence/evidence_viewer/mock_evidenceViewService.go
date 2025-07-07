package evidence_viewer

import (
	"github.com/stretchr/testify/mock"
)

type MockEvidenceViewer struct {
	mock.Mock
}

func (m *MockEvidenceViewer) GetEvidenceByCase(caseID string) ([]EvidenceResponse, error) {
	args := m.Called(caseID)
	return args.Get(0).([]EvidenceResponse), args.Error(1)
}

func (m *MockEvidenceViewer) GetEvidenceByID(evidenceID string) (*EvidenceResponse, error) {
	args := m.Called(evidenceID)
	return args.Get(0).(*EvidenceResponse), args.Error(1)
}

func (m *MockEvidenceViewer) SearchEvidence(query string) ([]EvidenceResponse, error) {
	args := m.Called(query)
	return args.Get(0).([]EvidenceResponse), args.Error(1)
}

func (m *MockEvidenceViewer) GetFilteredEvidence(caseID string, filters map[string]interface{}, sortField string, sortOrder string) ([]EvidenceResponse, error) {
	args := m.Called(caseID, filters, sortField, sortOrder)
	return args.Get(0).([]EvidenceResponse), args.Error(1)
}