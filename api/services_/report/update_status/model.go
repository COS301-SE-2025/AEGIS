package update_status

import (
	"time"

	"github.com/google/uuid"
)

type ReportStatus string

const (
	ReportStatusDraft     ReportStatus = "draft"
	ReportStatusReview    ReportStatus = "review"
	ReportStatusPublished ReportStatus = "published"
)

type Report struct {
	ID           uuid.UUID    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	TenantID     uuid.UUID    `gorm:"type:uuid;not null;index" json:"tenant_id"`
	TeamID       uuid.UUID    `gorm:"type:uuid;not null;index" json:"team_id"`
	CaseID       uuid.UUID    `gorm:"type:uuid;not null;index" json:"case_id"`
	ExaminerID   uuid.UUID    `gorm:"type:uuid;not null;index" json:"examiner_id"`
	Name         string       `gorm:"type:varchar(255);not null" json:"name"`
	MongoID      string       `gorm:"type:char(24)" json:"mongo_id"`
	ReportNumber string       `gorm:"uniqueIndex" json:"report_number"`
	Status       ReportStatus `gorm:"type:report_status;default:'draft'" json:"status"`
	Version      int          `gorm:"not null;default:1" json:"version"`
	DateExamined *time.Time   `json:"date_examined,omitempty"`
	FilePath     string       `gorm:"type:varchar(255);not null" json:"file_path"`
	CreatedAt    time.Time    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time    `gorm:"autoUpdateTime" json:"updated_at"`
}
