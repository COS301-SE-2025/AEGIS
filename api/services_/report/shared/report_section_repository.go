package reportshared

import (
	"context"
	"time"
)

// Shared interface for Postgres section repository
// Move this out of report_ai_assistance to avoid import cycles

type ReportSectionRepository interface {
	GetSectionByID(ctx context.Context, id string) (*ReportSection, error)
	CreateSection(ctx context.Context, section *ReportSection) error
	UpdateSection(ctx context.Context, section *ReportSection) error
	ListSectionsByReport(ctx context.Context, reportID string) ([]*ReportSection, error)
}

// Minimal ReportSection struct for interface compliance
// You may want to expand this as needed

type ReportSection struct {
	ID        string    `gorm:"column:id;type:char(24);primaryKey"`
	ReportID  string    `gorm:"column:report_id;type:uuid;not null"`
	Title     string    `gorm:"column:section_name;type:varchar(255);not null"`
	Content   string    `gorm:"column:content;type:text"`
	Order     int       `gorm:"column:section_order;type:int;default:0"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp"`
}

// TableName sets the table name for Gorm
func (ReportSection) TableName() string {
	return "report_sections"
}
