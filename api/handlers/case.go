package handlers

import (
	"aegis-api/services/ListCases"
	"aegis-api/services/ListClosedCases"
	"aegis-api/services/case_assign"
	"aegis-api/services/case_creation"
	"aegis-api/services/case_status_update"
	"aegis-api/services/get_collaborators"
	"aegis-api/services/remove_user_from_case"
	"aegis-api/structs"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"strings"
)

type CaseServices struct {
	createCase         *case_creation.Service
	getCase            *ListCases.Service
	updateCaseStatus   *case_status_update.CaseStatusService
	getCollaborators   *get_collaborators.Service
	assignCase         *case_assign.CaseAssignmentService
	removeCollaborator *remove_user_from_case.Service
	closedCasesByUser  ListClosedCases.ClosedCaseRepository
}

func NewCaseServices(
	createCase *case_creation.Service,
	getCase *ListCases.Service,
	updateCaseStatus *case_status_update.CaseStatusService,
	getCollaborators *get_collaborators.Service,
	assignCase *case_assign.CaseAssignmentService,
	removeCollaborator *remove_user_from_case.Service,
	// closedCasesByUser ListClosedCases.ClosedCaseRepository, //active and closed cases functions to be removed
) *CaseServices {
	return &CaseServices{
		createCase:         createCase,
		getCase:            getCase,
		updateCaseStatus:   updateCaseStatus,
		getCollaborators:   getCollaborators,
		assignCase:         assignCase,
		removeCollaborator: removeCollaborator,
		//closedCasesByUser:  closedCasesByUser,
	}
}

// @Summary Create a new case
// @Description Creates a new case with the provided details. Only users with 'Admin' role can perform this action.
// @Tags Cases
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body case_creation.CreateCaseRequest true "Case Creation Request"
// @Success 201 {object} structs.SuccessResponse{data=case_creation.Case} "Case created successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid request payload"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/cases [post]
func (cs CaseServices) CreateCase(c *gin.Context) {
	var req case_creation.CreateCaseRequest //
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid case data",
			Details: err.Error(),
		})
		return
	}

	creatorID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
			Error:   "unauthorized",
			Message: "No authentication provided",
		})
		return
	}

	req.CreatedBy = creatorID.(string) //use userID from middleware

	newCase, err := cs.createCase.CreateCase(req)
	if err != nil {
		status := http.StatusInternalServerError
		errorType := "creation_failed"
		if strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "invalid") {
			status = http.StatusBadRequest
			errorType = "validation_failed"
		}
		c.JSON(status, structs.ErrorResponse{
			Error:   errorType,
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

// @Summary Get all cases
// @Description Retrieves all cases without any filtering
// @Tags Cases
// @Accept json
// @Produce json
// @Success 200 {object} structs.SuccessResponse{data=[]case_creation.Case} "Cases retrieved successfully"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/cases [get]
func (cs CaseServices) GetAllCases(c *gin.Context) {
	//admin-only privilege
	cases, err := cs.getCase.GetAllCases() //no filtering
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
func (cs CaseServices) GetFilteredCases(c *gin.Context) {
	// Extract query parameters
	status := c.Query("status")
	priority := c.Query("priority")
	createdBy := c.Query("created_by")
	teamName := c.Query("team_name")
	titleTerm := c.Query("title")
	sortBy := c.Query("sort_by")
	order := c.Query("order")

	cases, err := cs.getCase.GetFilteredCases(
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

// @Summary Get cases by user id
// @Description Retrieves cases created a user is assigned to
// @Tags Cases
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} structs.SuccessResponse{data=[]case_creation.Case} "Cases retrieved successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid user ID"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/cases/user/{user_id} [get]
func (cs CaseServices) GetCasesByUserID(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "User ID is required",
		})
		return
	}

	cases, err := cs.getCase.GetCasesByUser(userID)
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

// @Summary Update case status
// @Description Updates the status of a case
// @Tags Cases
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body case_status_update.UpdateCaseStatusRequest true "Status Update Request"
// @Success 200 {object} structs.SuccessResponse "Case status updated successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid request payload"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/cases/{id}/status [put]
func (cs CaseServices) UpdateCaseStatus(c *gin.Context) {
	caseID := c.Param("id")
	if caseID == "" {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Case ID is required",
		})
		return
	}

	var req case_status_update.UpdateCaseStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid status update data",
			Details: err.Error(),
		})
		return
	}

	req.CaseID = caseID //set case ID from URL parameter

	err := cs.updateCaseStatus.UpdateCaseStatus(req, "Admin") //hardcoded since the middleware checks for admin role
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

