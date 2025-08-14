package report

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ReportService defines the business logic for managing reports.
type ReportService interface {
	GenerateReport(ctx context.Context, caseID uuid.UUID, examinerID uuid.UUID) (*Report, error)
	SaveReport(ctx context.Context, report *Report) error
	GetReportByID(ctx context.Context, reportID uuid.UUID) (*Report, error)
	UpdateReport(ctx context.Context, report *Report) error
	GetAllReports(ctx context.Context) ([]Report, error)
	GetReportsByCaseID(ctx context.Context, caseID uuid.UUID) ([]ReportWithDetails, error)
	GetReportsByEvidenceID(ctx context.Context, evidenceID uuid.UUID) ([]Report, error)
	DeleteReportByID(ctx context.Context, reportID uuid.UUID) error
	DownloadReport(ctx context.Context, reportID uuid.UUID) (*ReportWithContent, error)
	DownloadReportAsPDF(ctx context.Context, reportID uuid.UUID) ([]byte, error)
	DownloadReportAsJSON(ctx context.Context, reportID uuid.UUID) ([]byte, error)
	UpdateCustomSectionContent(ctx context.Context, reportUUID uuid.UUID, sectionID primitive.ObjectID, newContent string) error
	//AddSection(ctx context.Context, reportID primitive.ObjectID, section ReportSection) error
	ReorderSection(ctx context.Context, reportUUID uuid.UUID, sectionID primitive.ObjectID, newOrder int) error

	AddCustomSection(ctx context.Context, reportUUID uuid.UUID, title, content string, order int) error
	DeleteCustomSection(ctx context.Context, reportUUID uuid.UUID, sectionID primitive.ObjectID) error
	UpdateSectionContent(ctx context.Context, reportUUID uuid.UUID, sectionID primitive.ObjectID, newContent string) error
	UpdateSectionTitle(ctx context.Context, reportUUID uuid.UUID, sectionID primitive.ObjectID, newTitle string) error
	ReorderCustomSection(ctx context.Context, reportUUID uuid.UUID, sectionID primitive.ObjectID, newOrder int) error
}

// ReportServiceImpl is the concrete implementation of ReportService.
type ReportServiceImpl struct {
	repo      ReportRepository
	mongoRepo ReportMongoRepository
	// artifactsRepo   ReportArtifactsRepository
	storage     Storage
	auditLogger AuditLogger
	authorizer  Authorizer
	coCRepo     GormCoCRepo
}

func NewReportService(
	repo ReportRepository,
	mongoRepo ReportMongoRepository,
	storage Storage,
	auditLogger AuditLogger,
	authorizer Authorizer,
	coCRepo GormCoCRepo,
) ReportService {
	return &ReportServiceImpl{
		repo:        repo,
		mongoRepo:   mongoRepo,
		storage:     storage,
		auditLogger: auditLogger,
		authorizer:  authorizer,
		coCRepo:     coCRepo,
	}
}

