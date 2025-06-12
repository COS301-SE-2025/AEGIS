package metadata

import (
	"gorm.io/gorm"
)

// Evidence represents the metadata for an uploaded evidence file.
// It includes fields for the case ID, uploader, filename, file type, IPFS CID,
type GormMetadataRepository struct {
	db *gorm.DB
}

// NewGormMetadataRepository creates a new instance of GormMetadataRepository with the provided gorm.DB instance.
// It initializes the repository for interacting with the metadata storage.
func NewGormMetadataRepository(db *gorm.DB) *GormMetadataRepository {
	return &GormMetadataRepository{db: db}
}

// SaveMetadata saves the evidence metadata to the database.
// It takes a pointer to Evidence and returns an error if the operation fails.
func (r *GormMetadataRepository) SaveMetadata(e *Evidence) error {
	return r.db.Create(e).Error
}

// AttachTags attaches tags to the given evidence.
// It takes a pointer to Evidence and a slice of tag names.
func (r *GormMetadataRepository) AttachTags(e *Evidence, tags []string) error {
	var tagEntities []Tag
	for _, tagName := range tags {
		var tag Tag
		if err := r.db.FirstOrCreate(&tag, Tag{Name: tagName}).Error; err != nil {
			return err
		}
		tagEntities = append(tagEntities, tag)
	}
	return r.db.Model(e).Association("Tags").Replace(tagEntities)
}
