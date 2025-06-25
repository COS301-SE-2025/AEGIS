package ListCases

import (
	"aegis-api/services_/case/case_creation"
)

// NewListCasesService constructs a new ListCases service.
func NewListCasesService(repo CaseQueryRepository) *Service {
	return &Service{repo: repo}
}

// GetAllCases returns all cases without filtering.
func (s *Service) GetAllCases() ([]case_creation.Case, error) {
	return s.repo.GetAllCases()
}

// GetCasesByUser returns cases created by a specific user.
func (s *Service) GetCasesByUser(userID string) ([]case_creation.Case, error) {
	return s.repo.GetCasesByUser(userID)
}

// GetFilteredCases applies multiple filters, including status, priority, creator,
// team name, title search term, sorting field and order.
func (s *Service) GetFilteredCases(
	status,
	priority,
	createdBy,
	teamName,
	titleTerm,
	sortBy,
	order string,
) ([]Case, error) {
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
