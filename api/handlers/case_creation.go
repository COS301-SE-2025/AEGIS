package handlers

import (
	"aegis-api/services_/auditlog"
	"aegis-api/services_/case/ListActiveCases"
	"aegis-api/services_/case/ListCases"
	"aegis-api/services_/case/ListClosedCases"
	"aegis-api/services_/case/case_assign"
	"aegis-api/services_/case/case_creation"
	"aegis-api/services_/case/listArchiveCases"

	"fmt"
	"net/http"

	"aegis-api/cache"
	update_case "aegis-api/services_/case/case_update"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CaseHandler struct {
	CaseService              CaseServiceInterface
	ListCasesService         ListCasesService
	ListActiveCasesServ      ListActiveCasesService
	auditLogger              *auditlog.AuditLogger
	ListClosedCasesService   ListClosedCasesService
	UpdateCaseService        *update_case.Service
	UserRepo                 case_assign.UserRepo // Add UserRepo here
	ListArchivedCasesService listArchiveCases.ArchiveCaseLister
	Cache                    cache.Client
}
type ListActiveCasesService interface {
	ListActiveCases(userID, tenantID, teamID string) ([]ListActiveCases.ActiveCase, error)
}

type ListClosedCasesService interface {
	ListClosedCases(userID, tenantID, teamID string) ([]ListClosedCases.ClosedCase, error)
}
type ListArchivedCasesService interface {
	ListArchivedCases(userID, tenantID, teamID string) ([]listArchiveCases.ArchivedCase, error)
}

type CaseServices struct {
	createCase *case_creation.Service
	//listCase           *ListCases.Service
	// updateCaseStatus   *case_status_update.CaseStatusService
	// getCollaborators   *get_collaborators.Service
	listCase     *ListCases.Service
	listActive   *ListActiveCases.Service
	listClosed   *ListClosedCases.Service
	listArchived *listArchiveCases.ArchiveCaseService // ✅ added here

	assignCase *case_assign.CaseAssignmentService
	// removeCollaborator *remove_user_from_case.Service
	UpdateCaseService *update_case.Service
}

func NewCaseServices(
	createCase *case_creation.Service,
	listCase *ListCases.Service,
	listActive *ListActiveCases.Service,
	assignCase *case_assign.CaseAssignmentService,
	listClosed *ListClosedCases.Service,
	listArchived *listArchiveCases.ArchiveCaseService,
	updateCaseService *update_case.Service,
) *CaseServices {
	return &CaseServices{
		createCase:        createCase,
		listCase:          listCase,
		listActive:        listActive,
		assignCase:        assignCase,
		listClosed:        listClosed,
		listArchived:      listArchived, // ✅ added here
		UpdateCaseService: updateCaseService,
	}
}

func (s *CaseServices) ListActiveCases(userID string, tenantID string, teamID string) ([]ListActiveCases.ActiveCase, error) {
	return s.listActive.ListActiveCases(userID, tenantID, teamID) // ✅ pass all args
}

func (s *CaseServices) ListClosedCases(userID string, tenantID string, teamID string) ([]ListClosedCases.ClosedCase, error) {
	return s.listClosed.ListClosedCases(userID, tenantID, teamID)
}
func (s *CaseServices) ListArchivedCases(userID string, tenantID string, teamID string) ([]listArchiveCases.ArchivedCase, error) {
	return s.listArchived.ListArchivedCases(userID, tenantID, teamID)
}

func (s *CaseServices) GetAllCases(userID string) ([]ListCases.Case, error) {
	return s.listCase.GetAllCases(userID)
}

type CaseServiceInterface interface {
	CreateCase(req *case_creation.CreateCaseRequest) (*case_creation.Case, error)

	AssignUserToCase(assignerRole string, assigneeID uuid.UUID, caseID uuid.UUID, assignerID uuid.UUID, role string, tenantID, teamID uuid.UUID) error
	ListActiveCases(userID string, tenantID string, teamID string) ([]ListActiveCases.ActiveCase, error)
	GetCaseByID(caseID string, tenantID string) (*ListCases.Case, error)
	UnassignUserFromCase(assignerID *gin.Context, assigneeID, caseID uuid.UUID) error // ← Add this
	ListClosedCases(userID string, tenantID string, teamID string) ([]ListClosedCases.ClosedCase, error)
	ListArchivedCases(userID string, tenantID string, teamID string) ([]listArchiveCases.ArchivedCase, error)
}

func NewCaseHandler(
	caseService CaseServiceInterface,
	listCasesService ListCasesService,
	listActiveCasesService ListActiveCasesService,
	listClosedCasesService ListClosedCasesService,
	listArchivedCasesService listArchiveCases.ArchiveCaseLister, // Add ListArchivedCasesService parameter
	auditLogger *auditlog.AuditLogger,
	userRepo case_assign.UserRepo, // Inject UserRepo here
	updateCaseService *update_case.Service,
	cacheClient cache.Client,
) *CaseHandler {
	return &CaseHandler{

		CaseService:              caseService,
		ListCasesService:         listCasesService,
		ListActiveCasesServ:      listActiveCasesService,
		ListClosedCasesService:   listClosedCasesService,
		ListArchivedCasesService: listArchivedCasesService, // Assign ListArchivedCasesService
		auditLogger:              auditLogger,
		UserRepo:                 userRepo, // Assign UserRepo
		UpdateCaseService:        updateCaseService,
		Cache:                    cacheClient,
	}
}

func (s *CaseServices) CreateCase(req *case_creation.CreateCaseRequest) (*case_creation.Case, error) {
	return s.createCase.CreateCase(req)
}

func (s *CaseServices) AssignUserToCase(assignerRole string, assigneeID uuid.UUID, caseID uuid.UUID, assignerID uuid.UUID, role string, tenantID, teamID uuid.UUID) error {
	return s.assignCase.AssignUserToCase(assignerRole, assigneeID, caseID, assignerID, role, tenantID, tenantID)
}

func (s *CaseServices) GetCaseByID(caseID string, tenantID string) (*ListCases.Case, error) {
	return s.listCase.GetCaseByID(caseID, tenantID)
}

func (s *CaseServices) UnassignUserFromCase(ctx *gin.Context, assigneeID, caseID uuid.UUID) error {
	return s.assignCase.UnassignUserFromCase(ctx, assigneeID, caseID)
}

func (h *CaseHandler) CreateCase(c *gin.Context) {
	var req case_creation.CreateCaseRequest

	// Grab user details from context
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	email, _ := c.Get("email") // Optional, if you have this set

	actor := auditlog.Actor{
		ID:        userID.(string),
		Role:      userRole.(string),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Email:     email.(string), // Optional, if you have this header set
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

	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing userID in context"})
		return
	}
	userID = userIDVal.(string) // if you need userID later

	tenantIDVal, exists := c.Get("tenantID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing tenantID in context"})
		return
	}
	tenantUUID, err := uuid.Parse(tenantIDVal.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID format"})
		return
	}

	req.TenantID = tenantUUID

	teamIDVal, exists := c.Get("teamID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing teamID in context"})
		return
	}
	teamUUID, err := uuid.Parse(teamIDVal.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID format"})
		return
	}
	req.TeamID = teamUUID

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
	// ✅ Invalidate caches
	cache.InvalidateTenantLists(c, h.Cache, req.TenantID.String())
	cache.InvalidateCaseHeader(c, h.Cache, req.TenantID.String(), newCase.ID.String())
	cache.InvalidateByUserLists(c, h.Cache, req.TenantID.String(), req.CreatedBy.String())
	cache.InvalidateDashboardTotals(c, h.Cache, req.TenantID.String(), userID.(string))
	cache.InvalidateEvidenceCount(c.Request.Context(), h.Cache, req.TenantID.String())

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Case created successfully",
		"case":    newCase,
	})
}
