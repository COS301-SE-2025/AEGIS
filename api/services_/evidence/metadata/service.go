package metadata

import (
	upload "aegis-api/services_/evidence/upload"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/google/uuid"
)

type Service struct {
	repo Repository
	ipfs upload.IPFSClientImp
	// IPFS client used for uploading evidence files
}

// NewService creates a new instance of the metadata service.
func NewService(repo Repository, ipfs upload.IPFSClientImp) *Service {
	return &Service{repo: repo, ipfs: ipfs}
}

// UploadEvidence uploads a file to IPFS and saves evidence data, including metadata.
// UploadEvidence streams the file to IPFS and computes checksum on-the-fly.
// UploadEvidence uploads evidence to IPFS and saves metadata into the database.
// Supports multi-tenancy (tenant & team).
func (s *Service) UploadEvidence(data UploadEvidenceRequest) error {
	// ✅ Compute SHA256 checksum while streaming to IPFS
	hash := sha256.New()
	tee := io.TeeReader(data.FileData, hash)

	// ✅ Upload to IPFS
	cid, err := s.ipfs.UploadFile(tee)
	if err != nil {
		return fmt.Errorf("IPFS upload failed: %w", err)
	}

	// ✅ Compute checksum
	checksum := fmt.Sprintf("%x", hash.Sum(nil))

	// ✅ Encode metadata as JSON string
	metadataJSON, err := json.Marshal(data.Metadata)
	if err != nil {
		return fmt.Errorf("metadata JSON marshal failed: %w", err)
	}

	// ✅ Build Evidence record with multi-tenancy
	e := &Evidence{
		ID:         uuid.New(),
		CaseID:     data.CaseID,
		UploadedBy: data.UploadedBy,
		TenantID:   data.TenantID, // ✅ new
		TeamID:     data.TeamID,   // ✅ new
		Filename:   data.Filename,
		FileType:   data.FileType,
		IpfsCID:    cid,
		FileSize:   data.FileSize,
		Checksum:   checksum,
		Metadata:   string(metadataJSON),
	}

	return s.repo.SaveEvidence(e)
}

// GetEvidenceByCaseID returns all evidence records for a given case.
func (s *Service) GetEvidenceByCaseID(caseID uuid.UUID) ([]Evidence, error) {
	return s.repo.FindEvidenceByCaseID(caseID)
}

// FindEvidenceByID retrieves an evidence record by its ID.
func (s *Service) FindEvidenceByID(id uuid.UUID) (*Evidence, error) {
	return s.repo.FindEvidenceByID(id)
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
