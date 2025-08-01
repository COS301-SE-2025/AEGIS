package ListUsers

import (
	"context"

	"github.com/google/uuid"
)

type ListUserService interface {
	ListUsers(ctx context.Context) ([]User, error)
	ListUsersByTenant(ctx context.Context, tenantID uuid.UUID) ([]User, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*User, error)
}

type ListUserRepository interface {
	GetAllUsers(ctx context.Context) ([]User, error)
	GetUsersByTenant(ctx context.Context, tenantID uuid.UUID) ([]User, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*User, error)
}
