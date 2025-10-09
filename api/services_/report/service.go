package report

import (
	reportshared "aegis-api/services_/report/shared"
	"log"

	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	// "image/jpeg"
	// "image/png"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	//"github.com/jung-kurt/gofpdf"

	"go.mongodb.org/mongo-driver/bson/primitive"
	//"golang.org/x/net/html"

	"html/template"
	"path/filepath"
	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	// "image"
	//"strconv"
	// "github.com/johnfercher/maroto/pkg/consts"
	// "github.com/johnfercher/maroto/pkg/pdf"
)

// ReportService defines the business logic for managing reports.
type ReportService interface {
	GenerateReport(ctx context.Context, caseID, examinerID, tenantID, teamID uuid.UUID) (*Report, error)
	SaveReport(ctx context.Context, report *Report) error
	GetReportByID(ctx context.Context, reportID string) (*Report, error)
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
	//GeneratePDFFromHTML(ctx context.Context, html string) ([]byte, error)
	AddCustomSection(ctx context.Context, reportUUID uuid.UUID, title, content string, order int) error
	DeleteCustomSection(ctx context.Context, reportUUID uuid.UUID, sectionID primitive.ObjectID) error
	UpdateSectionContent(ctx context.Context, reportUUID uuid.UUID, sectionID primitive.ObjectID, newContent string) error
	UpdateSectionTitle(ctx context.Context, reportUUID uuid.UUID, sectionID primitive.ObjectID, newTitle string) error
	ReorderCustomSection(ctx context.Context, reportUUID uuid.UUID, sectionID primitive.ObjectID, newOrder int) error // ... your existing methods ...
	ListRecentReports(ctx context.Context, opts RecentReportsOptions) ([]RecentReport, error)
	UpdateReportName(ctx context.Context, reportID uuid.UUID, name string) (*Report, error) // NEW
	GetReportsByTeamID(ctx context.Context, tenantID, teamID uuid.UUID) ([]ReportWithDetails, error)
}

// ReportServiceImpl is the concrete implementation of ReportService.
type ReportServiceImpl struct {
	repo          ReportRepository
	mongoRepo     ReportMongoRepository
	pgSectionRepo reportshared.ReportSectionRepository // Postgres section repository
	// artifactsRepo   ReportArtifactsRepository
	// auditLogger AuditLogger
	// authorizer  Authorizer
	//coCRepo     GormCoCRepo
}

func NewReportService(
	repo ReportRepository,
	mongoRepo ReportMongoRepository,
	pgSectionRepo reportshared.ReportSectionRepository,
	// storage Storage,
	// auditLogger AuditLogger,
	// authorizer Authorizer,
	//coCRepo GormCoCRepo,
) ReportService {
	return &ReportServiceImpl{
		repo:          repo,
		mongoRepo:     mongoRepo,
		pgSectionRepo: pgSectionRepo,
		// storage:     storage,
		// auditLogger: auditLogger,
		// authorizer:  authorizer,
		//coCRepo:     coCRepo,
	}
}

