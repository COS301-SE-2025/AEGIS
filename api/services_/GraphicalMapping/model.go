package graphicalmapping

import "time"

type IOC struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TenantID  uint      `gorm:"index" json:"tenant_id"`
	CaseID    uint      `gorm:"index" json:"case_id"`
	Type      string    `gorm:"size:50;index" json:"type"` // e.g., IP, Email, Domain
	Value     string    `gorm:"size:255;index" json:"value"`
	CreatedAt time.Time `json:"created_at"`
}
