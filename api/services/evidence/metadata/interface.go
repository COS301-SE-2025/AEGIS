package metadata

// Repository defines the interface for interacting with the metadata storage.
// It includes methods for saving evidence metadata and attaching tags to evidence.
type Repository interface {
	SaveMetadata(e *Evidence) error
	AttachTags(e *Evidence, tags []string) error
}
