package metadata

import (
	upload "aegis-api/services_/evidence/upload"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/google/uuid"
)

// VerifyEvidenceLogChain checks the hash chain integrity for a given evidence_id
func (s *Service) VerifyEvidenceLogChain(evidenceID uuid.UUID) (bool, string, error) {
	log.Printf("[DEBUG] Service entered for VerifyEvidenceLogChain")

	log.Printf("[DEBUG] VerifyEvidenceLogChain called with evidenceID: %s\n", evidenceID.String())
	var logs []EvidenceLog
	err := s.repo.(*GormRepository).db.Where("evidence_id = ?", evidenceID).Order("created_at ASC").Find(&logs).Error
	if err != nil {
		log.Printf("[ERROR] DB error: %v\n", err)
		return false, "database error", err
	}
	log.Printf("[DEBUG] Retrieved %d log entries\n", len(logs))
	var prevHash string
	for i, log := range logs {
		fmt.Printf("[DEBUG] Checking log #%d: %+v\n", i, log)
		if i == 0 {
			if log.PreviousHash != "" {
				fmt.Printf("[ERROR] First log entry has non-empty previous_hash\n")
				return false, "First log entry has non-empty previous_hash", fmt.Errorf("First log entry has non-empty previous_hash")
			}
		} else {
			prevLog := logs[i-1]
			hashInput := prevLog.Sha256 + prevLog.Sha512 + prevLog.Action + fmt.Sprintf("%v", prevLog.Result) + prevLog.Timestamp.String() + prevLog.Details + prevLog.CreatedAt.String()
			hashBytes := sha256.Sum256([]byte(hashInput))
			prevHash = hex.EncodeToString(hashBytes[:])
			if log.PreviousHash != prevHash {
				fmt.Printf("[ERROR] Hash chain broken at log #%d: expected %s, got %s\n", i, prevHash, log.PreviousHash)
				return false, fmt.Sprintf("Hash chain broken at log #%d", i), fmt.Errorf("Hash chain broken at log #%d", i)
			}
		}
	}
	fmt.Printf("[DEBUG] Hash chain valid for evidenceID: %s\n", evidenceID.String())
	return true, "Hash chain valid", nil
}

type Service struct {
	repo Repository
	ipfs upload.IPFSClientImp
	// IPFS client used for uploading evidence files
}

// FindEvidenceByCaseID satisfies the interface for context autofill
func (s *Service) FindEvidenceByCaseID(caseID uuid.UUID) ([]Evidence, error) {
	return s.repo.FindEvidenceByCaseID(caseID)
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
	// Compute SHA256 and SHA512 checksums while streaming to IPFS
	sha256Hasher := sha256.New()
	sha512Hasher := sha512.New()
	tee := io.TeeReader(data.FileData, io.MultiWriter(sha256Hasher, sha512Hasher))

	// Upload to IPFS
	cid, err := s.ipfs.UploadFile(tee)
	if err != nil {
		return fmt.Errorf("IPFS upload failed: %w", err)
	}
	sha256Sum := hex.EncodeToString(sha256Hasher.Sum(nil))
	sha512Sum := hex.EncodeToString(sha512Hasher.Sum(nil))
	// Merge into metadata JSON
	if data.Metadata == nil {
		data.Metadata = make(map[string]string)
	}
	data.Metadata["sha256"] = sha256Sum
	data.Metadata["sha512"] = sha512Sum

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

	// Save evidence
	if err := s.repo.SaveEvidence(e); err != nil {
		return err
	}

	// Compute previous_hash for hash chain
	var previousHash string
	lastLog, err := s.repo.GetLastEvidenceLog(e.ID)
	if err == nil && lastLog != nil {
		hashInput := lastLog.Sha256 + lastLog.Sha512 + lastLog.Action + fmt.Sprintf("%v", lastLog.Result) + lastLog.Timestamp.String() + lastLog.Details + lastLog.CreatedAt.String()
		hashBytes := sha256.Sum256([]byte(hashInput))
		previousHash = hex.EncodeToString(hashBytes[:])
	}

	// Append to evidence_log (append-only)
	log := &EvidenceLog{
		ID:           uuid.New(),
		EvidenceID:   e.ID,
		Sha256:       sha256Sum,
		Sha512:       sha512Sum,
		Action:       "upload",
		Result:       true,
		PreviousHash: previousHash,
	}
	return s.repo.AppendEvidenceLog(log)
}

// GetEvidenceByCaseID returns all evidence records for a given case.
func (s *Service) GetEvidenceByCaseID(caseID uuid.UUID) ([]Evidence, error) {
	return s.repo.FindEvidenceByCaseID(caseID)
}

// FindEvidenceByID retrieves an evidence record by its ID.
func (s *Service) FindEvidenceByID(id uuid.UUID) (*Evidence, error) {
	return s.repo.FindEvidenceByID(id)
}
