package ListUsers

import (
	"context"
)

type ListUserService interface {
	ListUsers(ctx context.Context) ([]User, error)
}

type listUserService struct {
	repo ListUserRepository
}

func NewListUserService(repo ListUserRepository) ListUserService {
	return &listUserService{repo: repo}
}

func (s *listUserService) ListUsers(ctx context.Context) ([]User, error) {
	return s.repo.GetAllUsers(ctx)
}
