// in integration_bootstrap_test.go
// add import:
package integration_test

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	case_assign "aegis-api/services_/case/case_assign"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	// ...existing imports
)

// mount test-only assignment endpoints
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

	// POST /cases/assign
	r.POST("/cases/assign", func(c *gin.Context) {
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
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing tenant"})
			return
		}

		if err := repo.AssignRole(userID, caseID, p.Role, tenantID); err != nil {
			// handle duplicate assignment gracefully
			if strings.Contains(err.Error(), "duplicate key value") {
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
			"assigned":  time.Now().Format(time.RFC3339),
		})
	})

	// POST /cases/unassign
	r.POST("/cases/unassign", func(c *gin.Context) {
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

		// delete via repo
		if err := repo.UnassignRole(userID, caseID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// verify deletion
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
		// still present
		c.JSON(http.StatusConflict, gin.H{"error": "role still present"})
	})
}
