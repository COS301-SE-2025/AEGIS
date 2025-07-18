package handlers

// import (
// 	"aegis-api/middleware"
// 	"aegis-api/services/case_status_update"
// 	"aegis-api/services/get_collaborators"
// 	"aegis-api/services/remove_user_from_case"
// 	"aegis-api/services_/case/ListCases"
// 	"aegis-api/services_/case/case_assign"
// 	"aegis-api/services_/case/case_creation"
// 	"aegis-api/structs"
// 	"net/http"
// 	"strings"

// 	"github.com/gin-gonic/gin"
// 	"github.com/google/uuid"
// )

// type CaseServices struct {
// 	createCase         *case_creation.Service
// 	listCase           *ListCases.Service
// 	updateCaseStatus   *case_status_update.CaseStatusService
// 	getCollaborators   *get_collaborators.Service
// 	assignCase         *case_assign.CaseAssignmentService
// 	removeCollaborator *remove_user_from_case.Service
// }

// func NewCaseServices(
// 	createCase *case_creation.Service,
// 	listCase *ListCases.Service,
// 	updateCaseStatus *case_status_update.CaseStatusService,
// 	getCollaborators *get_collaborators.Service,
// 	assignCase *case_assign.CaseAssignmentService,
// 	removeCollaborator *remove_user_from_case.Service,
// ) *CaseServices {
// 	return &CaseServices{
// 		createCase:         createCase,
// 		listCase:           listCase,
// 		updateCaseStatus:   updateCaseStatus,
// 		getCollaborators:   getCollaborators,
// 		assignCase:         assignCase,
// 		removeCollaborator: removeCollaborator,
// 	}
// }

// @Summary Create a new case
// @Description Creates a new case with the provided details. Only users with 'Admin' role can perform this action.
// @Tags Cases
// @Accept json
// @Accept x-www-form-urlencoded
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param request body structs.CreateCaseRequest true "Case Creation Request"
// @Success 201 {object} structs.SuccessResponse{data=case_creation.Case} "Case created successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid request payload"
// @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// @Failure 403 {object} structs.ErrorResponse "Forbidden - insufficient permissions"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/cases [post]

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
// 	newCase, err := cs.createCase.CreateCase(serviceReq)
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

// // handlers/case.go

// func (s *CaseServices) AssignCase(c *gin.Context) {
// 	// You can implement logic later or return a stub response
// 	c.JSON(501, gin.H{"message": "AssignCase not implemented"})
// }

// func (s *CaseServices) GetCaseByID(c *gin.Context) {
// 	c.JSON(501, gin.H{"message": "GetCaseByID not implemented"})
// }

// // @Summary List all cases
// // @Description Retrieves all cases without any filtering. Accessible by all authenticated users.
// // @Tags Cases
// // @Accept json
// // @Produce json
// // @Security ApiKeyAuth
// // @Success 200 {object} structs.SuccessResponse{data=[]case_creation.Case} "Cases retrieved successfully"
// // @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// // @Failure 500 {object} structs.ErrorResponse "Internal server error"
// // @Router /api/v1/cases [get]
// func (cs CaseServices) ListAllCases(c *gin.Context) {
// 	//might be more accurate to call it getlistofcases
// 	//admin-only privilege
// 	cases, err := cs.listCase.GetAllCases() //no filtering
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 			Error:   "fetch_failed",
// 			Message: "Could not fetch cases",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, structs.SuccessResponse{
// 		Success: true,
// 		Message: "Cases retrieved successfully",
// 		Data:    cases,
// 	})
// }

// // @Summary List filtered cases
// // @Description Retrieves cases based on multiple filter criteria. Only administrators can access this endpoint.
// // @Tags Cases
// // @Accept json
// // @Produce json
// // @Security ApiKeyAuth
// // @Param status query string false "Case status"
// // @Param priority query string false "Case priority"
// // @Param created_by query string false "Creator's user ID"
// // @Param team_name query string false "Team name"
// // @Param title query string false "Title search term"
// // @Param sort_by query string false "Field to sort by"
// // @Param order query string false "Sort order (asc/desc)"
// // @Success 200 {object} structs.SuccessResponse{data=[]case_creation.Case} "Cases retrieved successfully"
// // @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// // @Failure 403 {object} structs.ErrorResponse "Forbidden - insufficient permissions"
// // @Failure 500 {object} structs.ErrorResponse "Internal server error"
// // @Router /api/v1/cases/filter [get]
// func (cs CaseServices) ListFilteredCases(c *gin.Context) {
// 	// Extract query parameters
// 	status := c.Query("status")
// 	priority := c.Query("priority")
// 	createdBy := c.Query("created_by")
// 	teamName := c.Query("team_name")
// 	titleTerm := c.Query("title")
// 	sortBy := c.Query("sort_by")
// 	order := c.Query("order")

