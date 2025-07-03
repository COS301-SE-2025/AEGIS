package evidence_viewer



type EvidenceService struct {
	Repo       EvidenceViewer
	IPFSClient IPFSClient
}

func (s *EvidenceService) GetEvidenceByCase(caseID string) ([]EvidenceResponse, error) {
	return s.Repo.GetEvidenceByCase(caseID)
}

func (s *EvidenceService) GetEvidenceByID(evidenceID string) (*EvidenceResponse, error) {
	return s.Repo.GetEvidenceByID(evidenceID)
}

func (s *EvidenceService) SearchEvidence(query string) ([]EvidenceResponse, error) {
	return s.Repo.SearchEvidence(query)
}

func (s *EvidenceService) GetFilteredEvidence(caseID string, filters map[string]interface{}, sortField string, sortOrder string) ([]EvidenceResponse, error) {
	return s.Repo.GetFilteredEvidence(caseID, filters, sortField, sortOrder)
}
