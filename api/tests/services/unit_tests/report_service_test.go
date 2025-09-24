// file: report_service_test.go
package unit_tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"

	// 👇 change to your real module path
	report "aegis-api/services_/report"
	reportshared "aegis-api/services_/report/shared"
)

/*
	These are PURE UNIT TESTS for ReportServiceImpl.

	We mock BOTH dependencies:
	  - ReportRepository (Postgres/GORM layer)
	  - ReportMongoRepository (MongoDB layer)

	There are NO real databases, files, or networks involved.
	We just verify:
	  - method calls, arguments, and return values
	  - business logic performed by the service (validation, time zone conversion, PDF/json creation, etc.)
*/

/* ----------------------------- Mocks ----------------------------- */

// MockRepo stubs the ReportRepository. Each method delegates to testify/mock.
//type MockRepo struct{ mock.Mock }

func (m *MockRepo) SaveReport(ctx context.Context, r *report.Report) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}
func (m *MockRepo) GetByID(ctx context.Context, id string) (*report.Report, error) {
	args := m.Called(ctx, id)
	if v := args.Get(0); v != nil {
		return v.(*report.Report), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockRepo) GetAllReports(ctx context.Context) ([]report.Report, error) {
	args := m.Called(ctx)
	if v := args.Get(0); v != nil {
		return v.([]report.Report), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockRepo) GetReportsByCaseID(ctx context.Context, caseID uuid.UUID) ([]report.ReportWithDetails, error) {
	args := m.Called(ctx, caseID)
	if v := args.Get(0); v != nil {
		return v.([]report.ReportWithDetails), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockRepo) GetReportsByEvidenceID(ctx context.Context, evidenceID uuid.UUID) ([]report.Report, error) {
	args := m.Called(ctx, evidenceID)
	if v := args.Get(0); v != nil {
		return v.([]report.Report), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockRepo) DeleteReportByID(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockRepo) DownloadReport(ctx context.Context, id uuid.UUID) (*report.Report, error) {
	// Not used by the service (service calls GetByID instead), but implemented for interface completeness.
	args := m.Called(ctx, id)
	if v := args.Get(0); v != nil {
		return v.(*report.Report), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockRepo) UpdateReportName(ctx context.Context, id uuid.UUID, name string) (*report.Report, error) {
	args := m.Called(ctx, id, name)
	if v := args.Get(0); v != nil {
		return v.(*report.Report), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockRepo) ListRecentCandidates(ctx context.Context, opts report.RecentReportsOptions, candidateLimit int) ([]report.Report, error) {
	args := m.Called(ctx, opts, candidateLimit)
	if v := args.Get(0); v != nil {
		return v.([]report.Report), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockRepo) GetReportsByTeamID(ctx context.Context, tenantID, teamID uuid.UUID) ([]report.ReportWithDetails, error) {
	args := m.Called(ctx, tenantID, teamID)
	if v := args.Get(0); v != nil {
		return v.([]report.ReportWithDetails), args.Error(1)
	}
	return nil, args.Error(1)
}

// MockMongo stubs the ReportMongoRepository. Same idea as MockRepo.
type MockMongo struct{ mock.Mock }

func (m *MockMongo) SaveReportContent(ctx context.Context, c *report.ReportContentMongo) error {
	return m.Called(ctx, c).Error(0)
}
func (m *MockMongo) GetReportContent(ctx context.Context, id primitive.ObjectID, tenantID, teamID string) (*report.ReportContentMongo, error) {
	args := m.Called(ctx, id, tenantID, teamID)
	if v := args.Get(0); v != nil {
		return v.(*report.ReportContentMongo), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockMongo) UpdateSection(ctx context.Context, reportID, sectionID primitive.ObjectID, newContent, tenantID, teamID string) error {
	return m.Called(ctx, reportID, sectionID, newContent, tenantID, teamID).Error(0)
}
func (m *MockMongo) AddSection(ctx context.Context, reportID primitive.ObjectID, section report.ReportSection, tenantID, teamID string) error {
	return m.Called(ctx, reportID, section, tenantID, teamID).Error(0)
}
func (m *MockMongo) DeleteSection(ctx context.Context, reportID, sectionID primitive.ObjectID, tenantID, teamID string) error {
	return m.Called(ctx, reportID, sectionID, tenantID, teamID).Error(0)
}
func (m *MockMongo) UpdateSections(ctx context.Context, reportID primitive.ObjectID, sections []report.ReportSection, tenantID, teamID string) error {
	return m.Called(ctx, reportID, sections, tenantID, teamID).Error(0)
}
func (m *MockMongo) FindByReportUUID(ctx context.Context, reportUUID uuid.UUID) (*report.ReportContentMongo, error) {
	args := m.Called(ctx, reportUUID)
	if v := args.Get(0); v != nil {
		return v.(*report.ReportContentMongo), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockMongo) UpdateSectionTitle(ctx context.Context, reportID, sectionID primitive.ObjectID, newTitle string, tenantID, teamID string) error {
	return m.Called(ctx, reportID, sectionID, newTitle, tenantID, teamID).Error(0)
}
func (m *MockMongo) ReorderSection(ctx context.Context, reportID, sectionID primitive.ObjectID, newOrder int, tenantID, teamID string) error {
	return m.Called(ctx, reportID, sectionID, newOrder, tenantID, teamID).Error(0)
}
func (m *MockMongo) BulkUpdateSections(ctx context.Context, reportID primitive.ObjectID, sections []report.ReportSection) error {
	return m.Called(ctx, reportID, sections).Error(0)
}
func (m *MockMongo) LatestUpdateByReportIDs(ctx context.Context, reportIDs []string, tenantID, teamID string) (map[string]time.Time, error) {
	args := m.Called(ctx, reportIDs, tenantID, teamID)
	if v := args.Get(0); v != nil {
		return v.(map[string]time.Time), args.Error(1)
	}
	return nil, args.Error(1)
}

/* --------------------------- Constructor --------------------------- */

// MockSectionRepo stubs the ReportSectionRepository interface.
type MockSectionRepo struct{ mock.Mock }

// Implement CreateSection to satisfy the interface.
func (m *MockSectionRepo) CreateSection(ctx context.Context, section *reportshared.ReportSection) error {
	return m.Called(ctx, section).Error(0)
}
func (m *MockSectionRepo) GetSectionByID(ctx context.Context, id string) (*reportshared.ReportSection, error) {
	args := m.Called(ctx, id)
	if v := args.Get(0); v != nil {
		return v.(*reportshared.ReportSection), args.Error(1)
	}
	return nil, args.Error(1)
}

// Implement ListSectionsByReport to satisfy the interface.
func (m *MockSectionRepo) ListSectionsByReport(ctx context.Context, reportID string) ([]*reportshared.ReportSection, error) {
	args := m.Called(ctx, reportID)
	if v := args.Get(0); v != nil {
		return v.([]*reportshared.ReportSection), args.Error(1)
	}
	return nil, args.Error(1)
}

// Stub UpdateSection to satisfy the interface
func (m *MockSectionRepo) UpdateSection(ctx context.Context, section *reportshared.ReportSection) error {
	return m.Called(ctx, section).Error(0)
}

// newSvc wires the service under test with our mocks.
func newSvc(repo *MockRepo, mongo *MockMongo, sectionRepo *MockSectionRepo) report.ReportService {
	return report.NewReportService(repo, mongo, sectionRepo)
}

/* ----------------------------- Tests ------------------------------ */

// TestGenerateReport_HappyPath verifies the service:
//  1. builds a Report with correct fields (IDs, tenant/team, status/version)
//  2. persists it via repo
//  3. creates default sections and saves a Mongo content doc with proper tenant/team scoping
func TestGenerateReport_HappyPath(t *testing.T) {
	ctx := context.Background()
	repo := new(MockRepo)
	mongo := new(MockMongo)
	sectionRepo := new(MockSectionRepo)
	svc := newSvc(repo, mongo, sectionRepo)

	caseID := uuid.New()
	examinerID := uuid.New()
	tenantID := uuid.New()
	teamID := uuid.New()

	var captured *report.Report

	// Expect Postgres save; capture the report to validate what the service populated.
	repo.On("SaveReport", ctx, mock.MatchedBy(func(r *report.Report) bool {
		captured = r
		return r.ID != uuid.Nil &&
			r.CaseID == caseID &&
			r.ExaminerID == examinerID &&
			r.TenantID == tenantID &&
			r.TeamID == teamID &&
			r.MongoID != "" &&
			r.Status == "draft" &&
			r.Version == 1
	})).Return(nil).Once()

	// Expect SectionRepo to be called for each default section
	sectionRepo.On("CreateSection", ctx, mock.AnythingOfType("*reportshared.ReportSection")).Return(nil).Times(10)

	// Expect Mongo save and validate multi-tenancy + default sections (10, ordered, with timestamps/IDs).
	mongo.On("SaveReportContent", ctx, mock.MatchedBy(func(c *report.ReportContentMongo) bool {
		if captured == nil {
			return false
		}
		if c.ReportID != captured.ID.String() {
			return false
		}
		if c.TenantID != tenantID.String() || c.TeamID != teamID.String() {
			return false
		}
		if c.ID.Hex() != captured.MongoID {
			return false
		}
		if len(c.Sections) != 10 {
			return false
		}
		for i, s := range c.Sections {
			if s.Order != i+1 {
				return false
			}
			if s.ID.IsZero() || s.CreatedAt.IsZero() || s.UpdatedAt.IsZero() {
				return false
			}
		}
		return true
	})).Return(nil).Once()

	out, err := svc.GenerateReport(ctx, caseID, examinerID, tenantID, teamID)
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, tenantID, out.TenantID)
	assert.Equal(t, teamID, out.TeamID)

	repo.AssertExpectations(t)
	mongo.AssertExpectations(t)
	sectionRepo.AssertExpectations(t)
}

// TestGenerateReport_SaveFails ensures we surface repo save errors and DO NOT attempt Mongo writes afterward.
func TestGenerateReport_SaveFails(t *testing.T) {
	ctx := context.Background()
	repo := new(MockRepo)
	mongo := new(MockMongo)
	sectionRepo := new(MockSectionRepo)
	svc := newSvc(repo, mongo, sectionRepo)

	repo.On("SaveReport", ctx, mock.AnythingOfType("*report.Report")).
		Return(errors.New("db down")).Once()

	_, err := svc.GenerateReport(ctx, uuid.New(), uuid.New(), uuid.New(), uuid.New())
	require.ErrorContains(t, err, "failed to generate report metadata")

	// Ensure Mongo save is never invoked.
	mongo.AssertNotCalled(t, "SaveReportContent", mock.Anything, mock.Anything)
	repo.AssertExpectations(t)
}

// TestGenerateReport_MongoSaveFails ensures Mongo errors are returned after a successful repo save.
func TestGenerateReport_MongoSaveFails(t *testing.T) {
	ctx := context.Background()
	repo := new(MockRepo)
	mongo := new(MockMongo)
	sectionRepo := new(MockSectionRepo)
	svc := newSvc(repo, mongo, sectionRepo)

	repo.On("SaveReport", ctx, mock.AnythingOfType("*report.Report")).Return(nil).Once()
	sectionRepo.On("CreateSection", ctx, mock.AnythingOfType("*reportshared.ReportSection")).Return(nil).Times(10)
	mongo.On("SaveReportContent", ctx, mock.AnythingOfType("*report.ReportContentMongo")).Return(errors.New("mongo err")).Once()

	_, err := svc.GenerateReport(ctx, uuid.New(), uuid.New(), uuid.New(), uuid.New())
	require.ErrorContains(t, err, "failed to save report content in Mongo")

	repo.AssertExpectations(t)
	mongo.AssertExpectations(t)
	sectionRepo.AssertExpectations(t)
}

// TestSaveReport_AssignsIDIfNil checks the service generates an ID if the caller didn't set one.
func TestSaveReport_AssignsIDIfNil(t *testing.T) {
	ctx := context.Background()
	repo := new(MockRepo)
	mongo := new(MockMongo)
	sectionRepo := new(MockSectionRepo)
	svc := newSvc(repo, mongo, sectionRepo)

	// The matcher asserts the ID was set by the service.
	repo.On("SaveReport", ctx, mock.MatchedBy(func(r *report.Report) bool { return r.ID != uuid.Nil })).
		Return(nil).Once()

	in := &report.Report{ID: uuid.Nil}
	require.NoError(t, svc.SaveReport(ctx, in))
	repo.AssertExpectations(t)
}

// TestGet_By_All_Delete_Passthroughs confirms the service delegates these calls directly to the repo.
func TestGet_By_All_Delete_Passthroughs(t *testing.T) {
	ctx := context.Background()
	repo := new(MockRepo)
	mongo := new(MockMongo)
	sectionRepo := new(MockSectionRepo)
	svc := newSvc(repo, mongo, sectionRepo)

	id := uuid.New()
	rep := &report.Report{ID: id, Name: "x"}

	repo.On("GetByID", ctx, id.String()).Return(rep, nil).Once()
	repo.On("GetAllReports", ctx).Return([]report.Report{{ID: id}}, nil).Once()
	repo.On("GetReportsByEvidenceID", ctx, id).Return([]report.Report{{ID: id}}, nil).Once()
	repo.On("DeleteReportByID", ctx, id).Return(nil).Once()

	got, err := svc.GetReportByID(ctx, id.String())
	require.NoError(t, err)
	assert.Equal(t, "x", got.Name)

	all, err := svc.GetAllReports(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 1)

	byEv, err := svc.GetReportsByEvidenceID(ctx, id)
	require.NoError(t, err)
	assert.Len(t, byEv, 1)

	require.NoError(t, svc.DeleteReportByID(ctx, id))

	repo.AssertExpectations(t)
	mongo.AssertExpectations(t)
}

// TestGetReportsByCaseID_ConvertsTimezone verifies the service converts RFC3339 UTC timestamps
// to Africa/Johannesburg and formats them as "YYYY-MM-DD HH:MM:SS".
func TestGetReportsByCaseID_ConvertsTimezone(t *testing.T) {
	ctx := context.Background()
	repo := new(MockRepo)
	mongo := new(MockMongo)
	sectionRepo := new(MockSectionRepo)
	svc := newSvc(repo, mongo, sectionRepo)

	caseID := uuid.New()
	// Given UTC inputs…
	in := []report.ReportWithDetails{
		{ID: uuid.New(), LastModified: "2025-08-19T06:00:00Z"},
		{ID: uuid.New(), LastModified: "2025-08-19T13:45:30Z"},
	}
	repo.On("GetReportsByCaseID", ctx, caseID).Return(in, nil).Once()

	// …expect UTC+2 formatted outputs.
	out, err := svc.GetReportsByCaseID(ctx, caseID)
	require.NoError(t, err)
	require.Len(t, out, 2)
	assert.Equal(t, "2025-08-19 08:00:00", out[0].LastModified)
	assert.Equal(t, "2025-08-19 15:45:30", out[1].LastModified)
}

// TestDownloadReport_NoMongoID_ReturnsEmptyContent ensures that when the Report has no MongoID,
// the service returns metadata with an empty content slice and never hits Mongo.
func TestDownloadReport_NoMongoID_ReturnsEmptyContent(t *testing.T) {
	ctx := context.Background()
	repo := new(MockRepo)
	mongo := new(MockMongo)
	sectionRepo := new(MockSectionRepo)
	svc := newSvc(repo, mongo, sectionRepo)

	meta := &report.Report{
		ID:       uuid.New(),
		Name:     "R",
		TenantID: uuid.New(),
		TeamID:   uuid.New(),
		MongoID:  "", // no content to fetch
	}
	repo.On("GetByID", ctx, meta.ID.String()).Return(meta, nil).Once()

	rc, err := svc.DownloadReport(ctx, meta.ID)
	require.NoError(t, err)
	assert.Equal(t, "R", rc.Metadata.Name)
	assert.Empty(t, rc.Content)

	// Ensure Mongo was not queried.
	mongo.AssertNotCalled(t, "GetReportContent", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	repo.AssertExpectations(t)
}

// TestDownloadReport_WithMongo_Success verifies happy path where service fetches metadata then content.
func TestDownloadReport_WithMongo_Success(t *testing.T) {
	ctx := context.Background()
	repo := new(MockRepo)
	mongo := new(MockMongo)
	sectionRepo := new(MockSectionRepo)
	svc := newSvc(repo, mongo, sectionRepo)

	tenant := uuid.New()
	team := uuid.New()
	mid := primitive.NewObjectID()

	meta := &report.Report{
		ID:       uuid.New(),
		Name:     "R",
		TenantID: tenant,
		TeamID:   team,
		MongoID:  mid.Hex(),
	}
	repo.On("GetByID", ctx, meta.ID.String()).Return(meta, nil).Once()

	mongoDoc := &report.ReportContentMongo{
		ID:       mid,
		ReportID: meta.ID.String(),
		TenantID: tenant.String(),
		TeamID:   team.String(),
		Sections: []report.ReportSection{{ID: primitive.NewObjectID(), Title: "T", Content: "C", Order: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()}},
	}
	mongo.On("GetReportContent", ctx, mid, tenant.String(), team.String()).Return(mongoDoc, nil).Once()

	rc, err := svc.DownloadReport(ctx, meta.ID)
	require.NoError(t, err)
	require.Len(t, rc.Content, 1)
	assert.Equal(t, "T", rc.Content[0].Title)

	repo.AssertExpectations(t)
	mongo.AssertExpectations(t)
}

// TestDownloadReport_InvalidMongoID_Error ensures invalid hex MongoID surfaces a friendly error.
func TestDownloadReport_InvalidMongoID_Error(t *testing.T) {
	ctx := context.Background()
	repo := new(MockRepo)
	mongo := new(MockMongo)
	sectionRepo := new(MockSectionRepo)
	svc := newSvc(repo, mongo, sectionRepo)

	meta := &report.Report{
		ID:       uuid.New(),
		Name:     "R",
		TenantID: uuid.New(),
		TeamID:   uuid.New(),
		MongoID:  "not-an-oid",
	}
	repo.On("GetByID", ctx, meta.ID.String()).Return(meta, nil).Once()

	_, err := svc.DownloadReport(ctx, meta.ID)
	require.ErrorContains(t, err, "failed to convert MongoID")
}

// TestDownloadReport_MongoError_Propagates ensures Mongo lookup errors are wrapped and returned.
func TestDownloadReport_MongoError_Propagates(t *testing.T) {
	ctx := context.Background()
	repo := new(MockRepo)
	mongo := new(MockMongo)
	sectionRepo := new(MockSectionRepo)
	svc := newSvc(repo, mongo, sectionRepo)

	tenant := uuid.New()
	team := uuid.New()
	mid := primitive.NewObjectID()

	meta := &report.Report{
		ID:       uuid.New(),
		Name:     "R",
		TenantID: tenant,
		TeamID:   team,
		MongoID:  mid.Hex(),
	}
	repo.On("GetByID", ctx, meta.ID.String()).Return(meta, nil).Once()

	mongo.On("GetReportContent", ctx, mid, tenant.String(), team.String()).
		Return(nil, errors.New("mongo boom")).Once()

	_, err := svc.DownloadReport(ctx, meta.ID)
	require.ErrorContains(t, err, "failed to fetch content from MongoDB")
}

// TestDownloadReportAsJSON_Works verifies JSON marshaling of the composed report.
func TestDownloadReportAsJSON_Works(t *testing.T) {
	ctx := context.Background()
	repo := new(MockRepo)
	mongo := new(MockMongo)
	sectionRepo := new(MockSectionRepo)
	svc := newSvc(repo, mongo, sectionRepo)

	tenant := uuid.New()
	team := uuid.New()
	mid := primitive.NewObjectID()
	id := uuid.New()

	meta := &report.Report{ID: id, Name: "R", TenantID: tenant, TeamID: team, MongoID: mid.Hex()}
	repo.On("GetByID", ctx, id.String()).Return(meta, nil).Once()
	mongo.On("GetReportContent", ctx, mid, tenant.String(), team.String()).
		Return(&report.ReportContentMongo{ID: mid, Sections: []report.ReportSection{}}, nil).Once()

	b, err := svc.DownloadReportAsJSON(ctx, id)
	require.NoError(t, err)
	require.True(t, json.Valid(b))
	require.Contains(t, string(b), `"name":"R"`)
}

// TestDownloadReportAsPDF_BasicPDF ensures the service returns a valid PDF byte stream (starts with %PDF).
func TestDownloadReportAsPDF_BasicPDF(t *testing.T) {
	ctx := context.Background()
	repo := new(MockRepo)
	mongo := new(MockMongo)
	sectionRepo := new(MockSectionRepo)
	svc := newSvc(repo, mongo, sectionRepo)

	tenant := uuid.New()
	team := uuid.New()
	mid := primitive.NewObjectID()
	id := uuid.New()

	meta := &report.Report{ID: id, Name: "Sample", TenantID: tenant, TeamID: team, MongoID: mid.Hex()}
	repo.On("GetByID", ctx, id.String()).Return(meta, nil).Once()
	mongo.On("GetReportContent", ctx, mid, tenant.String(), team.String()).
		Return(&report.ReportContentMongo{
			ID: mid,
			Sections: []report.ReportSection{
				{ID: primitive.NewObjectID(), Title: "Section 1", Content: "<p>Hello world</p>", Order: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			},
		}, nil).Once()

	pdf, err := svc.DownloadReportAsPDF(ctx, id)
	require.NoError(t, err)
	require.Greater(t, len(pdf), 100)
	require.True(t, bytes.HasPrefix(pdf, []byte("%PDF")))
}

// TestUpdateCustomSectionContent_Success verifies getMongoID path + UpdateSection with tenant/team scoping.
func TestUpdateCustomSectionContent_Success(t *testing.T) {
	ctx := context.Background()
	repo := new(MockRepo)
	mongo := new(MockMongo)
	sectionRepo := new(MockSectionRepo)
	svc := newSvc(repo, mongo, sectionRepo)

	rid := uuid.New()
	tenant := uuid.New()
	team := uuid.New()
	mid := primitive.NewObjectID()
	secID := primitive.NewObjectID()

	// getMongoID: service reads report metadata to discover MongoID + tenant/team
	repo.On("GetByID", ctx, rid.String()).Return(&report.Report{
		ID:       rid,
		TenantID: tenant,
		TeamID:   team,
		MongoID:  mid.Hex(),
	}, nil).Once()

	// Then calls UpdateSection with scoped tenant/team
	mongo.On("UpdateSection", ctx, mid, secID, "new content", tenant.String(), team.String()).
		Return(nil).Once()

	require.NoError(t, svc.UpdateCustomSectionContent(ctx, rid, secID, "new content"))

	repo.AssertExpectations(t)
	mongo.AssertExpectations(t)
}

// TestUpdateCustomSectionContent_MissingMongoID verifies we error when the report has no MongoID.
func TestUpdateCustomSectionContent_MissingMongoID(t *testing.T) {
	ctx := context.Background()
	repo := new(MockRepo)
	mongo := new(MockMongo)
	sectionRepo := new(MockSectionRepo)
	svc := newSvc(repo, mongo, sectionRepo)

	rid := uuid.New()
	repo.On("GetByID", ctx, rid.String()).Return(&report.Report{
		ID:       rid,
		TenantID: uuid.New(),
		TeamID:   uuid.New(),
		MongoID:  "",
	}, nil).Once()

	err := svc.UpdateCustomSectionContent(ctx, rid, primitive.NewObjectID(), "x")
	require.ErrorContains(t, err, report.ErrMongoReportNotFound.Error())
}

// TestUpdateSectionContent_Delegates confirms UpdateSectionContent just forwards to UpdateCustomSectionContent.
func TestUpdateSectionContent_Delegates(t *testing.T) {

	var (
		ctx         context.Context
		repo        *MockRepo
		mongo       *MockMongo
		sectionRepo *MockSectionRepo
		svc         report.ReportService
		rid         uuid.UUID
		tenant      uuid.UUID
		team        uuid.UUID
		mid         primitive.ObjectID
		secID       primitive.ObjectID
	)

	// First test: UpdateSectionContent delegates to UpdateCustomSectionContent
	ctx = context.Background()
	repo = new(MockRepo)
	mongo = new(MockMongo)
	sectionRepo = new(MockSectionRepo)
	svc = newSvc(repo, mongo, sectionRepo)

	rid = uuid.New()
	tenant = uuid.New()
	team = uuid.New()
	mid = primitive.NewObjectID()
	secID = primitive.NewObjectID()

	// The service calls GetByID with rid.String(), not rid
	repo.On("GetByID", ctx, rid.String()).Return(&report.Report{
		ID:       rid,
		TenantID: tenant,
		TeamID:   team,
		MongoID:  mid.Hex(),
	}, nil).Once()
	mongo.On("UpdateSection", ctx, mid, secID, "c", tenant.String(), team.String()).
		Return(nil).Once()

	require.NoError(t, svc.UpdateSectionContent(ctx, rid, secID, "c"))

	repo.AssertExpectations(t)
	mongo.AssertExpectations(t)

	// Second test: UpdateSectionTitle
	ctx = context.Background()
	repo = new(MockRepo)
	mongo = new(MockMongo)
	sectionRepo = new(MockSectionRepo)
	svc = newSvc(repo, mongo, sectionRepo)

	rid = uuid.New()
	tenant = uuid.New()
	team = uuid.New()
	mid = primitive.NewObjectID()
	secID = primitive.NewObjectID()

	repo.On("GetByID", ctx, rid.String()).Return(&report.Report{
		ID:       rid,
		TenantID: tenant,
		TeamID:   team,
		MongoID:  mid.Hex(),
	}, nil).Once()
	mongo.On("UpdateSectionTitle", ctx, mid, secID, "Title", tenant.String(), team.String()).
		Return(nil).Once()

	require.NoError(t, svc.UpdateSectionTitle(ctx, rid, secID, "Title"))
}

// TestReorderCustomSection_Success verifies reorder calls Mongo with the new order and scoping.
func TestReorderCustomSection_Success(t *testing.T) {
	ctx := context.Background()
	repo := new(MockRepo)
	mongo := new(MockMongo)
	sectionRepo := new(MockSectionRepo)
	svc := newSvc(repo, mongo, sectionRepo)

	rid := uuid.New()
	tenant := uuid.New()
	team := uuid.New()
	mid := primitive.NewObjectID()
	secID := primitive.NewObjectID()

	repo.On("GetByID", ctx, rid.String()).Return(&report.Report{
		ID:       rid,
		TenantID: tenant,
		TeamID:   team,
		MongoID:  mid.Hex(),
	}, nil).Once()
	mongo.On("ReorderSection", ctx, mid, secID, 3, tenant.String(), team.String()).
		Return(nil).Once()

	require.NoError(t, svc.ReorderCustomSection(ctx, rid, secID, 3))
}

// TestUpdateReportName_ValidationAndRepoCall validates input trimming/length then repo call is made.
func TestUpdateReportName_ValidationAndRepoCall(t *testing.T) {
	ctx := context.Background()
	repo := new(MockRepo)
	mongo := new(MockMongo)
	sectionRepo := new(MockSectionRepo)
	svc := newSvc(repo, mongo, sectionRepo)

	id := uuid.New()

	// invalid: empty/whitespace
	_, err := svc.UpdateReportName(ctx, id, "   ")
	require.ErrorIs(t, err, report.ErrInvalidReportName)

	// invalid: too long
	_, err = svc.UpdateReportName(ctx, id, strings.Repeat("a", 256))
	require.ErrorIs(t, err, report.ErrInvalidReportName)

	// valid path: expect repo call with trimmed name
	updated := &report.Report{ID: id, Name: "Good"}
	repo.On("UpdateReportName", ctx, id, "Good").Return(updated, nil).Once()

	got, err := svc.UpdateReportName(ctx, id, " Good ")
	require.NoError(t, err)
	assert.Equal(t, "Good", got.Name)

	repo.AssertExpectations(t)
}

// TestGetReportsByTeamID_Passthrough confirms the service just forwards to repo.
func TestGetReportsByTeamID_Passthrough(t *testing.T) {
	ctx := context.Background()
	repo := new(MockRepo)
	mongo := new(MockMongo)
	sectionRepo := new(MockSectionRepo)
	svc := newSvc(repo, mongo, sectionRepo)

	tenant := uuid.New()
	team := uuid.New()
	rows := []report.ReportWithDetails{{ID: uuid.New(), Name: "N"}}

	repo.On("GetReportsByTeamID", ctx, tenant, team).Return(rows, nil).Once()

	out, err := svc.GetReportsByTeamID(ctx, tenant, team)
	require.NoError(t, err)
	require.Len(t, out, 1)
	assert.Equal(t, "N", out[0].Name)
}
