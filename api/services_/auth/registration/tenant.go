package registration

import (
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Tenant struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name      string    `gorm:"unique;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
type GormTenantRepository struct {
	db *gorm.DB
}

func NewGormTenantRepository(db *gorm.DB) *GormTenantRepository {
	if db == nil {
		log.Fatal("DB is nil in NewGormTenantRepository")
	}
	return &GormTenantRepository{db: db}
}

func (r *GormTenantRepository) Exists(id uuid.UUID) bool {
	var count int64
	r.db.Model(&Tenant{}).Where("id = ?", id).Count(&count)
	return count > 0
}
func (r *GormTenantRepository) GetAll() ([]Tenant, error) {
	var tenants []Tenant
	if err := r.db.Find(&tenants).Error; err != nil {
		return nil, err
	}
	return tenants, nil
}
