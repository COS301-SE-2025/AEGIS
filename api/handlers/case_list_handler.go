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

// Add the CaseHandler struct definition if not present, or update it:

func getActorFromContext(c *gin.Context) auditlog.Actor {
	userIDVal, _ := c.Get("userID")
	userRoleVal, _ := c.Get("userRole")
	emailVal, _ := c.Get("email")

	var userID, userRole, email string
	if v, ok := userIDVal.(string); ok && v != "" {
		userID = v
	}
	if v, ok := userRoleVal.(string); ok && v != "" {
		userRole = v
	}
	if v, ok := emailVal.(string); ok && v != "" {
		email = v
	}

	return auditlog.Actor{
		ID:        userID,
		Role:      userRole,
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Email:     email,
	}
}

// GET /cases/all
func (h *CaseHandler) GetAllCasesHandler(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	actor := getActorFromContext(c)

	// Optional paging/sorting to future-proof keys
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "20")
	sort := c.DefaultQuery("sort", "created_at")
	order := c.DefaultQuery("order", "desc")

	// Build qSig & key
	qSig := cache.BuildQuerySig(page, pageSize, sort, order, map[string]any{
		"scope": "all",
	})
	key := cache.ListKey(tenantID, cache.ScopeAll, qSig)

	ctx := c.Request.Context()

	// 1) Cache hit
	if raw, ok, _ := h.Cache.Get(ctx, key); ok {
		etag := cache.ListETag([]byte(raw))
		if c.GetHeader("If-None-Match") == etag {
			c.Header("ETag", etag)
			c.Status(http.StatusNotModified)
			return
		}
		c.Header("ETag", etag)
		c.Header("Cache-Control", "private, max-age=120")
		c.Data(http.StatusOK, "application/json", []byte(raw))
		return
	}

	// 2) Miss → service
	cases, err := h.ListCasesService.GetAllCases(tenantID)
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "LIST_ALL_CASES",
			Actor:       actor,
			Target:      auditlog.Target{Type: "case_listing"},
			Service:     "case",
			Status:      "FAILED",
			Description: "Failed to list all cases: " + err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not retrieve cases"})
		return
	}

	// 3) Build payload + cache it
	body, _ := json.Marshal(gin.H{
		"cases": cases,
		"meta":  gin.H{"page": page, "pageSize": pageSize, "sort": sort, "order": order},
	})
	_ = h.Cache.Set(ctx, key, string(body), 120*time.Second)

	// 4) ETag + headers + body
	etag := cache.ListETag(body)
	if c.GetHeader("If-None-Match") == etag {
		c.Header("ETag", etag)
		c.Status(http.StatusNotModified)
		return
	}
	c.Header("ETag", etag)
	c.Header("Cache-Control", "private, max-age=120")
	c.Data(http.StatusOK, "application/json", body)

	// 5) Friendlier activity text
	h.auditLogger.Log(c, auditlog.AuditLog{
		Action:      "LIST_ALL_CASES",
		Actor:       actor,
		Target:      auditlog.Target{Type: "case_listing"},
		Service:     "case",
		Status:      "SUCCESS",
		Description: fmt.Sprintf("You viewed all cases (%d found).", len(cases)),
	})
}

// GET /cases/user/:user_id
func (h *CaseHandler) GetCasesByUserHandler(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	actor := getActorFromContext(c)
	paramUserID := c.Param("user_id")

	if paramUserID == "" {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "LIST_USER_CASES",
			Actor:       actor,
			Target:      auditlog.Target{Type: "case_listing_by_user"},
			Service:     "case",
			Status:      "FAILED",
			Description: "Missing user_id parameter",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	cases, err := h.ListCasesService.GetCasesByUser(tenantID, paramUserID)
	if err != nil {
		fmt.Printf("[GetCasesByUserHandler] failed: %v\n", err)
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "LIST_USER_CASES",
			Actor:       actor,
			Target:      auditlog.Target{Type: "case_listing_by_user", ID: paramUserID},
			Service:     "case",
			Status:      "FAILED",
			Description: "Failed to get cases for user: " + err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not retrieve cases for user"})
		return
	}

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action:      "LIST_USER_CASES",
		Actor:       actor,
		Target:      auditlog.Target{Type: "case_listing_by_user", ID: paramUserID},
		Service:     "case",
		Status:      "SUCCESS",
		Description: fmt.Sprintf("Retrieved %d cases for user %s", len(cases), paramUserID),
	})

	c.JSON(http.StatusOK, gin.H{"cases": cases})
}

