package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"aegis-api/cache"
	"aegis-api/services_/auditlog"

	"github.com/gin-gonic/gin"
)

func (h *CaseHandler) ListClosedCasesHandler(c *gin.Context) {
	// ── Auth / context ────────────────────────────────────────────────────────────
	userIDv, uok := c.Get("userID")
	tenantIDv, tok := c.Get("tenantID")
	teamIDv, mok := c.Get("teamID")
	rolev, _ := c.Get("userRole")

	if !(uok && tok && mok) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, tenantID, teamID := userIDv.(string), tenantIDv.(string), teamIDv.(string)
	if userID == "" || tenantID == "" || teamID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid token data"})
		return
	}

	// ── Query params (kept to stabilize cache keys if UI adds controls) ───────────
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "20")
	sort := c.DefaultQuery("sort", "closed_at")
	order := c.DefaultQuery("order", "desc")

	// ── Cache key (tenant + scope + deterministic qSig) ───────────────────────────
	qSig := cache.BuildQuerySig(page, pageSize, sort, order, map[string]any{
		"scope":  "closed",
		"userId": userID, // keep if list is user/team scoped; remove if tenant-wide
		"teamId": teamID,
	})
	key := cache.ListKey(tenantID, cache.ScopeClosed, qSig)

	// ── Actor for audit ───────────────────────────────────────────────────────────
	roleStr, _ := rolev.(string)
	actor := auditlog.Actor{
		ID:        userID,
		Role:      roleStr,
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}

	ctx := c.Request.Context()
	cacheResult := "MISS" // will flip to HIT/REVALIDATED if applicable

	// ── 1) Try cache ──────────────────────────────────────────────────────────────
	if raw, ok, _ := h.Cache.Get(ctx, key); ok {
		etag := cache.ListETag([]byte(raw))
		if inm := c.GetHeader("If-None-Match"); inm != "" && inm == etag {
			// Revalidated
			cacheResult = "REVALIDATED"
			c.Header("X-Cache", cacheResult)
			c.Header("ETag", etag)
			c.Header("Cache-Control", "private, max-age=120")
			c.Status(http.StatusNotModified)
			// Audit success (optional to include revalidation outcome)
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
		c.Data(http.StatusOK, "application/json", []byte(raw))
		// (Optional) Audit on HIT as well; inexpensive and helpful for activity feeds
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list closed cases"})
		return
	}

	// Map progress for each closed case
	for i := range cases {
		cases[i].Progress = getProgressForStage(cases[i].InvestigationStage)
	}

	// ── 3) Build payload & write to cache ─────────────────────────────────────────
	body, _ := json.Marshal(gin.H{
		"closed_cases": cases,
		"meta": gin.H{
			"page": page, "pageSize": pageSize, "sort": sort, "order": order,
		},
	})

	if err := h.Cache.Set(ctx, key, string(body), 120*time.Second); err != nil {
		// Non-fatal: log for visibility
		fmt.Printf("[WARN] cache.set failed for key %s: %v\n", key, err)
	}

	// ── 4) ETag + headers + body ──────────────────────────────────────────────────
	etag := cache.ListETag(body)
	if inm := c.GetHeader("If-None-Match"); inm != "" && inm == etag {
		cacheResult = "REVALIDATED"
		c.Header("X-Cache", cacheResult)
		c.Header("ETag", etag)
		c.Header("Cache-Control", "private, max-age=120")
		c.Status(http.StatusNotModified)
		// Audit success
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
			Description: fmt.Sprintf("You viewed %d closed case(s) for your team. (cache=REVALIDATED)", len(cases)),
		})
		return
	}

	cacheResult = "MISS"
	c.Header("X-Cache", cacheResult)
	c.Header("ETag", etag)
	c.Header("Cache-Control", "private, max-age=120")
	c.Data(http.StatusOK, "application/json", body)

	// ── 5) Audit success (friendly wording for activity feed) ─────────────────────
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

	c.JSON(http.StatusOK, gin.H{"closed_cases": cases})

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
