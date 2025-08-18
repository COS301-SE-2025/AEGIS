package ListCases

import (
	"github.com/google/uuid"
)

func NewListCasesService(repo CaseQueryRepository) *Service {
	return &Service{repo: repo}
}

// CaseQueryRepository should have the new method signature

func (s *Service) GetAllCases(tenantID string) ([]Case, error) {
	cases, err := s.repo.GetAllCases(tenantID)
	if err != nil {
		return nil, err
	}
	result := make([]Case, len(cases))
	for i, c := range cases {
		result[i] = Case{
			ID:                 c.ID,
			Title:              c.Title,
			Description:        c.Description,
			Status:             c.Status,
			Priority:           c.Priority,
			InvestigationStage: c.InvestigationStage,
			CreatedBy:          c.CreatedBy,
			TeamName:           c.TeamName,
			CreatedAt:          c.CreatedAt,
			TenantID:           c.TenantID, // Ensure TenantID is included
		}
	}

	return result, nil
}

func (s *Service) GetCasesByUser(userID string, tenantID string) ([]Case, error) {
	cases, err := s.repo.GetCasesByUser(userID, tenantID)
	if err != nil {
		return nil, err
	}
	result := make([]Case, len(cases))
	for i, c := range cases {
		result[i] = Case{
			ID:                 c.ID,
			Title:              c.Title,
			Description:        c.Description,
			Status:             c.Status,
			Priority:           c.Priority,
			InvestigationStage: c.InvestigationStage,
			CreatedBy:          c.CreatedBy,
			TeamName:           c.TeamName,
			CreatedAt:          c.CreatedAt,
			TenantID:           c.TenantID, // Ensure TenantID is included
		}
	}

	return result, nil
}

func (s *Service) GetFilteredCases(TenantID, status, priority, createdBy, teamName, titleTerm, sortBy, order string) ([]Case, error) {
	var tenantUUID uuid.UUID
	var err error
	if TenantID != "" {
		tenantUUID, err = uuid.Parse(TenantID)
		if err != nil {
			return nil, err
		}
	}
	filter := CaseFilter{
		TenantID:  tenantUUID,
		Status:    status,
		Priority:  priority,
		CreatedBy: createdBy,
		TeamName:  teamName,
		TitleTerm: titleTerm,
		SortBy:    sortBy,
		SortOrder: order,
	}
	return s.repo.QueryCases(filter)
}

func (s *Service) GetCaseByID(caseID string, tenantID string) (*Case, error) {
	c, err := s.repo.GetCaseByID(caseID, tenantID)
	if err != nil {
		return nil, err
	}
	result := &Case{
		ID:                 c.ID,
		Title:              c.Title,
		Description:        c.Description,
		Status:             c.Status,
		Priority:           c.Priority,
		InvestigationStage: c.InvestigationStage,
		CreatedBy:          c.CreatedBy,
		TeamName:           c.TeamName,
		CreatedAt:          c.CreatedAt,
		TenantID:           c.TenantID, // Ensure TenantID is included
	}
	return result, nil
}
