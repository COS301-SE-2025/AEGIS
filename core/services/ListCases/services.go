package ListCases

import (
	"aegis-api/db"
	"github.com/google/uuid"
	"aegis-api/services/case_creation" // assuming models live there
)

type Service struct{}

func NewListCasesService() *Service {
	return &Service{}
}

// Fetches all cases from the database
func (s *Service) GetAllCases() ([]case_creation.Case, error) {
	var cases []case_creation.Case
	if err := db.DB.Find(&cases).Error; err != nil {
		return nil, err
	}
	return cases, nil
}

// Optionally: Fetch cases created by a specific user
func (s *Service) GetCasesByUser(userID string) ([]case_creation.Case, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	var cases []case_creation.Case
	if err := db.DB.Where("created_by = ?", uid).Find(&cases).Error; err != nil {
		return nil, err
	}
	return cases, nil
}
