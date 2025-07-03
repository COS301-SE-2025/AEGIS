package update_case_Investigation_stage

import (
	"errors"
	 "github.com/google/uuid"
	 "fmt"
)

type CaseService interface {
    UpdateCaseStage(caseID string, newStage InvestigationStage) error
}

type caseService struct {
    repo UpdateCaseStageRepository
}

func NewCaseService(r UpdateCaseStageRepository) CaseService {
    return &caseService{repo: r}
}

//validates the new stage and checks if the case exists before updating 
func (s *caseService) UpdateCaseStage(caseID string, newStage InvestigationStage) error {
   
	if !newStage.IsValid() {// Assuming IsValid is a method on InvestigationStage that checks if the stage is valid
		return errors.New("invalid investigation stage")
	}

	id, err := uuid.Parse(caseID) //Convert string to uuid.UUID
    if err != nil {
        return fmt.Errorf("invalid UUID: %w", err) // Return an error if the caseID is not a valid UUID
    }

    exists, err := s.repo.CaseExists(id) // Check if the case exists in the db
    if err != nil {
        return err // Return an error if there was an issue checking the case existence
    }
    if !exists {
        return errors.New("case not found") // Return an error if the case does not exist
    }

    return s.repo.UpdateStage(caseID, newStage) // Update the case stage in the db
}
