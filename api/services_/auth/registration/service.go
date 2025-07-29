package registration

import (
	// Assuming verifyemail package exists for email verification
	"aegis-api/services_/auditlog"
	"context"
	"fmt"
	"log"
	"time"

	"aegis-api/services_/auth/accept_terms"
	verifyemail "aegis-api/services_/auth/verify_email"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// This file contains the service layer of the registration module.
// The service layer handles business logic such as password hashing
// and mapping incoming DTOs to persistence-ready entities.

type RegistrationService struct {
	// repo is an interface to the user repository — used to persist users
	repo       UserRepository
	tenantRepo TenantRepository // Assuming TenantRepository exists
	teamRepo   TeamRepository   // Assuming TeamRepository exists
}

// VerifyUser method verifies the user's email using the provided token
func (s *RegistrationService) VerifyUser(token string) error {
	return verifyemail.VerifyEmail(s.repo.GetDB(), token)
}

// AcceptTerms method handles the acceptance of terms and conditions
func (s *RegistrationService) AcceptTerms(token string) error {
	// Look up token and get the associated user ID
	validToken, err := verifyemail.GetValidToken(s.repo.GetDB(), token)
	if err != nil {
		return err // invalid or expired token
	}
	//Use the UserID to mark the user as having accepted terms
	return accept_terms.AcceptTerms(s.repo.GetDB(), validToken.UserID.String())
}

// NewRegistrationService returns a new instance of the RegistrationService,
// injecting the appropriate UserRepository implementation (e.g. in-memory or GORM).
// func NewRegistrationService(repo UserRepository) *RegistrationService {
// 	return &RegistrationService{repo: repo}
// }

// // Register takes in a RegistrationRequest DTO, hashes the password,
// // builds a domain model, converts it to an entity, and saves it via the repository.
// func (s *RegistrationService) Register(req RegistrationRequest) (User, error) {
// 	//  Check for existing user first
// 	existingUser, err := s.repo.GetUserByEmail(req.Email)
// 	if err == nil && existingUser != nil {
// 		return User{}, fmt.Errorf("user already exists")
// 	}
// 	// Only allow continue if error is "not found"
// 	if err != nil && err != gorm.ErrRecordNotFound {
// 		return User{}, err
// 	}

// 	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
// 	if err != nil {
// 		log.Printf(" Registration failed (hash error) for %s: %v", req.Email, err)
// 		return User{}, err
// 	}

// 	model := NewUserModel(req, string(hash))
// 	id := generateID()
// 	entity := ModelToEntity(model, id)

// 	token := generateToken()
// 	entity.VerificationToken = token
// 	entity.IsVerified = false

// 	err = s.repo.CreateUser(&entity)
// 	if err != nil {
// 		log.Printf(" Registration failed (duplicate?) for %s: %v", req.Email, err)
// 		return User{}, err
// 	}

// 	log.Printf("✅ Registered new user: %s (%s %s)", entity.Email, entity.FullName, entity.Role)

// 	sendVerificationEmail(entity.Email, token)

// 	return entity, nil
// }

// func sendVerificationEmail(email string, token string) {
// 	// TODO: Replace with real SMTP integration

// }

func (s *RegistrationService) Register(req RegistrationRequest) (User, error) {
	// Validate tenant exists if TenantID provided
	if req.TenantID != nil {
		tenantExists := s.tenantRepo.Exists(*req.TenantID)
		if !tenantExists {
			return User{}, fmt.Errorf("tenant not found")
		}
	}
	// Validate team exists if TeamID provided
	if req.TeamID != nil {
		teamExists := s.teamRepo.Exists(*req.TeamID)
		if !teamExists {
			return User{}, fmt.Errorf("team not found")
		}
	}
	// Check for existing user
	existingUser, err := s.repo.GetUserByEmail(req.Email)
	if err == nil && existingUser != nil {
		return User{}, fmt.Errorf("user already exists")
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return User{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf(" Registration failed (hash error) for %s: %v", req.Email, err)
		return User{}, err
	}

	model := NewUserModel(req, string(hash))
	id := generateID()
	entity := ModelToEntity(model, id)
	entity.IsVerified = true

	err = s.repo.CreateUser(&entity)
	if err != nil {
		log.Printf(" Registration failed (duplicate?) for %s: %v", req.Email, err)
		return User{}, err
	}

	// 🔕 Email verification disabled
	token, err := verifyemail.CreateEmailVerificationToken(s.repo.GetDB(), id)
	if err != nil {
		log.Printf("❌ Failed to create email verification token for %s: %v", req.Email, err)
		return User{}, err
	}

	if err := verifyemail.SendVerificationEmail(entity.Email, token); err != nil {
		log.Printf("❌ Failed to send verification email to %s: %v", entity.Email, err)
	}
	log.Printf("✅ Sent email verification to new user: %s", entity.Email)
	log.Printf("✅ Registered new user: %s (%s %s)", entity.Email, entity.FullName, entity.Role)
	return entity, nil
}
func (s *RegistrationService) RegisterTenantUser(req RegistrationRequest) (User, error) {
	tenant := &Tenant{
		ID:        uuid.New(),
		Name:      req.OrganizationName,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := s.tenantRepo.CreateTenant(tenant)
	if err != nil {
		return User{}, fmt.Errorf("failed to create tenant: %w", err)
	}

	req.TenantID = &tenant.ID
	return s.Register(req)
}

func (s *RegistrationService) RegisterTeamUser(req RegistrationRequest) (User, error) {
	team := &Team{
		ID:        uuid.New(),
		Name:      req.TeamName,
		TenantID:  req.TenantID, // Use TenantID from request
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.teamRepo.CreateTeam(team); err != nil {
		return User{}, fmt.Errorf("failed to create team: %w", err)
	}
	req.TeamID = &team.ID
	return s.Register(req)
}

func (r *GormUserRepository) UpdatePassword(userID uuid.UUID, hashedPassword string) error {
	return r.db.
		Model(&User{}).
		Where("id = ?", userID).
		Update("password_hash", hashedPassword).
		Error
}

func (s *RegistrationService) GetAllUsers() ([]User, error) {
	return s.repo.FindAll()
}

func (r *GormUserRepository) GetByID(ctx context.Context, userID string) (*auditlog.User, error) {
	var user User
	if err := r.db.WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		return nil, err
	}

	return &auditlog.User{
		ID:    user.ID.String(),
		Email: user.Email,
		//Role:  user.Role, // Only include this if auditlog.User struct has a Role field
	}, nil
}
func (s *RegistrationService) CreateTenant(name string) (*Tenant, error) {
	tenant := &Tenant{
		ID:        uuid.New(),
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := s.tenantRepo.CreateTenant(tenant)
	if err != nil {
		log.Printf("❌ Failed to create tenant %s: %v", name, err)
		return nil, err
	}
	log.Printf("✅ Created new tenant: %s", name)
	return tenant, nil
}

func (s *RegistrationService) CreateTeam(name string, tenantID *uuid.UUID) (*Team, error) {
	// Optionally validate tenantID
	if tenantID != nil {
		if !s.tenantRepo.Exists(*tenantID) {
			return nil, fmt.Errorf("tenant not found")
		}
	}

	team := &Team{
		ID:        uuid.New(),
		Name:      name,
		TenantID:  tenantID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := s.teamRepo.CreateTeam(team)
	if err != nil {
		log.Printf("❌ Failed to create team %s: %v", name, err)
		return nil, err
	}
	log.Printf("✅ Created new team: %s", name)
	return team, nil
}
