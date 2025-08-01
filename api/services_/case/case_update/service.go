package update_case

import (
	"context"
)

type Service struct {
	repo UpdateCaseRepository
}

func NewService(repo UpdateCaseRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) UpdateCaseDetails(ctx context.Context, req *UpdateCaseRequest) (*UpdateCaseResponse, error) {
	if err := s.repo.UpdateCase(ctx, req); err != nil {
		return nil, err
	}
	return &UpdateCaseResponse{Success: true}, nil
}
