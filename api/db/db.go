package db

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() error {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not loaded")
	}

	// Set defaults for local development
	host := getEnvOrDefault("DB_HOST", "postgres")
	port := getEnvOrDefault("DB_PORT", "5432")
	user := getEnvOrDefault("DB_USER", "app_user")
	password := getEnvOrDefault("DB_PASSWORD", "dev_secure_password123")
	dbname := getEnvOrDefault("DB_NAME", "app_database")

	log.Printf("Attempting database connection to %s:%s as user %s for database %s", host, port, user, dbname)

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to initialize database, got error %w", err)
	}

	//err = DB.AutoMigrate(&registration.User{})
	//if err != nil {
	//	return fmt.Errorf("failed to auto-migrate database, got error %w", err)
	//}

	log.Println("✅ Connected to the database!")
	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// package db

// import (
// 	"fmt"
// 	"gorm.io/driver/postgres"
// 	"gorm.io/gorm"
// 	"gorm.io/gorm/logger"

// )

// var DB *gorm.DB

// func Connect() error {
// 	dsn := fmt.Sprintf("host=localhost user=app_user password=dev_secure_password123 dbname=app_database port=5432 sslmode=disable")
// 	var err error
// 		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
// 		Logger: logger.Default.LogMode(logger.Silent), // Suppress info/debug logs
// 	})
// 	return err
// }