// GenerateReport creates a new report for a given case and examiner.
// Here you could include more logic such as fetching case data, formatting content, etc.
func (s *ReportServiceImpl) GenerateReport(ctx context.Context, caseID, examinerID uuid.UUID) (*Report, error) {
	// 1. Create Postgres report metadata
	report := &Report{
		ID:         uuid.New(),
		CaseID:     caseID,
		ExaminerID: examinerID,
		Status:     "draft",
		Version:    1,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// 2. Generate MongoID for content
	mongoID := primitive.NewObjectID()
	report.MongoID = mongoID.Hex() // store mapping in Postgres

	// 3. Save metadata in Postgres
	if err := s.repo.SaveReport(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to generate report metadata: %w", err)
	}

	// 4. Save default sections in Mongo
	defaultSections := []ReportSection{
		{ID: primitive.NewObjectID(), Title: "Case Identification", Content: "", Order: 1},
		{ID: primitive.NewObjectID(), Title: "Scope and Objectives", Content: "", Order: 2},
		{ID: primitive.NewObjectID(), Title: "Evidence Summary", Content: "", Order: 3},
		{ID: primitive.NewObjectID(), Title: "Tools and Methodologies", Content: "", Order: 4},
		{ID: primitive.NewObjectID(), Title: "Findings", Content: "", Order: 5},
		{ID: primitive.NewObjectID(), Title: "Interpretation and Analysis", Content: "", Order: 6},
		{ID: primitive.NewObjectID(), Title: "Limitations", Content: "", Order: 7},
		{ID: primitive.NewObjectID(), Title: "Conclusion", Content: "", Order: 8},
		{ID: primitive.NewObjectID(), Title: "Appendices", Content: "", Order: 9},
		{ID: primitive.NewObjectID(), Title: "Certification", Content: "", Order: 10},
	}

	reportContent := &ReportContentMongo{
		ID:        mongoID,            // Use same Mongo ObjectID
		ReportID:  report.ID.String(), // store Postgres UUID as string
		Sections:  defaultSections,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.mongoRepo.SaveReportContent(ctx, reportContent); err != nil {
		return nil, fmt.Errorf("failed to save report content in Mongo: %w", err)
	}

	return report, nil
}

// SaveReport persists a report to the repository.
func (s *ReportServiceImpl) SaveReport(ctx context.Context, report *Report) error {
	if report.ID == uuid.Nil {
		report.ID = uuid.New()
	}
	return s.repo.SaveReport(ctx, report)
}

// GetReportByID retrieves a report by its ID.
func (s *ReportServiceImpl) GetReportByID(ctx context.Context, reportID uuid.UUID) (*Report, error) {
	return s.repo.GetByID(ctx, reportID)
}

// UpdateReport updates an existing report in the repository.
func (s *ReportServiceImpl) UpdateReport(ctx context.Context, report *Report) error {
	// You could add business logic like checking if the report exists first
	return s.repo.SaveReport(ctx, report) // assuming SaveReport handles both insert/update
}

// GetAllReports retrieves all reports.
func (s *ReportServiceImpl) GetAllReports(ctx context.Context) ([]Report, error) {
	return s.repo.GetAllReports(ctx)
}

// GetReportsByCaseID retrieves all reports for a specific case.
// Service layer: convert timestamps to Africa/Johannesburg
func (s *ReportServiceImpl) GetReportsByCaseID(ctx context.Context, caseID uuid.UUID) ([]ReportWithDetails, error) {
	reports, err := s.repo.GetReportsByCaseID(ctx, caseID)
	if err != nil {
		return nil, err
	}

	// Load timezone once
	loc, _ := time.LoadLocation("Africa/Johannesburg")

	for i := range reports {
		t, err := time.Parse(time.RFC3339, reports[i].LastModified) // or use the actual format your DB returns
		if err != nil {
			continue // or handle error
		}
		reports[i].LastModified = t.In(loc).Format("2006-01-02 15:04:05")
	}

	return reports, nil
}

// GetReportsByEvidenceID retrieves all reports for a specific evidence item.
func (s *ReportServiceImpl) GetReportsByEvidenceID(ctx context.Context, evidenceID uuid.UUID) ([]Report, error) {
	return s.repo.GetReportsByEvidenceID(ctx, evidenceID)
}

// DeleteReportByID deletes a report by ID.
func (s *ReportServiceImpl) DeleteReportByID(ctx context.Context, reportID uuid.UUID) error {
	return s.repo.DeleteReportByID(ctx, reportID)
}

// DownloadReport fetches the report for downloading.
// func (s *ReportServiceImpl) DownloadReport(ctx context.Context, reportID uuid.UUID) (*Report, error) {
// 	report, err := s.repo.DownloadReport(ctx, reportID)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to download report: %w", err)
// 	}
// 	return report, nil
// }

func (s *ReportServiceImpl) DownloadReport(ctx context.Context, reportID uuid.UUID) (*ReportWithContent, error) {
	// 1. Fetch report metadata from Postgres (this part seems fine).
	meta, err := s.repo.GetByID(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch report metadata: %w", err)
	}

	// 2. If MongoID is not empty, try to fetch content from MongoDB.
	var contentSections []ReportSection
	if meta.MongoID != "" {
		// Convert MongoID string to ObjectID.
		mongoID, err := primitive.ObjectIDFromHex(meta.MongoID) // Convert string to ObjectID
		if err != nil {
			return nil, fmt.Errorf("failed to convert MongoID to ObjectID: %w", err)
		}

		// Fetch the content from MongoDB using the valid ObjectID.
		content, err := s.mongoRepo.GetReportContent(ctx, mongoID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch content from MongoDB: %w", err)
		}

		// If content is found, assign it to sections.
		if content != nil {
			contentSections = content.Sections
		}
	}

	// Fallback to an empty slice if no content is found.
	if contentSections == nil {
		contentSections = []ReportSection{}
	}

	// Return both metadata and content sections.
	return &ReportWithContent{
		Metadata: meta,
		Content:  contentSections,
	}, nil
}

func (s *ReportServiceImpl) UpdateReportSection(ctx context.Context, reportID uuid.UUID, sectionID primitive.ObjectID, newContent string) error {
	// Convert reportID to Mongo ObjectID if you store a mapping
	mongoID := primitive.NewObjectID() // Replace with actual mapping
	return s.mongoRepo.UpdateSection(ctx, mongoID, sectionID, newContent)
}

// func (s *ReportServiceImpl) DownloadReport(ctx context.Context, reportID uuid.UUID) (*Report, error) {
// 	return s.repo.DownloadReport(ctx, reportID)
// }

// func (s *ReportServiceImpl) DownloadReportWithContent(ctx context.Context, reportID uuid.UUID) (*ReportWithContent, error) {
// 	meta, err := s.repo.GetByID(ctx, reportID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	mongoID := primitive.NewObjectID() // Map Postgres UUID -> Mongo ObjectID
// 	content, err := s.mongoRepo.GetReportContent(ctx, mongoID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &ReportWithContent{
// 		Metadata: meta,
// 		Content:  content.Sections,
// 	}, nil
// }

func (s *ReportServiceImpl) DownloadReportAsJSON(ctx context.Context, reportID uuid.UUID) ([]byte, error) {
	report, err := s.DownloadReport(ctx, reportID)
	if err != nil {
		return nil, err
	}
	return json.Marshal(report)
}

func (s *ReportServiceImpl) DownloadReportAsPDF(ctx context.Context, reportID uuid.UUID) ([]byte, error) {
	// Always go through the unified fetch method
	report, err := s.DownloadReport(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch report for PDF: %w", err)
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// Title
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, fmt.Sprintf("Report: %s", report.Metadata.Name))
	pdf.Ln(12)

	// Sections from Mongo
	for _, sec := range report.Content {
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(0, 8, sec.Title)
		pdf.Ln(8)

		pdf.SetFont("Arial", "", 11)
		if sec.Content == "" {
			pdf.MultiCell(0, 6, "(No content provided)", "", "", false)
		} else {
			pdf.MultiCell(0, 6, sec.Content, "", "", false)
		}
		pdf.Ln(4)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return buf.Bytes(), nil
}

func (s *ReportServiceImpl) UpdateCustomSectionContent(ctx context.Context, reportUUID uuid.UUID, sectionID primitive.ObjectID, newContent string) error {
	mongoID, err := s.getMongoID(ctx, reportUUID)
	if err != nil {
		return err
	}
	return s.mongoRepo.UpdateSection(ctx, mongoID, sectionID, newContent)
}

// Add a custom section

// // Delete a custom section
// func (s *ReportServiceImpl) DeleteCustomSection(ctx context.Context, reportUUID uuid.UUID, sectionID primitive.ObjectID) error {
// 	mongoID, err := s.getMongoID(ctx, reportUUID)
// 	if err != nil {
// 		return err
// 	}
// 	return s.mongoRepo.DeleteSection(ctx, mongoID, sectionID)
// }

// Update content of a section
func (s *ReportServiceImpl) UpdateSectionContent(ctx context.Context, reportUUID uuid.UUID, sectionID primitive.ObjectID, newContent string) error {
	mongoID, err := s.getMongoID(ctx, reportUUID)
	if err != nil {
		return err
	}
	return s.mongoRepo.UpdateSection(ctx, mongoID, sectionID, newContent)
}

// Update the title of a section
func (s *ReportServiceImpl) UpdateSectionTitle(ctx context.Context, reportUUID uuid.UUID, sectionID primitive.ObjectID, newTitle string) error {
	mongoID, err := s.getMongoID(ctx, reportUUID)
	if err != nil {
		return err
	}
	return s.mongoRepo.UpdateSectionTitle(ctx, mongoID, sectionID, newTitle)
}

// Reorder a section
func (s *ReportServiceImpl) ReorderCustomSection(ctx context.Context, reportUUID uuid.UUID, sectionID primitive.ObjectID, newOrder int) error {
	mongoID, err := s.getMongoID(ctx, reportUUID)
	if err != nil {
		return err
	}
	return s.mongoRepo.ReorderSection(ctx, mongoID, sectionID, newOrder)
}
