package evidence

import (
	"context"
	//"fmt"
	//"github.com/google/uuid"
	"aegis-api/db"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoCollection defines the subset of mongo.Collection methods used by the repository.
type MongoCollection interface {
	InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
}

// MongoEvidenceRepository handles persistence in MongoDB.
type MongoEvidenceRepository struct {
	Collection MongoCollection
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

// SaveEvidence inserts an Evidence document into MongoDB.
func (r *MongoEvidenceRepository) SaveEvidence(e *Evidence) error {
	_, err := r.Collection.InsertOne(context.Background(), e)
	return err
}

// AttachTags etc... (other methods remain unchanged)
