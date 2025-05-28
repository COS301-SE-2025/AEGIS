package get_collaborators

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) GetCollaboratorsByCaseID(caseID uuid.UUID) ([]Collaborator, error) {
	var collaborators []Collaborator

	err := r.db.
		Table("case_user_roles").
		Select("users.id, users.full_name, users.email, case_user_roles.role, case_user_roles.assigned_at").
		Joins("JOIN users ON users.id = case_user_roles.user_id").
		Where("case_user_roles.case_id = ?", caseID).
		Scan(&collaborators).Error

	if err != nil {
		return nil, err
	}

	return collaborators, nil
}

