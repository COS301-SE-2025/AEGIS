package unit_tests

import (
	"errors"
	"testing"

	"aegis-api/models"
	service "aegis-api/services_/case/update_case_Investigation_stage"

	"github.com/google/uuid"
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

func (f *fakeRepo) CaseExists(id uuid.UUID) (bool, error) {
	return f.shouldExist, nil
}

func TestUpdateCaseStage_ValidStage(t *testing.T) {
	repo := &fakeRepo{shouldExist: true}
	svc := service.NewCaseService(repo)

	validUUID := uuid.New().String()
	err := svc.UpdateCaseStage(validUUID, models.StageAnalysis)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestUpdateCaseStage_InvalidStage(t *testing.T) {
	repo := &fakeRepo{shouldExist: true}
	svc := service.NewCaseService(repo)

	invalidStage := models.InvestigationStage("badstage")
	validUUID := uuid.New().String()
	err := svc.UpdateCaseStage(validUUID, invalidStage)
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
