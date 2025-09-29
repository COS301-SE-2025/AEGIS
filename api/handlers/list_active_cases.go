package handlers

import (
	"aegis-api/cache"
	"aegis-api/middleware"
	"aegis-api/services_/auditlog"
	"aegis-api/services_/case/ListCases"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Assuming your ListCases service looks like this:
type ListCasesService interface {
	//ListActiveCases(userID string) ([]ListActiveCases.ActiveCase, error)
	GetAllCases(tenantID string) ([]ListCases.Case, error)
	GetCasesByUser(userID string, tenantID string) ([]ListCases.Case, error)
	GetFilteredCases(TenantID, status, priority, createdBy, teamName, titleTerm, sortBy, order string, userID, teamID string) ([]ListCases.Case, error)
	GetCaseByID(caseID string, tenantID string) (*ListCases.Case, error)
}

type CaseListHandler struct {
	Service ListCasesService
}

func NewCaseListHandler(service ListCasesService) *CaseListHandler {
	return &CaseListHandler{Service: service}
}

func (h *CaseHandler) ListActiveCasesHandler(c *gin.Context) {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user or tenant/team ID in token"})
		return
	}

	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "20")
	sort := c.DefaultQuery("sort", "created_at")
	order := c.DefaultQuery("order", "desc")

	qSig := cache.BuildQuerySig(page, pageSize, sort, order, map[string]any{
		"scope": "active", "userId": userID, "teamId": teamID,
	})
	key := cache.ListKey(tenantID, cache.ScopeActive, qSig)

	roleStr, _ := rolev.(string)
	actor := auditlog.Actor{ID: userID, Role: roleStr, IPAddress: c.ClientIP(), UserAgent: c.Request.UserAgent()}

	ctx := c.Request.Context()

	// 1) Try cache (HIT)
	if raw, ok, _ := h.Cache.Get(ctx, key); ok {
		etag := cache.ListETag([]byte(raw))
		if middleware.IfNoneMatch(c.Writer, c.Request, etag) {
			c.Header("X-Cache", "REVALIDATED")
			return
		}
		middleware.SetCacheControl(c.Writer, 120)
		c.Header("X-Cache", "HIT")
		c.Data(http.StatusOK, "application/json", []byte(raw))
		return
	}

	// 2) MISS → service
	cases, err := h.CaseService.ListActiveCases(userID, tenantID, teamID)
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "LIST_ACTIVE_CASES", Actor: actor,
			Target: auditlog.Target{Type: "active_case_listing", ID: userID,
				AdditionalInfo: map[string]string{"tenant_id": tenantID, "team_id": teamID}},
			Service: "case", Status: "FAILED", Description: "Failed to list active cases: " + err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list active cases"})
		return
	}

	// Map progress for each active case
	for i := range cases {
		cases[i].Progress = getProgressForStage(cases[i].InvestigationStage)
	}

	// ✅ Audit successful request
	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "LIST_ACTIVE_CASES",
		Actor:  actor,
		Target: auditlog.Target{
			Type: "active_case_listing",
			ID:   userID,
			AdditionalInfo: map[string]string{
				"tenant_id": tenantID,
				"team_id":   teamID,
			},
		},

		Service:     "case",
		Status:      "SUCCESS",
		Description: fmt.Sprintf("Retrieved %d active cases for user %s", len(cases), userID)})

	body, _ := json.Marshal(gin.H{
		"cases": cases,
		"meta":  gin.H{"page": page, "pageSize": pageSize, "sort": sort, "order": order},
	})
	_ = h.Cache.Set(ctx, key, string(body), 120*time.Second)

	etag := cache.ListETag(body)
	if middleware.IfNoneMatch(c.Writer, c.Request, etag) {
		c.Header("X-Cache", "REVALIDATED")
		return
	}
	middleware.SetCacheControl(c.Writer, 120)
	c.Header("X-Cache", "MISS")
	c.Data(http.StatusOK, "application/json", body)

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "LIST_ACTIVE_CASES", Actor: actor,
		Target: auditlog.Target{Type: "active_case_listing", ID: userID,
			AdditionalInfo: map[string]string{"tenant_id": tenantID, "team_id": teamID}},
		Service: "case", Status: "SUCCESS",
		Description: fmt.Sprintf("You viewed your active cases (%d found).", len(cases)),
	})
}
