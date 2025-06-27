package handlers

import (
	"aegis-api/services_/case/case_creation"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CaseHandler struct {
	CaseService CaseServiceInterface
}
type CaseServices struct {
	createCase *case_creation.Service
	// listCase           *ListCases.Service
	// updateCaseStatus   *case_status_update.CaseStatusService
	// getCollaborators   *get_collaborators.Service
	// assignCase         *case_assign.CaseAssignmentService
	// removeCollaborator *remove_user_from_case.Service
}

func NewCaseServices(
	createCase *case_creation.Service,
	// listCase *ListCases.Service,
	// updateCaseStatus *case_status_update.CaseStatusService,
	// getCollaborators *get_collaborators.Service,
	// assignCase *case_assign.CaseAssignmentService,
	// removeCollaborator *remove_user_from_case.Service,
) *CaseServices {
	return &CaseServices{
		createCase: createCase,
		// listCase:           listCase,
		// updateCaseStatus:   updateCaseStatus,
		// getCollaborators:   getCollaborators,
		// assignCase:         assignCase,
		// removeCollaborator: removeCollaborator,
	}
}

type CaseServiceInterface interface {
	CreateCase(req *case_creation.CreateCaseRequest) (*case_creation.Case, error)
}

func NewCaseHandler(service CaseServiceInterface) *CaseHandler {
	return &CaseHandler{CaseService: service}
}

// func (cs CaseServices) CreateCase(c *gin.Context) {
// 	var apiReq structs.CreateCaseRequest //
// 	if err := c.ShouldBind(&apiReq); err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: "Invalid case data",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	fullName, exists := c.Get("fullName") //should be set by middleware
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
// 			Error:   "unauthorized",
// 			Message: "No authentication provided",
// 		})
// 		return
// 	}

// 	serviceReq := case_creation.CreateCaseRequest{ //map
// 		Title:              apiReq.Title,
// 		Description:        apiReq.Description,
// 		Status:             apiReq.Status,
// 		Priority:           apiReq.Priority,
// 		InvestigationStage: apiReq.InvestigationStage,
// 		CreatedByFullName:  fullName.(string),
// 		TeamName:           apiReq.TeamName,
// 	}
// 	newCase, err := cs.createCase.CreateCase(&serviceReq)
// 	if err != nil {
// 		status := http.StatusInternalServerError
// 		errorType := "creation_failed"
// 		if strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "invalid") {
// 			status = http.StatusBadRequest
// 			errorType = "validation_failed"
// 		}
// 		c.JSON(status, structs.ErrorResponse{
// 			Error:   errorType,
// 			Message: "Could not create case",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusCreated, structs.SuccessResponse{
// 		Success: true,
// 		Message: "Case created successfully",
// 		Data:    newCase,
// 	})
// }

func (h *CaseHandler) CreateCase(c *gin.Context) {
	var req case_creation.CreateCaseRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	newCase, err := h.CaseService.CreateCase(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create case", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Case created successfully",
		"case":    newCase,
	})
}
