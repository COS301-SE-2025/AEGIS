package handlers

import (
	"aegis-api/services/ListCases"
	"aegis-api/services/ListClosedCases"
	"aegis-api/services/case_assign"
	"aegis-api/services/case_creation"
	"aegis-api/services/case_status_update"
	"aegis-api/services/get_collaborators"
	"aegis-api/structs"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CaseServices struct {
	createCase         *case_creation.Service
	getCase            *ListCases.Service
	updateCaseStatus   *case_status_update.CaseStatusService
	getCollaborators   *get_collaborators.Service
	assignCase         *case_assign.CaseAssignmentService
	removeCollaborator *case_assign.CaseAssignmentService
	closedCasesByUser  ListClosedCases.ClosedCaseRepository
}

func NewCaseServices(
	createCase *case_creation.Service,
	getCase *ListCases.Service,
	updateCaseStatus *case_status_update.CaseStatusService,
	getCollaborators *get_collaborators.Service,
	assignCase *case_assign.CaseAssignmentService,
	removeCollaborator *case_assign.CaseAssignmentService,
	closedCasesByUser ListClosedCases.ClosedCaseRepository,
) *CaseServices {
	return &CaseServices{
		createCase:         createCase,
		getCase:            getCase,
		updateCaseStatus:   updateCaseStatus,
		getCollaborators:   getCollaborators,
		assignCase:         assignCase,
		removeCollaborator: removeCollaborator,
		closedCasesByUser:  closedCasesByUser,
	}
}

// @Summary Get all cases
// @Description Retrieves all cases without any filtering
// @Tags Cases
// @Accept json
// @Produce json
// @Success 200 {object} structs.SuccessResponse{data=[]case_creation.Case} "Cases retrieved successfully"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/cases [get]
func (m CaseServices) GetAllCases(c *gin.Context) {
	//admin-only privilege
	cases, err := m.getCase.GetAllCases()
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "fetch_failed",
			Message: "Could not fetch cases",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Cases retrieved successfully",
		Data:    cases,
	})
}

// @Summary Get cases by user
// @Description Retrieves cases created by a specific user
// @Tags Cases
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} structs.SuccessResponse{data=[]case_creation.Case} "Cases retrieved successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid user ID"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/cases/user/{user_id} [get]
func (m CaseServices) GetCasesByUser(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "User ID is required",
		})
		return
	}

	cases, err := m.getCase.GetCasesByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "fetch_failed",
			Message: "Could not fetch user cases",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "User cases retrieved successfully",
		Data:    cases,
	})
}

// @Summary Get filtered cases
// @Description Retrieves cases based on multiple filter criteria
// @Tags Cases
// @Accept json
// @Produce json
// @Param status query string false "Case status"
// @Param priority query string false "Case priority"
// @Param created_by query string false "Creator's user ID"
// @Param team_name query string false "Team name"
// @Param title query string false "Title search term"
// @Param sort_by query string false "Field to sort by"
// @Param order query string false "Sort order (asc/desc)"
// @Success 200 {object} structs.SuccessResponse{data=[]case_creation.Case} "Cases retrieved successfully"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/cases/filter [get]
func (m CaseServices) GetFilteredCases(c *gin.Context) {
	// Extract query parameters
	status := c.Query("status")
	priority := c.Query("priority")
	createdBy := c.Query("created_by")
	teamName := c.Query("team_name")
	titleTerm := c.Query("title")
	sortBy := c.Query("sort_by")
	order := c.Query("order")

	cases, err := m.getCase.GetFilteredCases(
		status,
		priority,
		createdBy,
		teamName,
		titleTerm,
		sortBy,
		order,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "fetch_failed",
			Message: "Could not fetch filtered cases",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Filtered cases retrieved successfully",
		Data:    cases,
	})
}

// @Summary Create a new case
// @Description Creates a new case with the provided details
// @Tags Cases
// @Accept json
// @Produce json
// @Param request body case_creation.CreateCaseRequest true "Case Creation Request"
// @Success 201 {object} structs.SuccessResponse{data=case_creation.Case} "Case created successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid request payload"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/cases [post]
func (m CaseServices) CreateCase(c *gin.Context) {
	var req case_creation.CreateCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid case data",
			Details: err.Error(),
		})
		return
	}

	newCase, err := m.createCase.CreateCase(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "creation_failed",
			Message: "Could not create case",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, structs.SuccessResponse{
		Success: true,
		Message: "Case created successfully",
		Data:    newCase,
	})
}