// 	cases, err := cs.listCase.GetFilteredCases(
// 		status,
// 		priority,
// 		createdBy,
// 		teamName,
// 		titleTerm,
// 		sortBy,
// 		order,
// 	)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 			Error:   "fetch_failed",
// 			Message: "Could not fetch filtered cases",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, structs.SuccessResponse{
// 		Success: true,
// 		Message: "Filtered cases retrieved successfully",
// 		Data:    cases,
// 	})
// }

// // @Summary Get cases by user ID
// // @Description Retrieves cases associated with a user. Regular users can only access their own cases, while admins can access any user's cases.
// // @Tags Cases, Users, Admin
// // @Accept json
// // @Produce json
// // @Security ApiKeyAuth
// // @Param user_id path string false "User ID (required for admin route, automatically determined for user route)"
// // @Success 200 {object} structs.SuccessResponse{data=[]case_creation.Case} "Cases retrieved successfully"
// // @Failure 400 {object} structs.ErrorResponse "Invalid user ID"
// // @Failure 401 {object} structs.ErrorResponse "Unauthorized - Authentication required"
// // @Failure 403 {object} structs.ErrorResponse "Forbidden - Insufficient permissions"
// // @Failure 500 {object} structs.ErrorResponse "Internal server error"
// // @Router /api/v1/users/cases [get]
// // @Router /api/v1/admin/users/{user_id}/cases [get]
// func (cs CaseServices) ListCasesByUserID(c *gin.Context) {
// 	userID, ok := middleware.GetTargetUserID(c)
// 	if !ok {
// 		return
// 	}

// 	cases, err := cs.listCase.GetCasesByUser(userID)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 			Error:   "fetch_failed",
// 			Message: "Could not fetch user cases",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, structs.SuccessResponse{
// 		Success: true,
// 		Message: "User cases retrieved successfully",
// 		Data:    cases,
// 	})
// }

// // @Summary Update case status
// // @Description Updates the status of a case. Only users with 'Admin' role can perform this action.
// // @Tags Cases
// // @Accept json
// // @Accept x-www-form-urlencoded
// // @Accept multipart/form-data
// // @Produce json
// // @Security ApiKeyAuth
// // @Param case_id path string true "Case ID"
// // @Param request body structs.UpdateCaseStatusRequest true "Status Update Request"
// // @Success 200 {object} structs.SuccessResponse "Case status updated successfully"
// // @Failure 400 {object} structs.ErrorResponse "Invalid request payload"
// // @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// // @Failure 403 {object} structs.ErrorResponse "Forbidden - insufficient permissions"
// // @Failure 500 {object} structs.ErrorResponse "Internal server error"
// // @Router /api/v1/cases/{case_id}/status [put]
// func (cs CaseServices) UpdateCaseStatus(c *gin.Context) {
// 	caseID := c.Param("case_id")
// 	if caseID == "" {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: "Case ID is required",
// 		})
// 		return
// 	}

// 	var apiReq structs.UpdateCaseStatusRequest
// 	if err := c.ShouldBind(&apiReq); err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: "Invalid status update data",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	serviceReq := case_status_update.UpdateCaseStatusRequest{
// 		CaseID: caseID,
// 		Status: apiReq.Status,
// 	}

// 	err := cs.updateCaseStatus.UpdateCaseStatus(serviceReq, "Admin") //hardcoded since the middleware checks for admin role
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 			Error:   "update_failed",
// 			Message: "Could not update case status",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, structs.SuccessResponse{
// 		Success: true,
// 		Message: "Case status updated successfully",
// 	})
// }

// // @Summary Assign user to case
// // @Description Assigns a user to a case with a specific role. Only users with 'Admin' role can perform this action.
// // @Tags Cases
// // @Accept json
// // @Accept x-www-form-urlencoded
// // @Accept multipart/form-data
// // @Produce json
// // @Security ApiKeyAuth
// // @Param case_id path string true "Case ID"
// // @Param request body structs.AssignCaseRequest true "Assignment Request"
// // @Success 201 {object} structs.SuccessResponse "User assigned successfully"
// // @Failure 400 {object} structs.ErrorResponse "Invalid request payload"
// // @Failure 401 {object} structs.ErrorResponse "Unauthorized - Authentication required"
// // @Failure 403 {object} structs.ErrorResponse "Forbidden - Admin privileges required"
// // @Failure 500 {object} structs.ErrorResponse "Internal server error"
// // @Router /api/v1/cases/{case_id}/collaborators [post]
// func (cs CaseServices) CreateCollaborator(c *gin.Context) {
// 	caseID := c.Param("case_id")
// 	var req structs.AssignCaseRequest
// 	if err := c.ShouldBind(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: "Invalid assignment data",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	assignerID, exists := c.Get("userID") //should be set by middleware
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
// 			Error:   "unauthorized",
// 			Message: "No authentication provided",
// 		})
// 		return
// 	}

