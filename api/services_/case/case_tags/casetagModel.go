package case_tags

import(
	"github.com/google/uuid"
)

type CaseTag struct {
	CaseID uuid.UUID `gorm:"type:uuid;primaryKey"`
	TagID  int       `gorm:"primaryKey"`
}

type Tag struct {
	ID   int    `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"uniqueIndex;not null"`
}
