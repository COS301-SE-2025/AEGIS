package handlers

import (
	"aegis-api/cache"
	"aegis-api/middleware"
	"aegis-api/services_/evidence/evidence_viewer"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type EvidenceViewerHandler struct {
	Service *evidence_viewer.EvidenceService
	Cache   cache.Client // <-- use your cache.Client
}

func NewEvidenceViewerHandler(svc *evidence_viewer.EvidenceService, c cache.Client) *EvidenceViewerHandler {
	return &EvidenceViewerHandler{Service: svc, Cache: c}

}

// ----- helpers -----

func tenantIDFromCtx(c *gin.Context) string {
	if v, ok := c.Get("tenantID"); ok {
		if s, ok2 := v.(string); ok2 && s != "" {
			return s
		}
	}
	return "public"
}

// qsig builder for GET list without explicit filters (extensible later)
func qsigFromQueryMinimal(c *gin.Context) string {
	// If you later add pagination/sort, pull them here and into BuildQuerySig.
	return cache.BuildQuerySig(
		c.DefaultQuery("page", "1"),
		c.DefaultQuery("pageSize", "20"),
		c.DefaultQuery("sort", "created_at"),
		c.DefaultQuery("order", "desc"),
		map[string]any{}, // filters empty for now
	)
}

// ----- 1) LIST: GET /evidence/case/:case_id -----
// Key: ev:list:<tenantId>:<caseId>:q=<sha> ; TTL 60–120s ; ETag+304 ; Cache-Control: private, max-age=120
func (h *EvidenceViewerHandler) GetEvidenceByCaseID(c *gin.Context) {
	caseID := c.Param("case_id")
	if caseID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing case ID"})
		return
	}
	tenantID := tenantIDFromCtx(c)

	qsig := qsigFromQueryMinimal(c)
	key := cache.EvidenceListKey(tenantID, caseID, qsig)

	ctx := c.Request.Context()

	// HIT
	if raw, ok, _ := h.Cache.Get(ctx, key); ok && raw != "" {
		etag := cache.ListETag([]byte(raw))
		if middleware.IfNoneMatch(c.Writer, c.Request, etag) {
			c.Header("X-Cache", "REVALIDATED")
			return
		}
		middleware.SetCacheControl(c.Writer, 120)
		c.Header("ETag", etag)
		c.Header("X-Cache", "HIT")
		c.Data(http.StatusOK, "application/json", []byte(raw))
		return
	}

	// MISS
	files, err := h.Service.GetEvidenceFilesByCaseID(caseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get evidence files by case"})
		return
	}
	if len(files) == 0 {
		_ = h.Cache.Set(ctx, key, `{"files":[]}`, 30*time.Second) // short TTL for empty
		c.JSON(http.StatusNotFound, gin.H{"error": "No evidence files found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"files": files})

	body, _ := json.Marshal(gin.H{"files": files})
	_ = h.Cache.Set(ctx, key, string(body), 120*time.Second)

	etag := cache.ListETag(body)
	if middleware.IfNoneMatch(c.Writer, c.Request, etag) {
		c.Header("X-Cache", "REVALIDATED")
		return
	}
	middleware.SetCacheControl(c.Writer, 120)
	c.Header("ETag", etag)
	c.Header("X-Cache", "MISS")
	c.Data(http.StatusOK, "application/json", body)

}

// ----- 2) ITEM (binary): GET /evidence/:evidence_id -----
// Key: ev:item:<tenantId>:<evidenceId> ; TTL 5–15m ; ETag+304 ; Cache-Control: private, max-age=300
// NOTE: Caching binaries is optional; we store as JSON {"data": "<base64>"} to fit your string cache API.
func (h *EvidenceViewerHandler) GetEvidenceByID(c *gin.Context) {
	evidenceID := c.Param("evidence_id")
	if evidenceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing evidence ID"})
		return
	}
	tenantID := tenantIDFromCtx(c)
	key := cache.EvidenceItemKey(tenantID, evidenceID)
	ctx := c.Request.Context()

	// HIT
	if raw, ok, _ := h.Cache.Get(ctx, key); ok && raw != "" {
		etag := cache.ListETag([]byte(raw))
		if middleware.IfNoneMatch(c.Writer, c.Request, etag) {
			c.Header("X-Cache", "REVALIDATED")
			return
		}
		var wire struct {
			Data []byte `json:"data"`
		}
		_ = json.Unmarshal([]byte(raw), &wire)
		middleware.SetCacheControl(c.Writer, 300)
		c.Header("ETag", etag)
		c.Header("X-Cache", "HIT")
		c.Data(http.StatusOK, "application/octet-stream", wire.Data)
		return
	}

	// MISS
	file, err := h.Service.GetEvidenceFileByID(evidenceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get evidence file by ID"})
		return
	}
	if file == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Evidence file not found"})
		return
	}

	wire := struct {
		Data []byte `json:"data"`
	}{Data: file.Data}
	b, _ := json.Marshal(wire)
	_ = h.Cache.Set(ctx, key, string(b), 15*time.Minute)

	etag := cache.ListETag(b)
	if middleware.IfNoneMatch(c.Writer, c.Request, etag) {
		c.Header("X-Cache", "REVALIDATED")
		return
	}
	middleware.SetCacheControl(c.Writer, 300)
	c.Header("ETag", etag)
	c.Header("X-Cache", "MISS")
	c.Data(http.StatusOK, "application/octet-stream", file.Data)
}

// ----- 3) SEARCH: GET /evidence/search?query= -----
// Reuse list-style key with caseId = "-" ; TTL 60–120s ; ETag+304 ; Cache-Control: 120
func (h *EvidenceViewerHandler) SearchEvidence(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing search query"})
		return
	}
	tenantID := tenantIDFromCtx(c)

	qsig := cache.BuildQuerySig("1", "50", "relevance", "desc", map[string]any{"q": query})
	key := cache.EvidenceListKey(tenantID, "-", qsig)

	ctx := c.Request.Context()

	if raw, ok, _ := h.Cache.Get(ctx, key); ok && raw != "" {
		etag := cache.ListETag([]byte(raw))
		if middleware.IfNoneMatch(c.Writer, c.Request, etag) {
			c.Header("X-Cache", "REVALIDATED")
			return
		}
		middleware.SetCacheControl(c.Writer, 120)
		c.Header("ETag", etag)
		c.Header("X-Cache", "HIT")
		c.Data(http.StatusOK, "application/json", []byte(raw))
		return
	}

	files, err := h.Service.SearchEvidenceFiles(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search evidence files"})
		return
	}
	if len(files) == 0 {
		_ = h.Cache.Set(ctx, key, `{"files":[]}`, 30*time.Second)
		c.JSON(http.StatusNotFound, gin.H{"error": "No matching evidence files found"})
		return
	}

	body, _ := json.Marshal(gin.H{"files": files})
	_ = h.Cache.Set(ctx, key, string(body), 120*time.Second)

	etag := cache.ListETag(body)
	if middleware.IfNoneMatch(c.Writer, c.Request, etag) {
		c.Header("X-Cache", "REVALIDATED")
		return
	}
	middleware.SetCacheControl(c.Writer, 120)
	c.Header("ETag", etag)
	c.Header("X-Cache", "MISS")
	c.Data(http.StatusOK, "application/json", body)
}

