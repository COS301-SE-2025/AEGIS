package services

import (
	"testing"
	"aegis-api/services/evidence"
	"aegis-api/db"
)

func init() {
	if err := db.Connect(); err != nil {
		panic("DB connection failed: " + err.Error())
	}
}

func TestListEvidenceByCase(t *testing.T) {
	service := evidence.NewEvidenceService()
	caseID := "3f6fb56c-1fb6-49d9-8917-918c6331d643" // Replace with actual case ID

	results, err := service.ListEvidenceByCase(caseID)
	if err != nil {
		t.Fatalf("❌ Failed to list evidence: %v", err)
	}

	t.Logf("✅ Found %d evidence items for case %s", len(results), caseID)
	for _, ev := range results {
		t.Logf("- %s (%s)", ev.Filename, ev.IpfsCID)
	}
}

func TestListEvidenceByUser(t *testing.T) {
	service := evidence.NewEvidenceService()
	userID := "ded0a1b3-4712-46b5-8d01-fafbaf3f8236" // Replace with actual user ID

	results, err := service.ListEvidenceByUser(userID)
	if err != nil {
		t.Fatalf("❌ Failed to list user evidence: %v", err)
	}

	t.Logf("✅ Found %d evidence items for user %s", len(results), userID)
	for _, ev := range results {
		t.Logf("- %s (CID: %s)", ev.Filename, ev.IpfsCID)
	}
}
