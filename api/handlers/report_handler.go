package handlers

import (
	"aegis-api/services_/auditlog"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	graphicalmapping "aegis-api/services_/GraphicalMapping"
	"aegis-api/services_/evidence/metadata"
	"aegis-api/services_/report"
	"aegis-api/services_/timeline"

	// removed duplicate import
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ContextAutofillResponse is the JSON payload for context autofill

// GetSectionContext returns structured context for a report section (case info, IOCs, evidence, timeline)
func (h *ReportHandler) GetSectionContext(c *gin.Context) {
	// Grab user details from context
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	email, _ := c.Get("email") // Optional, if you have this set

	actor := auditlog.Actor{
		ID:        userID.(string),
		Role:      userRole.(string),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Email:     email.(string), // Optional, if you have this header set
	}

	_, exists := c.Get("userID")
	if !exists {
		fmt.Printf("[GetSectionContext] Missing userID in context\n")

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "GET_SECTION_CONTEXT",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report_section",
				ID:   "",
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Unauthorized: missing userID in context",
		})

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing userID in context"})
		return
	}

	reportIDStr := c.Param("reportID")
	sectionIDStr := c.Param("sectionID")
	ctx := c.Request.Context()

	log.Printf("[DEBUG] GetSectionContext: reportID=%s sectionID=%s", reportIDStr, sectionIDStr)

	// Fetch report (for validation only)
	rep, err := h.ReportService.DownloadReport(ctx, uuid.MustParse(reportIDStr))
	if err != nil || rep == nil {
		log.Printf("[DEBUG] Report not found: %s", reportIDStr)
		fmt.Printf("[GetSectionContext] Report not found: %s\n", reportIDStr)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "GET_SECTION_CONTEXT",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report",
				ID:   reportIDStr,
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Report not found",
		})

		writeError(c, http.StatusNotFound, "report_not_found", "report not found")
		return
	}

	// Validate section existence
	found := false
	for _, sec := range rep.Content {
		log.Printf("[DEBUG] Checking section: %s", sec.ID.Hex())
		if sec.ID.Hex() == sectionIDStr {
			found = true
			break
		}
	}
	if !found {
		log.Printf("[DEBUG] Section not found: %s in report %s", sectionIDStr, reportIDStr)
		fmt.Printf("[GetSectionContext] Section not found: %s in report %s\n", sectionIDStr, reportIDStr)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "GET_SECTION_CONTEXT",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report_section",
				ID:   sectionIDStr,
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Section not found",
		})

		writeError(c, http.StatusNotFound, "section_not_found", "section not found")
		return
	}

	// --- Fetch Case Info (join case + report metadata) ---
	var caseInfo any
	caseDetails := map[string]any{}
	// Example: fetch from case service (pseudo-code, replace with actual call)
	if h.CaseService != nil {
		caseObj, err := h.CaseService.GetCaseByID(ctx, rep.Metadata.CaseID.String())
		if err == nil && caseObj != nil {
			// Try to type assert to struct, else marshal/unmarshal to map
			switch v := caseObj.(type) {
			case map[string]any:
				caseDetails = v
			default:
				b, _ := json.Marshal(v)
				json.Unmarshal(b, &caseDetails)
			}
		}
	}
	// Merge report metadata
	if rep.Metadata != nil {
		caseDetails["report_name"] = rep.Metadata.Name
		caseDetails["report_status"] = rep.Metadata.Status
		caseDetails["examiner_id"] = rep.Metadata.ExaminerID.String()
		caseDetails["team_id"] = rep.Metadata.TeamID.String()
		caseDetails["tenant_id"] = rep.Metadata.TenantID.String()
		caseDetails["report_created_at"] = rep.Metadata.CreatedAt
		caseDetails["report_updated_at"] = rep.Metadata.UpdatedAt
	}
	caseInfo = caseDetails

	// --- Fetch Evidence ---
	var evidence []any
	if h.EvidenceService != nil {
		evidences, err := h.EvidenceService.FindEvidenceByCaseID(rep.Metadata.CaseID)
		if err == nil {
			for _, ev := range evidences {
				var hashes map[string]string
				if err := json.Unmarshal([]byte(ev.Metadata), &hashes); err != nil {
					hashes = map[string]string{}
				}
				evidence = append(evidence, map[string]any{
					"filename": ev.Filename,
					"sha512":   hashes["sha512"],
					"sha256":   hashes["sha256"],
				})
			}
		}
	}

	// --- Fetch IOCs from IOC service ---
	var iocs []any
	if h.IOCService != nil {
		iocList, err := h.IOCService.ListIOCsByCase(rep.Metadata.CaseID.String())
		if err == nil {
			for _, ioc := range iocList {
				iocs = append(iocs, map[string]any{
					"id":         ioc.ID,
					"type":       ioc.Type,
					"value":      ioc.Value,
					"created_at": ioc.CreatedAt,
				})
			}
		}
	}

	// --- Fetch Timeline Events ---
	var timeline []any
	if h.TimelineService != nil {
		events, err := h.TimelineService.ListEvents(rep.Metadata.CaseID.String())
		if err == nil {
			for _, ev := range events {
				timeline = append(timeline, map[string]any{
					"id":          ev.ID,
					"description": ev.Description,
					"severity":    ev.Severity,
					"analyst":     ev.AnalystName,
					"date":        ev.Date,
					"time":        ev.Time,
					"tags":        ev.Tags,
				})
			}
		}
	}

	resp := ContextAutofillResponse{
		CaseInfo: caseInfo,
		IOCs:     iocs,
		Evidence: evidence,
		Timeline: timeline,
	}
	log.Printf("[DEBUG] Returning rich context autofill for section %s in report %s", sectionIDStr, reportIDStr)
	fmt.Printf("[GetSectionContext] Successfully retrieved section context for section %s in report %s\n", sectionIDStr, reportIDStr)

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "GET_SECTION_CONTEXT",
		Actor:  actor,
		Target: auditlog.Target{
			Type: "report_section",
			ID:   sectionIDStr,
		},
		Service:     "report",
		Status:      "SUCCESS",
		Description: "Section context retrieved successfully",
	})

	c.JSON(http.StatusOK, resp)
}

// ContextAutofillResponse is the JSON payload for context autofill
type ContextAutofillResponse struct {
	CaseInfo any `json:"case_info"`
	IOCs     any `json:"iocs"`
	Evidence any `json:"evidence"`
	Timeline any `json:"timeline"`
}

// ReportHandler handles HTTP requests for reports.
type ReportHandler struct {
	ReportService   report.ReportService
	EvidenceService interface {
		FindEvidenceByCaseID(caseID uuid.UUID) ([]metadata.Evidence, error)
	}
	TimelineService interface {
		ListEvents(caseID string) ([]*timeline.TimelineEventResponse, error)
	}
	CaseService interface {
		GetCaseByID(ctx context.Context, caseID string) (any, error)
	}
	IOCService interface {
		ListIOCsByCase(caseID string) ([]*graphicalmapping.IOC, error)
	}
	auditLogger *auditlog.AuditLogger
}

// GetReportByID returns a report by its ID
func (h *ReportHandler) GetReportByID(c *gin.Context) {
	// Grab user details from context
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	email, _ := c.Get("email") // Optional, if you have this set

	actor := auditlog.Actor{
		ID:        userID.(string),
		Role:      userRole.(string),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Email:     email.(string), // Optional, if you have this header set
	}

	_, exists := c.Get("userID")
	if !exists {
		fmt.Printf("[GetReportByID] Missing userID in context\n")

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "GET_REPORT",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report",
				ID:   "",
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Unauthorized: missing userID in context",
		})

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing userID in context"})
		return
	}

	reportIDStr := c.Param("reportID")
	ctx := c.Request.Context()
	reportID, err := uuid.Parse(reportIDStr)
	if err != nil {
		fmt.Printf("[GetReportByID] Invalid report ID: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "GET_REPORT",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report",
				ID:   reportIDStr,
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Invalid report ID",
		})

		writeError(c, http.StatusBadRequest, "invalid_report_id", "Invalid report ID")
		return
	}
	rep, err := h.ReportService.DownloadReport(ctx, reportID)
	if err != nil || rep == nil {
		fmt.Printf("[GetReportByID] Report not found: %s\n", reportIDStr)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "GET_REPORT",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report",
				ID:   reportIDStr,
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Report not found",
		})

		writeError(c, http.StatusNotFound, "report_not_found", "Report not found")
		return
	}
	fmt.Printf("[GetReportByID] Successfully retrieved report: %s\n", reportIDStr)

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "GET_REPORT",
		Actor:  actor,
		Target: auditlog.Target{
			Type: "report",
			ID:   reportIDStr,
		},
		Service:     "report",
		Status:      "SUCCESS",
		Description: "Report retrieved successfully",
	})

	c.JSON(http.StatusOK, rep)
}

