package ListClosedCases

import "context"

type Service struct {
	repo ListClosedCasesRepository
}

func NewService(repo ListClosedCasesRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListClosedCases(userID string, tenantID string, teamID string) ([]ClosedCase, error) {
	return s.repo.GetClosedCasesByUserID(context.TODO(), userID, tenantID, teamID)
}
