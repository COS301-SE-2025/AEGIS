package handlers

import (
	"aegis-api/cache"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	evidencecount "aegis-api/services_/evidence/evidence_count"

	"github.com/gin-gonic/gin"
)

type EvidenceHandler struct {
	service evidencecount.EvidenceService
	cache   cache.Client
}

func NewEvidenceHandler(service evidencecount.EvidenceService, cacheClient cache.Client) *EvidenceHandler {
	return &EvidenceHandler{service: service, cache: cacheClient}
}

func (h *EvidenceHandler) GetEvidenceCount(c *gin.Context) {
	tenantIDFromToken, exists := c.Get("tenantID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing tenantID in token context"})
		return
	}
	tenantIDStr := tenantIDFromToken.(string)
	tenantID := c.Param("tenantId")

	fmt.Printf("[DEBUG] Received tenantID: %s (token=%s)\n", tenantID, tenantIDStr)
	if tenantID != tenantIDStr {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant ID mismatch"})
		return
	}

	// ---- Cache key (stable & scoped) ----
	key := "evidence:count:tenant:" + tenantID
	ctx := c.Request.Context()

	// ---- Try cache ----
	if raw, ok, _ := h.cache.Get(ctx, key); ok {
		etag := cache.EntityETag([]byte(raw))
		// Conditional GET support
		if ifNoneMatch := c.GetHeader("If-None-Match"); ifNoneMatch != "" && ifNoneMatch == etag {
			c.Header("X-Cache", "REVALIDATED")
			c.Header("ETag", etag)
			c.Header("Cache-Control", "public, max-age=60")
			c.Status(http.StatusNotModified)
			return
		}
		c.Header("X-Cache", "HIT")
		c.Header("ETag", etag)
		c.Header("Cache-Control", "public, max-age=60")
		c.Data(http.StatusOK, "application/json", []byte(raw))
		return
	}

	// ---- Miss → compute ----
	count, err := h.service.GetEvidenceCount(tenantID)
	fmt.Printf("[DEBUG] Evidence count for tenantID %s: %d (err=%v)\n", tenantID, count, err)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get evidence count"})
		return
	}

	body, _ := json.Marshal(gin.H{"count": count})

	// ---- Write to cache (TTL 5min) ----
	if err := h.cache.Set(ctx, key, string(body), 5*time.Minute); err != nil {
		// Non-fatal: log and continue
		fmt.Printf("[WARN] cache set failed for key %s: %v\n", key, err)
	}

	etag := cache.EntityETag(body)
	// Conditional GET support (first response after MISS)
	if ifNoneMatch := c.GetHeader("If-None-Match"); ifNoneMatch != "" && ifNoneMatch == etag {
		c.Header("X-Cache", "REVALIDATED")
		c.Header("ETag", etag)
		c.Header("Cache-Control", "public, max-age=60")
		c.Status(http.StatusNotModified)
		return
	}

	c.Header("X-Cache", "MISS")
	c.Header("ETag", etag)
	c.Header("Cache-Control", "public, max-age=60")
	c.Data(http.StatusOK, "application/json", body)
}
