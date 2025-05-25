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

func TestGetEvidenceByID_Success(t *testing.T) {
	service := evidence.NewEvidenceService()
	existingID := "5d5856ff-7eb9-410f-9e82-4924e72e4e45"

	ev, err := service.GetEvidenceByID(existingID)
	if err != nil {
		t.Fatalf("❌ Failed to retrieve evidence: %v", err)
	}

	t.Logf("✅ Retrieved evidence: %s (CID: %s)", ev.Filename, ev.IpfsCID)
}

func TestGetEvidenceByID_NotFound(t *testing.T) {
	service := evidence.NewEvidenceService()
	missingID := "00000000-0000-0000-0000-000000000000"

	_, err := service.GetEvidenceByID(missingID)
	if err == nil {
		t.Fatalf("❌ Expected error for non-existent evidence ID")
	}

	t.Logf("✅ Correctly handled missing evidence: %v", err)
}

func TestGetEvidenceByID_InvalidUUID(t *testing.T) {
	service := evidence.NewEvidenceService()

	_, err := service.GetEvidenceByID("not-a-uuid")
	if err == nil {
		t.Fatalf("❌ Expected error for invalid UUID format")
	}

	t.Logf("✅ Correctly rejected invalid UUID: %v", err)
}
