package ListCases

import (
	"aegis-api/services_/case/case_creation"

	"gorm.io/gorm"
)

type GormCaseQueryRepository struct {
	db *gorm.DB
}

// NewGormCaseQueryRepository creates a new GormCaseQueryRepository
func NewGormCaseQueryRepository(db *gorm.DB) *GormCaseQueryRepository {
	return &GormCaseQueryRepository{
		db: db,
	}
}

// Implements GetAllCases
func (r *GormCaseQueryRepository) GetAllCases() ([]case_creation.Case, error) {
	var cases []case_creation.Case
	err := r.db.Table("cases").Scan(&cases).Error
	return cases, err
}

// Implements GetCasesByUser
func (r *GormCaseQueryRepository) GetCasesByUser(userID string) ([]case_creation.Case, error) {
	var cases []case_creation.Case
	err := r.db.Table("cases").Where("created_by = ?", userID).Scan(&cases).Error
	return cases, err
}

// Implements QueryCases with basic filters
func (r *GormCaseQueryRepository) QueryCases(filter CaseFilter) ([]Case, error) {
	var cases []Case
	query := r.db.Table("cases")

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Priority != "" {
		query = query.Where("priority = ?", filter.Priority)
	}
	if filter.CreatedBy != "" {
		query = query.Where("created_by = ?", filter.CreatedBy)
	}
	if filter.TeamName != "" {
		query = query.Where("team_name = ?", filter.TeamName)
	}
	if filter.TitleTerm != "" {
		query = query.Where("title ILIKE ?", "%"+filter.TitleTerm+"%")
	}
	if filter.SortBy != "" && (filter.SortOrder == "asc" || filter.SortOrder == "desc") {
		query = query.Order(filter.SortBy + " " + filter.SortOrder)
	}

	err := query.Scan(&cases).Error
	return cases, err
}
