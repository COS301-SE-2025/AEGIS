package report_ai_assistance

import "time"

type ReportSection struct {
	ID        string `gorm:"primaryKey;type:char(24)"`
	ReportID  string `gorm:"type:uuid;not null;index"`
	Title     string `gorm:"column:section_name;size:255;not null"`
	Content   string `gorm:"type:text"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AISuggestion struct {
	ID             string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ReportID       string    `gorm:"type:uuid;not null;index"`
	SectionID      string    `gorm:"type:char(24);not null;index"`
	SuggestionText string    `gorm:"column:suggestion_text;type:text;not null"`
	Version        int       `gorm:"type:int;default:1"`
	Status         string    `gorm:"type:varchar(20);default:'pending'"`
	CreatedByAI    bool      `gorm:"type:boolean;default:true"`
	CreatedAt      time.Time `gorm:"type:timestamp;default:now()"`
}

type SectionRef struct {
	ID        string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	SectionID string `gorm:"type:uuid;not null;index"`
	RefType   string `gorm:"size:100;not null"` // e.g., "evidence", "IOC"
	RefID     string `gorm:"type:uuid;not null"`
}

type AIFeedback struct {
	ID           string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	SuggestionID string `gorm:"type:uuid;not null;index"`
	UserID       string `gorm:"type:uuid;not null"`
	Approved     bool
	Comment      string `gorm:"type:text"`
	CreatedAt    time.Time
}
