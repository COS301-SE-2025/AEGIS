package messages

import (
    "context"
    "fmt"
    "os"
    "testing"
    "time"

    "aegis-api/pkg/encryption"
    "github.com/google/uuid"
    "github.com/testcontainers/testcontainers-go"
    tcPostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
    "github.com/testcontainers/testcontainers-go/wait"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

var (
    db          *gorm.DB
    pgContainer testcontainers.Container
)

func TestMain(m *testing.M) {
    ctx := context.Background()
    
    // Set encryption key
    os.Setenv("ENCRYP_REST_MASTER_KEY", "0123456789abcdef0123456789abcdef")
    if err := encryption.Init(); err != nil {
        fmt.Println("Failed to initialize encryption:", err)
        os.Exit(1)
    }
    
    // Start PostgreSQL container
    var err error
    pgContainer, db, err = startPostgres(ctx)
    if err != nil {
        fmt.Println("Failed to start PostgreSQL:", err)
        os.Exit(1)
    }
    
    // Enable UUID extension
    if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
        fmt.Println("Failed to enable UUID extension:", err)
        terminateContainer(pgContainer, ctx)
        os.Exit(1)
    }
    
    // Run migrations
    if err := db.AutoMigrate(&ThreadMessage{}); err != nil {
        fmt.Println("Failed to migrate database:", err)
        terminateContainer(pgContainer, ctx)
        os.Exit(1)
    }
    
    // Run tests
    code := m.Run()
    
    // Cleanup
    terminateContainer(pgContainer, ctx)
    os.Exit(code)
}

func startPostgres(ctx context.Context) (testcontainers.Container, *gorm.DB, error) {
    // Start PostgreSQL container
    container, err := tcPostgres.Run(ctx,
        "postgres:15-alpine",
        tcPostgres.WithDatabase("testdb"),
        tcPostgres.WithUsername("postgres"),
        tcPostgres.WithPassword("password"),
        testcontainers.WithWaitStrategy(
            wait.ForLog("database system is ready to accept connections").
                WithOccurrence(2).
                WithStartupTimeout(30*time.Second)),
    )
    if err != nil {
        return nil, nil, err
    }
    
    // Get connection string
    connStr, err := container.ConnectionString(ctx)
    if err != nil {
        return container, nil, err
    }
    
    // Connect to database
    gormDB, err := gorm.Open(postgres.Open(connStr), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Silent),
    })
    if err != nil {
        return container, nil, err
    }
    
    return container, gormDB, nil
}

func terminateContainer(container testcontainers.Container, ctx context.Context) {
    if container != nil {
        if err := container.Terminate(ctx); err != nil {
            fmt.Println("Failed to terminate container:", err)
        }
    }
}

func TestMessageEncryptionHooks(t *testing.T) {
	
	original := "This is a secret"
	msg := &ThreadMessage{
		ID:        uuid.New(),
		ThreadID:  uuid.New(),
		UserID:    uuid.New(),
		Message:   original,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	if err := db.Create(msg).Error; err != nil {
		t.Fatalf("failed to create: %v", err)
	}

	// Query raw from DB (bypass hooks)
	var raw string
	db.Raw("SELECT message FROM thread_messages WHERE id = ?", msg.ID).Scan(&raw)
	if raw == original {
		t.Error("message was not encrypted in DB")
	}

	// Query via GORM (hooks trigger AfterFind)
	var fetched ThreadMessage
	if err := db.First(&fetched, "id = ?", msg.ID).Error; err != nil {
		t.Fatalf("failed to fetch: %v", err)
	}
	if fetched.Message != original {
		t.Errorf("expected %s, got %s", original, fetched.Message)
	}
}