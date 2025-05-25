package evidence

import (
	"aegis-api/db"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"errors"
	"time"
)

type Service struct{}

func NewEvidenceService() *Service {
	return &Service{}
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

	evidence := Evidence{
		CaseID:     caseID,
		UploadedBy: userID,
		Filename:   req.Filename,
		FileType:   req.FileType,
		IpfsCID:    req.IpfsCID,
		FileSize:   req.FileSize,
		Checksum:   req.Checksum,
		Metadata:   req.Metadata,
	}

	// First insert the evidence
	if err := db.DB.Create(&evidence).Error; err != nil {
		return nil, fmt.Errorf("failed to store evidence: %w", err)
	}

	// Handle tags
	if len(req.Tags) > 0 {
		var tags []Tag
		for _, tagName := range req.Tags {
			tag := Tag{Name: tagName}
			// Insert or fetch existing tag
			if err := db.DB.FirstOrCreate(&tag, Tag{Name: tagName}).Error; err != nil {
				return nil, fmt.Errorf("failed to store tag: %w", err)
			}
			tags = append(tags, tag)
		}
		if err := db.DB.Model(&evidence).Association("Tags").Append(&tags); err != nil {
			return nil, fmt.Errorf("failed to link tags: %w", err)
		}
	}
		// Log the upload to MongoDB
	err = LogEvidenceUpload(userID.String(), evidence.ID.String(), evidence.Filename)
	if err != nil {
		fmt.Printf("⚠️ Failed to log upload: %v\n", err)
	}

	return &evidence, nil
}

func (s *Service) ListEvidenceByCase(caseID string) ([]Evidence, error) {
	id, err := uuid.Parse(caseID)
	if err != nil {
		return nil, fmt.Errorf("invalid case ID: %w", err)
	}

	var results []Evidence
	if err := db.DB.Where("case_id = ?", id).Find(&results).Error; err != nil {
		return nil, err
	}

	return results, nil
}

func (s *Service) ListEvidenceByUser(userID string) ([]Evidence, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	var results []Evidence
	if err := db.DB.Where("uploaded_by = ?", id).Find(&results).Error; err != nil {
		return nil, err
	}

	return results, nil
}

func (s *Service) GetEvidenceByID(evidenceID string) (*Evidence, error) {
	id, err := uuid.Parse(evidenceID)
	if err != nil {
		return nil, fmt.Errorf("invalid evidence ID: %w", err)
	}

	var ev Evidence
	if err := db.DB.First(&ev, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("evidence not found")
		}
		return nil, err
	}

	return &ev, nil
}
func (s *Service) DeleteEvidenceByID(evidenceID string) error {
	id, err := uuid.Parse(evidenceID)
	if err != nil {
		return fmt.Errorf("invalid evidence ID: %w", err)
	}

	result := db.DB.Delete(&Evidence{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("evidence not found")
	}

	return nil
}

func (s *Service) GetEvidenceMetadata(evidenceID string) (*EvidenceMetadata, error) {
	id, err := uuid.Parse(evidenceID)
	if err != nil {
		return nil, fmt.Errorf("invalid evidence ID: %w", err)
	}

	var ev Evidence
	if err := db.DB.Preload("Tags").First(&ev, "id = ?", id).Error; err != nil {
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
