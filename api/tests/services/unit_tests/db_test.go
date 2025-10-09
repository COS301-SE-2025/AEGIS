package unit_tests

import (
	"aegis-api/db"
	"bytes"
	"database/sql"
	"errors"
	"log"
	"os"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// mockEnv sets environment variables for testing and returns a cleanup function
func mockEnv(vars map[string]string) func() {
	original := make(map[string]string)
	for k, v := range vars {
		original[k] = os.Getenv(k)
		if v == "" {
			os.Unsetenv(k)
		} else {
			os.Setenv(k, v)
		}
	}
	return func() {
		for k, v := range original {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}
}

// captureLogOutput captures log output to a buffer and returns the buffer and a cleanup function
func captureLogOutput() (*bytes.Buffer, func()) {
	var buf bytes.Buffer
	original := log.Writer()
	log.SetOutput(&buf)
	return &buf, func() {
		log.SetOutput(original)
	}
}

// createMockDB creates a mock database and GORM dialector for testing
func createMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, gorm.Dialector) {
	// Use MonitorPingsOption to prevent automatic ping
	sqlDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)

	dialector := postgres.New(postgres.Config{
		Conn:                 sqlDB,
		DriverName:           "postgres",
		PreferSimpleProtocol: true,
	})

	return sqlDB, mock, dialector
}

func TestInitDB(t *testing.T) {
	// Reset global DB before each test
	defer func() { db.DB = nil }()

	// Test case 1: Missing environment variables (invalid DSN)
	t.Run("MissingEnvVars", func(t *testing.T) {
		db.DB = nil

		cleanup := mockEnv(map[string]string{
			"DB_HOST":     "",
			"DB_USER":     "",
			"DB_PASSWORD": "",
			"DB_NAME":     "",
			"DB_PORT":     "",
		})
		defer cleanup()

		buf, logCleanup := captureLogOutput()
		defer logCleanup()

		err := db.InitDB()
		require.Error(t, err, "InitDB should return an error for invalid DSN")
		require.Contains(t, buf.String(), "Warning: .env file not loaded", "Log should contain warning")
	})

	// Test case 2: Successful initialization with mock DB
	t.Run("Success", func(t *testing.T) {
		db.DB = nil

		cleanup := mockEnv(map[string]string{
			"DB_HOST":     "localhost",
			"DB_USER":     "testuser",
			"DB_PASSWORD": "testpass",
			"DB_NAME":     "testdb",
			"DB_PORT":     "5432",
		})
		defer cleanup()

		sqlDB, mock, dialector := createMockDB(t)
		defer sqlDB.Close()

		// GORM will ping the database to check connection
		mock.ExpectPing()

		buf, logCleanup := captureLogOutput()
		defer logCleanup()

		err := db.InitDBWithDialector(dialector)
		require.NoError(t, err, "InitDB should succeed with mocked DB")
		require.NotNil(t, db.DB, "DB should be initialized")
		require.Contains(t, buf.String(), "✅ Connected to the database!", "Log should contain success message")
		require.NoError(t, mock.ExpectationsWereMet(), "All mock expectations should be met")
	})

	// Test case 3: Connection failure (mock returns error on ping)
	t.Run("ConnectionFailure", func(t *testing.T) {
		db.DB = nil

		cleanup := mockEnv(map[string]string{
			"DB_HOST":     "localhost",
			"DB_USER":     "testuser",
			"DB_PASSWORD": "testpass",
			"DB_NAME":     "testdb",
			"DB_PORT":     "5432",
		})
		defer cleanup()

		sqlDB, mock, dialector := createMockDB(t)
		defer sqlDB.Close()

		// Expect ping to fail with connection error
		mock.ExpectPing().WillReturnError(errors.New("connection refused"))

		buf, logCleanup := captureLogOutput()
		defer logCleanup()

		err := db.InitDBWithDialector(dialector)
		require.Error(t, err, "InitDB should return an error for connection failure")
		require.Contains(t, err.Error(), "connection refused", "Error should indicate connection failure")
		require.Contains(t, buf.String(), "Warning: .env file not loaded", "Log should contain warning")
		require.NoError(t, mock.ExpectationsWereMet(), "All mock expectations should be met")
	})

	// Test case 4: Database closed immediately
	t.Run("DatabaseClosed", func(t *testing.T) {
		db.DB = nil

		cleanup := mockEnv(map[string]string{
			"DB_HOST":     "localhost",
			"DB_USER":     "testuser",
			"DB_PASSWORD": "testpass",
			"DB_NAME":     "testdb",
			"DB_PORT":     "5432",
		})
		defer cleanup()

		sqlDB, mock, dialector := createMockDB(t)

		// Expect ping, but return "database is closed" error
		mock.ExpectPing().WillReturnError(sql.ErrConnDone)

		buf, logCleanup := captureLogOutput()
		defer logCleanup()

		err := db.InitDBWithDialector(dialector)
		require.Error(t, err, "InitDB should return an error when database is closed")
		require.Contains(t, buf.String(), "Warning: .env file not loaded", "Log should contain warning")

		sqlDB.Close()
		require.NoError(t, mock.ExpectationsWereMet(), "All mock expectations should be met")
	})
}

// Integration test (requires actual database)
func TestInitDB_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db.DB = nil

	cleanup := mockEnv(map[string]string{
		"DB_HOST":     "localhost",
		"DB_USER":     "postgres",
		"DB_PASSWORD": "postgres",
		"DB_NAME":     "testdb",
		"DB_PORT":     "5432",
	})
	defer cleanup()

	err := db.InitDB()
	if err != nil {
		t.Skipf("Skipping integration test: database not available: %v", err)
	}

	require.NotNil(t, db.DB, "DB should be initialized")

	// Test that we can actually query the database
	sqlDB, err := db.DB.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Ping())
}
