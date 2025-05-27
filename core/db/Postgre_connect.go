package db

import (
	"context"
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)



// ConnectPostgres establishes a connection to PostgreSQL
func ConnectPostgres(connStr string) *sql.DB {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Open connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("❌ PostgreSQL connection failed: %v", err)
	}

	// Verify connection
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("❌ PostgreSQL ping failed: %v", err)
	}

	log.Println("✅ Connected to PostgreSQL")
	return db
}