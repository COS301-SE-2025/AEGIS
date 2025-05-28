package Evidence_Viewer

import (
	"context"
	"log"
	"time"

	"aegis-api/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"


)

// Ensure MongoEvidenceRepository implements the interface
var _ EvidenceViewer = (*MongoEvidenceRepository)(nil)

type MongoEvidenceRepository struct {
	Collection EvidenceCollection
}



func NewMongoEvidenceRepository(client *mongo.Client, dbName, collectionName string) *MongoEvidenceRepository {
    collection := client.Database(dbName).Collection(collectionName)
    return &MongoEvidenceRepository{Collection: &RealCollection{Collection: collection}}
}

// GetEvidenceByCase returns all evidence items for a specific case ID.
func (repo *MongoEvidenceRepository) GetEvidenceByCase(caseID string) ([]models.EvidenceResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"case_id": caseID}
	cursor, err := repo.Collection.Find(ctx, filter)
	if err != nil {
		log.Printf("Error finding evidence by case: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var evidences []models.EvidenceResponse
	if err = cursor.All(ctx, &evidences); err != nil {
		log.Printf("Error decoding evidence results: %v", err)
		return nil, err
	}

	return evidences, nil
}

// GetEvidenceByID returns a single evidence item by ID.
func (repo *MongoEvidenceRepository) GetEvidenceByID(evidenceID string) (*models.EvidenceResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"id": evidenceID}
	var ev models.EvidenceResponse
	err := repo.Collection.FindOne(ctx, filter).Decode(&ev)

	if err == mongo.ErrNoDocuments {
		log.Printf("No evidence found for ID: %s", evidenceID)
		return nil, nil
	} else if err != nil {
		log.Printf("Error retrieving evidence by ID: %v", err)
		return nil, err
	}

	return &ev, nil
}

// SearchEvidence performs a case-insensitive search on multiple fields.
func (repo *MongoEvidenceRepository) SearchEvidence(query string) ([]models.EvidenceResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	regex := bson.M{"$regex": query, "$options": "i"}
	filter := bson.M{
		"$or": []bson.M{
			{"filename": regex},
			{"file_type": regex},
			{"metadata": regex},
		},
	}

	projection := bson.M{
		"id":        1,
		"filename":  1,
		"file_type": 1,
		"ipfs_cid":  1,
	}

	opts := options.Find().SetProjection(projection)
	cursor, err := repo.Collection.Find(ctx, filter, opts)
	if err != nil {
		log.Printf("Search error: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []models.EvidenceResponse
	if err = cursor.All(ctx, &results); err != nil {
		log.Printf("Error decoding search results: %v", err)
		return nil, err
	}

	return results, nil
}

// GetFilteredEvidence returns evidence with filters and sorting.
func (repo *MongoEvidenceRepository) GetFilteredEvidence(caseID string, filters map[string]interface{}, sortField string, sortOrder string) ([]models.EvidenceResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Defensive copy of filters
	queryFilter := bson.M{"case_id": caseID}
	for k, v := range filters {
		queryFilter[k] = v
	}

	findOptions := options.Find()

	// Add sorting if provided
	if sortField != "" && (sortOrder == "asc" || sortOrder == "desc") {
		direction := 1
		if sortOrder == "desc" {
			direction = -1
		}
		findOptions.SetSort(bson.D{{Key: sortField, Value: direction}})
	}

	projection := bson.M{
		"id":        1,
		"filename":  1,
		"file_type": 1,
		"ipfs_cid":  1,
	}
	findOptions.SetProjection(projection)

	cursor, err := repo.Collection.Find(ctx, queryFilter, findOptions)
	if err != nil {
		log.Printf("Filter query failed: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []models.EvidenceResponse
	if err = cursor.All(ctx, &results); err != nil {
		log.Printf("Error decoding filtered results: %v", err)
		return nil, err
	}

	return results, nil
}


