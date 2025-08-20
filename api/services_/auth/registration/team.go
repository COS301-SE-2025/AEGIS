package registration

import (
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Team struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Name      string     `gorm:"column:team_name"`
	TenantID  *uuid.UUID `gorm:"type:uuid;index"` // Teams optionally belong to Tenants
	CreatedAt time.Time
	UpdatedAt time.Time
}
type GormTeamRepository struct {
	db *gorm.DB
}

func NewGormTeamRepository(db *gorm.DB) *GormTeamRepository {
	if db == nil {
		log.Fatal("DB is nil in NewGormTeamRepository")
	}
	return &GormTeamRepository{db: db}
}

func (r *GormTeamRepository) Exists(id uuid.UUID) bool {
	var count int64
	r.db.Model(&Team{}).Where("id = ?", id).Count(&count)
	return count > 0
}