func NewReportHandler(s report.ReportService, auditLogger *auditlog.AuditLogger) *ReportHandler {
	// Usage: NewReportHandler(reportService)
	return &ReportHandler{ReportService: s, auditLogger: auditLogger}
}

// Use this constructor to inject dependencies for context autofill
func NewReportHandlerWithDeps(
	reportService report.ReportService,
	evidenceService interface {
		FindEvidenceByCaseID(caseID uuid.UUID) ([]metadata.Evidence, error)
	},
	timelineService interface {
		ListEvents(caseID string) ([]*timeline.TimelineEventResponse, error)
	},
	caseService interface {
		GetCaseByID(ctx context.Context, caseID string) (any, error)
	},
	iocService interface {
		ListIOCsByCase(caseID string) ([]*graphicalmapping.IOC, error)
	},
	auditLogger *auditlog.AuditLogger,
) *ReportHandler {
	return &ReportHandler{
		ReportService:   reportService,
		EvidenceService: evidenceService,
		TimelineService: timelineService,
		CaseService:     caseService,
		IOCService:      iocService,
		auditLogger:     auditLogger,
	}
}

// handlers/report_handler.go

type Claims struct {
	UserID   uuid.UUID
	TeamID   uuid.UUID
	TenantID uuid.UUID
	Role     string
	Email    string
	FullName string
}

// helper to pull a string from gin context
func getStr(c *gin.Context, key string) string {
	if v, ok := c.Get(key); ok {
		if s, ok2 := v.(string); ok2 {
			return s
		}
	}
	return ""
}

func ClaimsFromCtx(c *gin.Context) (Claims, error) {
	// 1) If a fully-formed "claims" was set, prefer it
	if v, ok := c.Get("claims"); ok {
		if cl, ok2 := v.(Claims); ok2 {
			return cl, nil
		}
	}

	// 2) Rebuild from individual keys set by AuthMiddleware
	userIDStr := getStr(c, "userID")
	tenantIDStr := getStr(c, "tenantID")
	teamIDStr := getStr(c, "teamID")

	if userIDStr == "" || tenantIDStr == "" {
		return Claims{}, errors.New("missing required claims")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return Claims{}, err
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return Claims{}, err
	}

	// team can be optional for some roles; use uuid.Nil if empty
	var teamID uuid.UUID
	if teamIDStr != "" {
		teamID, err = uuid.Parse(teamIDStr)
		if err != nil {
			return Claims{}, err
		}
	}

	return Claims{
		UserID:   userID,
		TenantID: tenantID,
		TeamID:   teamID,
		Role:     getStr(c, "userRole"),
		Email:    getStr(c, "email"),
		FullName: getStr(c, "fullName"),
	}, nil
}

// Keep MustClaims as a thin wrapper, but now it won’t panic on normal paths
func MustClaims(c *gin.Context) Claims {
	cl, err := ClaimsFromCtx(c)
	if err != nil {
		panic(err)
	}
	return cl
}

// 23505 is Postgres unique_violation
func IsUniqueViolation(err error) bool {
	var pgxErr *pgconn.PgError
	if errors.As(err, &pgxErr) {
		return pgxErr.Code == "23505"
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return string(pqErr.Code) == "23505"
	}
	return false
}

// GenerateReport creates a new report for a case.
// handlers/report.go
func (h *ReportHandler) GenerateReport(c *gin.Context) {
	// Grab user details from context
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	email, _ := c.Get("email") // Optional, if you have this set

	actor := auditlog.Actor{
		ID:        userID.(string),
		Role:      userRole.(string),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Email:     email.(string), // Optional, if you have this header set
	}

	_, exists := c.Get("userID")
	if !exists {
		fmt.Printf("[GenerateReport] Missing userID in context\n")

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "GENERATE_REPORT",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report",
				ID:   "",
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Unauthorized: missing userID in context",
		})

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing userID in context"})
		return
	}

	caseIDStr := c.Param("caseID")
	caseID, err := uuid.Parse(caseIDStr)
	if err != nil {
		fmt.Printf("[GenerateReport] Invalid case ID: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "GENERATE_REPORT",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "case",
				ID:   caseIDStr,
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Invalid case ID",
		})

		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case id"})
		return
	}
	ctx := c.Request.Context()

	// 1) If a report already exists for this case, return it
	if list, err := h.ReportService.GetReportsByCaseID(ctx, caseID); err == nil && len(list) > 0 {
		fmt.Printf("[GenerateReport] Report already exists for case %s\n", caseIDStr)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "GENERATE_REPORT",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report",
				ID:   list[0].ID.String(),
			},
			Service:     "report",
			Status:      "SUCCESS",
			Description: "Report already exists, returned existing",
		})

		c.JSON(http.StatusOK, gin.H{"id": list[0].ID})
		return
	}

	// 2) Otherwise create it
	claims := MustClaims(c)
	rep, genErr := h.ReportService.GenerateReport(ctx, caseID, claims.UserID, claims.TenantID, claims.TeamID)
	if genErr != nil {
		// If a parallel request created it, surface the existing one
		if IsUniqueViolation(genErr) {
			if again, e2 := h.ReportService.GetReportsByCaseID(ctx, caseID); e2 == nil && len(again) > 0 {
				fmt.Printf("[GenerateReport] Race condition resolved, returned existing report for case %s\n", caseIDStr)

				h.auditLogger.Log(c, auditlog.AuditLog{
					Action: "GENERATE_REPORT",
					Actor:  actor,
					Target: auditlog.Target{
						Type: "report",
						ID:   again[0].ID.String(),
					},
					Service:     "report",
					Status:      "SUCCESS",
					Description: "Race condition resolved, returned existing report",
				})

				c.JSON(http.StatusOK, gin.H{"id": again[0].ID})
				return
			}
		}
		fmt.Printf("[GenerateReport] Failed to generate report: %v\n", genErr)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "GENERATE_REPORT",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "case",
				ID:   caseIDStr,
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Failed to generate report: " + genErr.Error(),
		})

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate report"})
		return
	}

	// ADD CRITICAL NIL CHECKS HERE
	if rep == nil {
		fmt.Printf("[GenerateReport] Generated report is nil for case %s\n", caseIDStr)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "GENERATE_REPORT",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "case",
				ID:   caseIDStr,
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Generated report is nil",
		})

		c.JSON(http.StatusInternalServerError, gin.H{"error": "generated report is nil"})
		return
	}

	if rep.CaseID == uuid.Nil {
		fmt.Printf("[GenerateReport] Report has invalid case ID for case %s\n", caseIDStr)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "GENERATE_REPORT",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "case",
				ID:   caseIDStr,
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Report has invalid case ID",
		})

		c.JSON(http.StatusInternalServerError, gin.H{"error": "report has invalid case ID"})
		return
	}

	// Use rep.CaseID (not rep.Metadata)
	caseObj, caseErr := h.CaseService.GetCaseByID(ctx, rep.CaseID.String())
	var caseMap map[string]any
	if caseErr == nil && caseObj != nil {
		switch v := caseObj.(type) {
		case map[string]any:
			caseMap = v
		default:
			b, _ := json.Marshal(v)
			json.Unmarshal(b, &caseMap)
		}
	} else {
		caseMap = map[string]any{}
	}

	// Safely extract case properties with fallbacks
	caseIDVal := caseMap["ID"]
	caseName := caseMap["Name"]
	caseStatus := caseMap["Status"]
	description := caseMap["Description"]
	createdAt := caseMap["CreatedAt"]
	updatedAt := caseMap["UpdatedAt"]

	fmt.Printf("[GenerateReport] Successfully generated report %s for case %s\n", rep.ID.String(), caseIDStr)

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "GENERATE_REPORT",
		Actor:  actor,
		Target: auditlog.Target{
			Type: "report",
			ID:   rep.ID.String(),
		},
		Service:     "report",
		Status:      "SUCCESS",
		Description: "Report generated successfully",
	})

	c.JSON(http.StatusOK, gin.H{
		"id":            rep.ID,
		"name":          rep.Name,
		"status":        rep.Status,
		"last_modified": rep.UpdatedAt,
		"case_id":       caseIDVal,
		"case_name":     caseName,
		"case_status":   caseStatus,
		"description":   description,
		"created_at":    createdAt,
		"updated_at":    updatedAt,
	})
}