// @Summary Assign user to case
// @Description Assigns a user to a case with a specific role (requires admin privileges)
// @Tags Cases
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Case ID"
// @Param request body structs.AssignCaseRequest true "Assignment Request"
// @Success 200 {object} structs.SuccessResponse "User assigned successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid request payload"
// @Failure 401 {object} structs.ErrorResponse "Unauthorized - Authentication required"
// @Failure 403 {object} structs.ErrorResponse "Forbidden - Admin privileges required"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/cases/{id}/collaborators [post]
func (cs CaseServices) CreateCollaborator(c *gin.Context) {
	caseID := c.Param("id")
	var req structs.AssignCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid assignment data",
			Details: err.Error(),
		})
		return
	}

	assignerID, exists := c.Get("userID") //should be set by middleware
	if !exists {
		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
			Error:   "unauthorized",
			Message: "No authentication provided",
		})
		return
	}

	// Parse UUIDs
	caseUUID, err := uuid.Parse(caseID) //from url
	if err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_case_id",
			Message: "Invalid case ID format",
		})
		return
	}

	assignerUUID, err := uuid.Parse(assignerID.(string)) //from middleware
	if err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_assigner_id",
			Message: "Invalid assigner ID format",
		})
		return
	}

	assigneeUUID, err := uuid.Parse(req.UserID) //from request body
	if err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
		})
		return
	}

	err = cs.assignCase.AssignUserToCase(assignerUUID, assigneeUUID, caseUUID, req.Role) //create request struct COME BACK TO THIS
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "forbidden: admin privileges required" { //middleware already checks this - could remove
			status = http.StatusForbidden
		}
		c.JSON(status, structs.ErrorResponse{
			Error:   "assignment_failed",
			Message: "Could not assign user to case",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, structs.SuccessResponse{
		Success: true,
		Message: "User assigned to case successfully",
	})
}

// @Summary Get case collaborators
// @Description Retrieves all collaborators (users with roles) for a specific case
// @Tags Cases
// @Accept json
// @Produce json
// @Param id path string true "Case ID"
// @Success 200 {object} structs.SuccessResponse{data=[]get_collaborators.Collaborator} "Collaborators retrieved successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid case ID"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/cases/{id}/collaborators [get]
func (cs CaseServices) GetCollaborators(c *gin.Context) {
	caseID := c.Param("id")
	if caseID == "" {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Case ID is required",
		})
		return
	}

	caseUUID, err := uuid.Parse(caseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_case_id",
			Message: "Invalid case ID format",
		})
		return
	}

	collaborators, err := cs.getCollaborators.GetCollaborators(caseUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "fetch_failed",
			Message: "Could not fetch collaborators",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Collaborators retrieved successfully",
		Data:    collaborators,
	})
}

// @Summary Unassign user from case
// @Description Removes a user from a case (requires admin privileges)
// @Tags Cases
// @Accept json
// @Produce json
// @Param id path string true "Case ID"
// @Param user_id path string true "User ID"
// @Success 200 {object} structs.SuccessResponse "User successfully removed from case"
// @Failure 400 {object} structs.ErrorResponse "Invalid request payload"
// @Failure 403 {object} structs.ErrorResponse "Forbidden - Admin privileges required"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/cases/{id}/collaborators/{user_id} [delete]
func (cs CaseServices) RemoveCollaborator(c *gin.Context) {
	caseID := c.Param("id")
	caseUUID, err := uuid.Parse(caseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_case_id",
			Message: "Invalid case ID format",
		})
		return
	}

	userID := c.Param("user_id")
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
		})
		return
	}

	adminID, exists := c.Get("userID") // from middleware
	if !exists {
		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
			Error:   "unauthorized",
			Message: "No authentication provided",
		})
		return
	}

	adminUUID, err := uuid.Parse(adminID.(string)) // from middleware

	if err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_admin_id",
			Message: "Invalid admin ID format",
		})
		return
	}

	req := remove_user_from_case.RemoveUserRequest{
		CaseID:  caseUUID,
		UserID:  userUUID,
		AdminID: adminUUID,
	}

	err = cs.removeCollaborator.RemoveUser(req)
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
		Message: "User successfully removed from case",
	})
}

/*
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
func (cs CaseServices) GetClosedCasesByUserID(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "User ID is required",
		})
		return
	}

	cases, err := cs.closedCasesByUser.GetClosedCasesByUserID(c.Request.Context(), userID)
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
}*/
