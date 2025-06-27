package case_creation

import (
	"errors"
	"log"
	"time"

	"aegis-api/services_/auth/registration"

	"github.com/google/uuid"
)

// Service handles business logic for case creation.
type Service struct {
	repo     CaseRepository
	userRepo registration.UserRepository // Inject user repository to resolve full name to user ID
}

// NewCaseService constructs a new CaseService.
func NewCaseService(repo CaseRepository, userRepo registration.UserRepository) *Service {
	if userRepo == nil {
		log.Fatal("userRepo is not initialized")
	}
	return &Service{repo: repo, userRepo: userRepo}
}

// CaseRepository defines persistence operations for cases

// CreateCase validates and creates a new case.
func (s *Service) CreateCase(req *CreateCaseRequest) (*Case, error) {
	// Validate title and team name
	if req.Title == "" {
		return nil, errors.New("title is required")
	}
	if req.TeamName == "" {
		return nil, errors.New("team name is required")
	}

	// Check if userRepo is nil
	if s.userRepo == nil {
		return nil, errors.New("userRepo is nil in CreateCase service")
	}
	// Fetch the user's UUID by their full name
	user, err := s.userRepo.GetUserByFullName(req.CreatedByFullName)
	if err != nil {
		log.Printf("Error fetching user by full name: %v", err)
		return nil, errors.New("user not found with full name: " + req.CreatedByFullName)
	}
	if user == nil {
		log.Printf("User object is nil for full name: %s", req.CreatedByFullName)
		return nil, errors.New("user object is nil")
	}

	// Use user.ID directly as it is already uuid.UUID
	createdByUUID := user.ID

	// Construct new Case
	newCase := &Case{
		ID:                 uuid.New(),
		Title:              req.Title,
		Description:        req.Description,
		Status:             req.Status,
		Priority:           req.Priority,
		InvestigationStage: req.InvestigationStage,
		CreatedBy:          createdByUUID, // Use the resolved user ID as uuid.UUID
		TeamName:           req.TeamName,
		CreatedAt:          time.Now(),
	}

	// Persist the case in the repository
	if err := s.repo.CreateCase(newCase); err != nil {
		return nil, err
	}

	return newCase, nil
}
