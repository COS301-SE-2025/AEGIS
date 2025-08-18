// services_/report/recent.go
package report

import (
	"time"

	"github.com/google/uuid"
)

type RecentReport struct {
	ID           uuid.UUID `json:"id"`
	Title        string    `json:"title"`        // maps from Report.Name
	Status       string    `json:"status"`       // 'draft' | 'review' | 'published'
	LastModified time.Time `json:"lastModified"` // RFC3339 in handler
}

type RecentReportsOptions struct {
	Limit      int
	MineOnly   bool
	ExaminerID uuid.UUID
	CaseID     *uuid.UUID
	Status     *string
	TenantID   uuid.UUID  // ← add
	TeamID     *uuid.UUID // ← add (nil means “all teams in tenant”)
}
