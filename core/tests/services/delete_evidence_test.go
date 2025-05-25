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

func TestDeleteEvidenceByID_Success(t *testing.T) {
	service := evidence.NewEvidenceService()

	// Insert test record first
	req := evidence.UploadEvidenceRequest{
	CaseID:     "3c336456-aa9a-4298-8b24-cadaa5e41bf2",
	UploadedBy: "1d6dfa7d-166d-4a77-9364-7f7e6145ec42",
	Filename:   "delete_me.txt",
	FileType:   "text/plain",
	IpfsCID:    "QmDummyCIDDeleteTest",
	FileSize:   100,
	Checksum:   "abc123fakechecksum",
	Metadata:   map[string]interface{}{"test": true},
	Tags:       []string{"test", "delete"},
}

	ev, err := service.UploadEvidence(req)
	if err != nil {
		t.Fatalf("❌ Setup failed: %v", err)
	}

	// Try delete
	err = service.DeleteEvidenceByID(ev.ID.String())
	if err != nil {
		t.Fatalf("❌ Failed to delete evidence: %v", err)
	}

	t.Logf("✅ Successfully deleted evidence: %s", ev.ID)
}

func TestDeleteEvidenceByID_NotFound(t *testing.T) {
	service := evidence.NewEvidenceService()

	err := service.DeleteEvidenceByID("00000000-0000-0000-0000-000000000000")
	if err == nil {
		t.Fatalf("❌ Expected error for non-existent evidence ID")
	}

	t.Logf("✅ Correctly handled missing record: %v", err)
}

func TestDeleteEvidenceByID_InvalidUUID(t *testing.T) {
	service := evidence.NewEvidenceService()

	err := service.DeleteEvidenceByID("not-a-uuid")
	if err == nil {
		t.Fatalf("❌ Expected error for invalid UUID")
	}

	t.Logf("✅ Correctly rejected invalid UUID: %v", err)
}
