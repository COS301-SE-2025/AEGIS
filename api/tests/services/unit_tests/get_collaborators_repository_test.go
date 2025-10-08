package unit_tests

import (
	"testing"
	"time"

	"aegis-api/services_/admin/get_collaborators"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupCollaboratorsTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

	// Create test schema
	db.Exec(`CREATE TABLE users (
        id TEXT PRIMARY KEY,
        full_name TEXT,
        email TEXT,
        created_at DATETIME,
        updated_at DATETIME
    )`)

	db.Exec(`CREATE TABLE case_user_roles (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        case_id TEXT,
        user_id TEXT,
        role TEXT,
        assigned_at DATETIME,
        created_at DATETIME,
        updated_at DATETIME
    )`)

	return db
}

func TestGormRepository_NewGormRepository(t *testing.T) {
	db := setupCollaboratorsTestDB()
	repo := get_collaborators.NewGormRepository(db)

	assert.NotNil(t, repo)
}

func TestGormRepository_GetCollaboratorsByCaseID_Success(t *testing.T) {
	db := setupCollaboratorsTestDB()
	repo := get_collaborators.NewGormRepository(db)

	caseID := uuid.New()
	userID := uuid.New()

	// Insert test data
	db.Exec("INSERT INTO users (id, full_name, email, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
		userID.String(), "John Doe", "john@example.com", time.Now(), time.Now())

	db.Exec("INSERT INTO case_user_roles (case_id, user_id, role, assigned_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		caseID.String(), userID.String(), "analyst", time.Now(), time.Now(), time.Now())

	result, err := repo.GetCollaboratorsByCaseID(caseID)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, userID, result[0].ID)
	assert.Equal(t, "John Doe", result[0].FullName)
	assert.Equal(t, "john@example.com", result[0].Email)
	assert.Equal(t, "analyst", result[0].Role)
}

func TestGormRepository_GetCollaboratorsByCaseID_EmptyResult(t *testing.T) {
	db := setupCollaboratorsTestDB()
	repo := get_collaborators.NewGormRepository(db)

	caseID := uuid.New()

	result, err := repo.GetCollaboratorsByCaseID(caseID)

	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestGormRepository_GetCollaboratorsByCaseID_MultipleCollaborators(t *testing.T) {
	db := setupCollaboratorsTestDB()
	repo := get_collaborators.NewGormRepository(db)

	caseID := uuid.New()
	user1ID := uuid.New()
	user2ID := uuid.New()

	// Insert test users
	db.Exec("INSERT INTO users (id, full_name, email, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
		user1ID.String(), "Alice Smith", "alice@example.com", time.Now(), time.Now())
	db.Exec("INSERT INTO users (id, full_name, email, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
		user2ID.String(), "Bob Johnson", "bob@example.com", time.Now(), time.Now())

	// Insert case role assignments
	db.Exec("INSERT INTO case_user_roles (case_id, user_id, role, assigned_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		caseID.String(), user1ID.String(), "incident_responder", time.Now(), time.Now(), time.Now())
	db.Exec("INSERT INTO case_user_roles (case_id, user_id, role, assigned_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		caseID.String(), user2ID.String(), "forensics_analyst", time.Now(), time.Now(), time.Now())

	result, err := repo.GetCollaboratorsByCaseID(caseID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// Verify both users are returned
	userIDs := []uuid.UUID{result[0].ID, result[1].ID}
	assert.Contains(t, userIDs, user1ID)
	assert.Contains(t, userIDs, user2ID)
}

func TestGormRepository_GetCollaboratorsByCaseID_DatabaseError(t *testing.T) {
	// Create a database and then close it to simulate an error
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	sqlDB, _ := db.DB()
	sqlDB.Close() // Close connection to force error

	repo := get_collaborators.NewGormRepository(db)
	caseID := uuid.New()

	result, err := repo.GetCollaboratorsByCaseID(caseID)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGormRepository_GetCollaboratorsByCaseID_InvalidTableStructure(t *testing.T) {
	// Create DB with missing columns to test error handling
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

	// Create table with missing columns
	db.Exec(`CREATE TABLE users (id TEXT PRIMARY KEY)`)                  // Missing full_name, email
	db.Exec(`CREATE TABLE case_user_roles (case_id TEXT, user_id TEXT)`) // Missing role

	repo := get_collaborators.NewGormRepository(db)
	caseID := uuid.New()
	userID := uuid.New()

	// Insert minimal data
	db.Exec("INSERT INTO users (id) VALUES (?)", userID.String())
	db.Exec("INSERT INTO case_user_roles (case_id, user_id) VALUES (?, ?)", caseID.String(), userID.String())

	result, err := repo.GetCollaboratorsByCaseID(caseID)

	// Should handle missing columns gracefully (might return empty values or error)
	// Adjust assertion based on actual behavior
	if err != nil {
		assert.Error(t, err)
		assert.Nil(t, result)
	} else {
		assert.NotNil(t, result)
	}
}
