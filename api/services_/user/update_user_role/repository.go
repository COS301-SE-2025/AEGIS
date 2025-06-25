// File: services/user/repo_gorm.go
package update_user_role

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GormUserRepo implements the atomic role update.
type GormUserRepo struct {
	db *gorm.DB
}

// NewGormUserRepo constructs the repo.
func NewGormUserRepo(db *gorm.DB) *GormUserRepo {
	return &GormUserRepo{db: db}
}

// User maps to your `users` table.
type User struct {
	ID   uuid.UUID `gorm:"type:uuid;primaryKey"`
	Role string    `gorm:"type:user_role;not null"`
}

// UserRole maps to your `user_roles` table.
type UserRole struct {
	UserID uuid.UUID `gorm:"type:uuid;primaryKey"`
	Role   string    `gorm:"type:user_role;primaryKey"`
}

// UpdateRoleAndMirror updates the enum on users and upserts into user_roles.
func (r *GormUserRepo) UpdateRoleAndMirror(userID uuid.UUID, newRole string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1️ Update the users.role column
		if err := tx.
			Model(&User{}).
			Where("id = ?", userID).
			Update("role", newRole).
			Error; err != nil {
			return err
		}

		// 2️ Upsert into the join table
		ur := UserRole{UserID: userID, Role: newRole}
		if err := tx.
			Clauses(clause.OnConflict{DoNothing: true}).
			Create(&ur).
			Error; err != nil {
			return err
		}
		return nil
	})
}
