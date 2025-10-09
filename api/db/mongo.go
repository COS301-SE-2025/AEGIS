// package db

// import (
// 	"context"
// 	"fmt"
// 	"os"

// 	"go.mongodb.org/mongo-driver/mongo"
// 	"go.mongodb.org/mongo-driver/mongo/options"
// )

// var MongoClient *mongo.Client
// var MongoDatabase *mongo.Database

// func ConnectMongo() error {
// 	uri := os.Getenv("MONGO_URI")
// 	if uri == "" {
// 		return fmt.Errorf("MONGO_URI not set")
// 	}

// 	clientOptions := options.Client().ApplyURI(uri)
// 	client, err := mongo.Connect(context.TODO(), clientOptions)
// 	if err != nil {
// 		return fmt.Errorf("failed to connect to MongoDB: %w", err)
// 	}

// 	// Verify the connection
// 	err = client.Ping(context.TODO(), nil)
// 	if err != nil {
// 		return fmt.Errorf("MongoDB ping failed: %w", err)
// 	}

//		MongoClient = client
//		MongoDatabase = client.Database("app_database")
//		fmt.Println("✅ Connected to MongoDB successfully")
//		return nil
//	}
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

// MongoConnector interface for testing
type MongoConnector interface {
	Connect(ctx context.Context, opts ...*options.ClientOptions) (*mongo.Client, error)
	Ping(ctx context.Context, client *mongo.Client) error
}

// DefaultMongoConnector implements MongoConnector for production use
type DefaultMongoConnector struct{}

func (d *DefaultMongoConnector) Connect(ctx context.Context, opts ...*options.ClientOptions) (*mongo.Client, error) {
	return mongo.Connect(ctx, opts...)
}

func (d *DefaultMongoConnector) Ping(ctx context.Context, client *mongo.Client) error {
	return client.Ping(ctx, nil)
}

var connector MongoConnector = &DefaultMongoConnector{}

func ConnectMongo() error {
	return ConnectMongoWithConnector(connector)
}

// ConnectMongoWithConnector allows testing with a custom connector
func ConnectMongoWithConnector(conn MongoConnector) error {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		return fmt.Errorf("MONGO_URI not set")
	}

	clientOptions := options.Client().ApplyURI(uri)
	client, err := conn.Connect(context.TODO(), clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Verify the connection
	err = conn.Ping(context.TODO(), client)
	if err != nil {
		return fmt.Errorf("MongoDB ping failed: %w", err)
	}

	MongoClient = client
	MongoDatabase = client.Database("app_database")
	fmt.Println("✅ Connected to MongoDB successfully")
	return nil
}