// writeError is a helper to send error responses in a consistent format.
func writeError(c *gin.Context, status int, code, msg string) {
	c.AbortWithStatusJSON(status, gin.H{
		"error": gin.H{
			"code":    code,
			"message": msg,
		},
	})
}

func logWithCtx(level, msg string, c *gin.Context, kv map[string]any) {
	log.Printf("%s %s path=%s method=%s ip=%s ctx=%v",
		strings.ToUpper(level), msg, c.FullPath(), c.Request.Method, c.ClientIP(), kv)
}

func (h *ReportHandler) UpdateSectionContent(c *gin.Context) {
	// Grab user details from context
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	email, _ := c.Get("email") // Optional, if you have this set

	actor := auditlog.Actor{
		ID:        userID.(string),
		Role:      userRole.(string),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Email:     email.(string), // Optional, if you have this header set
	}

	_, exists := c.Get("userID")
	if !exists {
		fmt.Printf("[UpdateSectionContent] Missing userID in context\n")

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "UPDATE_SECTION_CONTENT",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report_section",
				ID:   "",
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Unauthorized: missing userID in context",
		})

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing userID in context"})
		return
	}

	reportIDStr := c.Param("reportID")
	reportUUID, err := uuid.Parse(reportIDStr)
	if err != nil {
		logWithCtx("warn", "invalid report ID", c, map[string]any{"reportID": reportIDStr, "err": err.Error()})
		fmt.Printf("[UpdateSectionContent] Invalid report ID: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "UPDATE_SECTION_CONTENT",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report",
				ID:   reportIDStr,
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Invalid report ID",
		})

		writeError(c, http.StatusBadRequest, "invalid_report_id", "invalid report ID")
		return
	}

	sectionIDStr := c.Param("sectionID")
	sectionID, err := primitive.ObjectIDFromHex(sectionIDStr)
	if err != nil {
		logWithCtx("warn", "invalid section ID", c, map[string]any{"reportID": reportUUID.String(), "sectionID": sectionIDStr, "err": err.Error()})
		fmt.Printf("[UpdateSectionContent] Invalid section ID: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "UPDATE_SECTION_CONTENT",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report_section",
				ID:   sectionIDStr,
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Invalid section ID",
		})

		writeError(c, http.StatusBadRequest, "invalid_section_id", "invalid section ID")
		return
	}

	var req struct {
		Content string `json:"content"`
	}
	// ✅ allow empty content (users may clear a section)
	if err := c.ShouldBindJSON(&req); err != nil {
		logWithCtx("warn", "invalid body", c, map[string]any{
			"reportID":  reportUUID.String(),
			"sectionID": sectionID.Hex(),
			"err":       err,
		})
		fmt.Printf("[UpdateSectionContent] Invalid request body: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "UPDATE_SECTION_CONTENT",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report_section",
				ID:   sectionID.Hex(),
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Invalid request body",
		})

		writeError(c, http.StatusBadRequest, "invalid_body", "invalid request body")
		return
	}

	if err := h.ReportService.UpdateCustomSectionContent(c.Request.Context(), reportUUID, sectionID, req.Content); err != nil {
		switch {
		case errors.Is(err, report.ErrReportNotFound), errors.Is(err, report.ErrMongoReportNotFound):
			logWithCtx("info", "report not found", c, map[string]any{"reportID": reportUUID.String(), "sectionID": sectionID.Hex(), "err": err.Error()})
			fmt.Printf("[UpdateSectionContent] Report not found: %s\n", reportUUID.String())

			h.auditLogger.Log(c, auditlog.AuditLog{
				Action: "UPDATE_SECTION_CONTENT",
				Actor:  actor,
				Target: auditlog.Target{
					Type: "report",
					ID:   reportUUID.String(),
				},
				Service:     "report",
				Status:      "FAILED",
				Description: "Report not found",
			})

			writeError(c, http.StatusNotFound, "report_not_found", "report not found")
			return
		case errors.Is(err, report.ErrSectionNotFound):
			logWithCtx("info", "section not found", c, map[string]any{"reportID": reportUUID.String(), "sectionID": sectionID.Hex(), "err": err.Error()})
			fmt.Printf("[UpdateSectionContent] Section not found: %s\n", sectionID.Hex())

			h.auditLogger.Log(c, auditlog.AuditLog{
				Action: "UPDATE_SECTION_CONTENT",
				Actor:  actor,
				Target: auditlog.Target{
					Type: "report_section",
					ID:   sectionID.Hex(),
				},
				Service:     "report",
				Status:      "FAILED",
				Description: "Section not found",
			})

			writeError(c, http.StatusNotFound, "section_not_found", "section not found")
			return
		default:
			low := strings.ToLower(err.Error())
			if strings.Contains(low, "not found") {
				logWithCtx("info", "resource not found", c, map[string]any{"reportID": reportUUID.String(), "sectionID": sectionID.Hex(), "err": err.Error()})
				fmt.Printf("[UpdateSectionContent] Resource not found: %v\n", err)

				h.auditLogger.Log(c, auditlog.AuditLog{
					Action: "UPDATE_SECTION_CONTENT",
					Actor:  actor,
					Target: auditlog.Target{
						Type: "report_section",
						ID:   sectionID.Hex(),
					},
					Service:     "report",
					Status:      "FAILED",
					Description: "Resource not found",
				})

				writeError(c, http.StatusNotFound, "not_found", "resource not found")
				return
			}
			logWithCtx("error", "update section failed", c, map[string]any{"reportID": reportUUID.String(), "sectionID": sectionID.Hex(), "err": err.Error()})
			fmt.Printf("[UpdateSectionContent] Failed to update section: %v\n", err)

			h.auditLogger.Log(c, auditlog.AuditLog{
				Action: "UPDATE_SECTION_CONTENT",
				Actor:  actor,
				Target: auditlog.Target{
					Type: "report_section",
					ID:   sectionID.Hex(),
				},
				Service:     "report",
				Status:      "FAILED",
				Description: "Failed to update section content: " + err.Error(),
			})

			writeError(c, http.StatusInternalServerError, "update_failed", "failed to update section content")
			return
		}
	}

	logWithCtx("info", "section content updated", c, map[string]any{"reportID": reportUUID.String(), "sectionID": sectionID.Hex()})
	fmt.Printf("[UpdateSectionContent] Successfully updated section content for section %s in report %s\n", sectionID.Hex(), reportUUID.String())

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "UPDATE_SECTION_CONTENT",
		Actor:  actor,
		Target: auditlog.Target{
			Type: "report_section",
			ID:   sectionID.Hex(),
		},
		Service:     "report",
		Status:      "SUCCESS",
		Description: "Section content updated successfully",
	})

	c.Status(http.StatusNoContent)
}

