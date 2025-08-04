package ListActiveCases

type Service struct {
	repo ActiveCaseQueryRepository
}

func NewService(repo ActiveCaseQueryRepository) *Service {
	return &Service{repo: repo}
}

// Updated to accept tenantID and teamID for multi-tenancy
func (s *Service) ListActiveCases(userID string, tenantID string, teamID string) ([]ActiveCase, error) {
	// Repo will filter by user, tenant, and team
	return s.repo.GetActiveCasesByUserID(nil, userID, tenantID, teamID)
}
