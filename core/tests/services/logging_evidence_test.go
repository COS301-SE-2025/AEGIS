package services

import (
	"testing"
	"time"
	"github.com/google/uuid"
	"aegis-api/db"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"context"
//	"github.com/joho/godotenv"
	//"log"
	"os"
	"aegis-api/services/evidence"
	//"go.mongodb.org/mongo-driver/bson/primitive"
)

func init() {
	// Load .env file
if os.Getenv("ENV") == "local" {
   os.Setenv("MONGO_URI", "mongodb://admin:mongo_secure_password123@localhost:27017/app_database?authSource=admin")

} else {
   os.Setenv("MONGO_URI", "mongodb://admin:mongo_secure_password123@localhost:27017/app_database?authSource=admin")

}

	if err := db.ConnectMongo(); err != nil {
		panic("❌ MongoDB connection failed: " + err.Error())
	}
}


func TestMongo_InsertAndLogEvidence(t *testing.T) {
	repo := evidence.NewMongoEvidenceRepository()

	// Prepare test evidence
	ev := evidence.Evidence{
		ID:         uuid.New(),
		CaseID:     uuid.MustParse("91c24784-b2d4-496c-b4a3-ba6b0443eea2"),
		UploadedBy: uuid.MustParse("8fb89568-3c52-4535-af33-d2f1266def52"),
		Filename:   "mongo_test_delete_me.txt",
		FileType:   "text/plain",
		IpfsCID:    "QmTestCIDMongoUnit123",
		FileSize:   2048,
		Checksum:   "fakechecksum123456",
		Metadata: map[string]interface{}{
			"test": true,
		},
		UploadedAt: time.Now(),
	}

	// Insert
	res, err := repo.Collection.InsertOne(context.TODO(), ev)
	if err != nil {
		t.Fatalf("❌ Failed to insert test evidence: %v", err)
	}

	insertedID := res.InsertedID.(primitive.ObjectID)
	t.Logf("✅ Inserted test evidence with Mongo _id: %s | App UUID: %s", insertedID.Hex(), ev.ID)
}
