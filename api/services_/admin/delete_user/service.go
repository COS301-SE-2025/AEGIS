package delete_user

import (
	"fmt"

	"github.com/google/uuid"
)

type UserDeleteService struct {
	repo UserRepository
}

func NewUserDeleteService(repo UserRepository) *UserDeleteService {
	return &UserDeleteService{repo: repo}
}

// DeleteUser deletes a user by ID only if the requester is an Admin.
func (s *UserDeleteService) DeleteUser(req DeleteUserRequest, requesterRole string) error {
	if requesterRole != "DFIR Admin" {
		return fmt.Errorf("unauthorized: only DFIR Admins can delete users")
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return fmt.Errorf("invalid user UUID: %w", err)
	}

	err = s.repo.DeleteUserByID(userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}
