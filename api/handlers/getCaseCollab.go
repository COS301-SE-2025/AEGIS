package handlers

import (
	"aegis-api/services_/admin/get_collaborators"
	"aegis-api/services_/auditlog"
	"aegis-api/structs"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GetCollaboratorsHandler struct {
	service     *get_collaborators.Service
	auditLogger *auditlog.AuditLogger
}

func NewGetCollaboratorsHandler(service *get_collaborators.Service, auditLogger *auditlog.AuditLogger) *GetCollaboratorsHandler {
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
