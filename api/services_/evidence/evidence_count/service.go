package evidencecount

type evidenceService struct {
	repo EvidenceRepository
}

func NewEvidenceService(repo EvidenceRepository) *evidenceService {
	return &evidenceService{repo: repo}
}

func (s *evidenceService) GetEvidenceCount(tenantID string) (int64, error) {
	return s.repo.GetEvidenceCountByTenantID(tenantID)
}
