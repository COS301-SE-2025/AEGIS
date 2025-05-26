package case_creation

import (
	"fmt"

	"github.com/google/uuid"
)

type CaseService struct {
	repo CaseRepository
}

// NewCaseService injects a repository (can be real or mock)
func NewCaseService(repo CaseRepository) *CaseService {
	return &CaseService{repo: repo}
}

func (s *CaseService) CreateCase(req CreateCaseRequest) (Case, error) {
	createdByUUID, err := uuid.Parse(req.CreatedBy)
	if err != nil {
		return Case{}, fmt.Errorf("invalid UUID: %w", err)
	}

	newCase := Case{
		ID:                 uuid.New(),
		Title:              req.Title,
		Description:        req.Description,
		Status:             req.Status,
		Priority:           req.Priority,
		InvestigationStage: req.InvestigationStage,
		CreatedBy:          createdByUUID,
	}

	// Save to repository
	if err := s.repo.CreateCase(&newCase); err != nil {
		return Case{}, err
	}

	return newCase, nil
}