// @Summary Assign user to case
// @Description Assigns a user to a case with a specific role (requires admin privileges)
// @Tags Cases
// @Accept json
// @Produce json
// @Param case_id path string true "Case ID"
// @Param request body structs.AssignCaseRequest true "Assignment Request"
// @Success 200 {object} structs.SuccessResponse "User assigned successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid request payload"
// @Failure 403 {object} structs.ErrorResponse "Forbidden - Admin privileges required"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/cases/{case_id}/assign [post]
func (m CaseServices) CreateCollaborator(c *gin.Context) {
	caseID := c.Param("case_id")
	var req structs.AssignCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid assignment data",
			Details: err.Error(),
		})
		return
	}

	// Get assigner ID from context (assuming it's set by auth middleware)
	assignerID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
			Error:   "unauthorized",
			Message: "No authentication provided",
		})
		return
	}

	// Parse UUIDs
	caseUUID, err := uuid.Parse(caseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_case_id",
			Message: "Invalid case ID format",
		})
		return
	}

	assignerUUID, err := uuid.Parse(assignerID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_assigner_id",
			Message: "Invalid assigner ID format",
		})
		return
	}

	assigneeUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
		})
		return
	}

	err = m.assignCase.AssignUserToCase(assignerUUID, assigneeUUID, caseUUID, req.Role)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "forbidden: admin privileges required" {
			status = http.StatusForbidden
		}
		c.JSON(status, structs.ErrorResponse{
			Error:   "assignment_failed",
			Message: "Could not assign user to case",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "User assigned to case successfully",
	})
}

// @Summary Unassign user from case
// @Description Removes a user from a case (requires admin privileges)
// @Tags Cases
// @Accept json
// @Produce json
// @Param case_id path string true "Case ID"
// @Param user_id path string true "User ID"
// @Success 200 {object} structs.SuccessResponse "User unassigned successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid request payload"
// @Failure 403 {object} structs.ErrorResponse "Forbidden - Admin privileges required"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/cases/{case_id}/users/{user_id} [delete]
func (m CaseServices) RemoveCollaborator(c *gin.Context) {
	caseID := c.Param("case_id")
	userID := c.Param("user_id")

	// Get assigner ID from context
	assignerID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
			Error:   "unauthorized",
			Message: "No authentication provided",
		})
		return
	}

	// Parse UUIDs
	caseUUID, err := uuid.Parse(caseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_case_id",
			Message: "Invalid case ID format",
		})
		return
	}

	assignerUUID, err := uuid.Parse(assignerID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_assigner_id",
			Message: "Invalid assigner ID format",
		})
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
		})
		return
	}

	err = m.removeCollaborator.UnassignUserFromCase(assignerUUID, userUUID, caseUUID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "forbidden: admin privileges required" {
			status = http.StatusForbidden
		}
		c.JSON(status, structs.ErrorResponse{
			Error:   "unassignment_failed",
			Message: "Could not unassign user from case",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "User unassigned from case successfully",
	})
}

// @Summary Update case status
// @Description Updates the status of a case
// @Tags Cases
// @Accept json
// @Produce json
// @Param case_id path string true "Case ID"
// @Param request body structs.UpdateCaseStatusRequest true "Status Update Request"
// @Success 200 {object} structs.SuccessResponse "Case status updated successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid request payload"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/cases/{case_id}/status [put]
func (m CaseServices) UpdateCaseStatus(c *gin.Context) {
	var req case_status_update.UpdateCaseStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid status update data",
			Details: err.Error(),
		})
		return
	}

	err := m.updateCaseStatus.UpdateCaseStatus(req, "Admin") //cant hardcode COME BACK
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "update_failed",
			Message: "Could not update case status",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Case status updated successfully",
	})
}

// @Summary Get closed cases by user
// @Description Retrieves all closed cases associated with a specific user
// @Tags Cases
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} structs.SuccessResponse{data=[]case_creation.Case} "Closed cases retrieved successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid user ID"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/cases/closed/{user_id} [get]
func (m CaseServices) GetClosedCasesByUserID(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "User ID is required",
		})
		return
	}

	cases, err := m.closedCasesByUser.GetClosedCasesByUserID(c.Request.Context(), userID) //why is the context needed in this function if it is not being used by it?
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "fetch_failed",
			Message: "Could not fetch closed cases",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Closed cases retrieved successfully",
		Data:    cases,
	})
}
