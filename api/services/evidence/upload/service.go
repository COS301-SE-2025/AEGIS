package evidence

// IPFSClient is a mock interface for the IPFS client
type Service struct {
	ipfs *IPFSClient
}

func NewEvidenceService(ipfs *IPFSClient) *Service {
	return &Service{
		ipfs: ipfs,
	}
}

// Service provides methods to interact with the IPFS client for uploading files.
func (s *Service) UploadFile(path string) (string, error) {
	return s.ipfs.UploadFile(path)
}