// ----- 4) FILTER: POST /evidence/case/:case_id/filter -----
// Key per case+filters; TTL 60–120s ; ETag+304 ; Cache-Control: 120
type FilterRequest struct {
	Filters   map[string]interface{} `json:"filters"`
	SortField string                 `json:"sort_field"`
	SortOrder string                 `json:"sort_order"`
}

func (h *EvidenceViewerHandler) GetFilteredEvidence(c *gin.Context) {
	caseID := c.Param("case_id")
	if caseID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing case ID"})
		return
	}
	tenantID := tenantIDFromCtx(c)

	var req FilterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	qsig := cache.BuildQuerySig(
		c.DefaultQuery("page", "1"),
		c.DefaultQuery("pageSize", "20"),
		req.SortField,
		req.SortOrder,
		map[string]any{"filters": req.Filters},
	)
	key := cache.EvidenceListKey(tenantID, caseID, qsig)

	ctx := c.Request.Context()

	if raw, ok, _ := h.Cache.Get(ctx, key); ok && raw != "" {
		etag := cache.ListETag([]byte(raw))
		if middleware.IfNoneMatch(c.Writer, c.Request, etag) {
			c.Header("X-Cache", "REVALIDATED")
			return
		}
		middleware.SetCacheControl(c.Writer, 120)
		c.Header("ETag", etag)
		c.Header("X-Cache", "HIT")
		c.Data(http.StatusOK, "application/json", []byte(raw))
		return
	}

	files, err := h.Service.GetFilteredEvidenceFiles(caseID, req.Filters, req.SortField, req.SortOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to filter evidence files"})
		return
	}
	if len(files) == 0 {
		_ = h.Cache.Set(ctx, key, `{"files":[]}`, 30*time.Second)
		c.JSON(http.StatusNotFound, gin.H{"error": "No matching evidence files found"})
		return
	}

	body, _ := json.Marshal(gin.H{"files": files})
	_ = h.Cache.Set(ctx, key, string(body), 120*time.Second)

	etag := cache.ListETag(body)
	if middleware.IfNoneMatch(c.Writer, c.Request, etag) {
		c.Header("X-Cache", "REVALIDATED")
		return
	}
	middleware.SetCacheControl(c.Writer, 120)
	c.Header("ETag", etag)
	c.Header("X-Cache", "MISS")
	c.Data(http.StatusOK, "application/json", body)
}
