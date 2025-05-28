package Evidence_Viewer

import "aegis-api/models"

type EvidenceViewer interface {
    GetEvidenceByCase(caseID string) ([]models.EvidenceResponse, error)
    GetEvidenceByID(evidenceID string) (*models.EvidenceResponse, error)
    SearchEvidence(query string) ([]models.EvidenceResponse, error)
	GetFilteredEvidence(caseID string, filters map[string]interface{}, sortField string, sortOrder string) ([]models.EvidenceResponse, error)
}


