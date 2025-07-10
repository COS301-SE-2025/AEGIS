package evidence_viewer

type EvidenceViewer interface {
    GetEvidenceFileByID(evidenceID string) (*EvidenceFile, error)
    GetEvidenceFilesByCaseID(caseID string) ([]EvidenceFile, error)
    GetFilteredEvidenceFiles(caseID string, filters map[string]interface{}, sortField, sortOrder string) ([]EvidenceFile, error)
    SearchEvidenceFiles(query string) ([]EvidenceFile, error)
}

