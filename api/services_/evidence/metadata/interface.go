package metadata

import (
	"github.com/google/uuid"
)

type MetadataService interface {
	UploadEvidence(UploadEvidenceRequest) error
	GetEvidenceByCaseID(caseID uuid.UUID) ([]Evidence, error)
	FindEvidenceByID(id uuid.UUID) (*Evidence, error)
	VerifyEvidenceLogChain(evidenceID uuid.UUID) (bool, string, error)
}
