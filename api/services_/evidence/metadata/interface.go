package metadata

import (
	"github.com/google/uuid"
)

// Repository defines methods for storing and retrieving evidence data.
type Repository interface {
	// SaveEvidence inserts a new evidence record with embedded metadata.
	SaveEvidence(e *Evidence) error

	// FindEvidenceByID retrieves an evidence record by its ID.
	FindEvidenceByID(id uuid.UUID) (*Evidence, error)
}

type MetadataService interface {
	UploadEvidence(UploadEvidenceRequest) error
}
