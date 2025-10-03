package unit_tests

import (
	"aegis-api/db"
	"bytes"
	"context"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// mockMongoConnector is a mock implementation of MongoConnector
type mockMongoConnector struct {
	connectErr error
	pingErr    error
	client     *mongo.Client
}

func (m *mockMongoConnector) Connect(ctx context.Context, opts ...*options.ClientOptions) (*mongo.Client, error) {
	if m.connectErr != nil {
		return nil, m.connectErr
	}
	// Return a mock client (it won't be usable, but we don't need it for tests)
	if m.client != nil {
		return m.client, nil
	}
	// Create a minimal client for testing
	client, _ := mongo.NewClient(opts...)
	return client, nil
}

func (m *mockMongoConnector) Ping(ctx context.Context, client *mongo.Client) error {
	return m.pingErr
}

// captureStdout captures stdout output and returns the buffer and cleanup function
func captureStdout() (*bytes.Buffer, func()) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	buf := new(bytes.Buffer)

	return buf, func() {
		w.Close()
		buf.ReadFrom(r)
		os.Stdout = old
	}
}

func TestConnectMongo(t *testing.T) {
	// Reset global variables before each test
	defer func() {
		db.MongoClient = nil
		db.MongoDatabase = nil
	}()

	// Test case 1: Missing MONGO_URI environment variable
	t.Run("MissingMongoURI", func(t *testing.T) {
		db.MongoClient = nil
		db.MongoDatabase = nil

		cleanup := mockEnv(map[string]string{
			"MONGO_URI": "",
		})
		defer cleanup()

		mock := &mockMongoConnector{}
		err := db.ConnectMongoWithConnector(mock)

		require.Error(t, err, "ConnectMongo should return an error when MONGO_URI is not set")
		require.Equal(t, "MONGO_URI not set", err.Error(), "Error message should indicate missing MONGO_URI")
		require.Nil(t, db.MongoClient, "MongoClient should not be set")
		require.Nil(t, db.MongoDatabase, "MongoDatabase should not be set")
	})

	// Test case 2: Successful connection
	t.Run("Success", func(t *testing.T) {
		db.MongoClient = nil
		db.MongoDatabase = nil

		cleanup := mockEnv(map[string]string{
			"MONGO_URI": "mongodb://localhost:27017",
		})
		defer cleanup()

		// Capture stdout to verify success message
		r, w, _ := os.Pipe()
		old := os.Stdout
		os.Stdout = w

		mock := &mockMongoConnector{
			connectErr: nil,
			pingErr:    nil,
		}
		err := db.ConnectMongoWithConnector(mock)

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		require.NoError(t, err, "ConnectMongo should succeed with valid URI and mock")
		require.NotNil(t, db.MongoClient, "MongoClient should be set")
		require.NotNil(t, db.MongoDatabase, "MongoDatabase should be set")
		require.Equal(t, "app_database", db.MongoDatabase.Name(), "Database name should be 'app_database'")
		require.Contains(t, output, "✅ Connected to MongoDB successfully", "Should print success message")
	})

	// Test case 3: Connection failure
	t.Run("ConnectionFailure", func(t *testing.T) {
		db.MongoClient = nil
		db.MongoDatabase = nil

		cleanup := mockEnv(map[string]string{
			"MONGO_URI": "mongodb://localhost:27017",
		})
		defer cleanup()

		mock := &mockMongoConnector{
			connectErr: errors.New("connection refused"),
		}
		err := db.ConnectMongoWithConnector(mock)

		require.Error(t, err, "ConnectMongo should return an error on connection failure")
		require.Contains(t, err.Error(), "failed to connect to MongoDB", "Error should indicate connection failure")
		require.Contains(t, err.Error(), "connection refused", "Error should contain original error message")
		require.Nil(t, db.MongoClient, "MongoClient should not be set on error")
		require.Nil(t, db.MongoDatabase, "MongoDatabase should not be set on error")
	})

	// Test case 4: Ping failure
	t.Run("PingFailure", func(t *testing.T) {
		db.MongoClient = nil
		db.MongoDatabase = nil

		cleanup := mockEnv(map[string]string{
			"MONGO_URI": "mongodb://localhost:27017",
		})
		defer cleanup()

		mock := &mockMongoConnector{
			connectErr: nil,
			pingErr:    errors.New("server selection timeout"),
		}
		err := db.ConnectMongoWithConnector(mock)

		require.Error(t, err, "ConnectMongo should return an error on ping failure")
		require.Contains(t, err.Error(), "MongoDB ping failed", "Error should indicate ping failure")
		require.Contains(t, err.Error(), "server selection timeout", "Error should contain original error message")
		require.Nil(t, db.MongoClient, "MongoClient should not be set on ping error")
		require.Nil(t, db.MongoDatabase, "MongoDatabase should not be set on ping error")
	})

	// Test case 5: Invalid URI format
	t.Run("InvalidURIFormat", func(t *testing.T) {
		db.MongoClient = nil
		db.MongoDatabase = nil

		cleanup := mockEnv(map[string]string{
			"MONGO_URI": "invalid://uri",
		})
		defer cleanup()

		mock := &mockMongoConnector{
			connectErr: errors.New("error parsing uri: invalid scheme"),
		}
		err := db.ConnectMongoWithConnector(mock)

		require.Error(t, err, "ConnectMongo should return an error for invalid URI")
		require.Contains(t, err.Error(), "failed to connect to MongoDB", "Error should indicate connection failure")
	})

	// Test case 6: Empty URI (different from missing)
	t.Run("EmptyURI", func(t *testing.T) {
		db.MongoClient = nil
		db.MongoDatabase = nil

		// Set the env var to empty string explicitly
		os.Setenv("MONGO_URI", "")
		defer os.Unsetenv("MONGO_URI")

		mock := &mockMongoConnector{}
		err := db.ConnectMongoWithConnector(mock)

		require.Error(t, err, "ConnectMongo should return an error when MONGO_URI is empty")
		require.Equal(t, "MONGO_URI not set", err.Error(), "Error message should indicate missing MONGO_URI")
	})
}

// Integration test (requires actual MongoDB instance)
func TestConnectMongo_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db.MongoClient = nil
	db.MongoDatabase = nil

	cleanup := mockEnv(map[string]string{
		"MONGO_URI": "mongodb://localhost:27017",
	})
	defer cleanup()

	err := db.ConnectMongo()
	if err != nil {
		t.Skipf("Skipping integration test: MongoDB not available: %v", err)
	}

	require.NotNil(t, db.MongoClient, "MongoClient should be initialized")
	require.NotNil(t, db.MongoDatabase, "MongoDatabase should be initialized")

	// Test that we can actually ping the database
	err = db.MongoClient.Ping(context.TODO(), nil)
	require.NoError(t, err, "Should be able to ping MongoDB")

	// Clean up connection
	if db.MongoClient != nil {
		db.MongoClient.Disconnect(context.TODO())
	}
}
