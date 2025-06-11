package evidence

type Service struct {
	ipfs *IPFSClient
}

func NewEvidenceService(ipfs *IPFSClient) *Service {
	return &Service{
		ipfs: ipfs,
	}
}

func (s *Service) UploadFile(path string) (string, error) {
	return s.ipfs.UploadFile(path)
}
