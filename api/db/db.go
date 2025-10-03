// package db

// import (
// 	"fmt"
// 	"log"
// 	"os"

// 	"github.com/joho/godotenv"

// 	"gorm.io/driver/postgres"
// 	"gorm.io/gorm"
// )

// var DB *gorm.DB

// func InitDB() error {
// 	if err := godotenv.Load(); err != nil {
// 		log.Println("Warning: .env file not loaded")
// 	}

// 	dsn := fmt.Sprintf(
// 		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
// 		os.Getenv("DB_HOST"),
// 		os.Getenv("DB_USER"),
// 		os.Getenv("DB_PASSWORD"),
// 		os.Getenv("DB_NAME"),
// 		os.Getenv("DB_PORT"),
// 	)

// 	var err error
// 	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
// 	if err != nil {
// 		return err
// 	}

//		log.Println("✅ Connected to the database!")
//		return nil
//	}
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
	return initDB(nil)
}

// InitDBWithDialector allows testing with a custom dialector
// This is exported specifically for unit testing purposes
func InitDBWithDialector(dialector gorm.Dialector) error {
	return initDB(dialector)
}

// initDB is the internal implementation
func initDB(dialector gorm.Dialector) error {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not loaded")
	}

	// If no dialector provided, create one from environment variables
	if dialector == nil {
		dsn := fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
			os.Getenv("DB_PORT"),
		)
		dialector = postgres.Open(dsn)
	}

	var err error
	DB, err = gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return err
	}

	log.Println("✅ Connected to the database!")
	return nil
}
