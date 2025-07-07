package case_creation

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Service handles business logic for case creation.
type Service struct {
	repo CaseRepository
}

// NewCaseService constructs a new CaseService.
func NewCaseService(repo CaseRepository) *Service {
	return &Service{repo: repo}
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

	creatorUUID := req.CreatedBy
	// Construct new Case
	newCase := &Case{
		ID:                 uuid.New(),
		Title:              req.Title,
		Description:        req.Description,
		Status:             req.Status,
		Priority:           req.Priority,
		InvestigationStage: req.InvestigationStage,
		CreatedBy:          creatorUUID, // Use the resolved user ID as uuid.UUID
		TeamName:           req.TeamName,
		CreatedAt:          time.Now(),
	}

	// Persist the case in the repository
	if err := s.repo.CreateCase(newCase); err != nil {
		return nil, err
	}

	return newCase, nil
}
