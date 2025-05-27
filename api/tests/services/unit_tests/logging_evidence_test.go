package unit_tests

import (
	"context"
	"testing"

	"aegis-api/services/evidence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"github.com/google/uuid"
	"time"
)

// MongoCollection defines only the methods we use on mongo.Collection for insertion
// This interface should live alongside your repository implementation.
type MongoCollection interface {
	InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
}

// MockCollection is a testify mock for MongoCollection
type MockCollection struct {
	mock.Mock
}

func (m *MockCollection) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	args := m.Called(ctx, document)
	return args.Get(0).(*mongo.InsertOneResult), args.Error(1)
}

// In your repository, add a constructor that accepts MongoCollection:
// func NewMongoEvidenceRepositoryWithCollection(coll MongoCollection) *MongoEvidenceRepository

func TestMongoEvidenceRepository_SaveEvidence(t *testing.T) {
	// Arrange
	mockColl := new(MockCollection)
	repo := evidence.NewMongoEvidenceRepositoryWithCollection(mockColl)

	e := &evidence.Evidence{
		ID:         uuid.New(),
		CaseID:     uuid.New(),
		UploadedBy: uuid.New(),
		Filename:   "unit_test.txt",
		FileType:   "text/plain",
		IpfsCID:    "QmTestCID",
		FileSize:   1024,
		Checksum:   "checksum123",
		Metadata:   map[string]interface{}{"unit": true},
		UploadedAt: time.Now(),
	}

	inserted := primitive.NewObjectID()
	mockColl.On("InsertOne", mock.Anything, e).Return(&mongo.InsertOneResult{InsertedID: inserted}, nil)

	// Act
	err := repo.SaveEvidence(e)

	// Assert
	assert.NoError(t, err)
	mockColl.AssertCalled(t, "InsertOne", mock.Anything, e)
}