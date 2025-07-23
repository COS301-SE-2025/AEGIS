package case_evidence_totals



import (
	"time"

	"github.com/google/uuid"
)

type InvestigationStage string

const (
	StageAnalysis     InvestigationStage = "analysis"
	StageResearch     InvestigationStage = "research"
	StageEvaluation   InvestigationStage = "evaluation"
	StageFinalization InvestigationStage = "finalization"
)

type Case struct {
	ID                 uuid.UUID          `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Title              string             `gorm:"not null"`
	Description        *string
	Status             string             `gorm:"default:'open'"`
	InvestigationStage InvestigationStage `gorm:"default:'analysis'"`
	Priority           string             `gorm:"default:'medium'"`
	TeamName           string             `gorm:"not null"`
	CreatedBy          uuid.UUID          `gorm:"type:uuid"`
	CreatedAt          time.Time          `gorm:"autoCreateTime"`
		Tags []*Tag `gorm:"many2many:case_tags;constraint:OnDelete:CASCADE;"`
}

func (s InvestigationStage) IsValid() bool {
	switch s {
	case StageAnalysis, StageResearch, StageEvaluation, StageFinalization:
		return true
	default:
		return false
	}
}

type CaseTag struct {
	CaseID uuid.UUID `gorm:"type:uuid;primaryKey"`
	TagID  int       `gorm:"primaryKey"`
}




type EvidenceDTO struct {
	ID         string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	CaseID     string    `gorm:"type:uuid;not null" json:"case_id"`
	UploadedBy string    `gorm:"not null" json:"uploaded_by"`
	Filename   string    `gorm:"not null" json:"filename"`
	FileType   string    `gorm:"not null" json:"file_type"`
	IPFSCID    string    `gorm:"not null" json:"ipfs_cid"`
	FileSize   int64     `gorm:"not null" json:"file_size"`
	Checksum   string    `gorm:"not null" json:"checksum"`
	Metadata   string    `gorm:"type:jsonb" json:"metadata"`
	UploadedAt time.Time `gorm:"autoCreateTime" json:"uploaded_at"`

}

type Tag struct {
	ID   int    `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"uniqueIndex;not null"`
}

func (EvidenceDTO) TableName() string {
	return "evidence"
}