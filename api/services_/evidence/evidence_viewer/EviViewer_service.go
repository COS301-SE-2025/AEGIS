package evidence_viewer

type EvidenceService struct {
    Repo EvidenceViewer
	
}

func NewEvidenceService(repo EvidenceViewer) *EvidenceService {
    return &EvidenceService{Repo: repo}
}

func (s *EvidenceService) GetEvidenceFileByID(evidenceID string) (*EvidenceFile, error) {
    return s.Repo.GetEvidenceFileByID(evidenceID)
}

func (s *EvidenceService) GetEvidenceFilesByCaseID(caseID string) ([]EvidenceFile, error) {
    return s.Repo.GetEvidenceFilesByCaseID(caseID)
}

func (s *EvidenceService) GetFilteredEvidenceFiles(
    caseID string,
    filters map[string]interface{},
    sortField, sortOrder string,
) ([]EvidenceFile, error) {
    return s.Repo.GetFilteredEvidenceFiles(caseID, filters, sortField, sortOrder)
}

func (s *EvidenceService) SearchEvidenceFiles(query string) ([]EvidenceFile, error) {
    return s.Repo.SearchEvidenceFiles(query)
}