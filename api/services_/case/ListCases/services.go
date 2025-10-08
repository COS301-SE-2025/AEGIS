package ListCases

import (
	"github.com/google/uuid"
)

// ListActiveCases returns all cases for a tenant with status 'active' and progress set
func (s *Service) ListActiveCases(tenantID string) ([]Case, error) {
	cases, err := s.repo.GetAllCases(tenantID)
	if err != nil {
		return nil, err
	}
	var activeCases []Case
	for _, c := range cases {
		if c.Status == "active" {
			activeCases = append(activeCases, Case{
				ID:                 c.ID,
				Title:              c.Title,
				Description:        c.Description,
				Status:             c.Status,
				Priority:           c.Priority,
				InvestigationStage: c.InvestigationStage,
				CreatedBy:          c.CreatedBy,
				TeamName:           c.TeamName,
				CreatedAt:          c.CreatedAt,
				TenantID:           c.TenantID,
				UpdatedAt:          c.UpdatedAt,
				Progress:           GetProgressForStage(c.InvestigationStage),
			})
		}
	}
	return activeCases, nil
}

func NewListCasesService(repo CaseQueryRepository) *Service {
	return &Service{repo: repo}
}

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
			TenantID:           c.TenantID,
			UpdatedAt:          c.UpdatedAt,
			Progress:           GetProgressForStage(c.InvestigationStage),
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
			TenantID:           c.TenantID,
			UpdatedAt:          c.UpdatedAt,
			Progress:           GetProgressForStage(c.InvestigationStage),
		}
	}
	return result, nil
}

func (s *Service) GetFilteredCases(
	TenantID, status, priority, createdBy, teamName, titleTerm, sortBy, order, userID, teamID string,
) ([]Case, error) {
	var tenantUUID uuid.UUID
	var teamUUID uuid.UUID
	var err error

	if TenantID != "" {
		tenantUUID, err = uuid.Parse(TenantID)
		if err != nil {
			return nil, err
		}
	}
	if teamID != "" {
		teamUUID, err = uuid.Parse(teamID)
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
		UserID:    userID,
		TeamID:    teamUUID,
	}

	// Fix: Map the cases and set Progress field
	cases, err := s.repo.QueryCases(filter)
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
			TenantID:           c.TenantID,
			UpdatedAt:          c.UpdatedAt,
			Progress:           GetProgressForStage(c.InvestigationStage), // Add this line
		}
	}
	return result, nil
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
		TenantID:           c.TenantID,
		UpdatedAt:          c.UpdatedAt,
		Progress:           GetProgressForStage(c.InvestigationStage),
	}
	return result, nil
}
