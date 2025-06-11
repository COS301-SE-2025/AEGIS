// file: mongo_repository.go
package evidence

import (
	"aegis-api/db"
	"context"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoCollection defines the subset of mongo.Collection methods used by the repository.
type MongoCollection interface {
	InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
	// Add other methods if needed, e.g. Find, Delete, etc.
}

// MongoEvidenceRepository handles persistence in MongoDB.
type MongoEvidenceRepository struct {
	Collection MongoCollection
}

// MongoEvidenceLogger is a struct for logging evidence actions.
type MongoEvidenceLogger struct{}

func NewMongoEvidenceLogger() *MongoEvidenceLogger {
	return &MongoEvidenceLogger{}
}

func (l *MongoEvidenceLogger) Log(userID, evidenceID, filename string) error {
	// calls your existing function in logs.go:
	return LogEvidenceUpload(userID, evidenceID, filename)
}

// NewMongoEvidenceRepository uses the live MongoDB collection.
func NewMongoEvidenceRepository() *MongoEvidenceRepository {
	return &MongoEvidenceRepository{
		Collection: db.MongoDatabase.Collection("evidence"),
	}
}

// NewMongoEvidenceRepositoryWithCollection allows injecting a mock collection for tests.
func NewMongoEvidenceRepositoryWithCollection(coll MongoCollection) *MongoEvidenceRepository {
	return &MongoEvidenceRepository{Collection: coll}
}

// SaveEvidence inserts an Evidence document into MongoDB, assigning a UUID if missing.
func (r *MongoEvidenceRepository) SaveEvidence(e *Evidence) error {
	// Ensure the evidence has a valid ID
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}

	_, err := r.Collection.InsertOne(context.Background(), e)
	return err
}

// AttachTags attaches tags to the evidence; implemented by repository.
func (r *MongoEvidenceRepository) AttachTags(e *Evidence, tags []string) error {
	// Implementation should insert into evidence_tags table
	// This is a stub for Mongo; adjust if using a different storage model.
	return nil
}

// FindByID retrieves an Evidence document by its UUID.
func (r *MongoEvidenceRepository) FindByID(id uuid.UUID) (*Evidence, error) {
	// stub: replace with real find logic
	return nil, nil
}

// DeleteByID deletes an Evidence document by its UUID.
func (r *MongoEvidenceRepository) DeleteByID(id uuid.UUID) error {
	// stub: replace with real delete logic
	return nil
}

// FindByCase retrieves all Evidence for a given case.
func (r *MongoEvidenceRepository) FindByCase(caseID uuid.UUID) ([]Evidence, error) {
	// stub: replace with real query logic
	return nil, nil
}

// FindByUser retrieves all Evidence uploaded by a given user.
func (r *MongoEvidenceRepository) FindByUser(userID uuid.UUID) ([]Evidence, error) {
	// stub: replace with real query logic
	return nil, nil
}

// PreloadMetadata loads an Evidence document with its metadata populated.
func (r *MongoEvidenceRepository) PreloadMetadata(id uuid.UUID) (*Evidence, error) {
	// stub: replace with real lookup and metadata processing
	return nil, nil
}