// -----------------
// DownloadReportPDF returns the report as PDF.
func (h *ReportHandler) DownloadReportPDF(c *gin.Context) {
	// Grab user details from context
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	email, _ := c.Get("email") // Optional, if you have this set

	actor := auditlog.Actor{
		ID:        userID.(string),
		Role:      userRole.(string),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Email:     email.(string), // Optional, if you have this header set
	}

	_, exists := c.Get("userID")
	if !exists {
		fmt.Printf("[DownloadReportPDF] Missing userID in context\n")

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "DOWNLOAD_REPORT_PDF",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report",
				ID:   "",
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Unauthorized: missing userID in context",
		})

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing userID in context"})
		return
	}

	reportIDStr := c.Param("reportID")
	reportID, err := uuid.Parse(reportIDStr)
	if err != nil {
		fmt.Printf("[DownloadReportPDF] Invalid report ID: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "DOWNLOAD_REPORT_PDF",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report",
				ID:   reportIDStr,
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Invalid report ID",
		})

		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report ID"})
		return
	}

	pdfBytes, err := h.ReportService.DownloadReportAsPDF(c.Request.Context(), reportID)
	if err != nil {
		fmt.Printf("[DownloadReportPDF] Failed to generate PDF: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "DOWNLOAD_REPORT_PDF",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report",
				ID:   reportIDStr,
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Failed to generate PDF: " + err.Error(),
		})

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate PDF"})
		return
	}
	fmt.Printf("[DownloadReportPDF] Successfully downloaded PDF for report %s\n", reportIDStr)

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "DOWNLOAD_REPORT_PDF",
		Actor:  actor,
		Target: auditlog.Target{
			Type: "report",
			ID:   reportIDStr,
		},
		Service:     "report",
		Status:      "SUCCESS",
		Description: "Report PDF downloaded successfully",
	})

	c.Header("Content-Disposition", "attachment; filename=report_"+reportIDStr+".pdf")
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// func (h *ReportHandler) DownloadReportPDF(c *gin.Context) {
// 	reportIDStr := c.Param("reportID")
// 	reportID, err := uuid.Parse(reportIDStr)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report ID"})
// 		return
// 	}

// 	// Check if HTML is provided in request body (for custom rendering)
// 	var requestBody struct {
// 		HTML string `json:"html"`
// 	}
	
// 	if c.Request.Method == "POST" {
// 		if err := c.ShouldBindJSON(&requestBody); err == nil && requestBody.HTML != "" {
// 			// Use HTML-based PDF generation (for cover page support)
// 			pdfBytes, err := h.ReportService.GeneratePDFFromHTML(c.Request.Context(), requestBody.HTML)
// 			if err != nil {
// 				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate PDF from HTML"})
// 				return
// 			}
// 			c.Header("Content-Disposition", "attachment; filename=report_"+reportIDStr+".pdf")
// 			c.Data(http.StatusOK, "application/pdf", pdfBytes)
// 			return
// 		}
// 	}

// 	// Default: use existing gofpdf generation
// 	pdfBytes, err := h.ReportService.DownloadReportAsPDF(c.Request.Context(), reportID)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate PDF"})
// 		return
// 	}

// 	c.Header("Content-Disposition", "attachment; filename=report_"+reportIDStr+".pdf")
// 	c.Data(http.StatusOK, "application/pdf", pdfBytes)
// }

// DownloadReportJSON returns the report as JSON.
func (h *ReportHandler) DownloadReportJSON(c *gin.Context) {
	// Grab user details from context
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	email, _ := c.Get("email") // Optional, if you have this set

	actor := auditlog.Actor{
		ID:        userID.(string),
		Role:      userRole.(string),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Email:     email.(string), // Optional, if you have this header set
	}

	_, exists := c.Get("userID")
	if !exists {
		fmt.Printf("[DownloadReportJSON] Missing userID in context\n")

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "DOWNLOAD_REPORT_JSON",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report",
				ID:   "",
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Unauthorized: missing userID in context",
		})

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing userID in context"})
		return
	}

	reportIDStr := c.Param("reportID")
	reportID, err := uuid.Parse(reportIDStr)
	if err != nil {
		fmt.Printf("[DownloadReportJSON] Invalid report ID: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "DOWNLOAD_REPORT_JSON",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report",
				ID:   reportIDStr,
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Invalid report ID",
		})

		writeError(c, http.StatusBadRequest, "invalid_report_id", "invalid report ID")
		return
	}

	jsonBytes, err := h.ReportService.DownloadReportAsJSON(c.Request.Context(), reportID)
	if err != nil {
		logWithCtx("error", "download json failed", c, map[string]any{"reportID": reportIDStr, "err": err.Error()})
		fmt.Printf("[DownloadReportJSON] Failed to generate JSON: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "DOWNLOAD_REPORT_JSON",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report",
				ID:   reportIDStr,
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Failed to generate JSON: " + err.Error(),
		})

		writeError(c, http.StatusInternalServerError, "generate_json_failed", "failed to generate JSON")
		return
	}
	fmt.Printf("[DownloadReportJSON] Successfully downloaded JSON for report %s\n", reportIDStr)

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "DOWNLOAD_REPORT_JSON",
		Actor:  actor,
		Target: auditlog.Target{
			Type: "report",
			ID:   reportIDStr,
		},
		Service:     "report",
		Status:      "SUCCESS",
		Description: "Report JSON downloaded successfully",
	})

	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename=report_"+reportIDStr+".json")
	c.Header("Cache-Control", "no-store")
	c.Data(http.StatusOK, "application/json", jsonBytes)
}

// DeleteReport deletes a report by ID.
func (h *ReportHandler) DeleteReport(c *gin.Context) {
	// Grab user details from context
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	email, _ := c.Get("email") // Optional, if you have this set

	actor := auditlog.Actor{
		ID:        userID.(string),
		Role:      userRole.(string),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Email:     email.(string), // Optional, if you have this header set
	}

	_, exists := c.Get("userID")
	if !exists {
		fmt.Printf("[DeleteReport] Missing userID in context\n")

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "DELETE_REPORT",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report",
				ID:   "",
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Unauthorized: missing userID in context",
		})

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing userID in context"})
		return
	}

	reportIDStr := c.Param("reportID")
	reportID, err := uuid.Parse(reportIDStr)
	if err != nil {
		fmt.Printf("[DeleteReport] Invalid report ID: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "DELETE_REPORT",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report",
				ID:   reportIDStr,
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Invalid report ID",
		})

		writeError(c, http.StatusBadRequest, "invalid_report_id", "invalid report ID")
		return
	}

	if err := h.ReportService.DeleteReportByID(c.Request.Context(), reportID); err != nil {
		// map known errors if you expose them from the repo/service
		if errors.Is(err, report.ErrReportNotFound) {
			fmt.Printf("[DeleteReport] Report not found: %s\n", reportIDStr)

			h.auditLogger.Log(c, auditlog.AuditLog{
				Action: "DELETE_REPORT",
				Actor:  actor,
				Target: auditlog.Target{
					Type: "report",
					ID:   reportIDStr,
				},
				Service:     "report",
				Status:      "FAILED",
				Description: "Report not found",
			})

			writeError(c, http.StatusNotFound, "report_not_found", "report not found")
			return
		}
		logWithCtx("error", "delete report failed", c, map[string]any{"reportID": reportIDStr, "err": err.Error()})
		fmt.Printf("[DeleteReport] Failed to delete report: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "DELETE_REPORT",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report",
				ID:   reportIDStr,
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Failed to delete report: " + err.Error(),
		})

		writeError(c, http.StatusInternalServerError, "delete_failed", "failed to delete report")
		return
	}
	fmt.Printf("[DeleteReport] Successfully deleted report: %s\n", reportIDStr)

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "DELETE_REPORT",
		Actor:  actor,
		Target: auditlog.Target{
			Type: "report",
			ID:   reportIDStr,
		},
		Service:     "report",
		Status:      "SUCCESS",
		Description: "Report deleted successfully",
	})

	c.JSON(http.StatusOK, gin.H{"status": "report deleted"})
}

