package ListUsers

import (
	"context"

	"github.com/google/uuid"
)

func NewListUserService(repo ListUserRepository) ListUserService {
	return &listUserService{repo: repo}
}

func (s *listUserService) ListUsers(ctx context.Context) ([]User, error) {
	return s.repo.GetAllUsers(ctx)
}
func (s *listUserService) ListUsersByTenant(ctx context.Context, tenantID uuid.UUID, page int, pageSize int) ([]User, int64, error) {
	return s.repo.GetUsersByTenant(ctx, tenantID, page, pageSize)
}
func (s *listUserService) GetUserByID(ctx context.Context, userID uuid.UUID) (*User, error) {
	return s.repo.GetUserByID(ctx, userID)
}
