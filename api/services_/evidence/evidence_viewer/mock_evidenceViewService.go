package evidence_viewer

import (
	"github.com/stretchr/testify/mock"
)

// ✅ Updated Mock to return []EvidenceFile or *EvidenceFile
type MockEvidenceViewer struct {
	mock.Mock
}

func (m *MockEvidenceViewer) GetEvidenceFileByID(evidenceID string) (*EvidenceFile, error) {
	args := m.Called(evidenceID)
	return args.Get(0).(*EvidenceFile), args.Error(1)
}

func (m *MockEvidenceViewer) GetEvidenceFilesByCaseID(caseID string) ([]EvidenceFile, error) {
	args := m.Called(caseID)
	return args.Get(0).([]EvidenceFile), args.Error(1)
}

func (m *MockEvidenceViewer) GetFilteredEvidenceFiles(caseID string, filters map[string]interface{}, sortField string, sortOrder string) ([]EvidenceFile, error) {
	args := m.Called(caseID, filters, sortField, sortOrder)
	return args.Get(0).([]EvidenceFile), args.Error(1)
}

func (m *MockEvidenceViewer) SearchEvidenceFiles(query string) ([]EvidenceFile, error) {
	args := m.Called(query)
	return args.Get(0).([]EvidenceFile), args.Error(1)
}
