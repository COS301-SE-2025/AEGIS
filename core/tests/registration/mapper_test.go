package registration_test

import (
	"testing"
	"time"

	"aegis-api/services/registration"
)

func TestRegistrationRequestToModel(t *testing.T) {
	req := registration.RegistrationRequest{
		Name:     "Ofentse",
		Surname:  "Mokwena",
		Email:    "test@example.com",
		Password: "secret123",
	}

	hash := "hashed-password"
	model := registration.RegistrationRequestToModel(req, hash)

	if model.Name != req.Name {
		t.Errorf("expected Name %s, got %s", req.Name, model.Name)
	}
	if model.Surname != req.Surname {
		t.Errorf("expected Surname %s, got %s", req.Surname, model.Surname)
	}
	if model.Email != req.Email {
		t.Errorf("expected Email %s, got %s", req.Email, model.Email)
	}
	if model.PasswordHash != hash {
		t.Errorf("expected PasswordHash %s, got %s", hash, model.PasswordHash)
	}
}

func TestModelToEntity(t *testing.T) {
	model := registration.UserModel{
		Name:         "Ofentse",
		Surname:      "Mokwena",
		Email:        "test@example.com",
		PasswordHash: "hashed-pass",
	}

	id := "uuid-123"
	entity := registration.ModelToEntity(model, id)

	if entity.ID != id {
		t.Errorf("expected ID %s, got %s", id, entity.ID)
	}
	if entity.Name != model.Name {
		t.Errorf("expected Name %s, got %s", model.Name, entity.Name)
	}
	now := time.Now()
delta := now.Sub(entity.CreatedAt)

if delta < 0 || delta > 2*time.Second {
	t.Errorf("expected CreatedAt to be recent (within 2s), got: %v", entity.CreatedAt)
}

}

func TestEntityToResponse(t *testing.T) {
	entity := registration.UserEntity{
		ID:        "uuid-123",
		Name:      "Ofentse",
		Surname:   "Mokwena",
		Email:     "test@example.com",
		CreatedAt: time.Now(),
	}

	resp := registration.EntityToResponse(entity)

	if resp.ID != entity.ID {
		t.Errorf("expected ID %s, got %s", entity.ID, resp.ID)
	}
	expectedFullName := "Ofentse Mokwena"
	if resp.FullName != expectedFullName {
		t.Errorf("expected FullName %s, got %s", expectedFullName, resp.FullName)
	}
	if resp.Email != entity.Email {
		t.Errorf("expected Email %s, got %s", entity.Email, resp.Email)
	}
}
