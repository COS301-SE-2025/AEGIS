package report_ai_assistance

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"

	"aegis-api/services_/report"
)

type MongoSectionRepository struct {
	db   *mongo.Database
	pgDB *gorm.DB
}

func NewMongoSectionRepository(db *mongo.Database) *MongoSectionRepository {
	// Accept pgDB as second argument
	return &MongoSectionRepository{db: db, pgDB: nil}
}

func NewMongoSectionRepositoryWithPg(db *mongo.Database, pgDB *gorm.DB) *MongoSectionRepository {
	return &MongoSectionRepository{db: db, pgDB: pgDB}
}

// ReportContentMongo and ReportSection should match your MongoDB models
func (r *MongoSectionRepository) GetSectionByID(ctx context.Context, reportID primitive.ObjectID, sectionID primitive.ObjectID) (*report.ReportSection, error) {
	var report report.ReportContentMongo
	err := r.db.Collection("report_contents").FindOne(ctx, bson.M{"_id": reportID}).Decode(&report)
	if err != nil {
		return nil, err
	}
	for _, section := range report.Sections {
		if section.ID == sectionID {
			return &section, nil
		}
	}
	return nil, errors.New("section not found")
}
