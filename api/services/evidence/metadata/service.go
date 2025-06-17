package metadata

import (
	upload "aegis-api/services/evidence/upload"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"

	"gorm.io/datatypes"
)

type Service struct {
	repo Repository
	ipfs upload.IPFSClient // IPFS client used for uploading evidence files
}

// NewService creates a new instance of the metadata service.
func NewService(repo Repository, ipfs upload.IPFSClient) *Service {
	return &Service{repo: repo, ipfs: ipfs}
}

// UploadEvidence uploads a file to IPFS and saves evidence data, including metadata.
func (s *Service) UploadEvidence(data UploadEvidenceRequest) error {
	// Upload the file to IPFS and get its CID
	cid, err := s.ipfs.UploadFile(data.FilePath)
	if err != nil {
		return err
	}
	// Compute SHA-256 checksum
	checksum, err := computeChecksum(data.FilePath)
	if err != nil {
		return err
	}

	// Construct the evidence record
	e := &Evidence{
		CaseID:     data.CaseID,
		UploadedBy: data.UploadedBy,
		Filename:   data.Filename,
		FileType:   data.FileType,
		IpfsCID:    cid,
		FileSize:   data.FileSize,
		Checksum:   checksum,
		Metadata:   datatypes.JSONMap(convertToJSONMap(data.Metadata)), // Convert map[string]string to datatypes.JSONMap
	}

	// Save the record
	return s.repo.SaveEvidence(e)
}

// convertToJSONMap converts a map[string]string to map[string]interface{} for use with datatypes.JSONMap.
func convertToJSONMap(m map[string]string) map[string]interface{} {
	result := make(map[string]interface{}, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}

// computeChecksum reads the file at the given path and returns a SHA-256 checksum.
func computeChecksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
