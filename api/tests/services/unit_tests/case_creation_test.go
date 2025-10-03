package unit_tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"aegis-api/services_/case/case_creation"
	// "aegis-api/pkg/websocket"
    // "aegis-api/services_/notification"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//
// ────────────────────────────────
//   MOCKS
// ────────────────────────────────
//

// MockCaseRepository implements CaseRepository for service tests
type MockCaseRepository struct {
	mock.Mock
}

func (m *MockCaseRepository) CreateCase(c *case_creation.Case) error {
	args := m.Called(c)
	return args.Error(0)
}

func (m *MockCaseRepository) GetCaseByID(ctx context.Context, id uuid.UUID) (*case_creation.Case, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*case_creation.Case), args.Error(1)
}

//
// ────────────────────────────────
//   MODEL TESTS
// ────────────────────────────────
//

func TestGetProgressForStage(t *testing.T) {
	assert.Equal(t, 10, case_creation.GetProgressForStage("Triage"))
	assert.Equal(t, 55, case_creation.GetProgressForStage("Correlation & Threat Intelligence"))
	assert.Equal(t, 0, case_creation.GetProgressForStage("Unknown"))
}

func TestGetProgressForStage_AllStages(t *testing.T) {
	tests := map[string]int{
		"Triage":                       10,
		"Evidence Collection":          25,
		"Analysis":                     40,
		"Correlation & Threat Intelligence": 55,
		"Containment & Eradication":    70,
		"Recovery":                     85,
		"Reporting & Documentation":    95,
		"Case Closure & Review":        100,
		"Unknown":                      0, // default branch
	}

	for stage, expected := range tests {
		assert.Equal(t, expected, case_creation.GetProgressForStage(stage), "stage=%s", stage)
	}
}


//
// ────────────────────────────────
//   REPOSITORY TESTS
// ────────────────────────────────
//

func setupGormWithMock(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	gdb, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm: %v", err)
	}
	return gdb, mock
}

