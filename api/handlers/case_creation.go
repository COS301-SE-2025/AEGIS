package handlers

import (
	"aegis-api/services_/case/ListActiveCases"
	"aegis-api/services_/case/ListCases"
	"aegis-api/services_/case/case_assign"
	"aegis-api/services_/case/case_creation"
	"fmt"
	"net/http"

	"aegis-api/services_/auditlog"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CaseHandler struct {
	CaseService         CaseServiceInterface
	ListCasesService    ListCasesService
	ListActiveCasesServ ListActiveCasesService
	auditLogger         *auditlog.AuditLogger
}
type ListActiveCasesService interface {
	ListActiveCases(userID string) ([]ListActiveCases.ActiveCase, error)
}
type CaseServices struct {
	createCase *case_creation.Service
	//listCase           *ListCases.Service
	// updateCaseStatus   *case_status_update.CaseStatusService
	// getCollaborators   *get_collaborators.Service
	listCase   *ListCases.Service
	listActive *ListActiveCases.Service

	assignCase *case_assign.CaseAssignmentService
	// removeCollaborator *remove_user_from_case.Service
}

func NewCaseServices(
	createCase *case_creation.Service,
	// listCase *ListCases.Service,
	listCase *ListCases.Service,
	listActive *ListActiveCases.Service,
	// updateCaseStatus *case_status_update.CaseStatusService,
	// getCollaborators *get_collaborators.Service,
	assignCase *case_assign.CaseAssignmentService,
	// removeCollaborator *remove_user_from_case.Service,
) *CaseServices {
	return &CaseServices{
		createCase: createCase,
		// listCase:           listCase,
		// updateCaseStatus:   updateCaseStatus,
		// getCollaborators:   getCollaborators,
		listCase:   listCase,
		listActive: listActive,
		assignCase: assignCase,
		// removeCollaborator: removeCollaborator,
	}
}
func (s *CaseServices) ListActiveCases(userID string) ([]ListActiveCases.ActiveCase, error) {
	return s.listActive.ListActiveCases(userID)
}

func (s *CaseServices) GetAllCases() ([]ListCases.Case, error) {
	return s.listCase.GetAllCases()
}

type CaseServiceInterface interface {
	CreateCase(req *case_creation.CreateCaseRequest) (*case_creation.Case, error)
	AssignUserToCase(assignerRole string, assigneeID, caseID uuid.UUID, role string) error
	ListActiveCases(userID string) ([]ListActiveCases.ActiveCase, error)
	GetCaseByID(caseID string) (*ListCases.Case, error)
	UnassignUserFromCase(assignerID *gin.Context, assigneeID, caseID uuid.UUID) error // ← Add this

}

func NewCaseHandler(
	caseService CaseServiceInterface,
	listCasesService ListCasesService,
	listActiveCasesService ListActiveCasesService,
	auditLogger *auditlog.AuditLogger,
) *CaseHandler {
	return &CaseHandler{
		CaseService:         caseService,
		ListCasesService:    listCasesService,
		ListActiveCasesServ: listActiveCasesService,
		auditLogger:         auditLogger,
	}
}

func (s *CaseServices) CreateCase(req *case_creation.CreateCaseRequest) (*case_creation.Case, error) {
	return s.createCase.CreateCase(req)
}

func (s *CaseServices) AssignUserToCase(assignerRole string, assigneeID, caseID uuid.UUID, role string) error {
	return s.assignCase.AssignUserToCase(assignerRole, assigneeID, caseID, role)
}

func (s *CaseServices) GetCaseByID(caseID string) (*ListCases.Case, error) {
	return s.listCase.GetCaseByID(caseID)
}

func (s *CaseServices) UnassignUserFromCase(ctx *gin.Context, assigneeID, caseID uuid.UUID) error {
	return s.assignCase.UnassignUserFromCase(ctx, assigneeID, caseID)
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

	// Grab user details from context
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")

	actor := auditlog.Actor{
		ID:        userID.(string),
		Role:      userRole.(string),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("[CreateCase] Invalid JSON input: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "CREATE_CASE",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "case",
				ID:   "",
			},
			Service:     "case",
			Status:      "FAILED",
			Description: "Invalid JSON input for creating case",
		})

		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid JSON payload",
			"details": err.Error(),
		})
		return
	}

	fmt.Printf("[CreateCase] Received valid request payload: %+v\n", req)

	newCase, err := h.CaseService.CreateCase(&req)
	if err != nil {
		fmt.Printf("[CreateCase] Failed to create case: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "CREATE_CASE",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "case",
				ID:   "",
			},
			Service:     "case",
			Status:      "FAILED",
			Description: "Failed to create case: " + err.Error(),
		})

		status := http.StatusInternalServerError
		errorType := "creation_failed"
		if err.Error() == "title is required" || err.Error() == "team name is required" || err.Error() == "created_by is required" {
			status = http.StatusBadRequest
			errorType = "validation_failed"
		}

		c.JSON(status, gin.H{
			"error":   errorType,
			"message": "Could not create case",
			"details": err.Error(),
		})
		return
	}

	fmt.Printf("[CreateCase] Successfully created case: %+v\n", newCase)

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "CREATE_CASE",
		Actor:  actor,
		Target: auditlog.Target{
			Type: "case",
			ID:   newCase.ID.String(),
		},
		Service:     "case",
		Status:      "SUCCESS",
		Description: "Case created successfully",
	})

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Case created successfully",
		"case":    newCase,
	})
}
