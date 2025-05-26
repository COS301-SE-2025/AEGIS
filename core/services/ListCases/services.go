package ListCases

import (
	//"aegis-api/db"
	//"github.com/google/uuid"
	//"fmt"
	//"strings"
	
	"aegis-api/services/case_creation" // assuming models live there
)
type Service struct {
	repo CaseQueryRepository
}


func (s *Service) GetAllCases() ([]case_creation.Case, error) {
	return s.repo.GetAllCases()
}

func (s *Service) GetCasesByUser(userID string) ([]case_creation.Case, error) {
	return s.repo.GetCasesByUser(userID)
}



func NewListCasesService(repo CaseQueryRepository) *Service {
	return &Service{repo: repo}
}
func (s *Service) GetFilteredCases(status, priority, createdBy, titleTerm, sortBy, order string) ([]Case, error) {
	filter := CaseFilter{
		Status:    status,
		Priority:  priority,
		CreatedBy: createdBy,
		TitleTerm: titleTerm,
		SortBy:    sortBy,
		SortOrder: order,
	}
	return s.repo.QueryCases(filter)
}



// Fetches all cases from the database
// func (s *Service) GetAllCases() ([]case_creation.Case, error) {
// 	var cases []case_creation.Case
// 	if err := db.DB.Find(&cases).Error; err != nil {
// 		return nil, err
// 	}
// 	return cases, nil
// }

// // Optionally: Fetch cases created by a specific user
// func (s *Service) GetCasesByUser(userID string) ([]case_creation.Case, error) {
// 	uid, err := uuid.Parse(userID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var cases []case_creation.Case
// 	if err := db.DB.Where("created_by = ?", uid).Find(&cases).Error; err != nil {
// 		return nil, err
// 	}
// 	return cases, nil
// }


// func (s *Service) GetFilteredCases(status, priority, createdBy, titleSearch, sort, order string) ([]Case, error) {
// 	query := db.DB.Model(&Case{})

// 	if status != "" {
// 		query = query.Where("status = ?", status)
// 	}
// 	if priority != "" {
// 		query = query.Where("priority = ?", priority)
// 	}
// 	if createdBy != "" {
// 		query = query.Where("created_by = ?", createdBy)
// 	}
// 	if titleSearch != "" {
// 		query = query.Where("LOWER(title) LIKE ?", "%"+strings.ToLower(titleSearch)+"%")
// 	}

// 	allowedSorts := map[string]bool{"created_at": true, "title": true}
// 	allowedOrders := map[string]bool{"asc": true, "desc": true}
// 	if !allowedSorts[sort] {
// 		sort = "created_at"
// 	}
// 	if !allowedOrders[order] {
// 		order = "desc"
// 	}

// 	query = query.Order(fmt.Sprintf("%s %s", sort, order))

// 	var results []Case
// 	if err := query.Find(&results).Error; err != nil {
// 		return nil, err
// 	}

// 	return results, nil
// }
