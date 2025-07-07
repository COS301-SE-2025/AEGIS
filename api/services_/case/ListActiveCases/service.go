package ListActiveCases

type Service struct {
	repo ActiveCaseQueryRepository
}

func NewService(repo ActiveCaseQueryRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListActiveCases(userID string) ([]ActiveCase, error) {
	// You might parse/validate UUID here or just trust the repository.
	// The repo will handle querying the DB.
	return s.repo.GetActiveCasesByUserID(nil, userID)
}
