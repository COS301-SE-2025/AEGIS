package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"aegis-api/cache"
	"aegis-api/services_/auditlog"

	"github.com/gin-gonic/gin"
)

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

	// ── Query params ──────────────────────────────────────────────────────────────
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "20")
	sort := c.DefaultQuery("sort", "closed_at")
	order := c.DefaultQuery("order", "desc")

	// ── Cache key ─────────────────────────────────────────────────────────────────
	qSig := cache.BuildQuerySig(page, pageSize, sort, order, map[string]any{
		"scope":  "closed",
		"userId": userID,
		"teamId": teamID,
	})
	key := cache.ListKey(tenantID, cache.ScopeClosed, qSig)

	roleStr, _ := rolev.(string)
	actor := auditlog.Actor{
		ID:        userID,
		Role:      roleStr,
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}

	ctx := c.Request.Context()
	cacheResult := "MISS"

	// ── 1) Try cache ──────────────────────────────────────────────────────────────
	if raw, ok, _ := h.Cache.Get(ctx, key); ok {
		etag := cache.ListETag([]byte(raw))
		if inm := c.GetHeader("If-None-Match"); inm != "" && inm == etag {
			cacheResult = "REVALIDATED"
			c.Header("X-Cache", cacheResult)
			c.Header("ETag", etag)
			c.Header("Cache-Control", "private, max-age=120")
			log.Println(">>> Responding with 304 Not Modified (cache REVALIDATED)")
			c.Status(http.StatusNotModified)
			h.auditLogger.Log(c, auditlog.AuditLog{
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
				Status:      "SUCCESS",
				Description: "You viewed closed cases for your team. (cache=REVALIDATED)",
			})
			return
		}

		// Cache HIT
		cacheResult = "HIT"
		c.Header("X-Cache", cacheResult)
		c.Header("ETag", etag)
		c.Header("Cache-Control", "private, max-age=120")

		log.Println(">>> Responding with 200 OK (cache HIT)")
		c.Data(http.StatusOK, "application/json", []byte(raw))

		h.auditLogger.Log(c, auditlog.AuditLog{
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
			Status:      "SUCCESS",
			Description: "You viewed closed cases for your team. (cache=HIT)",
		})
		return
	}

	// ── 2) Cache MISS → fetch service ─────────────────────────────────────────────
	cases, err := h.CaseService.ListClosedCases(userID, tenantID, teamID)
	if err != nil {
		log.Println(">>> ERROR: CaseService.ListClosedCases failed:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list closed cases"})
		h.auditLogger.Log(c, auditlog.AuditLog{
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

	// ── 3) Build payload ─────────────────────────────────────────────────────────
	payload := gin.H{
		"closed_cases": cases,
		"meta": gin.H{
			"page": page, "pageSize": pageSize, "sort": sort, "order": order,
		},
	}

	// Cache store (non-fatal if fails)
	if b, err := json.Marshal(payload); err == nil {
		if err := h.Cache.Set(ctx, key, string(b), 120*time.Second); err != nil {
			log.Printf("[WARN] cache.set failed for key %s: %v\n", key, err)
		}
	}

	b, _ := json.Marshal(payload)
	etag := cache.ListETag(b)
	c.Header("X-Cache", cacheResult)
	c.Header("ETag", etag)
	c.Header("Cache-Control", "private, max-age=120")

	// ── 4) Final response ────────────────────────────────────────────────────────
	log.Println(">>> Responding with 200 OK (cache MISS)")
	c.JSON(http.StatusOK, payload)

	h.auditLogger.Log(c, auditlog.AuditLog{
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
			"You viewed %d closed case(s) for your team. (cache=%s)", len(cases), cacheResult,
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
