package handlers

import (
	"aegis-api/services_/admin/get_collaborators"
	"aegis-api/services_/auditlog"
	"aegis-api/structs"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Add interfaces for dependency injection
type CollaboratorService interface {
	GetCollaborators(caseID uuid.UUID) ([]get_collaborators.Collaborator, error)
}

type AuditService interface {
	Log(c *gin.Context, log auditlog.AuditLog) error
}

type GetCollaboratorsHandler struct {
	service     CollaboratorService // Changed from concrete type to interface
	auditLogger AuditService        // Changed from concrete type to interface
}

func NewGetCollaboratorsHandler(service CollaboratorService, auditLogger AuditService) *GetCollaboratorsHandler {
	return &GetCollaboratorsHandler{
		service:     service,
		auditLogger: auditLogger,
	}
}

func (h *GetCollaboratorsHandler) GetCollaboratorsByCaseID(c *gin.Context) {
	caseIDParam := c.Param("case_id")
	if caseIDParam == "" {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:  "GET_COLLABORATORS_FOR_CASE",
			Service: "cases",
			Target: auditlog.Target{
				Type: "case",
				ID:   "",
			},
			Status:      "FAILED",
			Description: "Missing case_id parameter",
		})
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "case_id parameter is required",
		})
		return
	}

	caseID, err := uuid.Parse(caseIDParam)
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:  "GET_COLLABORATORS_FOR_CASE",
			Service: "cases",
			Target: auditlog.Target{
				Type: "case",
				ID:   caseIDParam,
			},
			Status:      "FAILED",
			Description: "Invalid case_id format",
		})
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "invalid case_id format",
		})
		return
	}

	collaborators, err := h.service.GetCollaborators(caseID)
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:  "GET_COLLABORATORS_FOR_CASE",
			Service: "cases",
			Target: auditlog.Target{
				Type: "case",
				ID:   caseID.String(),
			},
			Status:      "FAILED",
			Description: "Could not retrieve collaborators: " + err.Error(),
		})
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "internal_error",
			Message: "could not retrieve collaborators: " + err.Error(),
		})
		return
	}

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action:  "GET_COLLABORATORS_FOR_CASE",
		Service: "cases",
		Target: auditlog.Target{
			Type: "case",
			ID:   caseID.String(),
		},
		Status:      "SUCCESS",
		Description: "Collaborators retrieved successfully",
	})

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Collaborators retrieved successfully",
		Data:    collaborators,
	})
}
