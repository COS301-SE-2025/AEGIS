package unit_tests

import (
	"errors"
	"testing"

	"aegis-api/models"
	"aegis-api/services_"
)

// Fake repository that implements the interface
type fakeRepo struct {
	shouldExist bool
}

func (f *fakeRepo) UpdateStage(id string, stage models.InvestigationStage) error {
	if !f.shouldExist {
		return errors.New("case not found")
	}
	return nil
}

func (f *fakeRepo) CaseExists(id string) (bool, error) {
	return f.shouldExist, nil
}

func TestUpdateCaseStage_ValidStage(t *testing.T) {
	repo := &fakeRepo{shouldExist: true}
	svc := service.NewCaseService(repo)

	err := svc.UpdateCaseStage("123", models.StageAnalysis)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestUpdateCaseStage_InvalidStage(t *testing.T) {
	repo := &fakeRepo{shouldExist: true}
	svc := service.NewCaseService(repo)

	invalidStage := models.InvestigationStage("badstage")
	err := svc.UpdateCaseStage("123", invalidStage)
	if err == nil {
		t.Error("expected error for invalid stage, got nil")
	}
}

func TestUpdateCaseStage_CaseNotFound(t *testing.T) {
	repo := &fakeRepo{shouldExist: false}
	svc := service.NewCaseService(repo)

	err := svc.UpdateCaseStage("123", models.StageFinalization)
	if err == nil {
		t.Error("expected error for case not found, got nil")
	}
}
