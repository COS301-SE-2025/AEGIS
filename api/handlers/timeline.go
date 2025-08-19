package handlers

import (
	"net/http"

	"aegis-api/services_/timeline"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

type TimelineHandler struct {
	service timeline.Service
}

func NewTimelineHandler(s timeline.Service) *TimelineHandler {
	return &TimelineHandler{service: s}
}

func (h *TimelineHandler) ListByCase(c *gin.Context) {
	caseID := c.Param("case_id")
	events, err := h.service.ListEvents(caseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, events)
}

func (h *TimelineHandler) Create(c *gin.Context) {
	caseID := c.Param("case_id")
	var req struct {
		Description string   `json:"description" binding:"required"`
		Evidence    []string `json:"evidence"`
		Tags        []string `json:"tags"`
		Severity    string   `json:"severity"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	analystID, _ := c.Get("userID")
	analystName, _ := c.Get("fullName")

	ev := &timeline.TimelineEvent{
		CaseID:      caseID,
		Description: req.Description,
		Severity:    req.Severity,
		AnalystID:   analystID.(string),
		AnalystName: analystName.(string),
	}

	// convert evidence/tags to JSON
	ev.Evidence = datatypes.JSON([]byte("[]"))
	ev.Tags = datatypes.JSON([]byte("[]"))
	if len(req.Evidence) > 0 {
		b, _ := json.Marshal(req.Evidence)
		ev.Evidence = datatypes.JSON(b)
	}
	if len(req.Tags) > 0 {
		b, _ := json.Marshal(req.Tags)
		ev.Tags = datatypes.JSON(b)
	}

	created, err := h.service.AddEvent(ev)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Transform to the same shape ListByCase uses
	resp := timeline.TimelineEventResponse{
		ID:          created.ID,
		Description: created.Description,
		Severity:    created.Severity,
		AnalystName: created.AnalystName,
		Date:        created.CreatedAt.Format("2006-01-02"),
		Time:        created.CreatedAt.Format("15:04"),
		Evidence:    created.Evidence,
		Tags:        created.Tags,
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *TimelineHandler) Reorder(c *gin.Context) {
	caseID := c.Param("case_id")
	var req struct {
		OrderedIDs []string `json:"ordered_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	if err := h.service.ReorderEvents(caseID, req.OrderedIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *TimelineHandler) Update(c *gin.Context) {
	eventID := c.Param("event_id")

	// Partial update request struct: fields are pointers to detect if provided
	var req struct {
		Description *string   `json:"description"`
		Evidence    *[]string `json:"evidence"`
		Tags        *[]string `json:"tags"`
		Severity    *string   `json:"severity"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	// Fetch existing event
	event, err := h.service.GetEventByID(eventID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}

	// Update fields only if provided
	if req.Description != nil {
		event.Description = *req.Description
	}
	if req.Severity != nil {
		event.Severity = *req.Severity
	}
	if req.Evidence != nil {
		b, err := json.Marshal(*req.Evidence)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid evidence format"})
			return
		}
		event.Evidence = datatypes.JSON(b)
	}
	if req.Tags != nil {
		b, err := json.Marshal(*req.Tags)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tags format"})
			return
		}
		event.Tags = datatypes.JSON(b)
	}

	// Save updated event
	updated, err := h.service.UpdateEvent(event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *TimelineHandler) Delete(c *gin.Context) {
	eventID := c.Param("event_id")
	if err := h.service.DeleteEvent(eventID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
