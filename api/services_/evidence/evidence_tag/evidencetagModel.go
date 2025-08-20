package evidence_tag

import "github.com/google/uuid"

type EvidenceTag struct {
	EvidenceID uuid.UUID `json:"evidence_id"`
	TagID      int       `json:"tag_id"`
}

type Tag struct {
	ID   int    `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"uniqueIndex;not null"`
}
