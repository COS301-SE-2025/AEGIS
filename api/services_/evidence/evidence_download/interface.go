package evidence_download

import (
	"io"

	"github.com/google/uuid"
)

type downloadInterface interface {
	DownloadEvidence(evidenceID uuid.UUID) (string, string, io.ReadCloser, error)
}
