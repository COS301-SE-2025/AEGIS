package ListCases

import (
    "aegis-api/services/case_creation"
)

// CaseFilter defines the fields available for querying cases, now including TeamName
// to support filtering by the team responsible for the case.
// type CaseFilter struct {
//     Status    string
//     Priority  string
//     CreatedBy string
//     TeamName  string
//     TitleTerm string
//     SortBy    string
//     SortOrder string
// }

// CaseQueryRepository abstracts the data access layer for case querying.
// type CaseQueryRepository interface {
//     GetAllCases() ([]case_creation.Case, error)
//     GetCasesByUser(userID string) ([]case_creation.Case, error)
//     QueryCases(filter CaseFilter) ([]Case, error)
// }

// Service provides operations for listing and filtering cases.
type Service struct {
    repo CaseQueryRepository
}

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
