// File: services/user/service.go
package update_user_role

import (
	"fmt"
	"github.com/google/uuid"
)

// ValidRoles lists allowable roles in the system.
var ValidRoles = []string{
	"Incident Responder",
	"Forensic Analyst",
	"Malware Analyst",
	"Threat Intelligent Analyst",
	"DFIR Manager",
	"Legal/Compliance Liaison",
	"Detection Engineer",
	"Generic user",
}

// UserRepository defines persistence operations to update roles atomically.
type UserRepository interface {
	// UpdateRoleAndMirror updates users.role and upserts into user_roles within a transaction.
	UpdateRoleAndMirror(userID uuid.UUID, newRole string) error
}

// UserService provides user-related business logic.
type UserService struct {
	repo UserRepository
}

// NewUserService constructs a UserService.
func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

// UpdateUserRole validates input and delegates to repository.
func (s *UserService) UpdateUserRole(userIDStr, newRole string) error {
	// Parse and validate UUID
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	// Validate role against allowed set
	valid := false
	for _, r := range ValidRoles {
		if r == newRole {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid role: %s", newRole)
	}

	// Delegate to repository for atomic update
	err = s.repo.UpdateRoleAndMirror(userID, newRole)
	if err != nil {
		return fmt.Errorf("failed to update user role: %w", err)
	}

	return nil
}
