package case_tags

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	
)

type CaseTagRepository interface {
	AddTagsToCase(ctx context.Context, userID uuid.UUID, caseID uuid.UUID, tags []string) error
	RemoveTagsFromCase(ctx context.Context, userID uuid.UUID, caseID uuid.UUID, tags []string) error
	GetTagsForCase(ctx context.Context, caseID uuid.UUID) ([]string, error)
}

type caseTagRepo struct {
	db *gorm.DB
}

func NewCaseTagRepository(db *gorm.DB) CaseTagRepository {
	return &caseTagRepo{db: db}
}

func normalizeTag(tag string) string {
	return strings.ToLower(strings.TrimSpace(tag))
}

func (r *caseTagRepo) AddTagsToCase(ctx context.Context, userID, caseID uuid.UUID, tags []string) error {
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, raw := range tags {
		tagName := normalizeTag(raw)

		var tag Tag
		if err := tx.FirstOrCreate(&tag, Tag{Name: tagName}).Error; err != nil {
			tx.Rollback()
			return err
		}

		caseTag := CaseTag{CaseID: caseID, TagID: tag.ID}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&caseTag).Error; err != nil {
			tx.Rollback()
			return err
		}

		// TODO: Insert audit log here
	}

	return tx.Commit().Error
}

func (r *caseTagRepo) RemoveTagsFromCase(ctx context.Context, userID, caseID uuid.UUID, tags []string) error {
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, raw := range tags {
		tagName := normalizeTag(raw)

		var tag Tag
		if err := tx.First(&tag, "name = ?", tagName).Error; err != nil {
			continue // Tag doesn't exist — skip
		}

		if err := tx.Delete(&CaseTag{}, "case_id = ? AND tag_id = ?", caseID, tag.ID).Error; err != nil {
			tx.Rollback()
			return err
		}

		// TODO: Insert audit log here
	}

	return tx.Commit().Error
}

func (r *caseTagRepo) GetTagsForCase(ctx context.Context, caseID uuid.UUID) ([]string, error) {
	var tags []string

	err := r.db.WithContext(ctx).
		Table("tags").
		Select("tags.name").
		Joins("JOIN case_tags ON tags.id = case_tags.tag_id").
		Where("case_tags.case_id = ?", caseID).
		Scan(&tags).Error

	if err != nil {
		return nil, err
	}

	return tags, nil
}
