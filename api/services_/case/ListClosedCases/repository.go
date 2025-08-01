package ListClosedCases

import (
	"context"
	//"aegis-api/db"
	"gorm.io/gorm"
)

type ClosedCaseRepository struct {
	// ActiveCaseQueryRepository is the repository for querying active cases.
	db *gorm.DB
}

func NewClosedCaseRepository(db *gorm.DB) *ClosedCaseRepository {
	return &ClosedCaseRepository{
		db: db,
	}
}

func (r *ClosedCaseRepository) GetClosedCasesByUserID(ctx context.Context, userID string, tenantID string, teamID string) ([]ClosedCase, error) {
	var cases []ClosedCase
	err := r.db.Table("cases").
		Select("DISTINCT cases.*").
		Joins("LEFT JOIN case_user_roles ON case_user_roles.case_id = cases.id").
		Where(`(case_user_roles.user_id = ? OR cases.created_by = ?)
               AND cases.status = ?
               AND cases.tenant_id = ?
               AND cases.team_id = ?`,
			userID, userID, "closed", tenantID, teamID).
		Scan(&cases).Error

	if err != nil {
		return nil, err
	}
	return cases, nil
}
