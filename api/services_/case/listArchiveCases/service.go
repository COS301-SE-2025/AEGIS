package listArchiveCases

type ArchiveCaseService struct {
	repo ArchiveCaseLister
}

func NewArchiveCaseService(repo ArchiveCaseLister) *ArchiveCaseService {
	return &ArchiveCaseService{repo: repo}
}

func (s *ArchiveCaseService) ListArchivedCases(userID, tenantID, teamID string) ([]ArchivedCase, error) {
	return s.repo.ListArchivedCases(userID, tenantID, teamID)
}