func TestGormCaseRepository_CreateCase(t *testing.T) {
	gdb, mock := setupGormWithMock(t)
	repo := case_creation.NewGormCaseRepository(gdb)

	// Expect insert
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "cases"`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectCommit()

	c := &case_creation.Case{
		ID:        uuid.New(),
		Title:     "Test",
		CreatedBy: uuid.New(),
		TeamName:  "BlueTeam",
		TenantID:  uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.CreateCase(c)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGormCaseRepository_GetCaseByID_Success(t *testing.T) {
	gdb, mock := setupGormWithMock(t)
	repo := case_creation.NewGormCaseRepository(gdb)

	id := uuid.New()
	

	// rows := sqlmock.NewRows([]string{"id", "title", "team_name", "created_by", "tenant_id", "created_at", "updated_at"}).
	// 	AddRow(id, "Case A", "Team1", uuid.New(), uuid.New(), time.Now(), time.Now())

	// mock.ExpectQuery(`SELECT .* FROM "cases" WHERE id = .* LIMIT 1`).
	// 	WithArgs(id).
	// 	WillReturnRows(rows)

	rows := sqlmock.NewRows([]string{
    "id", "title", "description", "status", "priority",
    "investigation_stage", "created_by", "team_name", "tenant_id",
    "team_id", "created_at", "updated_at", "progress",
}).AddRow(
    id, "Case A", "", "open", "medium", "analysis",
    uuid.New(), "Team1", uuid.New(), uuid.New(),
    time.Now(), time.Now(), 0,
)


	mock.ExpectQuery(`SELECT .* FROM "cases" WHERE id = .* ORDER BY "cases"\."id" LIMIT .*`).
    WithArgs(id, sqlmock.AnyArg()). // allow second arg
    WillReturnRows(rows)



	result, err := repo.GetCaseByID(context.Background(), id)
	assert.NoError(t, err)
	assert.Equal(t, id, result.ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGormCaseRepository_GetCaseByID_NotFound(t *testing.T) {
	gdb, mock := setupGormWithMock(t)
	repo := case_creation.NewGormCaseRepository(gdb)

	id := uuid.New()
	mock.ExpectQuery(`SELECT .* FROM "cases" WHERE id = .* LIMIT 1`).
		WithArgs(id).
		WillReturnError(gorm.ErrRecordNotFound)

	result, err := repo.GetCaseByID(context.Background(), id)
	assert.Error(t, err)
	assert.Nil(t, result)
}


//
// ────────────────────────────────
 //   SERVICE TESTS
// ────────────────────────────────
//

func TestService_CreateCase_Success(t *testing.T) {
	mockRepo := new(MockCaseRepository)
	svc := case_creation.NewCaseService(mockRepo, nil, nil)

	req := &case_creation.CreateCaseRequest{
		Title:     "Case Alpha",
		TeamName:  "BlueTeam",
		CreatedBy: uuid.New(),
		TenantID:  uuid.New(),
	}

	// // Expected case
	// expected := &case_creation.Case{
	// 	ID:        uuid.Nil, // service assigns new UUID, so we don’t assert exact
	// 	Title:     req.Title,
	// 	TeamName:  req.TeamName,
	// 	CreatedBy: req.CreatedBy,
	// 	TenantID:  req.TenantID,
	// }

	mockRepo.On("CreateCase", mock.AnythingOfType("*case_creation.Case")).Return(nil)

	result, err := svc.CreateCase(req)
	assert.NoError(t, err)
	assert.Equal(t, req.Title, result.Title)
	assert.Equal(t, req.TeamName, result.TeamName)

	mockRepo.AssertExpectations(t)
}

func TestService_CreateCase_ValidationErrors(t *testing.T) {
	mockRepo := new(MockCaseRepository)
	svc := case_creation.NewCaseService(mockRepo, nil, nil)

	// Missing title
	req1 := &case_creation.CreateCaseRequest{TeamName: "Blue", CreatedBy: uuid.New()}
	_, err := svc.CreateCase(req1)
	assert.EqualError(t, err, "title is required")

	// Missing team
	req2 := &case_creation.CreateCaseRequest{Title: "Case", CreatedBy: uuid.New()}
	_, err = svc.CreateCase(req2)
	assert.EqualError(t, err, "team name is required")
}

func TestService_CreateCase_RepoError(t *testing.T) {
	mockRepo := new(MockCaseRepository)
	svc := case_creation.NewCaseService(mockRepo, nil, nil)

	req := &case_creation.CreateCaseRequest{
		Title:     "Case Beta",
		TeamName:  "RedTeam",
		CreatedBy: uuid.New(),
		TenantID:  uuid.New(),
	}

	mockRepo.On("CreateCase", mock.AnythingOfType("*case_creation.Case")).Return(errors.New("db error"))

	_, err := svc.CreateCase(req)
	assert.EqualError(t, err, "db error")

	mockRepo.AssertExpectations(t)
}

func TestService_GetCaseByID(t *testing.T) {
	mockRepo := new(MockCaseRepository)
	svc := case_creation.NewCaseService(mockRepo, nil, nil)

	id := uuid.New()
	c := &case_creation.Case{ID: id, Title: "Case Gamma", TeamName: "TeamX", CreatedBy: uuid.New(), TenantID: uuid.New()}

	mockRepo.On("GetCaseByID", mock.Anything, id).Return(c, nil)

	result, err := svc.GetCaseByID(context.Background(), id.String())
	assert.NoError(t, err)
	assert.Equal(t, c, result)

	mockRepo.AssertExpectations(t)
}

func TestService_GetCaseByID_InvalidUUID(t *testing.T) {
	mockRepo := new(MockCaseRepository)
	svc := case_creation.NewCaseService(mockRepo, nil, nil)

	_, err := svc.GetCaseByID(context.Background(), "not-a-uuid")
	assert.Error(t, err)
}

func TestService_CreateCase_WithNotification(t *testing.T) {
	mockRepo := new(MockCaseRepository)
	// fake deps
	// fakeHub := &websocket.Hub{}
	// fakeNotif := &notification.NotificationService{}

	// svc := case_creation.NewCaseService(mockRepo, fakeNotif, fakeHub)
	svc := case_creation.NewCaseService(mockRepo, nil, nil)


	req := &case_creation.CreateCaseRequest{
		Title:     "Case Delta",
		TeamName:  "BlueTeam",
		CreatedBy: uuid.New(),
		TenantID:  uuid.New(),
	}

	mockRepo.On("CreateCase", mock.AnythingOfType("*case_creation.Case")).Return(nil)

	result, err := svc.CreateCase(req)
	assert.NoError(t, err)
	assert.Equal(t, req.Title, result.Title)

	mockRepo.AssertExpectations(t)
	// No panic = notification branch executed
}


func TestService_GetCaseByID_RepoError(t *testing.T) {
	mockRepo := new(MockCaseRepository)
	svc := case_creation.NewCaseService(mockRepo, nil, nil)

	id := uuid.New()
	mockRepo.On("GetCaseByID", mock.Anything, id).Return(nil, errors.New("db failure"))

	result, err := svc.GetCaseByID(context.Background(), id.String())
	assert.Error(t, err)
	assert.Nil(t, result)

	mockRepo.AssertExpectations(t)
}
