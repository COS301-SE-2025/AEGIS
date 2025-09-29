package handlers

import (
	"aegis-api/cache"
	"aegis-api/middleware"
	"aegis-api/services_/case/case_evidence_totals"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type CaseEvidenceTotalsHandler struct {
	DashboardService case_evidence_totals.DashboardService
	Cache            cache.Client
}

func NewCaseEvidenceTotalsHandler(service case_evidence_totals.DashboardService, cacheClient cache.Client) *CaseEvidenceTotalsHandler {
	return &CaseEvidenceTotalsHandler{
		DashboardService: service,
		Cache:            cacheClient,
	}
}

func (h *CaseEvidenceTotalsHandler) GetDashboardTotals(c *gin.Context) {
	// Parse inputs
	statusQuery := c.DefaultQuery("statuses", "open,ongoing,closed")
	rawStatuses := strings.Split(statusQuery, ",")
	// normalize statuses for a stable signature
	statuses := make([]string, 0, len(rawStatuses))
	for _, s := range rawStatuses {
		s = strings.ToLower(strings.TrimSpace(s))
		if s != "" {
			statuses = append(statuses, s)
		}
	}
	sort.Strings(statuses) // stable order

	userID := c.GetString("userID")
	tenantID := c.GetString("tenantID")
	if userID == "" || tenantID == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Build key: dashboard:<tenantId>:totals:user:<userId>:q=<sha(...)>
	qSig := cache.BuildQuerySig("1", "1", "n/a", "n/a", map[string]any{
		"scope":    "dashboard_totals",
		"statuses": statuses,
		"userId":   userID,
	})
	key := "dashboard:" + tenantID + ":totals:user:" + userID + ":q=" + cacheHash(qSig)

	ctx := c.Request.Context()

	// Try cache
	if raw, ok, _ := h.Cache.Get(ctx, key); ok {
		etag := cache.EntityETag([]byte(raw))
		if middleware.IfNoneMatch(c.Writer, c.Request, etag) {
			c.Header("X-Cache", "REVALIDATED")
			return
		}
		middleware.SetCacheControl(c.Writer, 60)
		c.Header("X-Cache", "HIT")
		c.Data(http.StatusOK, "application/json", []byte(raw))
		return
	}

	// Miss → service
	caseCount, evidenceCount, err := h.DashboardService.GetCounts(userID, statuses)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve dashboard totals"})
		return
	}

	body, _ := json.Marshal(gin.H{
		"case_count":     caseCount,
		"evidence_count": evidenceCount,
		"meta": gin.H{
			"statuses": statuses,
			"userId":   userID,
		},
	})

	// Set cache (TTL 60s)
	_ = h.Cache.Set(ctx, key, string(body), 60*time.Second)

	etag := cache.EntityETag(body)
	if middleware.IfNoneMatch(c.Writer, c.Request, etag) {
		c.Header("X-Cache", "REVALIDATED")
		return
	}
	middleware.SetCacheControl(c.Writer, 60)
	c.Header("X-Cache", "MISS")
	c.Data(http.StatusOK, "application/json", body)
}

// cacheHash reuses your sha helper; add this tiny wrapper or use cache.ListKey-style
func cacheHash(sig string) string {
	// same algorithm as shaQuery() you use in cache/keys.go
	// either export shaQuery or replicate here:
	// (replicated to keep this self-contained)
	h := sha256.Sum256([]byte(sig))
	return hex.EncodeToString(h[:])
}
