package handlers

import (
	"aegis-core/services"
	//"strconv"

	//"aegis-core/services"
	"aegis-core/structs"
	"github.com/gin-gonic/gin"
	"net/http"
	//"net/http"
)

//mock services

type AdminServiceInterface interface {
	RegisterUser(c *gin.Context)
	ListUsers(c *gin.Context)
	GetUserActivity(c *gin.Context)
	UpdateUserRole(c *gin.Context)
	DeleteUser(c *gin.Context)
	GetRoles(c *gin.Context)
}

type AuthServiceInterface interface {
	Login(c *gin.Context)
	Logout(c *gin.Context)
	ResetPassword(c *gin.Context)
}

type CaseServiceInterface interface {
	GetCases(c *gin.Context)
	CreateCase(c *gin.Context)
	GetCase(c *gin.Context)
	UpdateCase(c *gin.Context)
	AssignCase(c *gin.Context)
	GetCollaborators(c *gin.Context)
	CreateCollaborator(c *gin.Context)
	RemoveCollaborator(c *gin.Context)
}

type EvidenceServiceInterface interface {
	GetEvidence(c *gin.Context)
	UploadEvidence(c *gin.Context)
	GetEvidenceItem(c *gin.Context)
	PreviewEvidence(c *gin.Context)
}

type UserServiceInterface interface {
	GetUserInfo(c *gin.Context)
	UpdateUserInfo(c *gin.Context)
	GetUserCases(c *gin.Context)
}

type Handler struct {
	AdminService    AdminServiceInterface
	AuthService     AuthServiceInterface
	CaseService     CaseServiceInterface
	EvidenceService EvidenceServiceInterface
	UserService     UserServiceInterface
}

type MockAdminService struct {
	adminService services.AdminService
}

func (m MockAdminService) RegisterUser(c *gin.Context) {
	//struct to hold user data
	//binding and validation
	var req structs.RegisterUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid input",
			Details: err.Error(),
		})
		return
	}

	//call the service function
	user, err := m.adminService.Register(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "registration_failed",
			Message: "Could not register user",
			Details: err.Error(),
		})
		return
	}

	//http response
	c.JSON(http.StatusCreated, structs.SuccessResponse{
		Success: true,
		Message: "User registered successfully",
		Data:    user,
	})
}

func (m MockAdminService) ListUsers(c *gin.Context) {
	//binding and validation
	var req structs.UserFilter
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid query parameters",
			Details: err.Error(),
		})
		return
	}

	//call the service function
	users, err := m.adminService.ListUsers(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "list_users_failed",
			Message: "Could not retrieve users",
			Details: err.Error(),
		})
		return
	}

	//http response
	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Users retrieved successfully",
		Data:    users,
	})
}

func (m MockAdminService) GetUserActivity(c *gin.Context) {
	// Get user ID from URL parameter
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "User ID is required",
		})
		return
	}

	// Get query parameters for filtering
	//timeRange := c.Query("time_range")
	//activityType := c.Query("activity_type")

	activities, err := m.adminService.GetUserActivity(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "activity_fetch_failed",
			Message: "Could not fetch user activity",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "User activity retrieved successfully",
		Data:    activities,
	})
}

func (m MockAdminService) UpdateUserRole(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "User ID is required",
		})
		return
	}

	var req structs.UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid role data",
			Details: err.Error(),
		})
		return
	}

	err := m.adminService.UpdateUserRole(userID, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "role_update_failed",
			Message: "Could not update user role",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "User role updated successfully",
	})
}

func (m MockAdminService) DeleteUser(c *gin.Context) {
	// Get user ID from URL parameter
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "User ID is required",
		})
		return
	}

	err := m.adminService.DeleteUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "deletion_failed",
			Message: "Could not delete user",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "User deleted successfully",
	})
}

func (m MockAdminService) GetRoles(c *gin.Context) {
	roles, err := m.adminService.GetRoles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "roles_fetch_failed",
			Message: "Could not fetch roles",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Roles retrieved successfully",
		Data:    roles, //arr
	})
}

type MockAuthService struct {
	authService services.AuthService
}