// GenerateReport creates a new report for a given case and examiner.
// Here you could include more logic such as fetching case data, formatting content, etc.
// services_/report/service_impl.go
func (s *ReportServiceImpl) GenerateReport(
	ctx context.Context,
	caseID, examinerID, tenantID, teamID uuid.UUID,
) (*Report, error) {
	now := time.Now()

	// 1) Create Postgres report metadata (includes tenant/team)
	report := &Report{
		ID:         uuid.New(),
		TenantID:   tenantID, // NEW
		TeamID:     teamID,   // NEW
		CaseID:     caseID,
		ExaminerID: examinerID,
		Status:     "draft",
		Version:    1,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	// 2) Generate MongoID for content
	mongoID := primitive.NewObjectID()
	report.MongoID = mongoID.Hex()
	// 3) Save metadata in Postgres
	if err := s.repo.SaveReport(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to generate report metadata: %w", err)
	}

	// 4) Default sections (ensure timestamps)
	defaultSections := []ReportSection{
		{ID: primitive.NewObjectID(), Title: "Case Identification", Content: "", Order: 1, CreatedAt: now, UpdatedAt: now},
		{ID: primitive.NewObjectID(), Title: "Scope and Objectives", Content: "", Order: 2, CreatedAt: now, UpdatedAt: now},
		{ID: primitive.NewObjectID(), Title: "Evidence Summary", Content: "", Order: 3, CreatedAt: now, UpdatedAt: now},
		{ID: primitive.NewObjectID(), Title: "Tools and Methodologies", Content: "", Order: 4, CreatedAt: now, UpdatedAt: now},
		{ID: primitive.NewObjectID(), Title: "Findings", Content: "", Order: 5, CreatedAt: now, UpdatedAt: now},
		{ID: primitive.NewObjectID(), Title: "Interpretation and Analysis", Content: "", Order: 6, CreatedAt: now, UpdatedAt: now},
		{ID: primitive.NewObjectID(), Title: "Limitations", Content: "", Order: 7, CreatedAt: now, UpdatedAt: now},
		{ID: primitive.NewObjectID(), Title: "Conclusion", Content: "", Order: 8, CreatedAt: now, UpdatedAt: now},
		{ID: primitive.NewObjectID(), Title: "Appendices", Content: "", Order: 9, CreatedAt: now, UpdatedAt: now},
		{ID: primitive.NewObjectID(), Title: "Certification", Content: "", Order: 10, CreatedAt: now, UpdatedAt: now},
	}

	// 5) Save content in Mongo WITH tenant/team
	// 4b) Insert default sections into Postgres report_sections
	log.Printf("[DEBUG] pgSectionRepo is nil? %v (type: %T)", s.pgSectionRepo == nil, s.pgSectionRepo)
	for _, sec := range defaultSections {
		pgSection := &reportshared.ReportSection{
			ID:        sec.ID.Hex(),
			ReportID:  report.ID.String(),
			Title:     sec.Title,
			Content:   sec.Content,
			Order:     sec.Order,
			CreatedAt: sec.CreatedAt,
			UpdatedAt: sec.UpdatedAt,
		}
		log.Printf("[DEBUG] Inserting section into Postgres: ID=%s, ReportID=%s, Title=%s", pgSection.ID, pgSection.ReportID, pgSection.Title)
		if s.pgSectionRepo != nil {
			if err := s.pgSectionRepo.CreateSection(ctx, pgSection); err != nil {
				log.Printf("[ERROR] Failed to insert section into Postgres: ID=%s, err=%v", pgSection.ID, err)
				return nil, fmt.Errorf("failed to create default section '%s' in Postgres: %w", sec.Title, err)
			}
		} else {
			log.Printf("[WARN] pgSectionRepo is nil, skipping Postgres section insert for section: %s", pgSection.Title)
		}
	}

	reportContent := &ReportContentMongo{
		ID:        mongoID,
		ReportID:  report.ID.String(),
		Sections:  defaultSections,
		TenantID:  tenantID.String(),
		TeamID:    teamID.String(),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.mongoRepo.SaveReportContent(ctx, reportContent); err != nil {
		return nil, fmt.Errorf("failed to save report content in Mongo: %w", err)
	}
	return report, nil

	// SaveReport persists a report to the repository.
}
func (s *ReportServiceImpl) SaveReport(ctx context.Context, report *Report) error {
	if report.ID == uuid.Nil {
		report.ID = uuid.New()
	}
	return s.repo.SaveReport(ctx, report)
}

// GetReportByID retrieves a report by its hex string ID.
func (s *ReportServiceImpl) GetReportByID(ctx context.Context, reportID string) (*Report, error) {
	// If your repo expects a hex string, pass directly
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

func (s *ReportServiceImpl) DownloadReport(ctx context.Context, reportID uuid.UUID) (*ReportWithContent, error) {
	// 1) Fetch Postgres metadata (also gives us TenantID/TeamID)
	meta, err := s.repo.GetByID(ctx, reportID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch report metadata: %w", err)
	}

	// 2) Optionally fetch Mongo content (scoped by tenant/team)
	var contentSections []ReportSection
	if meta.MongoID != "" {
		mongoID, err := primitive.ObjectIDFromHex(meta.MongoID)
		if err != nil {
			return nil, fmt.Errorf("failed to convert MongoID to ObjectID: %w", err)
		}

		content, err := s.mongoRepo.GetReportContent(ctx, mongoID, meta.TenantID.String(), meta.TeamID.String())
		if err != nil {
			return nil, fmt.Errorf("failed to fetch content from MongoDB: %w", err)
		}
		if content != nil {
			contentSections = content.Sections
		}
	}

	// 3) Fallback: ensure non-nil slice
	if contentSections == nil {
		contentSections = []ReportSection{}
	}

	return &ReportWithContent{
		Metadata: meta,
		Content:  contentSections,
	}, nil
}

func (s *ReportServiceImpl) UpdateReportSection(
	ctx context.Context,
	reportUUID uuid.UUID,
	sectionID primitive.ObjectID,
	newContent string,
) error {
	return s.UpdateCustomSectionContent(ctx, reportUUID, sectionID, newContent)
}

func (s *ReportServiceImpl) DownloadReportAsJSON(ctx context.Context, reportID uuid.UUID) ([]byte, error) {
	report, err := s.DownloadReport(ctx, reportID)
	if err != nil {
		return nil, err
	}
	return json.Marshal(report)
}

// at file scope (or inside the same file above DownloadReportAsPDF)
type embeddedImage struct {
	Mimetype string
	Data     []byte
}

func extractDataURLImages(html string) (cleanHTML string, imgs []embeddedImage) {
	// local regex so it’s not a global unused symbol
	imgTagRe := regexp.MustCompile(`(?i)<img[^>]+src=["']data:(image/(?:png|jpeg|jpg));base64,([^"']+)["'][^>]*>`)
	out := imgTagRe.ReplaceAllStringFunc(html, func(tag string) string {
		m := imgTagRe.FindStringSubmatch(tag)
		if len(m) != 3 {
			return ""
		}
		mime := m[1]
		b64 := strings.NewReplacer(" ", "", "\n", "").Replace(m[2])
		data, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			return ""
		}
		imgs = append(imgs, embeddedImage{Mimetype: mime, Data: data})
		return "" // strip <img> from HTML; we’ll render images separately
	})
	return out, imgs
}



func (s *ReportServiceImpl) DownloadReportAsPDF(ctx context.Context, reportID uuid.UUID) ([]byte, error) {
	// Retrieve report data
	rpt, err := s.DownloadReport(ctx, reportID)
	if err != nil {
		return nil, err
	}

	// Resolve template path
	path, err := filepath.Abs("services_/report/report.html")
	if err != nil {
		log.Printf("Failed to resolve template path: %v", err)
		return nil, fmt.Errorf("template path: %w", err)
	}

	// Define safeHTML function to prevent escaping of embedded HTML
	funcMap := template.FuncMap{
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
	}

	// Parse template with function map
	tmpl := template.New(filepath.Base(path)).Funcs(funcMap)
	tmpl, err = tmpl.ParseFiles(path)
	if err != nil {
		log.Printf("Template parsing failed: %v", err)
		return nil, fmt.Errorf("template parse: %w", err)
	}

	// Render HTML into buffer
	var htmlBuf bytes.Buffer
	if err := tmpl.Execute(&htmlBuf, rpt); err != nil {
		log.Printf("Template execution failed: %v", err)
		return nil, fmt.Errorf("template render: %w", err)
	}

	// Create PDF generator
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		log.Printf("PDF generator creation failed: %v", err)
		return nil, fmt.Errorf("pdf generator: %w", err)
	}

	// Set global options
	pdfg.Dpi.Set(300)
	pdfg.PageSize.Set(wkhtmltopdf.PageSizeA4)
	pdfg.MarginLeft.Set(15)
	pdfg.MarginRight.Set(15)
	pdfg.MarginTop.Set(15)
	pdfg.MarginBottom.Set(15)

	// Create page from rendered HTML
	page := wkhtmltopdf.NewPageReader(&htmlBuf)
	page.EnableLocalFileAccess.Set(true)
	page.PrintMediaType.Set(true)
	page.NoBackground.Set(false)

	pdfg.AddPage(page)

	// Generate PDF
	if err := pdfg.Create(); err != nil {
		log.Printf("PDF creation failed: %v", err)
		return nil, fmt.Errorf("pdf create: %w", err)
	}

	return pdfg.Bytes(), nil
}


func (s *ReportServiceImpl) UpdateCustomSectionContent(
	ctx context.Context,
	reportUUID uuid.UUID,
	sectionID primitive.ObjectID,
	newContent string,
) error {
	mongoID, tenantID, teamID, err := s.getMongoID(ctx, reportUUID)
	if err != nil {
		return err
	}
	return s.mongoRepo.UpdateSection(ctx, mongoID, sectionID, newContent, tenantID, teamID)
}

func (s *ReportServiceImpl) UpdateSectionContent(
	ctx context.Context,
	reportUUID uuid.UUID,
	sectionID primitive.ObjectID,
	newContent string,
) error {
	// delegate to the same impl
	return s.UpdateCustomSectionContent(ctx, reportUUID, sectionID, newContent)
}
func (s *ReportServiceImpl) UpdateSectionTitle(
	ctx context.Context,
	reportUUID uuid.UUID,
	sectionID primitive.ObjectID,
	newTitle string,
) error {
	mongoID, tenantID, teamID, err := s.getMongoID(ctx, reportUUID)
	if err != nil {
		return err
	}
	return s.mongoRepo.UpdateSectionTitle(ctx, mongoID, sectionID, newTitle, tenantID, teamID)
}

func (s *ReportServiceImpl) ReorderCustomSection(
	ctx context.Context,
	reportUUID uuid.UUID,
	sectionID primitive.ObjectID,
	newOrder int,
) error {
	mongoID, tenantID, teamID, err := s.getMongoID(ctx, reportUUID)
	if err != nil {
		return err
	}
	return s.mongoRepo.ReorderSection(ctx, mongoID, sectionID, newOrder, tenantID, teamID)
}

var (
	ErrInvalidReportName      = errors.New("name must be 1..255 characters")
	ErrReportNotFoundWithName = errors.New("report not found")
)

func (s *ReportServiceImpl) UpdateReportName(ctx context.Context, reportID uuid.UUID, name string) (*Report, error) {
	// 1) Normalize + validate
	trimmed := strings.TrimSpace(name)
	if l := utf8.RuneCountInString(trimmed); l == 0 || l > 255 {
		return nil, ErrInvalidReportName
	}

	// 2) (Optional) Authorization & tenancy checks.
	//    If you store tenant/team on Report, you can fetch first and check.
	//    Example:
	//    rep, err := s.repo.GetReportByID(ctx, reportID)
	//    if err != nil { return nil, err }
	//    if rep == nil { return nil, ErrReportNotFoundWithName }
	//    // TODO: authorize current principal against rep.TenantID/TeamID

	// 3) Persist via repository
	updated, err := s.repo.UpdateReportName(ctx, reportID, trimmed)
	if err != nil {
		// Map repo "not found" to a stable service error if your repo returns it
		if errors.Is(err, ErrReportNotFoundWithName) {
			return nil, ErrReportNotFoundWithName
		}
		return nil, err
	}

	// 4) (Optional) Audit event
	// if s.aud != nil {
	//     _ = s.aud.ReportRenamed(ctx, updated.ID, oldName, updated.Name, actorID)
	// }

	return updated, nil
}

// service_impl.go
// services_/report/service_impl.go
func (s *ReportServiceImpl) GetReportsByTeamID(
	ctx context.Context,
	tenantID, teamID uuid.UUID,
) ([]ReportWithDetails, error) {
	return s.repo.GetReportsByTeamID(ctx, tenantID, teamID)
}

func (s *ReportServiceImpl) GeneratePDFFromHTML(ctx context.Context, html string) ([]byte, error) {
	// Create new PDF generator
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		return nil, fmt.Errorf("failed to create PDF generator: %w", err)
	}

	// Set global options
	pdfg.Dpi.Set(300)
	pdfg.NoCollate.Set(false)
	pdfg.PageSize.Set(wkhtmltopdf.PageSizeA4)
	pdfg.Orientation.Set(wkhtmltopdf.OrientationPortrait)
	pdfg.MarginTop.Set(15)
	pdfg.MarginBottom.Set(15)
	pdfg.MarginLeft.Set(15)
	pdfg.MarginRight.Set(15)

	// Create a new page from HTML string
	page := wkhtmltopdf.NewPageReader(strings.NewReader(html))
	page.EnableLocalFileAccess.Set(true)
	page.PrintMediaType.Set(true)
	page.NoBackground.Set(false)

	// Add page to PDF generator
	pdfg.AddPage(page)

	// Generate PDF
	if err := pdfg.Create(); err != nil {
		return nil, fmt.Errorf("failed to create PDF: %w", err)
	}

	return pdfg.Bytes(), nil
}

func safeHTML(s string) template.HTML {
	return template.HTML(s)
}

func loadReportTemplate(path string) (*template.Template, error) {
	funcMap := template.FuncMap{
		"safeHTML": safeHTML,
	}

	tmpl := template.New("report_pdf.html").Funcs(funcMap)
	return tmpl.ParseFiles(path)
}