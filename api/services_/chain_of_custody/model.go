package chain_of_custody

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type ChainOfCustody struct {
	ID              uuid.UUID      `json:"id" db:"id"`
	EvidenceID      uuid.UUID      `json:"evidence_id" db:"evidence_id"`
	Custodian       string         `json:"custodian" db:"custodian"`
	AcquisitionDate *time.Time     `json:"acquisition_date" db:"acquisition_date"`
	AcquisitionTool string         `json:"acquisition_tool" db:"acquisition_tool"`
	SystemInfo      datatypes.JSON `json:"system_info" db:"system_info"`
	ForensicInfo    datatypes.JSON `json:"forensic_info" db:"forensic_info"`
	CreatedAt       time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at" db:"updated_at"`
}

func (ChainOfCustody) TableName() string {
	return "chain_of_custody"
}
