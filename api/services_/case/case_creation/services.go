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
func (s *Service) CreateCase(req CreateCaseRequest) (*Case, error) {
	// Validate title
	if req.Title == "" {
		return nil, errors.New("title is required")
	}

	// Validate team name
	if req.TeamName == "" {
		return nil, errors.New("team name is required")
	}

	// Parse CreatedBy UUID
	creatorUUID, err := uuid.Parse(req.CreatedBy)
	if err != nil {
		return nil, errors.New("invalid CreatedBy UUID")
	}

	// Construct new Case
	newCase := &Case{
		ID:                 uuid.New(),
		Title:              req.Title,
		Description:        req.Description,
		Status:             req.Status,
		Priority:           req.Priority,
		InvestigationStage: req.InvestigationStage,
		CreatedBy:          creatorUUID,
		TeamName:           req.TeamName,
		CreatedAt:          time.Now(),
	}

	// Persist via repository
	if err := s.repo.CreateCase(newCase); err != nil {
		return nil, err
	}

	return newCase, nil
}
