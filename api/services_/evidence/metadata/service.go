package metadata

import (
	upload "aegis-api/services_/evidence/upload"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"

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
	//  Compute SHA256 checksum while streaming to IPFS
	sha256Hasher := sha256.New()
	md5Hasher := md5.New()

	tee := io.TeeReader(data.FileData, io.MultiWriter(sha256Hasher, md5Hasher))

	// Upload to IPFS
	cid, err := s.ipfs.UploadFile(tee)
	if err != nil {
		return fmt.Errorf("IPFS upload failed: %w", err)
	}
	sha256Sum := hex.EncodeToString(sha256Hasher.Sum(nil))
	md5Sum := hex.EncodeToString(md5Hasher.Sum(nil))
	//  Merge into metadata JSON
	if data.Metadata == nil {
		data.Metadata = make(map[string]string)
	}
	data.Metadata["sha256"] = sha256Sum
	data.Metadata["md5"] = md5Sum

	metadataJSON, err := json.Marshal(data.Metadata)
	if err != nil {
		return fmt.Errorf("metadata JSON marshal failed: %w", err)
	}

	// Build Evidence record with multi-tenancy
	e := &Evidence{
		ID:         uuid.New(),
		CaseID:     data.CaseID,
		UploadedBy: data.UploadedBy,
		TenantID:   data.TenantID,
		TeamID:     data.TeamID,
		Filename:   data.Filename,
		FileType:   data.FileType,
		IpfsCID:    cid,
		FileSize:   data.FileSize,
		Checksum:   sha256Sum,
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
