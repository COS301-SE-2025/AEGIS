package ListUsers

import (
	"context"
)

type ListUserService interface {
	ListUsers(ctx context.Context) ([]User, error)
}

type ListUserRepository interface {
	GetAllUsers(ctx context.Context) ([]User, error)
}
