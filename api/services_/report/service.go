package report

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image/jpeg"
	"image/png"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ReportService defines the business logic for managing reports.
type ReportService interface {
	GenerateReport(ctx context.Context, caseID, examinerID, tenantID, teamID uuid.UUID) (*Report, error)
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
	ReorderCustomSection(ctx context.Context, reportUUID uuid.UUID, sectionID primitive.ObjectID, newOrder int) error // ... your existing methods ...
	ListRecentReports(ctx context.Context, opts RecentReportsOptions) ([]RecentReport, error)
	UpdateReportName(ctx context.Context, reportID uuid.UUID, name string) (*Report, error) // NEW
	GetReportsByTeamID(ctx context.Context, tenantID, teamID uuid.UUID) ([]ReportWithDetails, error)
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
	reportContent := &ReportContentMongo{
		ID:        mongoID,
		ReportID:  report.ID.String(),
		TenantID:  tenantID.String(), // NEW
		TeamID:    teamID.String(),   // NEW
		Sections:  defaultSections,
		CreatedAt: now,
		UpdatedAt: now,
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

func (s *ReportServiceImpl) DownloadReport(ctx context.Context, reportID uuid.UUID) (*ReportWithContent, error) {
	// 1) Fetch Postgres metadata (also gives us TenantID/TeamID)
	meta, err := s.repo.GetByID(ctx, reportID)
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
	rpt, err := s.DownloadReport(ctx, reportID)
	if err != nil {
		return nil, err
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, fmt.Sprintf("Report: %s", rpt.Metadata.Name))
	pdf.Ln(12)

	// If you’re using gofpdf’s HTML renderer:
	html := pdf.HTMLBasicNew()

	for _, sec := range rpt.Content {
		// title...
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(0, 8, sec.Title)
		pdf.Ln(8)

		// (optional) sanitize first; otherwise use sec.Content directly
		// sanitized := policy.Sanitize(sec.Content)
		sanitized := sec.Content

		// ⬇️ call the extractor so the function & regex are USED
		cleaned, images := extractDataURLImages(sanitized)

		// text rendering
		pdf.SetFont("Arial", "", 11)
		trimmed := strings.TrimSpace(strings.ToLower(strings.ReplaceAll(cleaned, " ", "")))
		if trimmed == "" || trimmed == "<p><br></p>" || trimmed == "<p></p>" {
			pdf.MultiCell(0, 6, "(No content provided)", "", "", false)
		} else {
			html.Write(5, cleaned)
		}
		pdf.Ln(4)

		// image rendering (the block you already have)
		for _, im := range images {
			imgType := strings.ToUpper(strings.TrimPrefix(im.Mimetype, "image/"))
			if imgType == "JPEG" {
				imgType = "JPG"
			}

			r := bytes.NewReader(im.Data)
			if imgType == "PNG" && len(im.Data) > 1_000_000 {
				if pngImg, err := png.Decode(bytes.NewReader(im.Data)); err == nil {
					var buf bytes.Buffer
					_ = jpeg.Encode(&buf, pngImg, &jpeg.Options{Quality: 85})
					r = bytes.NewReader(buf.Bytes())
					imgType = "JPG"
				}
			}

			name := fmt.Sprintf("sec-%v-%d", sec.ID, time.Now().UnixNano()) // use sec.ID or sec.ID.Hex()
			opts := gofpdf.ImageOptions{ImageType: imgType, ReadDpi: true}
			info := pdf.RegisterImageOptionsReader(name, opts, r)

			w, h := info.Width(), info.Height()
			maxW := 180.0
			if w > maxW {
				scale := maxW / w
				w = maxW
				h *= scale
			}

			x := (210.0 - w) / 2.0
			y := pdf.GetY()
			pdf.ImageOptions(name, x, y, w, 0, false, opts, 0, "")
			pdf.Ln(h + 4)
		}

		if pdf.GetY() > 260 {
			pdf.AddPage()
		}
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("pdf output: %w", err)
	}
	return buf.Bytes(), nil
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
