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
	ExaminerID uuid.UUID  // who is asking; used when MineOnly=true
	CaseID     *uuid.UUID // optional filter
	Status     *string    // optional filter
	// If you carry multi-tenancy via Case → Tenant, filter in repo query with a JOIN.
}
