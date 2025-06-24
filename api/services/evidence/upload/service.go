package upload

// Service handles high-level file operations via IPFS
type Service struct {
	ipfs IPFSClientImp
}

// NewEvidenceService returns a new upload service using the provided IPFS client.
func NewEvidenceService(ipfs IPFSClientImp) *Service {
	return &Service{
		ipfs: ipfs,
	}
}

// UploadFile uploads a file to IPFS via the configured IPFS client.
func (s *Service) UploadFile(path string) (string, error) {
	return s.ipfs.UploadFile(path)
}
