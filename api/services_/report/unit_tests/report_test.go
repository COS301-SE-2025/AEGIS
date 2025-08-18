// services/report/report_service_test.go
package report

import (
	coc "aegis-api/services_/chain_of_custody"
	"aegis-api/services_/report"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mocking the dependencies
type MockCaseReportsRepo struct {
	mock.Mock
}

func (m *MockCaseReportsRepo) SaveReport(ctx context.Context, report *report.Report) error {
	args := m.Called(ctx, report)
	return args.Error(0)
}

func (m *MockCaseReportsRepo) GetByID(ctx context.Context, reportID string) (*report.CaseReportRow, error) {
	args := m.Called(ctx, reportID)
	return args.Get(0).(*report.CaseReportRow), args.Error(1)
}

func (m *MockCaseReportsRepo) GetAllReports(ctx context.Context) ([]report.CaseReportRow, error) {
	args := m.Called(ctx)
	return args.Get(0).([]report.CaseReportRow), args.Error(1)
}
func (m *MockCaseReportsRepo) GetReportsByCaseID(ctx context.Context, caseID string) ([]report.CaseReportRow, error) {
	args := m.Called(ctx, caseID)
	return args.Get(0).([]report.CaseReportRow), args.Error(1)
}

// Implementing GetReportsByEvidenceID for the mock
func (m *MockCaseReportsRepo) GetReportsByEvidenceID(ctx context.Context, evidenceID string) ([]report.CaseReportRow, error) {
	args := m.Called(ctx, evidenceID)
	return args.Get(0).([]report.CaseReportRow), args.Error(1)
}

// Mocking the AuditLogger
type MockAuditLogger struct {
	mock.Mock
}

func (m *MockAuditLogger) LogGenerateReport(ctx context.Context, caseID, reportID, artifactID, actorID, ip, ua string) error {
	args := m.Called(ctx, caseID, reportID, artifactID, actorID, ip, ua)
	return args.Error(0)
}

func (m *MockAuditLogger) LogDownloadReport(ctx context.Context, reportID, userID, userRole, userAgent, ipAddress, email string, timestamp time.Time) error {
	// Use the mock framework's Called method to simulate method calls
	args := m.Called(ctx, reportID, userID, userRole, userAgent, ipAddress, email, timestamp)
	return args.Error(0) // return the error (or nil if no error)
}

// Mocking CoCRepo
type MockCoCRepo struct {
	mock.Mock
}

func (m *MockCoCRepo) ListByCase(ctx context.Context, caseID string) ([]coc.Entry, error) {
	args := m.Called(ctx, caseID)
	return args.Get(0).([]coc.Entry), args.Error(1)
}

// Test function for GenerateReport
func TestGenerateReport(t *testing.T) {
	// Mock the dependencies
	mockCaseReportsRepo := new(MockCaseReportsRepo)
	mockAuditLogger := new(MockAuditLogger)
	mockCoCRepo := new(MockCoCRepo)

	// Prepare the report service
	reportService := &report.ReportService{
		CaseReportsRepo: mockCaseReportsRepo,
		AuditLogger:     mockAuditLogger,
		CoCRepo:         mockCoCRepo,
	}

	// Sample test data
	caseID := uuid.New()
	examinerID := uuid.New()

	// Mock the expected behavior for CoCRepo
	// Mock the expected behavior for CoCRepo
	// Mock the expected behavior for CoCRepo
	// Mock the expected behavior for CoCRepo
	// Mock the expected behavior for CoCRepo
	mockCoCRepo.On("ListByCase", mock.Anything, caseID.String()).Return([]coc.Entry{
		{
			Reason: stringPtr("Sample CoC Entry 1"),
		},
		{
			Reason: stringPtr("Sample CoC Entry 2"),
		},
	}, nil)

	// Mock the expected behavior for SaveReport
	mockCaseReportsRepo.On("SaveReport", mock.Anything, mock.Anything).Return(nil)

	// Mock the expected behavior for LogGenerateReport
	mockAuditLogger.On("LogGenerateReport", mock.Anything, caseID.String(), mock.Anything, "artifactID", examinerID.String(), "ip", "userAgent").Return(nil)

	// Call GenerateReport
	report, err := reportService.GenerateReport(context.Background(), caseID, examinerID)

	// Assertions
	assert.Nil(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, caseID, report.CaseID)
	assert.Equal(t, examinerID, report.ExaminerID)
	assert.Equal(t, "Sample CoC Entry 1\nSample CoC Entry 2\n", report.EvidenceSummary)

	// Assert that the methods were called
	mockCaseReportsRepo.AssertExpectations(t)
	mockAuditLogger.AssertExpectations(t)
	mockCoCRepo.AssertExpectations(t)
}

// Helper function to create string pointer for mock data (CoC entries)
func stringPtr(s string) *string {
	return &s
}
func TestSaveReport(t *testing.T) {
	// Mock the dependencies
	mockCaseReportsRepo := new(MockCaseReportsRepo)

	// Prepare the report service
	reportService := &report.ReportService{
		CaseReportsRepo: mockCaseReportsRepo,
	}

	// Sample report data
	report := &report.Report{
		CaseID:     uuid.New(),
		ExaminerID: uuid.New(),
		Scope:      "Scope of investigation",
		Objectives: "Objectives of investigation",
	}

	// Mock the expected behavior for SaveReport
	mockCaseReportsRepo.On("SaveReport", mock.Anything, report).Return(nil)

	// Call SaveReport
	err := reportService.SaveReport(context.Background(), report)

	// Assertions
	assert.Nil(t, err)

	// Assert that the SaveReport method was called
	mockCaseReportsRepo.AssertExpectations(t)
}

func TestUpdateReport(t *testing.T) {
	// Mock the dependencies
	mockCaseReportsRepo := new(MockCaseReportsRepo)

	// Prepare the report service
	reportService := &report.ReportService{
		CaseReportsRepo: mockCaseReportsRepo,
	}

	// Sample report data
	report := &report.Report{
		ID:         uuid.New(),
		CaseID:     uuid.New(),
		ExaminerID: uuid.New(),
		Scope:      "Updated scope",
		Objectives: "Updated objectives",
	}

	// Mock the expected behavior for SaveReport
	mockCaseReportsRepo.On("SaveReport", mock.Anything, report).Return(nil)

	// Call UpdateReport
	err := reportService.UpdateReport(context.Background(), report)

	// Assertions
	assert.Nil(t, err)

	// Assert that the SaveReport method was called
	mockCaseReportsRepo.AssertExpectations(t)
}

// func TestGetReportsByCaseID(t *testing.T) {
// 	// Mock the dependencies
// 	mockCaseReportsRepo := new(MockCaseReportsRepo)
// 	mockAuditLogger := new(MockAuditLogger)
// 	mockCoCRepo := new(MockCoCRepo)

// 	// Prepare the report service
// 	reportService := &report.ReportService{
// 		CaseReportsRepo: mockCaseReportsRepo,
// 		AuditLogger:     mockAuditLogger,
// 		CoCRepo:         mockCoCRepo,
// 	}

// 	// Sample test data
// 	caseID := uuid.New()

// 	// Sample reports to mock the database response
// 	mockReports := []report.CaseReportRow{
// 		{
// 			ID:         uuid.New().String(),
// 			CaseID:     caseID.String(),
// 			ExaminerID: uuid.New().String(),
// 		},
// 		{
// 			ID:         uuid.New().String(),
// 			CaseID:     caseID.String(),
// 			ExaminerID: uuid.New().String(),
// 		},
// 	}

// 	// Mock the expected behavior for GetReportsByCaseID
// 	mockCaseReportsRepo.On("GetReportsByCaseID", mock.Anything, caseID.String()).Return(mockReports, nil)

// 	// Call GetReportsByCaseID
// 	reports, err := reportService.GetReportsByCaseID(context.Background(), caseID)

// 	// Assertions
// 	assert.Nil(t, err)
// 	assert.Len(t, reports, 2)

// 	// Assert that the GetReportsByCaseID method was called
// 	mockCaseReportsRepo.AssertExpectations(t)
// }

// // Test function for GetReportsByEvidenceID
// func TestGetReportsByEvidenceID(t *testing.T) {
// 	// Mock the dependencies
// 	mockCaseReportsRepo := new(MockCaseReportsRepo)

// 	// Prepare the report service
// 	reportService := &report.ReportService{
// 		CaseReportsRepo: mockCaseReportsRepo,
// 	}

// 	// Sample evidence ID and expected report data
// 	evidenceID := uuid.New()
// 	mockReports := []report.CaseReportRow{
// 		{
// 			ID:         uuid.New().String(),
// 			CaseID:     uuid.New().String(),
// 			ExaminerID: uuid.New().String(),
// 		},
// 	}

// 	// Mock the expected behavior for GetReportsByEvidenceID
// 	mockCaseReportsRepo.On("GetReportsByEvidenceID", mock.Anything, evidenceID).Return(mockReports, nil)

// 	// Call GetReportsByEvidenceID
// 	reports, err := reportService.GetReportsByEvidenceID(context.Background(), evidenceID)

// 	// Assertions
// 	assert.Nil(t, err)
// 	assert.Len(t, reports, 1)

// 	// Assert that the GetReportsByEvidenceID method was called
// 	mockCaseReportsRepo.AssertExpectations(t)
// }
