package ListCases

func NewListCasesService(repo CaseQueryRepository) *Service {
	return &Service{repo: repo}
}

// CaseQueryRepository should have the new method signature

func (s *Service) GetAllCases() ([]Case, error) {
	cases, err := s.repo.GetAllCases()
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
		}
	}

	return result, nil
}

func (s *Service) GetCasesByUser(userID string) ([]Case, error) {
	cases, err := s.repo.GetCasesByUser(userID)
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
		}
	}

	return result, nil
}
func (s *Service) GetFilteredCases(status, priority, createdBy, teamName, titleTerm, sortBy, order string) ([]Case, error) {
	filter := CaseFilter{
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

func (s *Service) GetCaseByID(caseID string) (*Case, error) {
	c, err := s.repo.GetCaseByID(caseID)
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
	}
	return result, nil
}
