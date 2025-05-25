package main

import (
	"fmt"
	"aegis-api/db"
	"aegis-api/services/auditlog"
)

func main() {
	err := db.ConnectMongo()
	if err != nil {
		panic("Mongo connection failed: " + err.Error())
	}

	err = auditlog.Log(
		"test",
		"evidence",
		"123e4567-e89b-12d3-a456-426614174000",
		"ded0a1b3-4712-46b5-8d01-fafbaf3f8236",
		"This is a test log entry",
	)
	if err != nil {
		fmt.Println("❌ Failed to log:", err)
	} else {
		fmt.Println("✅ Log entry saved to MongoDB")
	}
}
