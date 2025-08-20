package case_assign

import (
	"aegis-api/services_/auth/registration"

	"github.com/google/uuid"
)

type UserAdapter struct {
	regRepo *registration.GormUserRepository
}

func NewUserAdapter(repo *registration.GormUserRepository) *UserAdapter {
	return &UserAdapter{
		regRepo: repo,
	}
}

// Ensure UserAdapter implements UserRepo
var _ UserRepo = (*UserAdapter)(nil)

func (ua *UserAdapter) GetUserByID(id uuid.UUID) (*User, error) {
	regUser, err := ua.regRepo.GetUserByID(id.String())
	if err != nil {
		return nil, err
	}

	var tenantID uuid.UUID
	if regUser.TenantID != nil {
		tenantID = *regUser.TenantID
	}

	// Convert registration.User to case_assign.User
	return &User{
		ID:        id,
		FullName:  regUser.FullName,
		Email:     regUser.Email,
		TenantID:  tenantID,
		CreatedAt: regUser.CreatedAt,
		UpdatedAt: regUser.UpdatedAt,
	}, nil
}
