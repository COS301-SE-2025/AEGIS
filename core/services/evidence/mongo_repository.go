package evidence

import (
	"aegis-api/db"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoEvidenceRepository struct {
	Collection *mongo.Collection
}

func NewMongoEvidenceRepository() *MongoEvidenceRepository {
	return &MongoEvidenceRepository{
		Collection: db.MongoDatabase.Collection("evidence"),
	}
}
