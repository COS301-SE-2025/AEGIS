package timeline

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// TimelineEvent represents an investigation timeline event for a case.
type TimelineEvent struct {
	ID          string         `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CaseID      string         `gorm:"type:uuid;index;not null" json:"case_id"`
	Description string         `gorm:"type:text;not null" json:"description"`
	Evidence    datatypes.JSON `gorm:"type:jsonb;default:'[]'::jsonb" json:"evidence"` // JSON array of strings or objects
	Tags        datatypes.JSON `gorm:"type:jsonb;default:'[]'::jsonb" json:"tags"`     // JSON array of strings
	Severity    string         `gorm:"size:20;index" json:"severity"`                  // low|medium|high|critical
	AnalystID   string         `gorm:"type:uuid" json:"analyst_id,omitempty"`
	AnalystName string         `gorm:"size:255" json:"analyst_name,omitempty"`

	Order     int            `gorm:"index" json:"order"` // used for ordering events in a case
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
