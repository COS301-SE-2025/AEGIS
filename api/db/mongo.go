package db

import (
	"context"
	"fmt"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client
var MongoDatabase *mongo.Database

func ConnectMongo() error {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		return fmt.Errorf("MONGO_URI not set")
	}

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Verify the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return fmt.Errorf("MongoDB ping failed: %w", err)
	}

	MongoClient = client
	MongoDatabase = client.Database("app_database")
	fmt.Println("✅ Connected to MongoDB successfully")
	return nil
}