// AddSection handles adding a new custom section to a report.
func (h *ReportHandler) AddSection(c *gin.Context) {
	// Grab user details from context
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	email, _ := c.Get("email") // Optional, if you have this set

	actor := auditlog.Actor{
		ID:        userID.(string),
		Role:      userRole.(string),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Email:     email.(string), // Optional, if you have this header set
	}

	_, exists := c.Get("userID")
	if !exists {
		fmt.Printf("[AddSection] Missing userID in context\n")

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "ADD_SECTION",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report_section",
				ID:   "",
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Unauthorized: missing userID in context",
		})

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing userID in context"})
		return
	}

	reportIDStr := c.Param("reportID")
	reportUUID, err := uuid.Parse(reportIDStr)
	if err != nil {
		fmt.Printf("[AddSection] Invalid report ID: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "ADD_SECTION",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report",
				ID:   reportIDStr,
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Invalid report ID",
		})

		writeError(c, http.StatusBadRequest, "invalid_report_id", "invalid report ID")
		return
	}

	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		Order   int    `json:"order"` // 1-based; service can clamp/append if <=0
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("[AddSection] Invalid request body: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "ADD_SECTION",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report",
				ID:   reportIDStr,
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Invalid request body",
		})

		writeError(c, http.StatusBadRequest, "invalid_body", "invalid request body")
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		fmt.Printf("[AddSection] Title is required\n")

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "ADD_SECTION",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report",
				ID:   reportIDStr,
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Title is required",
		})

		writeError(c, http.StatusBadRequest, "invalid_title", "title is required")
		return
	}
	// Optional title length guardrail
	if len(req.Title) > 200 {
		fmt.Printf("[AddSection] Title too long\n")

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "ADD_SECTION",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report",
				ID:   reportIDStr,
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Title is too long",
		})

		writeError(c, http.StatusBadRequest, "invalid_title", "title is too long")
		return
	}

	if err := h.ReportService.AddCustomSection(c.Request.Context(), reportUUID, req.Title, req.Content, req.Order); err != nil {
		switch {
		case errors.Is(err, report.ErrReportNotFound), errors.Is(err, report.ErrMongoReportNotFound):
			fmt.Printf("[AddSection] Report not found: %s\n", reportIDStr)

			h.auditLogger.Log(c, auditlog.AuditLog{
				Action: "ADD_SECTION",
				Actor:  actor,
				Target: auditlog.Target{
					Type: "report",
					ID:   reportIDStr,
				},
				Service:     "report",
				Status:      "FAILED",
				Description: "Report not found",
			})

			writeError(c, http.StatusNotFound, "report_not_found", "report not found")
		case errors.Is(err, report.ErrInvalidInput):
			fmt.Printf("[AddSection] Invalid input: %v\n", err)

			h.auditLogger.Log(c, auditlog.AuditLog{
				Action: "ADD_SECTION",
				Actor:  actor,
				Target: auditlog.Target{
					Type: "report",
					ID:   reportIDStr,
				},
				Service:     "report",
				Status:      "FAILED",
				Description: "Invalid input",
			})

			writeError(c, http.StatusBadRequest, "invalid_input", "invalid input")
		default:
			logWithCtx("error", "add section failed", c, map[string]any{
				"reportID": reportUUID.String(),
				"title":    req.Title,
				"order":    req.Order,
				"err":      err.Error(),
			})
			fmt.Printf("[AddSection] Failed to add section: %v\n", err)

			h.auditLogger.Log(c, auditlog.AuditLog{
				Action: "ADD_SECTION",
				Actor:  actor,
				Target: auditlog.Target{
					Type: "report",
					ID:   reportIDStr,
				},
				Service:     "report",
				Status:      "FAILED",
				Description: "Failed to add section: " + err.Error(),
			})

			writeError(c, http.StatusInternalServerError, "add_section_failed", "failed to add section")
		}
		return
	}
	fmt.Printf("[AddSection] Successfully added section '%s' to report %s\n", req.Title, reportIDStr)

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "ADD_SECTION",
		Actor:  actor,
		Target: auditlog.Target{
			Type: "report",
			ID:   reportIDStr,
		},
		Service:     "report",
		Status:      "SUCCESS",
		Description: "Section added successfully",
	})

	c.JSON(http.StatusOK, gin.H{"status": "section added successfully"})
}

// DeleteSection handles deleting a section from a report
func (h *ReportHandler) DeleteSection(c *gin.Context) {
	// Grab user details from context
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	email, _ := c.Get("email") // Optional, if you have this set

	actor := auditlog.Actor{
		ID:        userID.(string),
		Role:      userRole.(string),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Email:     email.(string), // Optional, if you have this header set
	}

	_, exists := c.Get("userID")
	if !exists {
		fmt.Printf("[DeleteSection] Missing userID in context\n")

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "DELETE_SECTION",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report_section",
				ID:   "",
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Unauthorized: missing userID in context",
		})

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing userID in context"})
		return
	}

	reportIDStr := c.Param("reportID")
	reportUUID, err := uuid.Parse(reportIDStr)
	if err != nil {
		fmt.Printf("[DeleteSection] Invalid report ID: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "DELETE_SECTION",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report",
				ID:   reportIDStr,
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Invalid report ID",
		})

		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report ID"})
		return
	}

	sectionIDStr := c.Param("sectionID")
	sectionID, err := primitive.ObjectIDFromHex(sectionIDStr)
	if err != nil {
		fmt.Printf("[DeleteSection] Invalid section ID: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "DELETE_SECTION",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report_section",
				ID:   sectionIDStr,
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Invalid section ID",
		})

		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid section ID"})
		return
	}

	err = h.ReportService.DeleteCustomSection(c.Request.Context(), reportUUID, sectionID)
	if err != nil {
		fmt.Printf("[DeleteSection] Failed to delete section: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "DELETE_SECTION",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report_section",
				ID:   sectionID.Hex(),
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Failed to delete section: " + err.Error(),
		})

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete section"})
		return
	}
	fmt.Printf("[DeleteSection] Successfully deleted section %s from report %s\n", sectionID.Hex(), reportIDStr)

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "DELETE_SECTION",
		Actor:  actor,
		Target: auditlog.Target{
			Type: "report_section",
			ID:   sectionID.Hex(),
		},
		Service:     "report",
		Status:      "SUCCESS",
		Description: "Section deleted successfully",
	})

	c.JSON(http.StatusOK, gin.H{"status": "section deleted successfully"})
}

func (h *ReportHandler) GetReportsByCaseID(c *gin.Context) {
	// Grab user details from context
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	email, _ := c.Get("email") // Optional, if you have this set

	actor := auditlog.Actor{
		ID:        userID.(string),
		Role:      userRole.(string),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Email:     email.(string), // Optional, if you have this header set
	}

	_, exists := c.Get("userID")
	if !exists {
		fmt.Printf("[GetReportsByCaseID] Missing userID in context\n")

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "GET_REPORTS_BY_CASE",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "reports",
				ID:   "",
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Unauthorized: missing userID in context",
		})

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing userID in context"})
		return
	}

	caseIDStr := c.Param("caseID")
	caseUUID, err := uuid.Parse(caseIDStr)
	if err != nil {
		fmt.Printf("[GetReportsByCaseID] Invalid case ID: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "GET_REPORTS_BY_CASE",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "case",
				ID:   caseIDStr,
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Invalid case ID",
		})

		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case ID"})
		return
	}

	reports, err := h.ReportService.GetReportsByCaseID(c.Request.Context(), caseUUID)
	if err != nil {
		fmt.Printf("[GetReportsByCaseID] Failed to fetch reports: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "GET_REPORTS_BY_CASE",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "case",
				ID:   caseIDStr,
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Failed to fetch reports: " + err.Error(),
		})

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch reports"})
		return
	}
	fmt.Printf("[GetReportsByCaseID] Successfully retrieved reports for case %s\n", caseIDStr)

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "GET_REPORTS_BY_CASE",
		Actor:  actor,
		Target: auditlog.Target{
			Type: "case",
			ID:   caseIDStr,
		},
		Service:     "report",
		Status:      "SUCCESS",
		Description: "Reports by case retrieved successfully",
	})

	c.JSON(http.StatusOK, reports)
}

