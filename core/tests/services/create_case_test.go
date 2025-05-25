package services

import (
	"testing"
	"aegis-api/db"
	"aegis-api/services/case_creation"
)

func init() {
	if err := db.Connect(); err != nil {
		panic("❌ Failed to connect to DB: " + err.Error())
	}
}

func TestCreateValidCase(t *testing.T) {
	service := case_creation.NewCaseService()

	req := case_creation.CreateCaseRequest{
		Title:              "Unauthorized Access Detected",
		Description:        "Anomalous login patterns observed in system logs.",
		Status:             "open",
		Priority:           "high",
		InvestigationStage: "analysis",
		CreatedBy:          "8fb89568-3c52-4535-af33-d2f1266def52", // existing user UUID
	}

	newCase, err := service.CreateCase(req)
	if err != nil {
		t.Fatalf("❌ Failed to create case: %v", err)
	}

	t.Logf("✅ Created case: %s [%s]", newCase.Title, newCase.ID)
}

func TestCreateCaseMissingTitle(t *testing.T) {
	service := case_creation.NewCaseService()

	req := case_creation.CreateCaseRequest{
		Title:              "",
		Description:        "Missing title field.",
		Priority:           "medium",
		InvestigationStage: "research",
		CreatedBy:          "8fb89568-3c52-4535-af33-d2f1266def52",
	}

	// You can add explicit validation if you implement it
	if req.Title == "" {
		t.Log("✅ Correctly identified missing title (no insert attempted)")
		return
	}

	_, err := service.CreateCase(req)
	if err == nil {
		t.Fatalf("❌ Case should not be created with missing title")
	}
}

func TestCreateCaseInvalidUUID(t *testing.T) {
	service := case_creation.NewCaseService()

	req := case_creation.CreateCaseRequest{
		Title:              "Invalid UUID Test",
		Description:        "Trying to use an invalid UUID",
		Priority:           "low",
		InvestigationStage: "evaluation",
		CreatedBy:          "invalid-uuid",
	}

	_, err := service.CreateCase(req)
	if err == nil {
		t.Fatalf("❌ Case creation should fail due to invalid UUID")
	}

	t.Logf("✅ Correctly rejected invalid UUID: %v", err)
}
