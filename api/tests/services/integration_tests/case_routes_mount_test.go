package integration_test

// --- add these imports at top of integration_bootstrap_test.go ---
import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"aegis-api/handlers"
	graphicalmapping "aegis-api/services_/GraphicalMapping"
	lac "aegis-api/services_/case/ListActiveCases"
	case_creation "aegis-api/services_/case/case_creation"
	timeline "aegis-api/services_/timeline"

	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	// existing imports...
)

// --- add this test-only repo somewhere at file-scope ---
type testCaseRepo struct{ db *sql.DB }

func (r *testCaseRepo) CreateCase(c *case_creation.Case) error {
	// Keep it minimal; let defaults fill status/stage/priority/created_at
	_, err := r.db.Exec(`
		INSERT INTO cases (id, title, description, team_name, created_by, tenant_id, team_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
	`, c.ID, c.Title, c.Description, c.TeamName, c.CreatedBy, c.TenantID, c.TeamID)
	return err
}

// Implement GetCaseByID to satisfy case_creation.CaseRepository
func (r *testCaseRepo) GetCaseByID(ctx context.Context, id uuid.UUID) (*case_creation.Case, error) {
	var (
		title, description, teamName, status, priority, stage string
		createdBy, tenantID, teamID                           uuid.UUID
		createdAt                                             time.Time
	)
	err := r.db.QueryRowContext(ctx, `
		SELECT title,
			   COALESCE(description,''), team_name,
			   COALESCE(status::text,'open'),
			   COALESCE(priority::text,'medium'),
			   COALESCE(investigation_stage::text,'Triage'),
			   created_by, tenant_id, team_id, created_at
		FROM cases WHERE id = $1
	`, id).Scan(&title, &description, &teamName, &status, &priority, &stage,
		&createdBy, &tenantID, &teamID, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &case_creation.Case{
		ID:                 id,
		Title:              title,
		Description:        description,
		TeamName:           teamName,
		Status:             status,
		Priority:           priority,
		InvestigationStage: stage,
		CreatedBy:          createdBy,
		TenantID:           tenantID,
		TeamID:             teamID,
		CreatedAt:          createdAt,
	}, nil
}

// --- helper to register the two endpoints on the test router ---
func registerCaseTestEndpoints(r *gin.Engine) {
	repo := &testCaseRepo{db: pgSQL}
	svc := case_creation.NewCaseService(repo, nil, nil) // nil notif/hub ok

	type createCasePayload struct {
		Title              string `json:"title" binding:"required"`
		Description        string `json:"description"`
		TeamName           string `json:"team_name" binding:"required"`
		Status             string `json:"status,omitempty"`
		Priority           string `json:"priority,omitempty"`
		InvestigationStage string `json:"investigation_stage,omitempty"`
	}

	// POST /cases (201 on success)
	r.POST("/cases", func(c *gin.Context) {
		var p createCasePayload
		if err := c.ShouldBindJSON(&p); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
			return
		}

		req := case_creation.CreateCaseRequest{
			Title:              p.Title,
			Description:        p.Description,
			TeamName:           p.TeamName,
			Status:             p.Status,
			Priority:           p.Priority,
			InvestigationStage: p.InvestigationStage,
		}

		var err error
		if req.CreatedBy, err = uuid.Parse(c.GetString("userID")); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user"})
			return
		}
		if req.TenantID, err = uuid.Parse(c.GetString("tenantID")); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing tenant"})
			return
		}
		if req.TeamID, err = uuid.Parse(c.GetString("teamID")); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing team"})
			return
		}

		entity, err := svc.CreateCase(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"id":         entity.ID.String(),
			"title":      entity.Title,
			"team_name":  entity.TeamName,
			"tenant_id":  entity.TenantID.String(),
			"created_by": entity.CreatedBy.String(),
			"status":     entity.Status,
			"created_at": entity.CreatedAt.Format(time.RFC3339),
		})
	})

	// GET /cases/:id (200 or 404)
	r.GET("/cases/:case_id", func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("case_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		var (
			title, description, teamName, status, priority, stage string
			createdBy, tenantID, teamID                           uuid.UUID
			createdAt                                             time.Time
		)
		err = pgSQL.QueryRow(`
            SELECT title,
                   COALESCE(description,''), team_name,
                   COALESCE(status::text,'open'),
                   COALESCE(priority::text,'medium'),
                   COALESCE(investigation_stage::text,'Triage'),
                   created_by, tenant_id, team_id, created_at
            FROM cases WHERE id = $1
        `, id).Scan(&title, &description, &teamName, &status, &priority, &stage,
			&createdBy, &tenantID, &teamID, &createdAt)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"case_id":             id.String(),
			"title":               title,
			"description":         description,
			"team_name":           teamName,
			"status":              status,
			"priority":            priority,
			"investigation_stage": stage,
			"created_by":          createdBy.String(),
			"tenant_id":           tenantID.String(),
			"team_id":             teamID.String(),
			"created_at":          createdAt.Format(time.RFC3339),
		})
	})

	r.GET("/cases/active", func(c *gin.Context) {
		uid := c.GetString("userID")
		tid := c.GetString("tenantID")
		gid := c.GetString("teamID")

		repo := lac.NewActiveCaseRepository(pgDB)
		items, err := repo.GetActiveCasesByUserID(tcCtx, uid, tid, gid)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, items)
	})

}
func registerTimelineTestEndpoints(r *gin.Engine) {
	// Initialize testTimelineService before using it
	testTimelineService := timeline.NewService(timeline.NewRepository(pgDB)) // Use the correct constructor for the timeline service

	// List all events for a case
	r.GET("/cases/:case_id/timeline", func(c *gin.Context) {
		caseID := c.Param("case_id")
		events, err := testTimelineService.ListEvents(caseID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, events)
	})

	// Create new event for a case
	r.POST("/cases/:case_id/timeline", func(c *gin.Context) {
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
		ev := &timeline.TimelineEvent{
			CaseID:      caseID,
			Description: req.Description,
			Severity:    req.Severity,
			AnalystID:   FixedUserID.String(),
			AnalystName: "Test Analyst",
		}
		// Convert evidence/tags to JSON
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
		created, err := testTimelineService.AddEvent(ev)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, created)
	})

	// Update a timeline event by ID
	r.PATCH("/timeline/:event_id", func(c *gin.Context) {
		eventID := c.Param("event_id")
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
		event, err := testTimelineService.GetEventByID(eventID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
			return
		}
		if req.Description != nil {
			event.Description = *req.Description
		}
		if req.Severity != nil {
			event.Severity = *req.Severity
		}
		if req.Evidence != nil {
			b, _ := json.Marshal(*req.Evidence)
			event.Evidence = datatypes.JSON(b)
		}
		if req.Tags != nil {
			b, _ := json.Marshal(*req.Tags)
			event.Tags = datatypes.JSON(b)
		}
		updated, err := testTimelineService.UpdateEvent(event)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, updated)
	})

	// Delete a timeline event by ID
	r.DELETE("/timeline/:event_id", func(c *gin.Context) {
		eventID := c.Param("event_id")
		if err := testTimelineService.DeleteEvent(eventID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	})

	// Reorder events for a case
	r.POST("/cases/:case_id/timeline/reorder", func(c *gin.Context) {
		caseID := c.Param("case_id")
		var req struct {
			OrderedIDs []string `json:"ordered_ids" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		if err := testTimelineService.ReorderEvents(caseID, req.OrderedIDs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	})
}
func registerGraphicalMappingTestEndpoints(r *gin.Engine) {
	// You need a test IOC service/repo. Adjust as needed for your setup:
	testIOCService := graphicalmapping.NewIOCService(graphicalmapping.NewIOCRepository(pgDB))
	testIOCHandler := handlers.NewIOCHandler(testIOCService)

	// GET /tenants/:tenantId/ioc-graph
	r.GET("/tenants/:tenantId/ioc-graph", testIOCHandler.GetTenantIOCGraph)

	// GET /tenants/:tenantId/cases/:case_id/ioc-graph
	r.GET("/tenants/:tenantId/cases/:case_id/ioc-graph", testIOCHandler.GetCaseIOCGraph)

	// GET /cases/:case_id/iocs
	r.GET("/cases/:case_id/iocs", testIOCHandler.GetIOCsByCase)

	// POST /cases/:case_id/iocs
	r.POST("/cases/:case_id/iocs", testIOCHandler.AddIOCToCase)
}
