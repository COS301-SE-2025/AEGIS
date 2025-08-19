package integration_test

// --- add these imports at top of integration_bootstrap_test.go ---
import (
	"database/sql"
	"net/http"
	"time"

	case_creation "aegis-api/services_/case/case_creation"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	r.GET("/cases/:id", func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
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
			"id":                  id.String(),
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
}
