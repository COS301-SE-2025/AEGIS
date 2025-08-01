package ListActiveCases

import (
	"context"
	//"aegis-api/db"
	"gorm.io/gorm"
)

type ActiveCaseRepository struct {
	// ActiveCaseQueryRepository is the repository for querying active cases.
	db *gorm.DB
}

func NewActiveCaseRepository(db *gorm.DB) *ActiveCaseRepository {
	return &ActiveCaseRepository{
		db: db,
	}
}
func (r *ActiveCaseRepository) GetActiveCasesByUserID(ctx context.Context, userID string) ([]ActiveCase, error) {
	var cases []ActiveCase
	err := r.db.Table("cases").
		Select("cases.*").
		Joins("JOIN case_user_roles ON case_user_roles.case_id = cases.id").
		Where("case_user_roles.user_id = ? AND cases.status != ?", userID, "closed").
		Scan(&cases).Error

	if err != nil {
		return nil, err
	}
	return cases, nil
}
