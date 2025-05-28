package mocks

import (
	"github.com/stretchr/testify/mock"
	"aegis-api/models"
)

type MockEvidenceViewer struct {
	mock.Mock
}

func (m *MockEvidenceViewer) GetEvidenceByCase(caseID string) ([]models.EvidenceResponse, error) {
	args := m.Called(caseID)
	return args.Get(0).([]models.EvidenceResponse), args.Error(1)
}

func (m *MockEvidenceViewer) GetEvidenceByID(evidenceID string) (*models.EvidenceResponse, error) {
	args := m.Called(evidenceID)
	return args.Get(0).(*models.EvidenceResponse), args.Error(1)
}

func (m *MockEvidenceViewer) SearchEvidence(query string) ([]models.EvidenceResponse, error) {
	args := m.Called(query)
	return args.Get(0).([]models.EvidenceResponse), args.Error(1)
}

func (m *MockEvidenceViewer) GetFilteredEvidence(caseID string, filters map[string]interface{}, sortField string, sortOrder string) ([]models.EvidenceResponse, error) {
	args := m.Called(caseID, filters, sortField, sortOrder)
	return args.Get(0).([]models.EvidenceResponse), args.Error(1)
}