// GET /cases/filter
func (h *CaseHandler) GetFilteredCasesHandler(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	actor := getActorFromContext(c)

	// Get user and team from context
	userID := c.GetString("userID")
	teamID := c.GetString("teamID")

	status := c.Query("status")
	priority := c.Query("priority")
	createdBy := c.Query("created_by")
	teamName := c.Query("team_name")
	titleTerm := c.Query("title_term")
	sortBy := c.Query("sort_by")
	order := c.Query("order")
	progress := c.Query("progress")

	// Always filter by user/team access
	cases, err := h.ListCasesService.GetFilteredCases(
		tenantID, status, priority, createdBy, teamName, titleTerm, sortBy, order, userID, teamID,
	)
	if err != nil {
		fmt.Printf("[GetFilteredCasesHandler] failed: %v\n", err)
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "LIST_FILTERED_CASES",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "case_filtered_listing",
				AdditionalInfo: map[string]string{
					"status":    status,
					"priority":  priority,
					"createdBy": createdBy,
					"teamName":  teamName,
					"titleTerm": titleTerm,
					"sortBy":    sortBy,
					"order":     order,
					"progress":  progress,
					"userID":    userID,
					"teamID":    teamID,
				},
			},
			Service:     "case",
			Status:      "FAILED",
			Description: "Failed to apply case filters: " + err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not apply filters"})
		return
	}

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "LIST_FILTERED_CASES",
		Actor:  actor,
		Target: auditlog.Target{
			Type: "case_filtered_listing",
			AdditionalInfo: map[string]string{
				"status":    status,
				"priority":  priority,
				"createdBy": createdBy,
				"teamName":  teamName,
				"titleTerm": titleTerm,
				"sortBy":    sortBy,
				"order":     order,
				"progress":  progress,
				"userID":    userID,
				"teamID":    teamID,
			},
		},
		Service:     "case",
		Status:      "SUCCESS",
		Description: fmt.Sprintf("Retrieved %d filtered cases", len(cases)),
	})

	c.JSON(http.StatusOK, gin.H{"cases": cases})
}

// GET /cases/:case_id
func (h *CaseHandler) GetCaseByIDHandler(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	actor := getActorFromContext(c)
	caseID := c.Param("case_id")

	if caseID == "" {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "GET_CASE_BY_ID", Actor: actor,
			Target:  auditlog.Target{Type: "case_details"},
			Service: "case", Status: "FAILED",
			Description: "Missing case_id parameter",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "case_id is required"})
		return
	}

	// --- Cache lookup ---
	key := cache.CaseHeaderKey(tenantID, caseID)
	ctx := c.Request.Context()

	if raw, ok, _ := h.Cache.Get(ctx, key); ok {
		etag := cache.EntityETag([]byte(raw))
		if c.GetHeader("If-None-Match") == etag {
			c.Header("ETag", etag)
			c.Status(http.StatusNotModified)
			return
		}
		c.Header("ETag", etag)
		c.Header("Cache-Control", "private, max-age=300")
		c.Data(http.StatusOK, "application/json", []byte(raw))
		return
	}

	// --- Cache miss → service ---
	caseDetails, err := h.ListCasesService.GetCaseByID(caseID, tenantID)
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "GET_CASE_BY_ID", Actor: actor,
			Target:  auditlog.Target{Type: "case_details", ID: caseID},
			Service: "case", Status: "FAILED",
			Description: "Couldn’t load case details: " + err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not retrieve case"})
		return
	}

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action:      "GET_CASE_BY_ID",
		Actor:       actor,
		Target:      auditlog.Target{Type: "case_details", ID: caseID},
		Service:     "case",
		Status:      "SUCCESS",
		Description: fmt.Sprintf("Retrieved case %s", caseID),
	})
	// Build payload & cache it (5 min)
	body, _ := json.Marshal(gin.H{"case": caseDetails})
	_ = h.Cache.Set(ctx, key, string(body), 5*time.Minute)

	// ETag + conditional + headers
	etag := cache.EntityETag(body)
	if c.GetHeader("If-None-Match") == etag {
		c.Header("ETag", etag)
		c.Status(http.StatusNotModified)
		return
	}
	c.Header("ETag", etag)
	c.Header("Cache-Control", "private, max-age=300")
	c.Data(http.StatusOK, "application/json", body)

	// Friendlier activity text for the feed
	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "GET_CASE_BY_ID", Actor: actor,
		Target:  auditlog.Target{Type: "case_details", ID: caseID},
		Service: "case", Status: "SUCCESS",
		Description: "You viewed the case details of" + caseID + ".",
	})
}

// GET /cases/archived
func (h *CaseHandler) ListArchivedCasesHandler(c *gin.Context) {
	actor := getActorFromContext(c)

	userIDVal, ok := c.Get("userID")
	if !ok || userIDVal == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userID missing"})
		return
	}
	tenantIDVal, ok := c.Get("tenantID")
	if !ok || tenantIDVal == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenantID missing"})
		return
	}
	teamIDVal, ok := c.Get("teamID")
	if !ok || teamIDVal == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "teamID missing"})
		return
	}

	userID, ok1 := userIDVal.(string)
	tenantID, ok2 := tenantIDVal.(string)
	teamID, ok3 := teamIDVal.(string)
	if !ok1 || !ok2 || !ok3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid context values"})
		return
	}

	cases, err := h.ListArchivedCasesService.ListArchivedCases(userID, tenantID, teamID)
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "LIST_ARCHIVED_CASES",
			Actor:       actor,
			Target:      auditlog.Target{Type: "case_archived_listing"},
			Service:     "case",
			Status:      "FAILED",
			Description: "Failed to list archived cases: " + err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not retrieve archived cases"})
		return
	}
	h.auditLogger.Log(c, auditlog.AuditLog{
		Action:      "LIST_ARCHIVED_CASES",
		Actor:       actor,
		Target:      auditlog.Target{Type: "case_archived_listing"},
		Service:     "case",
		Status:      "SUCCESS",
		Description: "Retrieved archived cases successfully",
	})
	c.JSON(http.StatusOK, gin.H{"archived_cases": cases})
}
