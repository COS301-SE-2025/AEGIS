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
func (r *ActiveCaseRepository) GetActiveCasesByUserID(ctx context.Context, userID string, tenantID string, teamID string) ([]ActiveCase, error) {
	var cases []ActiveCase
	err := r.db.Table("cases").
		Select("DISTINCT cases.*").
		Joins("LEFT JOIN case_user_roles ON case_user_roles.case_id = cases.id").
		Where(`(case_user_roles.user_id = ? OR cases.created_by = ?) 
			   AND cases.status NOT IN (?, ?) 
			   AND cases.tenant_id = ? 
			   AND cases.team_id = ?`,
			userID, userID, "closed", "archived", tenantID, teamID).
		Scan(&cases).Error

	if err != nil {
		return nil, err
	}
	return cases, nil
}
