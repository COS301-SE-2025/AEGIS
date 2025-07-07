package evidence_viewer



type EvidenceViewer interface {
    GetEvidenceByCase(caseID string) ([]EvidenceResponse, error)
    GetEvidenceByID(evidenceID string) (*EvidenceResponse, error)
    SearchEvidence(query string) ([]EvidenceResponse, error)
	GetFilteredEvidence(caseID string, filters map[string]interface{}, sortField string, sortOrder string) ([]EvidenceResponse, error)
}


