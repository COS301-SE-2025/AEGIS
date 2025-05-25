package services

import (
	"testing"
    	"aegis-api/db"
		"aegis-api/services/ListCases" // assuming models live there
)



func init() {
	if err := db.Connect(); err != nil {
		panic("❌ Failed to connect to Docker DB: " + err.Error())
	}
}


func TestGetAllCases(t *testing.T) {
	service := ListCases.NewListCasesService()

	cases, err := service.GetAllCases()
	if err != nil {
		t.Fatalf("Failed to get all cases: %v", err)
	}

	t.Logf("✅ Found %d cases", len(cases))
	for _, c := range cases {
		t.Logf("- %s: %s", c.ID, c.Title)
	}
}

func TestGetCasesByUser(t *testing.T) {
	service := ListCases.NewListCasesService()

	userID := "8fb89568-3c52-4535-af33-d2f1266def52"
	cases, err := service.GetCasesByUser(userID)
	if err != nil {
		t.Fatalf("Failed to get user cases: %v", err)
	}

	t.Logf("✅ Found %d cases for user %s", len(cases), userID)
	for _, c := range cases {
		t.Logf("- %s: %s", c.ID, c.Title)
	}
}


func TestGetCasesByNonexistentUser(t *testing.T) {
	service := ListCases.NewListCasesService()

	// This is a made-up UUID that should not exist
	nonexistentUserID := "00000000-0000-0000-0000-000000000999"

	cases, err := service.GetCasesByUser(nonexistentUserID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(cases) != 0 {
		t.Errorf("Expected 0 cases, got %d", len(cases))
	} else {
		t.Log("✅ Correctly returned 0 cases for nonexistent user ID")
	}
}
