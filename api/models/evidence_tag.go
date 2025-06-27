package models

import "github.com/google/uuid"

type EvidenceTag struct {
	EvidenceID uuid.UUID `json:"evidence_id"`
	TagID      int       `json:"tag_id"`
}

type Tag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
