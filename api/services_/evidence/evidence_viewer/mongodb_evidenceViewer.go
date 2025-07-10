package evidence_viewer

import (
	"context"
	"time"


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

func (repo *MongoEvidenceRepository) GetEvidenceFilesByCaseID(caseID string) ([]EvidenceFile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"case_id": caseID}
	projection := bson.M{
		"id":   1,
		"data": 1,
	}

	opts := options.Find().SetProjection(projection)
	cursor, err := repo.Collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rawResults []struct {
		ID   string `bson:"id"`
		Data []byte `bson:"data"`
	}
	if err := cursor.All(ctx, &rawResults); err != nil {
		return nil, err
	}

	var results []EvidenceFile
	for _, r := range rawResults {
		results = append(results, EvidenceFile{
			ID:   r.ID,
			Data: r.Data,
		})
	}
	return results, nil
}


func (repo *MongoEvidenceRepository) GetEvidenceFileByID(evidenceID string) (*EvidenceFile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"id": evidenceID}
	projection := bson.M{
		"id":   1,
		"data": 1,
	}

	opts := options.FindOne().SetProjection(projection)

	var result struct {
		ID   string `bson:"id"`
		Data []byte `bson:"data"`
	}

	err := repo.Collection.FindOne(ctx, filter, opts).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &EvidenceFile{
		ID:   result.ID,
		Data: result.Data,
	}, nil
}


func (repo *MongoEvidenceRepository) SearchEvidenceFiles(query string) ([]EvidenceFile, error) {
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
		"id":   1,
		"data": 1,
	}

	opts := options.Find().SetProjection(projection)
	cursor, err := repo.Collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rawResults []struct {
		ID   string `bson:"id"`
		Data []byte `bson:"data"`
	}
	if err := cursor.All(ctx, &rawResults); err != nil {
		return nil, err
	}

	var results []EvidenceFile
	for _, r := range rawResults {
		results = append(results, EvidenceFile{
			ID:   r.ID,
			Data: r.Data,
		})
	}
	return results, nil
}


func (repo *MongoEvidenceRepository) GetFilteredEvidenceFiles(caseID string, filters map[string]interface{}, sortField string, sortOrder string) ([]EvidenceFile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"case_id": caseID}
	for k, v := range filters {
		filter[k] = v
	}

	projection := bson.M{
		"id":   1,
		"data": 1,
	}

	opts := options.Find().SetProjection(projection)
	if sortField != "" && (sortOrder == "asc" || sortOrder == "desc") {
		direction := 1
		if sortOrder == "desc" {
			direction = -1
		}
		opts.SetSort(bson.D{{Key: sortField, Value: direction}})
	}

	cursor, err := repo.Collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rawResults []struct {
		ID   string `bson:"id"`
		Data []byte `bson:"data"`
	}
	if err := cursor.All(ctx, &rawResults); err != nil {
		return nil, err
	}

	var results []EvidenceFile
	for _, r := range rawResults {
		results = append(results, EvidenceFile{
			ID:   r.ID,
			Data: r.Data,
		})
	}
	return results, nil
}