func (h *ReportHandler) UpdateSectionTitle(c *gin.Context) {
	// Grab user details from context
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	email, _ := c.Get("email") // Optional, if you have this set

	actor := auditlog.Actor{
		ID:        userID.(string),
		Role:      userRole.(string),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Email:     email.(string), // Optional, if you have this header set
	}

	_, exists := c.Get("userID")
	if !exists {
		fmt.Printf("[UpdateSectionTitle] Missing userID in context\n")

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "UPDATE_SECTION_TITLE",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report_section",
				ID:   "",
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Unauthorized: missing userID in context",
		})

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing userID in context"})
		return
	}

	// IDs
	reportIDStr := c.Param("reportID")
	reportUUID, err := uuid.Parse(reportIDStr)
	if err != nil {
		fmt.Printf("[UpdateSectionTitle] Invalid report ID: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "UPDATE_SECTION_TITLE",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report",
				ID:   reportIDStr,
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Invalid report ID",
		})

		writeError(c, http.StatusBadRequest, "invalid_report_id", "invalid report ID")
		return
	}
	sectionIDStr := c.Param("sectionID")
	sectionID, err := primitive.ObjectIDFromHex(sectionIDStr)
	if err != nil {
		fmt.Printf("[UpdateSectionTitle] Invalid section ID: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "UPDATE_SECTION_TITLE",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report_section",
				ID:   sectionIDStr,
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Invalid section ID",
		})

		writeError(c, http.StatusBadRequest, "invalid_section_id", "invalid section ID")
		return
	}

	// Body
	var req struct {
		Title string `json:"title"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("[UpdateSectionTitle] Invalid request body: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "UPDATE_SECTION_TITLE",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report_section",
				ID:   sectionID.Hex(),
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Invalid request body",
		})

		writeError(c, http.StatusBadRequest, "invalid_body", "invalid request body")
		return
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		fmt.Printf("[UpdateSectionTitle] Title is required\n")

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "UPDATE_SECTION_TITLE",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report_section",
				ID:   sectionID.Hex(),
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Title is required",
		})

		writeError(c, http.StatusBadRequest, "invalid_title", "title is required")
		return
	}
	if len(title) > 200 { // guardrails; adjust if you want
		fmt.Printf("[UpdateSectionTitle] Title too long\n")

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "UPDATE_SECTION_TITLE",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report_section",
				ID:   sectionID.Hex(),
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Title is too long",
		})

		writeError(c, http.StatusBadRequest, "invalid_title", "title is too long")
		return
	}

	// Service
	if err := h.ReportService.UpdateSectionTitle(c.Request.Context(), reportUUID, sectionID, title); err != nil {
		switch {
		case errors.Is(err, report.ErrReportNotFound), errors.Is(err, report.ErrMongoReportNotFound):
			fmt.Printf("[UpdateSectionTitle] Report not found: %s\n", reportIDStr)

			h.auditLogger.Log(c, auditlog.AuditLog{
				Action: "UPDATE_SECTION_TITLE",
				Actor:  actor,
				Target: auditlog.Target{
					Type: "report",
					ID:   reportIDStr,
				},
				Service:     "report",
				Status:      "FAILED",
				Description: "Report not found",
			})

			writeError(c, http.StatusNotFound, "report_not_found", "report not found")
		case errors.Is(err, report.ErrSectionNotFound):
			fmt.Printf("[UpdateSectionTitle] Section not found: %s\n", sectionID.Hex())

			h.auditLogger.Log(c, auditlog.AuditLog{
				Action: "UPDATE_SECTION_TITLE",
				Actor:  actor,
				Target: auditlog.Target{
					Type: "report_section",
					ID:   sectionID.Hex(),
				},
				Service:     "report",
				Status:      "FAILED",
				Description: "Section not found",
			})

			writeError(c, http.StatusNotFound, "section_not_found", "section not found")
		case errors.Is(err, report.ErrInvalidInput):
			fmt.Printf("[UpdateSectionTitle] Invalid input: %v\n", err)

			h.auditLogger.Log(c, auditlog.AuditLog{
				Action: "UPDATE_SECTION_TITLE",
				Actor:  actor,
				Target: auditlog.Target{
					Type: "report_section",
					ID:   sectionID.Hex(),
				},
				Service:     "report",
				Status:      "FAILED",
				Description: "Invalid input",
			})

			writeError(c, http.StatusBadRequest, "invalid_input", "invalid input")
		default:
			fmt.Printf("[UpdateSectionTitle] Failed to update section title: %v\n", err)

			h.auditLogger.Log(c, auditlog.AuditLog{
				Action: "UPDATE_SECTION_TITLE",
				Actor:  actor,
				Target: auditlog.Target{
					Type: "report_section",
					ID:   sectionID.Hex(),
				},
				Service:     "report",
				Status:      "FAILED",
				Description: "Failed to update section title: " + err.Error(),
			})

			writeError(c, http.StatusInternalServerError, "update_failed", "failed to update section title")
		}
		return
	}
	fmt.Printf("[UpdateSectionTitle] Successfully updated section title for section %s in report %s\n", sectionID.Hex(), reportIDStr)

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "UPDATE_SECTION_TITLE",
		Actor:  actor,
		Target: auditlog.Target{
			Type: "report_section",
			ID:   sectionID.Hex(),
		},
		Service:     "report",
		Status:      "SUCCESS",
		Description: "Section title updated successfully",
	})

	c.JSON(http.StatusOK, gin.H{"status": "section title updated successfully"})
}

func (h *ReportHandler) ReorderSection(c *gin.Context) {
	// Grab user details from context
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	email, _ := c.Get("email") // Optional, if you have this set

	actor := auditlog.Actor{
		ID:        userID.(string),
		Role:      userRole.(string),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Email:     email.(string), // Optional, if you have this header set
	}

	_, exists := c.Get("userID")
	if !exists {
		fmt.Printf("[ReorderSection] Missing userID in context\n")

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "REORDER_SECTION",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report_section",
				ID:   "",
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Unauthorized: missing userID in context",
		})

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing userID in context"})
		return
	}

	// IDs
	reportIDStr := c.Param("reportID")
	reportUUID, err := uuid.Parse(reportIDStr)
	if err != nil {
		fmt.Printf("[ReorderSection] Invalid report ID: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "REORDER_SECTION",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report",
				ID:   reportIDStr,
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Invalid report ID",
		})

		writeError(c, http.StatusBadRequest, "invalid_report_id", "invalid report ID")
		return
	}
	sectionIDStr := c.Param("sectionID")
	sectionID, err := primitive.ObjectIDFromHex(sectionIDStr)
	if err != nil {
		fmt.Printf("[ReorderSection] Invalid section ID: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "REORDER_SECTION",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report_section",
				ID:   sectionIDStr,
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Invalid section ID",
		})

		writeError(c, http.StatusBadRequest, "invalid_section_id", "invalid section ID")
		return
	}

	// Body
	var req struct {
		NewOrder int `json:"order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("[ReorderSection] Invalid request body: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "REORDER_SECTION",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report_section",
				ID:   sectionID.Hex(),
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Invalid request body",
		})

		writeError(c, http.StatusBadRequest, "invalid_body", "invalid request body")
		return
	}
	if req.NewOrder < 1 {
		fmt.Printf("[ReorderSection] Invalid order: %d\n", req.NewOrder)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "REORDER_SECTION",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report_section",
				ID:   sectionID.Hex(),
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Order must be >= 1",
		})

		writeError(c, http.StatusBadRequest, "invalid_order", "order must be >= 1")
		return
	}

	// Service (rename here if your service method is ReorderCustomSection)
	if err := h.ReportService.ReorderSection(c.Request.Context(), reportUUID, sectionID, req.NewOrder); err != nil {
		switch {
		case errors.Is(err, report.ErrReportNotFound), errors.Is(err, report.ErrMongoReportNotFound):
			fmt.Printf("[ReorderSection] Report not found: %s\n", reportIDStr)

			h.auditLogger.Log(c, auditlog.AuditLog{
				Action: "REORDER_SECTION",
				Actor:  actor,
				Target: auditlog.Target{
					Type: "report",
					ID:   reportIDStr,
				},
				Service:     "report",
				Status:      "FAILED",
				Description: "Report not found",
			})

			writeError(c, http.StatusNotFound, "report_not_found", "report not found")
		case errors.Is(err, report.ErrSectionNotFound):
			fmt.Printf("[ReorderSection] Section not found: %s\n", sectionID.Hex())

			h.auditLogger.Log(c, auditlog.AuditLog{
				Action: "REORDER_SECTION",
				Actor:  actor,
				Target: auditlog.Target{
					Type: "report_section",
					ID:   sectionID.Hex(),
				},
				Service:     "report",
				Status:      "FAILED",
				Description: "Section not found",
			})

			writeError(c, http.StatusNotFound, "section_not_found", "section not found")
		case errors.Is(err, report.ErrInvalidInput):
			fmt.Printf("[ReorderSection] Invalid input: %v\n", err)

			h.auditLogger.Log(c, auditlog.AuditLog{
				Action: "REORDER_SECTION",
				Actor:  actor,
				Target: auditlog.Target{
					Type: "report_section",
					ID:   sectionID.Hex(),
				},
				Service:     "report",
				Status:      "FAILED",
				Description: "Invalid input",
			})

			writeError(c, http.StatusBadRequest, "invalid_input", "invalid input")
		default:
			// Optional: if you implement optimistic concurrency and return a "conflict" error,
			// map it to 409 here.
			low := strings.ToLower(err.Error())
			if strings.Contains(low, "conflict") {
				fmt.Printf("[ReorderSection] Conflict: %v\n", err)

				h.auditLogger.Log(c, auditlog.AuditLog{
					Action: "REORDER_SECTION",
					Actor:  actor,
					Target: auditlog.Target{
						Type: "report_section",
						ID:   sectionID.Hex(),
					},
					Service:     "report",
					Status:      "FAILED",
					Description: "Conflict: the section was modified by someone else",
				})

				writeError(c, http.StatusConflict, "conflict", "the section was modified by someone else")
				return
			}
			fmt.Printf("[ReorderSection] Failed to reorder section: %v\n", err)

			h.auditLogger.Log(c, auditlog.AuditLog{
				Action: "REORDER_SECTION",
				Actor:  actor,
				Target: auditlog.Target{
					Type: "report_section",
					ID:   sectionID.Hex(),
				},
				Service:     "report",
				Status:      "FAILED",
				Description: "Failed to reorder section: " + err.Error(),
			})

			writeError(c, http.StatusInternalServerError, "reorder_failed", "failed to reorder section")
		}
		return
	}
	fmt.Printf("[ReorderSection] Successfully reordered section %s to order %d in report %s\n", sectionID.Hex(), req.NewOrder, reportIDStr)

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "REORDER_SECTION",
		Actor:  actor,
		Target: auditlog.Target{
			Type: "report_section",
			ID:   sectionID.Hex(),
		},
		Service:     "report",
		Status:      "SUCCESS",
		Description: "Section reordered successfully",
	})

	c.JSON(http.StatusOK, gin.H{"status": "section reordered successfully"})
}

