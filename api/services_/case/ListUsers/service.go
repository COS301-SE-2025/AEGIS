package ListUsers

import (
	"context"
)

func NewListUserService(repo ListUserRepository) ListUserService {
	return &listUserService{repo: repo}
}

func (s *listUserService) ListUsers(ctx context.Context) ([]User, error) {
	return s.repo.GetAllUsers(ctx)
}
