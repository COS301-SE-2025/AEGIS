// File: services/evidence/logs.go
package evidence

import (
	"context"
	"time"
	"log"

	"aegis-api/db"
	//"go.mongodb.org/mongo-driver/mongo"
)

type UploadLog struct {
	UserID     string    `bson:"user_id"`
	EvidenceID string    `bson:"evidence_id"`
	Filename   string    `bson:"filename"`
	Timestamp  time.Time `bson:"timestamp"`
	Action     string    `bson:"action"`
}

func LogEvidenceUpload(userID, evidenceID, filename string) error {
	if db.MongoDatabase == nil {
		log.Println("⚠️ MongoDatabase is not initialized — skipping upload log.")
		return nil // or return an error if strict logging is needed
	}

	logEntry := UploadLog{
		UserID:     userID,
		EvidenceID: evidenceID,
		Filename:   filename,
		Timestamp:  time.Now(),
		Action:     "upload",
	}

	collection := db.MongoDatabase.Collection("upload_logs")
	_, err := collection.InsertOne(context.TODO(), logEntry)
	if err != nil {
		log.Printf("❌ Failed to insert upload log: %v", err)
		return err
	}

	log.Printf("📝 Logged upload event for evidence %s by user %s", evidenceID, userID)
	return nil
}

