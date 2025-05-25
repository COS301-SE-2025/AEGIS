package services

import (
	"testing"
	"aegis-api/db"
    "aegis-api/services/evidence"
)

func init() {
	if err := db.Connect(); err != nil {
		panic("DB connection failed: " + err.Error())
	}
}

func TestGetEvidenceMetadata(t *testing.T) {
	service := evidence.NewEvidenceService()

	evidenceID := "6b726119-4d34-4abf-9789-17b0cdc190c9"


	meta, err := service.GetEvidenceMetadata(evidenceID)
	if err != nil {
		t.Fatalf("❌ Failed to get metadata: %v", err)
	}

	t.Logf("✅ Metadata retrieved: %s (%s, %d bytes)", meta.Filename, meta.IpfsCID, meta.FileSize)
}