// GET /api/v1/reports/recent?limit=6&mine=true&caseId=<uuid>&status=<string>
func (h *ReportHandler) GetRecentReports(c *gin.Context) {
	// Grab user details from context
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	email, _ := c.Get("email") // Optional, if you have this set

	actor := auditlog.Actor{
		ID:        userID.(string),
		Role:      userRole.(string),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Email:     email.(string), // Optional, if you have this header set
	}

	userIDVal, exists := c.Get("userID")
	if !exists {
		fmt.Printf("[GetRecentReports] Missing userID in context\n")

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "GET_RECENT_REPORTS",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "reports",
				ID:   "",
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Unauthorized: missing userID in context",
		})

		writeError(c, http.StatusUnauthorized, "unauthorized", "user not authorized")
		return
	}

	// auth: same style as GenerateReport
	userIDVal, ok := c.Get("userID")
	if !ok {
		fmt.Printf("[GetRecentReports] User not authorized\n")

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "GET_RECENT_REPORTS",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "reports",
				ID:   "",
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "User not authorized",
		})

		writeError(c, http.StatusUnauthorized, "unauthorized", "user not authorized")
		return
	}
	examinerID, err := uuid.Parse(userIDVal.(string))
	if err != nil {
		fmt.Printf("[GetRecentReports] Invalid user ID format: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "GET_RECENT_REPORTS",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "reports",
				ID:   "",
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Invalid user ID format",
		})

		writeError(c, http.StatusInternalServerError, "invalid_user_id", "invalid user id format")
		return
	}

	// query params
	limitStr := c.DefaultQuery("limit", "6")
	mineStr := c.DefaultQuery("mine", "true")
	caseIDStr := c.Query("caseId")
	statusStr := strings.TrimSpace(c.Query("status"))

	limit := 6
	if v, err := strconv.Atoi(limitStr); err == nil && v > 0 {
		limit = v
	}
	mine := strings.EqualFold(mineStr, "true")

	var caseID *uuid.UUID
	if caseIDStr != "" {
		if cid, err := uuid.Parse(caseIDStr); err == nil {
			caseID = &cid
		} else {
			fmt.Printf("[GetRecentReports] Invalid case ID: %v\n", err)

			h.auditLogger.Log(c, auditlog.AuditLog{
				Action: "GET_RECENT_REPORTS",
				Actor:  actor,
				Target: auditlog.Target{
					Type: "case",
					ID:   caseIDStr,
				},
				Service:     "report",
				Status:      "FAILED",
				Description: "Invalid case ID",
			})

			writeError(c, http.StatusBadRequest, "invalid_case_id", "invalid caseId")
			return
		}
	}

	var status *string
	if statusStr != "" {
		status = &statusStr
	}

	// NEW: tenant & team from auth middleware (strings)
	tidStr := c.GetString("tenantID")
	if tidStr == "" {
		fmt.Printf("[GetRecentReports] Tenant missing\n")

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "GET_RECENT_REPORTS",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "reports",
				ID:   "",
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Tenant missing",
		})

		writeError(c, http.StatusUnauthorized, "tenant_missing", "tenant not found")
		return
	}
	tenantID, err := uuid.Parse(tidStr)
	if err != nil {
		fmt.Printf("[GetRecentReports] Invalid tenant ID: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "GET_RECENT_REPORTS",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "tenant",
				ID:   tidStr,
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Invalid tenant ID",
		})

		writeError(c, http.StatusBadRequest, "invalid_tenant_id", "invalid tenant id")
		return
	}

	var teamIDPtr *uuid.UUID
	if teamStr := c.GetString("teamID"); teamStr != "" {
		if t, err := uuid.Parse(teamStr); err == nil {
			teamIDPtr = &t
		} else {
			fmt.Printf("[GetRecentReports] Invalid team ID: %v\n", err)

			h.auditLogger.Log(c, auditlog.AuditLog{
				Action: "GET_RECENT_REPORTS",
				Actor:  actor,
				Target: auditlog.Target{
					Type: "team",
					ID:   teamStr,
				},
				Service:     "report",
				Status:      "FAILED",
				Description: "Invalid team ID",
			})

			writeError(c, http.StatusBadRequest, "invalid_team_id", "invalid team id")
			return
		}
	}

	opts := report.RecentReportsOptions{
		Limit:      limit,
		MineOnly:   mine,
		ExaminerID: examinerID,
		CaseID:     caseID,
		Status:     status,
		TenantID:   tenantID, // ← NEW
		TeamID:     teamIDPtr,
	}

	items, err := h.ReportService.ListRecentReports(c.Request.Context(), opts)
	if err != nil {
		logWithCtx("error", "list recent reports failed", c, map[string]any{"err": err.Error()})
		fmt.Printf("[GetRecentReports] Failed to load recent reports: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "GET_RECENT_REPORTS",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "reports",
				ID:   "",
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Failed to load recent reports: " + err.Error(),
		})

		writeError(c, http.StatusInternalServerError, "recent_reports_failed", "failed to load recent reports")
		return
	}
	fmt.Printf("[GetRecentReports] Successfully retrieved %d recent reports\n", len(items))

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "GET_RECENT_REPORTS",
		Actor:  actor,
		Target: auditlog.Target{
			Type: "reports",
			ID:   "",
		},
		Service:     "report",
		Status:      "SUCCESS",
		Description: "Recent reports retrieved successfully",
	})

	// shape the response to what your React expects
	type row struct {
		ID           uuid.UUID `json:"id"`
		Title        string    `json:"title"`
		Status       string    `json:"status"`
		LastModified string    `json:"lastModified"` // ISO time
	}

	resp := make([]row, len(items))
	for i, it := range items {
		resp[i] = row{
			ID:           it.ID,
			Title:        it.Title,
			Status:       it.Status,
			LastModified: it.LastModified.Format(time.RFC3339),
		}
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ReportHandler) UpdateReportName(c *gin.Context) {
	// Grab user details from context
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	email, _ := c.Get("email") // Optional, if you have this set

	actor := auditlog.Actor{
		ID:        userID.(string),
		Role:      userRole.(string),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Email:     email.(string), // Optional, if you have this header set
	}

	_, exists := c.Get("userID")
	if !exists {
		fmt.Printf("[UpdateReportName] Missing userID in context\n")

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "UPDATE_REPORT_NAME",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report",
				ID:   "",
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Unauthorized: missing userID in context",
		})

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing userID in context"})
		return
	}

	rid, err := uuid.Parse(c.Param("reportID"))
	if err != nil {
		fmt.Printf("[UpdateReportName] Invalid report ID: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "UPDATE_REPORT_NAME",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report",
				ID:   c.Param("reportID"),
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Invalid report ID",
		})

		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid reportID"})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("[UpdateReportName] Invalid payload: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "UPDATE_REPORT_NAME",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "report",
				ID:   rid.String(),
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Invalid payload",
		})

		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	updated, err := h.ReportService.UpdateReportName(c.Request.Context(), rid, req.Name)
	if err != nil {
		switch {
		case errors.Is(err, report.ErrInvalidReportName):
			fmt.Printf("[UpdateReportName] Invalid report name: %v\n", err)

			h.auditLogger.Log(c, auditlog.AuditLog{
				Action: "UPDATE_REPORT_NAME",
				Actor:  actor,
				Target: auditlog.Target{
					Type: "report",
					ID:   rid.String(),
				},
				Service:     "report",
				Status:      "FAILED",
				Description: "Invalid report name: " + err.Error(),
			})

			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, report.ErrReportNotFound):
			fmt.Printf("[UpdateReportName] Report not found: %s\n", rid.String())

			h.auditLogger.Log(c, auditlog.AuditLog{
				Action: "UPDATE_REPORT_NAME",
				Actor:  actor,
				Target: auditlog.Target{
					Type: "report",
					ID:   rid.String(),
				},
				Service:     "report",
				Status:      "FAILED",
				Description: "Report not found",
			})

			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			fmt.Printf("[UpdateReportName] Failed to update name: %v\n", err)

			h.auditLogger.Log(c, auditlog.AuditLog{
				Action: "UPDATE_REPORT_NAME",
				Actor:  actor,
				Target: auditlog.Target{
					Type: "report",
					ID:   rid.String(),
				},
				Service:     "report",
				Status:      "FAILED",
				Description: "Failed to update report name: " + err.Error(),
			})

			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update name"})
		}
		return
	}
	fmt.Printf("[UpdateReportName] Successfully updated report name for report %s\n", rid.String())

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "UPDATE_REPORT_NAME",
		Actor:  actor,
		Target: auditlog.Target{
			Type: "report",
			ID:   rid.String(),
		},
		Service:     "report",
		Status:      "SUCCESS",
		Description: "Report name updated successfully",
	})

	c.JSON(http.StatusOK, updated) // full Report JSON
}

