package registration

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)

func InitDB() *gorm.DB {
	// DSN = Data Source Name
	dsn := "host=localhost user=postgres password=your_password dbname=usersdb port=5432 sslmode=disable TimeZone=Africa/Johannesburg"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}

	// Auto-create the users table if not exists
	db.AutoMigrate(&UserEntity{})

	db.AutoMigrate(&registration.UserEntity{}) // ensures unique index is created


	return db
}
 