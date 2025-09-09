package integration_test

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	case_assign "aegis-api/services_/case/case_assign"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// (optional) test auth injector
func withTestAuth(tenantID, teamID uuid.UUID) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("tenantID", tenantID.String())
		c.Set("teamID", teamID.String())
		c.Next()
	}
}

func registerCaseAssignmentTestEndpoints(r *gin.Engine) {
	repo := case_assign.NewGormCaseAssignmentRepo(pgDB)

	type assignPayload struct {
		UserID string `json:"user_id" binding:"required"`
		CaseID string `json:"case_id" binding:"required"`
		Role   string `json:"role"    binding:"required"`
	}
	type unassignPayload struct {
		UserID string `json:"user_id" binding:"required"`
		CaseID string `json:"case_id" binding:"required"`
	}

	// ensure tenantID/teamID in ctx for all routes in this group
	api := r.Group("/", withTestAuth(FixedTenantID, FixedTeamID))

	api.POST("/cases/assign", func(c *gin.Context) {
		var p assignPayload
		if err := c.ShouldBindJSON(&p); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
			return
		}

		userID, err := uuid.Parse(p.UserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
			return
		}
		caseID, err := uuid.Parse(p.CaseID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case_id"})
			return
		}

		tenantID, err := uuid.Parse(c.GetString("tenantID"))
		if err != nil || tenantID == uuid.Nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing tenant"})
			return
		}

		// Resolve teamID: prefer ctx, else fall back to the case's team_id
		var teamID uuid.UUID
		if s := c.GetString("teamID"); s != "" {
			if t, err := uuid.Parse(s); err == nil {
				teamID = t
			}
		}
		if teamID == uuid.Nil {
			// Fallback: read from case row (your insertCaseRow sets this to FixedTeamID)
			if err := pgSQL.QueryRow(`SELECT team_id FROM cases WHERE id=$1`, caseID).Scan(&teamID); err != nil {
				if err == sql.ErrNoRows {
					c.JSON(http.StatusNotFound, gin.H{"error": "case not found"})
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				}
				return
			}
		}
		if teamID == uuid.Nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "case has no team_id"})
			return
		}

		if err := repo.AssignRole(userID, caseID, p.Role, tenantID, teamID); err != nil {
			if isUniqueViolation(err) {
				c.JSON(http.StatusConflict, gin.H{"error": "already assigned"})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"user_id":   userID.String(),
			"case_id":   caseID.String(),
			"role":      p.Role,
			"tenant_id": tenantID.String(),
			"team_id":   teamID.String(), // helpful for assertions
			"assigned":  time.Now().Format(time.RFC3339),
		})
	})

	api.POST("/cases/unassign", func(c *gin.Context) {
		var p unassignPayload
		if err := c.ShouldBindJSON(&p); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
			return
		}
		userID, err := uuid.Parse(p.UserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
			return
		}
		caseID, err := uuid.Parse(p.CaseID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case_id"})
			return
		}

		if err := repo.UnassignRole(userID, caseID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var dummy int
		err = pgSQL.QueryRow(`SELECT 1 FROM case_user_roles WHERE user_id=$1 AND case_id=$2`, userID, caseID).Scan(&dummy)
		if err == sql.ErrNoRows {
			c.Status(http.StatusNoContent)
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusConflict, gin.H{"error": "role still present"})
	})
}

// duplicate-key detection for pq/pgx/GORM
func isUniqueViolation(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && string(pqErr.Code) == "23505" {
		return true
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "duplicate key value") ||
		strings.Contains(strings.ToLower(err.Error()), "unique constraint")
}
