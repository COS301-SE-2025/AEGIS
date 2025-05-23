package case_creation

import (
	"fmt"
	"github.com/google/uuid"
	"aegis-api/db"
)

type CaseService struct{}

func NewCaseService() *CaseService {
	return &CaseService{}
}

func (s *CaseService) CreateCase(req CreateCaseRequest) (Case, error) {
	createdByUUID, err := uuid.Parse(req.CreatedBy)
	if err != nil {
		return Case{}, fmt.Errorf("invalid UUID: %w", err)
	}

	newCase := Case{
		Title:              req.Title,
		Description:        req.Description,
		Status:             req.Status,
		Priority:           req.Priority,
		InvestigationStage: req.InvestigationStage,
		CreatedBy:          createdByUUID,
	}

	if err := db.DB.Create(&newCase).Error; err != nil {
		return Case{}, err
	}

	return newCase, nil
}
