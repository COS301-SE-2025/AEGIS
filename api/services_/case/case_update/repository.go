package update_case

import (
	"context"
	"errors"

	"time"

	"gorm.io/gorm"
)

// GormUpdateCaseRepository implements UpdateCaseRepository
type GormUpdateCaseRepository struct {
	db *gorm.DB
}

// Constructor
func NewGormUpdateCaseRepository(db *gorm.DB) *GormUpdateCaseRepository {
	return &GormUpdateCaseRepository{db: db}
}

// UpdateCase updates case details in the DB
func (r *GormUpdateCaseRepository) UpdateCase(ctx context.Context, req *UpdateCaseRequest) error {
	result := r.db.WithContext(ctx).Model(&Case{}).
		Where("id = ? AND tenant_id = ? AND team_id = ?", req.CaseID, req.TenantID, req.TeamID).
		Updates(map[string]interface{}{
			"title":               req.Title,
			"description":         req.Description,
			"status":              req.Status,
			"investigation_stage": req.InvestigationStage,
			"updated_at":          time.Now(), // Update the timestamp
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("no rows updated — case may not exist")
	}

	return nil
}