func (m MockAuthService) Login(c *gin.Context) {
	var req structs.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid credentials",
			Details: err.Error(),
		})
		return
	}

	response, err := m.authService.Login(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
			Error:   "authentication_failed",
			Message: "Invalid credentials",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Login successful",
		Data:    response,
	})
}

func (m MockAuthService) Logout(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	err := m.authService.Logout(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "logout_failed",
			Message: "Could not log out user",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Logged out successfully",
	})
}

func (m MockAuthService) ResetPassword(c *gin.Context) {
	var req structs.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid reset password data",
			Details: err.Error(),
		})
		return
	}

	err := m.authService.ResetPassword(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "reset_failed",
			Message: "Could not reset password",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Password reset email sent successfully",
	})
}

type MockCaseService struct {
	caseService services.CaseService
}

func (m MockCaseService) GetCases(c *gin.Context) {
	var filter structs.CaseFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid query parameters",
			Details: err.Error(),
		})
		return
	}

	//if filter.Page == "" {
	//	filter.Page = "1"
	//}
	//if filter.PageSize == "" {
	//	filter.PageSize = "10"
	//}
	//if page, err := strconv.Atoi(filter.Page); err != nil || page < 1 {
	//	c.JSON(http.StatusBadRequest, structs.ErrorResponse{
	//		Error:   "invalid_request",
	//		Message: "Invalid page number",
	//	})
	//	return
	//}
	//
	//if pageSize, err := strconv.Atoi(filter.PageSize); err != nil || pageSize < 1 || pageSize > 100 {
	//	c.JSON(http.StatusBadRequest, structs.ErrorResponse{
	//		Error:   "invalid_request",
	//		Message: "Invalid page size (must be between 1 and 100)",
	//	})
	//	return
	//}

	cases, err := m.caseService.GetCases(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "cases_fetch_failed",
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

func (m MockCaseService) CreateCase(c *gin.Context) {
	var req structs.CreateCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid case data",
			Details: err.Error(),
		})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	case_, err := m.caseService.CreateCase(userID.(string), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "case_creation_failed",
			Message: "Could not create case",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, structs.SuccessResponse{
		Success: true,
		Message: "Case created successfully",
		Data:    case_,
	})
}

func (m MockCaseService) GetCase(c *gin.Context) {
	caseID := c.Param("id")
	if caseID == "" {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Case ID is required",
		})
		return
	}

	case_, err := m.caseService.GetCase(caseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "case_fetch_failed",
			Message: "Could not fetch case",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Case retrieved successfully",
		Data:    case_,
	})
}

func (m MockCaseService) UpdateCase(c *gin.Context) {
	caseID := c.Param("id")
	if caseID == "" {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Case ID is required",
		})
		return
	}

	var req structs.UpdateCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid case data",
			Details: err.Error(),
		})
		return
	}

	err := m.caseService.UpdateCase(caseID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "update_failed",
			Message: "Could not update case",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Case updated successfully",
	})
}

func (m MockCaseService) AssignCase(c *gin.Context) {
	caseID := c.Param("id")
	if caseID == "" {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Case ID is required",
		})
		return
	}

	var req structs.AssignCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid assignment data",
			Details: err.Error(),
		})
		return
	}

	err := m.caseService.AssignCase(caseID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "assignment_failed",
			Message: "Could not assign case",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Case assigned successfully",
	})
}

func (m MockCaseService) GetCollaborators(c *gin.Context) {
	caseID := c.Param("id")
	if caseID == "" {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Case ID is required",
		})
		return
	}

	collaborators, err := m.caseService.GetCollaborators(caseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "collaborators_fetch_failed",
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

func (m MockCaseService) CreateCollaborator(c *gin.Context) {
	caseID := c.Param("id")
	if caseID == "" {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Case ID is required",
		})
		return
	}

	var req structs.User
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid collaborator data",
			Details: err.Error(),
		})
		return
	}

	err := m.caseService.AddCollaborator(caseID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "collaborator_creation_failed",
			Message: "Could not add collaborator",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, structs.SuccessResponse{
		Success: true,
		Message: "Collaborator added successfully",
	})
}