func (h *ReportHandler) GetReportsForTeam(c *gin.Context) {
	// Grab user details from context
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	email, _ := c.Get("email") // Optional, if you have this set

	actor := auditlog.Actor{
		ID:        userID.(string),
		Role:      userRole.(string),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Email:     email.(string), // Optional, if you have this header set
	}

	_, exists := c.Get("userID")
	if !exists {
		fmt.Printf("[GetReportsForTeam] Missing userID in context\n")

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "GET_REPORTS_FOR_TEAM",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "reports",
				ID:   "",
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Unauthorized: missing userID in context",
		})

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing userID in context"})
		return
	}

	// tenant from token (AuthMiddleware sets this)
	tenantIDVal, _ := c.Get("tenantID")
	tenantIDStr := fmt.Sprint(tenantIDVal)
	tenantUUID, err := uuid.Parse(tenantIDStr)
	if err != nil || tenantUUID == uuid.Nil {
		fmt.Printf("[GetReportsForTeam] Invalid tenant in token: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "GET_REPORTS_FOR_TEAM",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "tenant",
				ID:   tenantIDStr,
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Invalid tenant in token",
		})

		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid tenant in token"})
		return
	}

	// team from URL
	teamIDParam := c.Param("teamID")
	teamUUID, err := uuid.Parse(teamIDParam)
	if err != nil || teamUUID == uuid.Nil {
		fmt.Printf("[GetReportsForTeam] Invalid team ID: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "GET_REPORTS_FOR_TEAM",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "team",
				ID:   teamIDParam,
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Invalid team ID",
		})

		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team id"})
		return
	}

	// (Optional) enforce team from token equals URL team for DFIR roles:
	if roleVal, ok := c.Get("userRole"); ok && (roleVal == "DFIR Admin" || roleVal == "DFIR User") {
		if claimTeamVal, ok := c.Get("teamID"); ok {
			if claimTeamStr := fmt.Sprint(claimTeamVal); claimTeamStr != "" {
				if claimTeamUUID, e := uuid.Parse(claimTeamStr); e == nil && claimTeamUUID != teamUUID {
					fmt.Printf("[GetReportsForTeam] Team mismatch\n")

					h.auditLogger.Log(c, auditlog.AuditLog{
						Action: "GET_REPORTS_FOR_TEAM",
						Actor:  actor,
						Target: auditlog.Target{
							Type: "team",
							ID:   teamIDParam,
						},
						Service:     "report",
						Status:      "FAILED",
						Description: "Team mismatch",
					})

					c.JSON(http.StatusForbidden, gin.H{"error": "team mismatch"})
					return
				}
			}
		}
	}

	reports, err := h.ReportService.GetReportsByTeamID(c.Request.Context(), tenantUUID, teamUUID)
	if err != nil {
		fmt.Printf("[GetReportsForTeam] Failed to fetch team reports: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "GET_REPORTS_FOR_TEAM",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "team",
				ID:   teamIDParam,
			},
			Service:     "report",
			Status:      "FAILED",
			Description: "Failed to fetch team reports: " + err.Error(),
		})

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch team reports"})
		return
	}
	fmt.Printf("[GetReportsForTeam] Successfully retrieved reports for team %s\n", teamIDParam)

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "GET_REPORTS_FOR_TEAM",
		Actor:  actor,
		Target: auditlog.Target{
			Type: "team",
			ID:   teamIDParam,
		},
		Service:     "report",
		Status:      "SUCCESS",
		Description: "Team reports retrieved successfully",
	})

	c.JSON(http.StatusOK, gin.H{"reports": reports})
}
