package db

import (
	"fmt"
	"log"
	"os"
	"github.com/joho/godotenv"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() error {
		if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not loaded")
	}
	
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)


	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	log.Println("✅ Connected to the database!")
	return nil
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