func (m MockCaseService) RemoveCollaborator(c *gin.Context) {
	caseID := c.Param("id")
	if caseID == "" {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Case ID is required",
		})
		return
	}

	userID := c.Param("user")
	if userID == "" {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "User ID is required",
		})
		return
	}

	err := m.caseService.RemoveCollaborator(caseID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "collaborator_removal_failed",
			Message: "Could not remove collaborator",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Collaborator removed successfully",
	})
}

type MockEvidenceService struct {
	evidenceService services.EvidenceService
}

func (m MockEvidenceService) GetEvidence(c *gin.Context) {
	caseID := c.Param("id")
	if caseID == "" {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Case ID is required",
		})
		return
	}

	var filter structs.EvidenceFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid query parameters",
			Details: err.Error(),
		})
		return
	}

	evidence, err := m.evidenceService.GetEvidence(caseID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "evidence_fetch_failed",
			Message: "Could not fetch evidence",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Evidence retrieved successfully",
		Data:    evidence,
	})
}

func (m MockEvidenceService) UploadEvidence(c *gin.Context) {
	caseID := c.Param("id")
	if caseID == "" {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Case ID is required",
		})
		return
	}

	// Handle file upload
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_file",
			Message: "No file uploaded",
			Details: err.Error(),
		})
		return
	}

	// Validate file size
	const maxFileSize = 10 << 30 // 10GB
	if file.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "file_too_large",
			Message: "File size exceeds 100MB limit",
		})
		return
	}

	req := structs.UploadEvidenceRequest{
		Name:        file.Filename,
		Type:        file.Header.Get("Content-Type"),
		Description: c.PostForm("description"),
	}

	evidence, err := m.evidenceService.UploadEvidence(caseID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "upload_failed",
			Message: "Could not upload evidence",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, structs.SuccessResponse{
		Success: true,
		Message: "Evidence uploaded successfully",
		Data:    evidence,
	})
}

func (m MockEvidenceService) GetEvidenceItem(c *gin.Context) {
	caseID := c.Param("id")
	evidenceID := c.Param("e_id")
	if caseID == "" || evidenceID == "" {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Case ID and Evidence ID are required",
		})
		return
	}

	evidence, err := m.evidenceService.GetEvidenceItem(caseID, evidenceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "evidence_fetch_failed",
			Message: "Could not fetch evidence item",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Evidence item retrieved successfully",
		Data:    evidence,
	})
}

func (m MockEvidenceService) PreviewEvidence(c *gin.Context) {
	caseID := c.Param("id")
	evidenceID := c.Param("e_id")
	if caseID == "" || evidenceID == "" {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Case ID and Evidence ID are required",
		})
		return
	}

	preview, err := m.evidenceService.PreviewEvidence(caseID, evidenceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "preview_generation_failed",
			Message: "Could not generate evidence preview",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Evidence preview generated successfully",
		Data:    preview,
	})
}

type MockUserService struct {
	userService services.UserService
}

func (m MockUserService) GetUserInfo(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	userInfo, err := m.userService.GetUserInfo(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "user_info_failed",
			Message: "Could not retrieve user information",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "User information retrieved successfully",
		Data:    userInfo,
	})
}

func (m MockUserService) UpdateUserInfo(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	var req structs.UpdateUserInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid user data",
			Details: err.Error(),
		})
		return
	}

	err := m.userService.UpdateUserInfo(userID.(string), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "update_failed",
			Message: "Could not update user information",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "User information updated successfully",
	})
}

func (m MockUserService) GetUserCases(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	// Get pagination parameters
	//page := c.DefaultQuery("page", "1")
	//pageSize := c.DefaultQuery("page_size", "10")

	cases, err := m.userService.GetUserCases(userID.(string)) //, page, pageSize
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "cases_fetch_failed",
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

func NewHandler() *Handler {
	return &Handler{
		AdminService:    &MockAdminService{},
		AuthService:     &MockAuthService{},
		CaseService:     &MockCaseService{},
		EvidenceService: &MockEvidenceService{},
		UserService:     &MockUserService{},
	}
}