// 	// Parse UUIDs
// 	caseUUID, err := uuid.Parse(caseID) //from url
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_case_id",
// 			Message: "Invalid case ID format",
// 		})
// 		return
// 	}

// 	assignerUUID, err := uuid.Parse(assignerID.(string)) //from middleware
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_assigner_id",
// 			Message: "Invalid assigner ID format",
// 		})
// 		return
// 	}

// 	assigneeUUID, err := uuid.Parse(req.UserID) //from request body
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_user_id",
// 			Message: "Invalid user ID format",
// 		})
// 		return
// 	}

// 	err = cs.assignCase.AssignUserToCase(assignerUUID, assigneeUUID, caseUUID, req.Role) //create request struct COME BACK TO THIS
// 	if err != nil {
// 		status := http.StatusInternalServerError
// 		if err.Error() == "forbidden: admin privileges required" { //middleware already checks this - could remove
// 			status = http.StatusForbidden
// 		}
// 		c.JSON(status, structs.ErrorResponse{
// 			Error:   "assignment_failed",
// 			Message: "Could not assign user to case",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusCreated, structs.SuccessResponse{
// 		Success: true,
// 		Message: "User assigned to case successfully",
// 	})
// }

// // @Summary Get case collaborators
// // @Description Retrieves all collaborators (users with roles) for a specific case. Requires authentication.
// // @Tags Cases
// // @Accept json
// // @Produce json
// // @Security ApiKeyAuth
// // @Param case_id path string true "Case ID"
// // @Success 200 {object} structs.SuccessResponse{data=[]get_collaborators.Collaborator} "Collaborators retrieved successfully"
// // @Failure 400 {object} structs.ErrorResponse "Invalid case ID"
// // @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// // @Failure 500 {object} structs.ErrorResponse "Internal server error"
// // @Router /api/v1/cases/{case_id}/collaborators [get]
// func (cs CaseServices) ListCollaborators(c *gin.Context) {
// 	caseID := c.Param("case_id")
// 	if caseID == "" {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: "Case ID is required",
// 		})
// 		return
// 	}

// 	caseUUID, err := uuid.Parse(caseID)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_case_id",
// 			Message: "Invalid case ID format",
// 		})
// 		return
// 	}

// 	collaborators, err := cs.getCollaborators.GetCollaborators(caseUUID)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 			Error:   "fetch_failed",
// 			Message: "Could not fetch collaborators",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, structs.SuccessResponse{
// 		Success: true,
// 		Message: "Collaborators retrieved successfully",
// 		Data:    collaborators,
// 	})
// }

// // @Summary Unassign user from case
// // @Description Removes a user from a case. Only users with 'Admin' role can perform this action.
// // @Tags Cases
// // @Accept json
// // @Produce json
// // @Security ApiKeyAuth
// // @Param case_id path string true "Case ID"
// // @Param user_id path string true "User ID"
// // @Success 200 {object} structs.SuccessResponse "User successfully removed from case"
// // @Failure 400 {object} structs.ErrorResponse "Invalid request payload"
// // @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// // @Failure 403 {object} structs.ErrorResponse "Forbidden - Admin privileges required"
// // @Failure 500 {object} structs.ErrorResponse "Internal server error"
// // @Router /api/v1/cases/{case_id}/collaborators/{user_id} [delete]
// func (cs CaseServices) RemoveCollaborator(c *gin.Context) {
// 	caseID := c.Param("case_id")
// 	caseUUID, err := uuid.Parse(caseID)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_case_id",
// 			Message: "Invalid case ID format",
// 		})
// 		return
// 	}

// 	userID := c.Param("user_id")
// 	userUUID, err := uuid.Parse(userID)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_user_id",
// 			Message: "Invalid user ID format",
// 		})
// 		return
// 	}

// 	adminID, exists := c.Get("userID") // from middleware
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
// 			Error:   "unauthorized",
// 			Message: "No authentication provided",
// 		})
// 		return
// 	}

// 	adminUUID, err := uuid.Parse(adminID.(string)) // from middleware

// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_admin_id",
// 			Message: "Invalid admin ID format",
// 		})
// 		return
// 	}

// 	req := remove_user_from_case.RemoveUserRequest{
// 		CaseID:  caseUUID,
// 		UserID:  userUUID,
// 		AdminID: adminUUID,
// 	}

// 	err = cs.removeCollaborator.RemoveUser(req)
// 	if err != nil {
// 		status := http.StatusInternalServerError
// 		if err.Error() == "forbidden: admin privileges required" {
// 			status = http.StatusForbidden
// 		}
// 		c.JSON(status, structs.ErrorResponse{
// 			Error:   "unassignment_failed",
// 			Message: "Could not unassign user from case",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, structs.SuccessResponse{
// 		Success: true,
// 		Message: "User successfully removed from case",
// 	})
// }
