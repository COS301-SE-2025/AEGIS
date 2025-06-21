package evidence_download

import (
	"io"

	"aegis-api/services/evidence/metadata"
	upload "aegis-api/services/evidence/upload"

	"github.com/google/uuid"
)

type Service struct {
	Repo metadata.Repository
	IPFS upload.IPFSClientImp
}

// NewService initializes the download service.
func NewService(repo metadata.Repository, ipfs upload.IPFSClientImp) *Service {
	return &Service{
		Repo: repo,
		IPFS: ipfs,
	}
}

// DownloadEvidence retrieves the file stream and metadata from IPFS.
func (s *Service) DownloadEvidence(evidenceID uuid.UUID) (filename string, reader io.ReadCloser, filetype string, err error) {
	e, err := s.Repo.FindEvidenceByID(evidenceID)
	if err != nil {
		return "", nil, "", err
	}

	stream, err := s.IPFS.Download(e.IpfsCID)
	if err != nil {
		return "", nil, "", err
	}

	return e.Filename, stream, e.FileType, nil
}
