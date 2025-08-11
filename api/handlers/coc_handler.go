package handlers

import (
	auditlog "aegis-api/services_/auditlog"
	coc "aegis-api/services_/chain_of_custody"
	"aegis-api/structs"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type CoCHandler struct {
	Service     coc.Service
	AuditLogger *auditlog.AuditLogger
}

func NewCoCHandler(svc coc.Service, audit *auditlog.AuditLogger) *CoCHandler {
	return &CoCHandler{Service: svc, AuditLogger: audit}
}

// POST /api/v1/coc/log
func (h *CoCHandler) Log(c *gin.Context) {
	var req struct {
		CaseID     string  `json:"caseId" binding:"required,uuid4"`
		EvidenceID string  `json:"evidenceId" binding:"required,uuid4"`
		Action     string  `json:"action" binding:"required,oneof=upload download archive view"`
		Reason     *string `json:"reason"`
		Location   *string `json:"location"`
		HashMD5    *string `json:"hashMd5"`
		HashSHA1   *string `json:"hashSha1"`
		HashSHA256 *string `json:"hashSha256"`
		OccurredAt *string `json:"occurredAt"` // optional ISO8601
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.AuditLogger.Log(c, auditlog.AuditLog{
			Action:      "CHAIN_OF_CUSTODY_LOG",
			Actor:       auditlog.Actor{}, // unknown until we read context
			Target:      auditlog.Target{Type: "evidence", ID: req.EvidenceID},
			Service:     "coc",
			Status:      "FAILED",
			Description: "Invalid CoC log payload: " + err.Error(),
		})
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// Validate/parse action
	a, ok := coc.ParseAction(req.Action)
	if !ok {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "action must be one of: upload, download, archive, view",
		})
		return
	}

	var occurred time.Time
	if req.OccurredAt != nil && *req.OccurredAt != "" {
		t, err := time.Parse(time.RFC3339, *req.OccurredAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, structs.ErrorResponse{
				Error:   "invalid_request",
				Message: "occurredAt must be RFC3339",
			})
			return
		}
		occurred = t
	}

	// actor from context (same pattern as your auth handlers)
	var actorID *string
	if v, ok := c.Get("userID"); ok {
		if s, ok2 := v.(string); ok2 && s != "" {
			actorID = &s
		}
	}

	id, err := h.Service.Log(c.Request.Context(), coc.LogParams{
		CaseID:     req.CaseID,
		EvidenceID: req.EvidenceID,
		ActorID:    actorID,
		Action:     a,
		Reason:     req.Reason,
		Location:   req.Location,
		HashMD5:    req.HashMD5,
		HashSHA1:   req.HashSHA1,
		HashSHA256: req.HashSHA256,
		OccurredAt: occurred, // can be zero; service will default to time.Now()
	})
	status := "SUCCESS"
	desc := "CoC entry logged"
	if err != nil {
		status = "FAILED"
		desc = "CoC log failed: " + err.Error()
	}

	h.AuditLogger.Log(c, auditlog.AuditLog{
		Action: "CHAIN_OF_CUSTODY_LOG",
		Actor: auditlog.Actor{
			ID:    deref(actorID),
			Email: "", // fill if you place email in context
			Role:  "", // fill if you place role in context
		},
		Target: auditlog.Target{
			Type: "evidence",
			ID:   req.EvidenceID,
		},
		Service:     "coc",
		Status:      status,
		Description: desc,
	})

	if err != nil {
		c.JSON(http.StatusForbidden, structs.ErrorResponse{
			Error:   "forbidden",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "CoC entry created",
		Data:    gin.H{"id": id},
	})
}

// GET /api/v1/coc/:evidenceId
func (h *CoCHandler) ListByEvidence(c *gin.Context) {
	eid := c.Param("evidenceId")
	if eid == "" {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "missing evidenceId",
		})
		return
	}

	var f coc.ListFilters
	if s := c.Query("action"); s != "" {
		a, ok := coc.ParseAction(s)
		if !ok {
			c.JSON(http.StatusBadRequest, structs.ErrorResponse{
				Error:   "invalid_request",
				Message: "invalid action",
			})
			return
		}
		f.Action = &a
	}
	if s := c.Query("actorId"); s != "" {
		f.ActorID = &s
	}
	if s := c.Query("since"); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			c.JSON(http.StatusBadRequest, structs.ErrorResponse{
				Error:   "invalid_request",
				Message: "since must be RFC3339",
			})
			return
		}
		f.Since = &t
	}
	if s := c.Query("until"); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			c.JSON(http.StatusBadRequest, structs.ErrorResponse{
				Error:   "invalid_request",
				Message: "until must be RFC3339",
			})
			return
		}
		f.Until = &t
	}
	if s := c.Query("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			f.Limit = n
		} else {
			c.JSON(http.StatusBadRequest, structs.ErrorResponse{
				Error:   "invalid_request",
				Message: "limit must be an integer",
			})
			return
		}
	}
	if s := c.Query("offset"); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			f.Offset = n
		} else {
			c.JSON(http.StatusBadRequest, structs.ErrorResponse{
				Error:   "invalid_request",
				Message: "offset must be an integer",
			})
			return
		}
	}

	entries, err := h.Service.ListByEvidence(c.Request.Context(), eid, f)
	if err != nil {
		h.AuditLogger.Log(c, auditlog.AuditLog{
			Action:      "VIEW_COC_LOG",
			Actor:       auditlog.Actor{}, // fill from context if you keep it
			Target:      auditlog.Target{Type: "evidence", ID: eid},
			Service:     "coc",
			Status:      "FAILED",
			Description: "Failed to list CoC: " + err.Error(),
		})
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "list_failed",
			Message: err.Error(),
		})
		return
	}

	h.AuditLogger.Log(c, auditlog.AuditLog{
		Action:      "VIEW_COC_LOG",
		Actor:       auditlog.Actor{}, // fill from context if needed
		Target:      auditlog.Target{Type: "evidence", ID: eid},
		Service:     "coc",
		Status:      "SUCCESS",
		Description: "CoC listed successfully",
	})

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "OK",
		Data:    gin.H{"entries": entries},
	})
}

// GET /api/v1/coc/:evidenceId/export.csv
func (h *CoCHandler) ExportCSV(c *gin.Context) {
	eid := c.Param("evidenceId")
	if eid == "" {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "missing evidenceId",
		})
		return
	}

	entries, err := h.Service.ListByEvidence(c.Request.Context(), eid, coc.ListFilters{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "export_failed",
			Message: err.Error(),
		})
		return
	}
	csvBytes, err := h.Service.ToCSV(entries)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "export_failed",
			Message: err.Error(),
		})
		return
	}

	// audit
	h.AuditLogger.Log(c, auditlog.AuditLog{
		Action:      "EXPORT_COC_CSV",
		Actor:       auditlog.Actor{},
		Target:      auditlog.Target{Type: "evidence", ID: eid},
		Service:     "coc",
		Status:      "SUCCESS",
		Description: "Exported CoC CSV",
	})

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="chain_of_custody_%s.csv"`, eid))
	c.Data(http.StatusOK, "text/csv", csvBytes)
}

func deref(p *string) string {
	if p != nil {
		return *p
	}
	return ""
}
