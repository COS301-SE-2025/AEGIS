package evidence

import (
	//"aegis-api/db"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"errors"
	"time"
)

type Service struct {
	ipfs   IPFSClient
	repo   EvidenceRepository
	logger EvidenceLogger
}

type IPFSClient interface {
	Upload(path string) (string, error)
	Download(cid string) ([]byte, error) 
}


type EvidenceRepository interface {
	SaveEvidence(e *Evidence) error
	AttachTags(e *Evidence, tags []string) error
	FindByID(id uuid.UUID) (*Evidence, error)
	DeleteByID(id uuid.UUID) error
	FindByCase(caseID uuid.UUID) ([]Evidence, error) 
	FindByUser(userID uuid.UUID) ([]Evidence, error)
	PreloadMetadata(id uuid.UUID) (*Evidence, error)
}



type EvidenceLogger interface {
	Log(userID, evidenceID, filename string) error
}


func NewEvidenceService(ipfs IPFSClient, repo EvidenceRepository, logger EvidenceLogger) *Service {
	return &Service{ipfs: ipfs, repo: repo, logger: logger}
}



func (s *Service) UploadEvidence(req UploadEvidenceRequest) (*Evidence, error) {
	caseID, err := uuid.Parse(req.CaseID)
	if err != nil {
		return nil, fmt.Errorf("invalid case ID: %w", err)
	}
	userID, err := uuid.Parse(req.UploadedBy)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	e := &Evidence{
		CaseID:     caseID,
		UploadedBy: userID,
		Filename:   req.Filename,
		FileType:   req.FileType,
		IpfsCID:    req.IpfsCID,
		FileSize:   req.FileSize,
		Checksum:   req.Checksum,
		Metadata:   req.Metadata,
	}

	if err := s.repo.SaveEvidence(e); err != nil {
		return nil, err
	}
	if err := s.repo.AttachTags(e, req.Tags); err != nil {
		return nil, err
	}
	if err := s.logger.Log(userID.String(), e.ID.String(), e.Filename); err != nil {
		fmt.Printf("⚠️ Failed to log upload: %v\n", err)
	}

	return e, nil
}

func (s *Service) ListEvidenceByCase(caseID string) ([]Evidence, error) {
	id, err := uuid.Parse(caseID)
	if err != nil {
		return nil, fmt.Errorf("invalid case ID: %w", err)
	}

	return s.repo.FindByCase(id)
}

func (s *Service) ListEvidenceByUser(userID string) ([]Evidence, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	return s.repo.FindByUser(id)
}


func (s *Service) GetEvidenceByID(evidenceID string) (*Evidence, error) {
	id, err := uuid.Parse(evidenceID)
	if err != nil {
		return nil, fmt.Errorf("invalid evidence ID: %w", err)
	}

	ev, err := s.repo.FindByID(id) // 
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("evidence not found")
		}
		return nil, err
	}

	return ev, nil
}

func (s *Service) DeleteEvidenceByID(evidenceID string) error {
	id, err := uuid.Parse(evidenceID)
	if err != nil {
		return fmt.Errorf("invalid evidence ID: %w", err)
	}

	err = s.repo.DeleteByID(id) // now using mockable interface
	if err != nil {
		return err
	}

	return nil
}


func (s *Service) GetEvidenceMetadata(evidenceID string) (*EvidenceMetadata, error) {
	id, err := uuid.Parse(evidenceID)
	if err != nil {
		return nil, fmt.Errorf("invalid evidence ID: %w", err)
	}

	ev, err := s.repo.PreloadMetadata(id)

	if err != nil {
	return nil, fmt.Errorf("evidence not found")
}


	var tagNames []string
	for _, tag := range ev.Tags {
		tagNames = append(tagNames, tag.Name)
	}

	metadata := &EvidenceMetadata{
		ID:         ev.ID.String(),
		Filename:   ev.Filename,
		FileType:   ev.FileType,
		IpfsCID:    ev.IpfsCID,
		FileSize:   ev.FileSize,
		Checksum:   ev.Checksum,
		Metadata:   ev.Metadata,
		Tags:       tagNames,
		CaseID:     ev.CaseID.String(),
		UploadedBy: ev.UploadedBy.String(),
		UploadedAt: ev.UploadedAt.Format(time.RFC3339),
	}

	return metadata, nil
}
func (s *Service) DownloadEvidenceByUser(userID string) ([]EvidenceFile, error) {
    // 1) Parse the incoming string
    id, err := uuid.Parse(userID)
    if err != nil {
        return nil, fmt.Errorf("invalid user ID: %w", err)
    }

    // 2) Fetch metadata records from Postgres
    records, err := s.repo.FindByUser(id)
    if err != nil {
        return nil, err
    }

    // 3) For each record, download from IPFS
    var out []EvidenceFile
    for _, ev := range records {
        blob, err := s.ipfs.Download(ev.IpfsCID)
        if err != nil {
            // optional: log and skip
            _ = s.logger.Log(userID, ev.ID.String(), "ipfs download failed")
            continue
        }
        out = append(out, EvidenceFile{
            Filename: ev.Filename,
            FileType: ev.FileType,
            IpfsCID:  ev.IpfsCID,
            Content:  blob,
        })
    }

    return out, nil
}
