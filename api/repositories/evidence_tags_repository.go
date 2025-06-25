package repositories

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"aegis-api/models"
)

type EvidenceTagRepository interface {
	AddTagsToEvidence(ctx context.Context, userID, evidenceID uuid.UUID, tags []string) error
	RemoveTagsFromEvidence(ctx context.Context, userID, evidenceID uuid.UUID, tags []string) error
	GetTagsForEvidence(ctx context.Context, evidenceID uuid.UUID) ([]string, error)
}

type evidenceTagRepository struct {
	db *gorm.DB
}

func NewEvidenceTagRepository(db *gorm.DB) EvidenceTagRepository {
	return &evidenceTagRepository{db: db}
}

func (r *evidenceTagRepository) AddTagsToEvidence(ctx context.Context, userID, evidenceID uuid.UUID, tags []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, tagName := range tags {
			normalized := strings.TrimSpace(strings.ToLower(tagName))
			var tag models.Tag
			if err := tx.Where("name = ?", normalized).First(&tag).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					tag = models.Tag{Name: normalized}
					if err := tx.Create(&tag).Error; err != nil {
						return err
					}
				} else {
					return err
				}
			}

			evidenceTag := models.EvidenceTag{
				EvidenceID: evidenceID,
				TagID:      tag.ID,
			}
			if err := tx.Clauses(
				gorm.Clauses{gorm.Expr("ON CONFLICT DO NOTHING")},
			).Create(&evidenceTag).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *evidenceTagRepository) RemoveTagsFromEvidence(ctx context.Context, userID, evidenceID uuid.UUID, tags []string) error {
	for _, tagName := range tags {
		normalized := strings.TrimSpace(strings.ToLower(tagName))
		var tag models.Tag
		if err := r.db.WithContext(ctx).Where("name = ?", normalized).First(&tag).Error; err != nil {
			continue // silently skip if tag doesn't exist
		}
		r.db.Where("evidence_id = ? AND tag_id = ?", evidenceID, tag.ID).
			Delete(&models.EvidenceTag{})
	}
	return nil
}

func (r *evidenceTagRepository) GetTagsForEvidence(ctx context.Context, evidenceID uuid.UUID) ([]string, error) {
	var tags []string
	err := r.db.WithContext(ctx).
		Table("tags").
		Select("tags.name").
		Joins("JOIN evidence_tags ON tags.id = evidence_tags.tag_id").
		Where("evidence_tags.evidence_id = ?", evidenceID).
		Scan(&tags).Error
	return tags, err
}
