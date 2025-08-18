package graphicalmapping

import "time"

type IOC struct {
	ID        string    `gorm:"type:uuid;default:gen_random_uuid()" json:"id"`
	TenantID  string    `gorm:"type:uuid;index" json:"tenant_id"`
	CaseID    string    `gorm:"type:uuid;index" json:"case_id"`
	Type      string    `gorm:"size:50;index" json:"type"` // e.g., IP, Email, Domain
	Value     string    `gorm:"size:255;index" json:"value"`
	CreatedAt time.Time `json:"created_at"`
}
