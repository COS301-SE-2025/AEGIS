package handlers

import (
	"fmt"
	"log"
	"net/http"

	"aegis-api/services_/auditlog"

	"github.com/gin-gonic/gin"
)

// Add this method to safely log audit events
func (h *CaseHandler) safeAuditLog(c *gin.Context, log auditlog.AuditLog) {
	if h.auditLogger != nil {
		h.auditLogger.Log(c, log)
	}
}

func (h *CaseHandler) ListClosedCasesHandler(c *gin.Context) {
	log.Println(">>> ListClosedCasesHandler called")

	// ── Auth / context ────────────────────────────────────────────────────────────
	userIDv, uok := c.Get("userID")
	tenantIDv, tok := c.Get("tenantID")
	teamIDv, mok := c.Get("teamID")
	rolev, _ := c.Get("userRole")

	if !(uok && tok && mok) {
		log.Println(">>> AUTH FAIL: missing userID / tenantID / teamID")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, tenantID, teamID := userIDv.(string), tenantIDv.(string), teamIDv.(string)
	if userID == "" || tenantID == "" || teamID == "" {
		log.Println(">>> AUTH FAIL: empty userID / tenantID / teamID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid token data"})
		return
	}

	roleStr, _ := rolev.(string)
	actor := auditlog.Actor{
		ID:        userID,
		Role:      roleStr,
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}

	// ── Fetch closed cases from service ──────────────────────────────────────────
	cases, err := h.CaseService.ListClosedCases(userID, tenantID, teamID)
	if err != nil {
		log.Println(">>> ERROR: CaseService.ListClosedCases failed:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list closed cases"})
		h.safeAuditLog(c, auditlog.AuditLog{
			Action: "LIST_CLOSED_CASES",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "closed_case_listing",
				ID:   userID,
				AdditionalInfo: map[string]string{
					"tenant_id": tenantID, "team_id": teamID,
				},
			},
			Service:     "case",
			Status:      "FAILED",
			Description: "Failed to list closed cases: " + err.Error(),
		})
		return
	}

	// Map progress for each closed case
	for i := range cases {
		cases[i].Progress = getProgressForStage(cases[i].InvestigationStage)
	}

	// ── Build response payload ───────────────────────────────────────────────────
	payload := gin.H{
		"closed_cases": cases,
	}

	// ── Final response ───────────────────────────────────────────────────────────
	log.Println(">>> Responding with 200 OK")
	c.JSON(http.StatusOK, payload)

	h.safeAuditLog(c, auditlog.AuditLog{
		Action: "LIST_CLOSED_CASES",
		Actor:  actor,
		Target: auditlog.Target{
			Type: "closed_case_listing",
			ID:   userID,
			AdditionalInfo: map[string]string{
				"tenant_id": tenantID, "team_id": teamID,
			},
		},
		Service: "case",
		Status:  "SUCCESS",
		Description: fmt.Sprintf(
			"You viewed %d closed case(s) for your team.", len(cases),
		),
	})
}

// getProgressForStage returns a progress value (0-100) based on investigation stage
func getProgressForStage(stage string) int {
	switch stage {
	case "Triage":
		return 10
	case "Evidence Collection":
		return 25
	case "Analysis":
		return 40
	case "Correlation & Threat Intelligence":
		return 55
	case "Containment & Eradication":
		return 70
	case "Recovery":
		return 85
	case "Reporting & Documentation":
		return 95
	case "Case Closure & Review":
		return 100
	default:
		return 0
	}
}
