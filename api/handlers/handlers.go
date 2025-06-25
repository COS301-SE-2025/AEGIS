package handlers

// import (
// 	//"strconv"

// 	//"aegis-core/services"
// 	"aegis-api/structs"
// 	"github.com/gin-gonic/gin"
// 	"time"
// 	"aegis-api/services/registration"
// 	"aegis-api/services/ListUsers"
// 	"aegis-api/services/update_user_role"
// 	"aegis-api/services/delete_user"
// 	"context"
// 	"net/http"
// 	"strings"
// 	"log"
// 	"aegis-api/services/login/auth"
// 	"github.com/google/uuid"
// 	"aegis-api/services/reset_password"
// 	"aegis-api/services/case_creation"
// 	"aegis-api/services/case_assign"

// )

// type Handler struct {
// 	AdminService    AdminServiceInterface
// 	AuthService     AuthServiceInterface
// 	CaseService     CaseServiceInterface
// 	EvidenceService EvidenceServiceInterface
// 	UserService     UserServiceInterface
// }

// //mock services

// type AdminServiceInterface interface {
// 	RegisterUser(c *gin.Context)
// 	ListUsers(c *gin.Context)
// 	//GetUserActivity(c *gin.Context)// I cant Find the Implementation
// 	UpdateUserRole(c *gin.Context)
// 	DeleteUser(c *gin.Context)
// 	//GetRoles(c *gin.Context)//I cant find implementation

// }
// type AuthServiceInterface interface {
//     LoginHandler(c *gin.Context)
//    // LogoutHandler(c *gin.Context)
//     ResetPasswordHandler(c *gin.Context)
// }

// type AdminService struct {
// 	registrationService *registration.RegistrationService
// 	listUserService     ListUsers.ListUserService
// 	userService         *update_user_role.UserService
// 	userDeleteService *delete_user.UserDeleteService
// }

// //constructor for your AdminService. It’s used to create a new instance of AdminService with its dependencies properly injected
// // NewAdminService constructs an AdminService with required dependencies.
// func NewAdminService(
//     regService *registration.RegistrationService,
//     listUserSvc ListUsers.ListUserService,
//     userService *update_user_role.UserService,
//     userDeleteService *delete_user.UserDeleteService,
// ) *AdminService {
//     return &AdminService{
//         registrationService: regService,
//         listUserService:     listUserSvc,
//         userService:         userService,
//         userDeleteService:   userDeleteService,
//     }
// }
// /*
// **
// **----------------- RegisterUser -----
// **
// */

// //Registration
// func (s *AdminService) RegisterUser(c *gin.Context) {
// 	var req registration.RegistrationRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: err.Error(),
// 		})
// 		return
// 	}

// 	user, err := s.registrationService.Register(req)
// 	if err != nil {
// 		log.Printf("Registration error: %v", err)
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "registration_failed",
// 			Message: err.Error(),
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusCreated, structs.SuccessResponse{
// 		Success: true,
// 		Message: "User registered successfully",
// 		Data:    user,
// 	})
// }

// /*
// **
// **----------------- ListUsers -----
// **
// */

// func (s *AdminService) ListUsers(c *gin.Context) {
// 	users, err := s.listUserService.ListUsers(context.Background())
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 			Error:   "list_users_failed",
// 			Message: err.Error(),
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, structs.SuccessResponse{
// 		Success: true,
// 		Message: "Users retrieved successfully",
// 		Data:    users,
// 	})
// }

// /*
// **
// **----------------- updateUserRole -----
// **
// */

// func (s *AdminService) UpdateUserRole(c *gin.Context) {
// 	userID := c.Param("user_id")
// 	if userID == "" {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: "User ID is required",
// 		})
// 		return
// 	}

// 	var req structs.UpdateUserRoleRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: err.Error(),
// 		})
// 		return
// 	}

// 	if err := s.userService.UpdateUserRole(userID, req.Role); err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "update_failed",
// 			Message: err.Error(),
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, structs.SuccessResponse{
// 		Success: true,
// 		Message: "User role updated successfully",
// 	})
// }

// /*
// **
// **----------------- DeleteUser -----
// **
// */

// // DeleteUser handles HTTP DELETE /admin/users/:user_id
// func (s *AdminService) DeleteUser(c *gin.Context) {
// 	// Bind URI parameter into your common structs type
// 	var reqStruct structs.DeleteUserRequest
// 	if err := c.ShouldBindUri(&reqStruct); err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: err.Error(),
// 		})
// 		return
// 	}

// 	// Extract requester role from context
// 	reqRoleIface, exists := c.Get("userRole")
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
// 			Error:   "unauthorized",
// 			Message: "User role missing",
// 		})
// 		return
// 	}

// 	reqRole, ok := reqRoleIface.(string)
// 	if !ok {
// 		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 			Error:   "internal_error",
// 			Message: "Invalid user role in context",
// 		})
// 		return
// 	}

// 	// Convert to service package's request type
// 	serviceReq := delete_user.DeleteUserRequest{UserID: reqStruct.UserID}

// 	// Call delete service
// 	err := s.userDeleteService.DeleteUser(serviceReq, reqRole)
// 	if err != nil {
// 		status := http.StatusInternalServerError
// 		if strings.Contains(err.Error(), "unauthorized") {
// 			status = http.StatusForbidden
// 		}
// 		c.JSON(status, structs.ErrorResponse{
// 			Error:   "deletion_failed",
// 			Message: err.Error(),
// 		})
// 		return
// 	}

// 	// Successful deletion
// 	c.JSON(http.StatusOK, structs.SuccessResponse{
// 		Success: true,
// 		Message: "User deleted successfully",
// 	})
// }

// /*
// **
// **----------------- Login -----
// **
// */

// type AuthHandler struct {
// 	authService           *auth.AuthService
// 	passwordResetService  *reset_password.PasswordResetService
// }
// type EmailSender interface {
// 	SendPasswordResetEmail(email string, token string) error
// }

// // NewAuthHandler constructs an AuthHandler with required dependencies

// func NewAuthHandler(authService *auth.AuthService, resetService *reset_password.PasswordResetService) *AuthHandler {
//     return &AuthHandler{
//         authService: authService,
//         passwordResetService: resetService,
//     }
// }

// func (h *AuthHandler) LoginHandler(c *gin.Context) {
// 	var req struct {
// 		Email    string `json:"email" binding:"required,email"`
// 		Password string `json:"password" binding:"required"`
// 	}

// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: err.Error(),
// 		})
// 		return
// 	}

// 	resp, err := h.authService.Login(req.Email, req.Password)
// 	if err != nil {
// 		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
// 			Error:   "authentication_failed",
// 			Message: "Invalid email or password",
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, structs.SuccessResponse{
// 		Success: true,
// 		Message: "Login successful",
// 		Data:    resp,
// 	})
// }

// // LogoutHandler handles POST /auth/logout
// // func (h *AuthHandler) LogoutHandler(c *gin.Context) {
// // 	// Assume userID is set in context by middleware
// // 	userID, exists := c.Get("userID")
// // 	if !exists {
// // 		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "User not authenticated"})
// // 		return
// // 	}

// // 	uid, ok := userID.(uuid.UUID)
// // 	if !ok {
// // 		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Invalid user ID"})
// // 		return
// // 	}

// // 	if err := h.authService.Logout(uid); err != nil {
// // 		c.JSON(http.StatusInternalServerError, gin.H{"error": "logout_failed", "message": err.Error()})
// // 		return
// // 	}

// // 	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Logged out successfully"})
// // }

// //ResetPasswordHandler handles POST /auth/password-reset
// func (h *AuthHandler) ResetPasswordHandler(c *gin.Context) {
// 	var req struct {
// 		Email       string `json:"email" binding:"required,email"`
// 		NewPassword string `json:"newPassword" binding:"required"`
// 		Token       string `json:"token" binding:"required"`
// 	}
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
// 		return
// 	}

// 	err := h.passwordResetService.ResetPassword(req.Token, req.NewPassword)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "reset_failed", "message": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Password reset successfully"})
// }

// type CaseServiceInterface interface {
// 	//GetCases(c *gin.Context)
// 	CreateCase(c *gin.Context)
// 	//GetCase(c *gin.Context)
// 	//UpdateCase(c *gin.Context)
// 	//GetCollaborators(c *gin.Context)
// 	CreateCollaborator(c *gin.Context)
// 	//RemoveCollaborator(c *gin.Context)
// }

// type CaseHandler struct {
//     caseService *case_creation.Service
// 	  caseAssignmentService *case_assign.CaseAssignmentService
// 	      service *case_creation.Service
// }

// func NewCaseHandler(
//     service *case_creation.Service,
//     caseAssignService *case_assign.CaseAssignmentService,
// ) *CaseHandler {
//     return &CaseHandler{
//         caseService: service,
//         caseAssignmentService: caseAssignService,
//     }
// }
// // func NewCaseHandler(caseService case_creation.ServiceInterface, caseAssignService *case_assign.CaseAssignmentService) *CaseHandler {
// //     return &CaseHandler{
// //         caseService:           caseService,
// //         caseAssignmentService: caseAssignService,
// //     }
// // }
// // CreateCase handles POST /api/v1/cases
// func (h *CaseHandler) CreateCase(c *gin.Context) {
//     var req case_creation.CreateCaseRequest

//     if err := c.ShouldBindJSON(&req); err != nil {
//         c.JSON(http.StatusBadRequest, structs.ErrorResponse{
//             Error:   "invalid_request",
//             Message: "Invalid case data",
//             Details: err.Error(),
//         })
//         return
//     }

//     newCase, err := h.service.CreateCase(req)
//     if err != nil {
//         log.Printf("CreateCase error: %v", err)
//         c.JSON(http.StatusBadRequest, structs.ErrorResponse{
//             Error:   "creation_failed",
//             Message: err.Error(),
//         })
//         return
//     }

//     c.JSON(http.StatusCreated, structs.SuccessResponse{
//         Success: true,
//         Message: "Case created successfully",
//         Data:    newCase,
//     })
// }

// func (h *CaseHandler) CreateCollaborator(c *gin.Context) {
//     caseIDStr := c.Param("case_id")
//     if caseIDStr == "" {
//         c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Case ID is required"})
//         return
//     }

//     var req struct {
//         UserID string `json:"user_id" binding:"required,uuid"`
//         Role   string `json:"role" binding:"required"`
//     }

//     if err := c.ShouldBindJSON(&req); err != nil {
//         c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
//         return
//     }

//     // Parse UUIDs
//     caseID, err := uuid.Parse(caseIDStr)
//     if err != nil {
//         c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Invalid case ID"})
//         return
//     }

//     assigneeID, err := uuid.Parse(req.UserID)
//     if err != nil {
//         c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Invalid user ID"})
//         return
//     }

//     // Get assigner ID from context (set by middleware)
//     assignerIDStr, exists := c.Get("userID")
//     if !exists {
//         c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "User ID missing from context"})
//         return
//     }

//     assignerID, err := uuid.Parse(assignerIDStr.(string))
//     if err != nil {
//         c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Invalid assigner user ID"})
//         return
//     }

//     // Call service to assign collaborator (role)
//     err = h.caseAssignmentService.AssignUserToCase(assignerID, assigneeID, caseID, req.Role)
//     if err != nil {
//         if strings.Contains(err.Error(), "forbidden") {
//             c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": err.Error()})
//             return
//         }
//         c.JSON(http.StatusInternalServerError, gin.H{"error": "assignment_failed", "message": err.Error()})
//         return
//     }

//     c.JSON(http.StatusOK, gin.H{
//         "success": true,
//         "message": "Collaborator assigned successfully",
//     })
// }

// type EvidenceServiceInterface interface {
// 	GetEvidence(c *gin.Context)
// 	UploadEvidence(c *gin.Context)
// 	GetEvidenceItem(c *gin.Context)
// 	PreviewEvidence(c *gin.Context)
// }

// type UserServiceInterface interface {
// 	GetUserInfo(c *gin.Context)
// 	UpdateUserInfo(c *gin.Context)
// 	GetUserCases(c *gin.Context)
// }

// func NewHandler(
//     adminSvc AdminServiceInterface,
//     authSvc AuthServiceInterface,
//     caseSvc CaseServiceInterface,
//     evidenceSvc EvidenceServiceInterface,
//     userSvc UserServiceInterface,
// ) *Handler {
//     return &Handler{
//         AdminService:    adminSvc,
//         AuthService:     authSvc,
//         CaseService:     caseSvc,
//         EvidenceService: evidenceSvc,
//         UserService:     userSvc,
//     }
// }

// type MockAdminService struct{}

// // @Summary Register a new user
// // @Description Registers a new user with the provided details. Only users with 'Admin' role can perform this action.
// // @Tags Admin
// // @Accept  json
// // @Produce  json
// // @Param   request body structs.RegisterUserRequest true "User Registration Request"
// // @Success 201 {object} structs.SuccessResponse{data=structs.User} "User registered successfully"
// // @Failure 400 {object} structs.ErrorResponse "Invalid request payload"
// // @Failure 500 {object} structs.ErrorResponse "Internal server error"
// // @Router /api/v1/admin/users [post]
// func (m MockAdminService) RegisterUser(c *gin.Context) {
// 	//struct to hold user data
// 	//binding and validation
// 	var req structs.RegisterUserRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: "Invalid input",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	//call the service function
// 	//user, err := m.adminService.Register(req)
// 	//if err != nil {
// 	//	c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 	//		Error:   "registration_failed",
// 	//		Message: "Could not register user",
// 	//		Details: err.Error(),
// 	//	})
// 	//	return
// 	//}

// 	//http response
// 	c.JSON(http.StatusCreated, structs.SuccessResponse{
// 		Success: true,
// 		Message: "User registered successfully",
// 		//Data:    user,
// 		Data: structs.User{ //hardcode results
// 			ID:       "mock-user-123",
// 			Email:    req.Email,
// 			FullName: req.FullName,
// 			Role: structs.UserRole{
// 				Name: req.Role,
// 			},
// 		},
// 	})
// }

// // @Summary List all users
// // @Description Retrieves a list of all registered users. Supports filtering by role, status, and creation date range.
// // @Tags Admin
// // @Accept json
// // @Produce json
// // @Param role query string false "Filter users by role (e.g., 'Forensic Analyst')"
// // @Param status query string false "Filter users by status (e.g., 'active', 'inactive')"
// // @Param start_date query string false "Filter users created after this date (YYYY-MM-DD)"
// // @Param end_date query string false "Filter users created before this date (YYYY-MM-DD)"
// // @Success 200 {object} structs.SuccessResponse{data=[]structs.User} "Users retrieved successfully"
// // @Failure 400 {object} structs.ErrorResponse "Invalid query parameters"
// // @Failure 500 {object} structs.ErrorResponse "Internal server error"
// // @Router /api/v1/admin/users [get]
// func (m MockAdminService) ListUsers(c *gin.Context) {
// 	//binding and validation
// 	var req structs.UserFilter
// 	if err := c.ShouldBindQuery(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: "Invalid query parameters",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	//call the service function
// 	//users, err := m.adminService.ListUsers(req)
// 	//if err != nil {
// 	//	c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 	//		Error:   "list_users_failed",
// 	//		Message: "Could not retrieve users",
// 	//		Details: err.Error(),
// 	//	})
// 	//	return
// 	//}

// 	mockUsers := []structs.User{
// 		{
// 			ID:       "user-1",
// 			Email:    "user1@example.com",
// 			FullName: "User One",
// 			Role:     structs.UserRole{Name: "Forensic Analyst"},
// 		},
// 		{
// 			ID:       "user-2",
// 			Email:    "user2@example.com",
// 			FullName: "User Two",
// 			Role:     structs.UserRole{Name: "DFIR Manager"},
// 		},
// 	}

// 	//http response
// 	c.JSON(http.StatusOK, structs.SuccessResponse{
// 		Success: true,
// 		Message: "Users retrieved successfully",
// 		//Data:    users,
// 		Data: mockUsers,
// 	})
// }

// // @Summary Get user activity
// // @Description Retrieves the activity log for a specific user.
// // @Tags Admin
// // @Accept json
// // @Produce json
// // @Param user_id path string true "User ID"
// // @Success 200 {object} structs.SuccessResponse{data=[]structs.UserActivity} "User activity retrieved successfully"
// // @Failure 400 {object} structs.ErrorResponse "Invalid request (e.g., missing user ID)"
// // @Failure 500 {object} structs.ErrorResponse "Internal server error"
// // @Router /api/v1/admin/users/{user_id} [get]
// func (m MockAdminService) GetUserActivity(c *gin.Context) {
// 	// Get user ID from URL parameter
// 	userID := c.Param("user_id")
// 	if userID == "" {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: "User ID is required",
// 		})
// 		return
// 	}

// 	// Get query parameters for filtering
// 	//timeRange := c.Query("time_range")
// 	//activityType := c.Query("activity_type")

// 	//activities, err := m.adminService.GetUserActivity(userID)
// 	//if err != nil {
// 	//	c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 	//		Error:   "activity_fetch_failed",
// 	//		Message: "Could not fetch user activity",
// 	//		Details: err.Error(),
// 	//	})
// 	//	return
// 	//}

// 	mockActivity := []structs.UserActivity{
// 		{
// 			UserID:   userID,
// 			Action:   "login",
// 			Resource: "system",
// 		},
// 		{
// 			UserID:   userID,
// 			Action:   "create_case",
// 			Resource: "case-123",
// 		},
// 	}

// 	c.JSON(http.StatusOK, structs.SuccessResponse{
// 		Success: true,
// 		Message: "User activity retrieved successfully",
// 		//Data:    activities,
// 		Data: mockActivity,
// 	})
// }

// // @Summary Update a user's role
// // @Description Updates the role of a specific user. Only 'Admin' can perform this action.
// // @Tags Admin
// // @Accept json
// // @Produce json
// // @Param user_id path string true "User ID"
// // @Param request body structs.UpdateUserRoleRequest true "User Role Update Request"
// // @Success 200 {object} structs.SuccessResponse "User role updated successfully"
// // @Failure 400 {object} structs.ErrorResponse "Invalid request payload or user ID"
// // @Failure 500 {object} structs.ErrorResponse "Internal server error"
// // @Router /api/v1/admin/users/{user_id} [put]
// func (m MockAdminService) UpdateUserRole(c *gin.Context) {
// 	userID := c.Param("user_id")
// 	if userID == "" {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: "User ID is required",
// 		})
// 		return
// 	}

// 	var req structs.UpdateUserRoleRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: "Invalid role data",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	//err := m.adminService.UpdateUserRole(userID, req.Role)
// 	//if err != nil {
// 	//	c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 	//		Error:   "role_update_failed",
// 	//		Message: "Could not update user role",
// 	//		Details: err.Error(),
// 	//	})
// 	//	return
// 	//}

// 	c.JSON(http.StatusOK, structs.SuccessResponse{
// 		Success: true,
// 		Message: "User role updated successfully",
// 	})
// }

// // @Summary Delete a user
// // @Description Deletes a specific user from the system. Only 'Admin' can perform this action.
// // @Tags Admin
// // @Accept json
// // @Produce json
// // @Param user_id path string true "User ID"
// // @Success 200 {object} structs.SuccessResponse "User deleted successfully"
// // @Failure 400 {object} structs.ErrorResponse "Invalid request (e.g., missing user ID)"
// // @Failure 500 {object} structs.ErrorResponse "Internal server error"
// // @Router /api/v1/admin/users/{user_id} [delete]
// func (m MockAdminService) DeleteUser(c *gin.Context) {
// 	// Get user ID from URL parameter
// 	userID := c.Param("user_id")
// 	if userID == "" {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: "User ID is required",
// 		})
// 		return
// 	}

// 	//err := m.adminService.DeleteUser(userID)
// 	//if err != nil {
// 	//	c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 	//		Error:   "deletion_failed",
// 	//		Message: "Could not delete user",
// 	//		Details: err.Error(),
// 	//	})
// 	//	return
// 	//}

// 	c.JSON(http.StatusOK, structs.SuccessResponse{
// 		Success: true,
// 		Message: "User deleted successfully",
// 	})
// }

// // @Summary Get all user roles
// // @Description Retrieves a list of all available user roles and their associated permissions.
// // @Tags Admin
// // @Accept json
// // @Produce json
// // @Success 200 {object} structs.SuccessResponse{data=[]structs.UserRole} "Roles retrieved successfully"
// // @Failure 500 {object} structs.ErrorResponse "Internal server error"
// // @Router /api/v1/admin/roles [get]
// func (m MockAdminService) GetRoles(c *gin.Context) {
// 	//roles, err := m.adminService.GetRoles()
// 	//if err != nil {
// 	//	c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 	//		Error:   "roles_fetch_failed",
// 	//		Message: "Could not fetch roles",
// 	//		Details: err.Error(),
// 	//	})
// 	//	return
// 	//}

// 	mockRoles := []structs.UserRole{
// 		{ID: "1", Name: "Incident Responder", Permissions: []string{"read_cases", "create_cases"}},
// 		{ID: "2", Name: "Forensic Analyst", Permissions: []string{"read_cases", "analyze_evidence"}},
// 		{ID: "3", Name: "DFIR Manager", Permissions: []string{"read_cases", "create_cases", "manage_users"}},
// 	}

// 	c.JSON(http.StatusOK, structs.SuccessResponse{
// 		Success: true,
// 		Message: "Roles retrieved successfully",
// 		//Data:    roles, //arr
// 		Data: mockRoles,
// 	})
// }

// type MockAuthService struct{}

// // @Summary User login
// // @Description Authenticates a user and returns a token upon successful login.
// // @Tags Authentication
// // @Accept json
// // @Produce json
// // @Param request body structs.LoginRequest true "User Login Credentials"
// // @Success 200 {object} structs.SuccessResponse{data=structs.LoginResponse} "Login successful"
// // @Failure 400 {object} structs.ErrorResponse "Invalid request payload or credentials"
// // @Failure 401 {object} structs.ErrorResponse "Authentication failed (invalid credentials)"
// // @Router /api/v1/auth/login [post]
// func (m MockAuthService) Login(c *gin.Context) {
// 	var req structs.LoginRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: "Invalid credentials",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	//response, err := m.authService.Login(req)
// 	//if err != nil {
// 	//	c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
// 	//		Error:   "authentication_failed",
// 	//		Message: "Invalid credentials",
// 	//		Details: err.Error(),
// 	//	})
// 	//	return
// 	//}

// 	response := structs.LoginResponse{
// 		Token: "mock-jwt-token-12345",
// 		User: structs.User{
// 			ID:       "user-123",
// 			Email:    req.Email,
// 			FullName: "Mock User",
// 			Role:     structs.UserRole{Name: "Forensic Analyst"},
// 		},
// 	}

// 	c.JSON(http.StatusOK, structs.SuccessResponse{
// 		Success: true,
// 		Message: "Login successful",
// 		Data:    response,
// 	})
// }

// // @Summary User logout
// // @Description Logs out the currently authenticated user by invalidating their session or token. Requires authentication.
// // @Tags Authentication
// // @Accept json
// // @Produce json
// // @Security ApiKeyAuth
// // @Success 200 {object} structs.SuccessResponse "Logged out successfully"
// // @Failure 401 {object} structs.ErrorResponse "Unauthorized (user not authenticated)"
// // @Failure 500 {object} structs.ErrorResponse "Internal server error"
// // @Router /api/v1/auth/logout [post]
// func (m MockAuthService) Logout(c *gin.Context) {
// 	_, exists := c.Get("userID") //_ -> userID
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
// 			Error:   "unauthorized",
// 			Message: "User not authenticated",
// 		})
// 		return
// 	}

// 	//err := m.authService.Logout(userID.(string))
// 	//if err != nil {
// 	//	c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 	//		Error:   "logout_failed",
// 	//		Message: "Could not log out user",
// 	//		Details: err.Error(),
// 	//	})
// 	//	return
// 	//}

// 	c.JSON(http.StatusOK, structs.SuccessResponse{
// 		Success: true,
// 		Message: "Logged out successfully",
// 	})
// }

// // @Summary Request password reset
// // @Description Initiates the password reset process by sending a reset email to the user's registered email address.
// // @Tags Authentication
// // @Accept json
// // @Produce json
// // @Param request body structs.ResetPasswordRequest true "Password Reset Request"
// // @Success 200 {object} structs.SuccessResponse "Password reset email sent successfully"
// // @Failure 400 {object} structs.ErrorResponse "Invalid request payload (e.g., malformed email)"
// // @Failure 500 {object} structs.ErrorResponse "Internal server error"
// // @Router /api/v1/auth/password-reset [post]
// func (m MockAuthService) ResetPassword(c *gin.Context) {
// 	var req structs.ResetPasswordRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: "Invalid reset password data",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	//err := m.authService.ResetPassword(req)
// 	//if err != nil {
// 	//	c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 	//		Error:   "reset_failed",
// 	//		Message: "Could not reset password",
// 	//		Details: err.Error(),
// 	//	})
// 	//	return
// 	//}

// 	c.JSON(http.StatusOK, structs.SuccessResponse{
// 		Success: true,
// 		Message: "Password reset email sent successfully",
// 	})
// }

// type MockCaseService struct{}

// // @Summary Get all cases
// // @Description Retrieves a paginated and filterable list of security cases.
// // @Tags Cases
// // @Accept json
// // @Produce json
// // @Param status query string false "Filter cases by status (e.g., 'open', 'closed')"
// // @Param start_date query string false "Filter cases created after this date (YYYY-MM-DD)"
// // @Param end_date query string false "Filter cases created before this date (YYYY-MM-DD)"
// // @Param page query int false "Page number for pagination (default: 1)" default(1)
// // @Param page_size query int false "Number of items per page (default: 10, max: 100)" default(10)
// // @Security ApiKeyAuth
// // @Success 200 {object} structs.SuccessResponse{data=[]structs.Case} "Cases retrieved successfully"
// // @Failure 400 {object} structs.ErrorResponse "Invalid query parameters"
// // @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// // @Failure 500 {object} structs.ErrorResponse "Internal server error"
// // @Router /api/v1/cases [get]
// func (m MockCaseService) GetCases(c *gin.Context) {
// 	var filter structs.CaseFilter
// 	if err := c.ShouldBindQuery(&filter); err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: "Invalid query parameters",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	//if filter.Page == "" {
// 	//	filter.Page = "1"
// 	//}
// 	//if filter.PageSize == "" {
// 	//	filter.PageSize = "10"
// 	//}
// 	//if page, err := strconv.Atoi(filter.Page); err != nil || page < 1 {
// 	//	c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 	//		Error:   "invalid_request",
// 	//		Message: "Invalid page number",
// 	//	})
// 	//	return
// 	//}
// 	//
// 	//if pageSize, err := strconv.Atoi(filter.PageSize); err != nil || pageSize < 1 || pageSize > 100 {
// 	//	c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 	//		Error:   "invalid_request",
// 	//		Message: "Invalid page size (must be between 1 and 100)",
// 	//	})
// 	//	return
// 	//}

// 	//cases, err := m.caseService.GetCases(filter)
// 	//if err != nil {
// 	//	c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 	//		Error:   "cases_fetch_failed",
// 	//		Message: "Could not fetch cases",
// 	//		Details: err.Error(),
// 	//	})
// 	//	return
// 	//}

// 	mockCases := []structs.Case{
// 		{
// 			ID:          "case-1",
// 			Title:       "Incident: Server Breach",
// 			Description: "Unauthorized access to prod server",
// 			Status:      "open",
// 			CreatedBy:   "admin-1",
// 			CreatedAt:   time.Now().Add(-48 * time.Hour),
// 			Collaborators: []structs.CollaboratorInfo{
// 				{ID: "user-1", FullName: "User One", Role: "Forensic Analyst"},
// 				{ID: "user-2", FullName: "User Two", Role: "DFIR Manager"},
// 			},
// 		},
// 	}

// 	c.JSON(http.StatusOK, structs.SuccessResponse{
// 		Success: true,
// 		Message: "Cases retrieved successfully",
// 		//Data:    cases,
// 		Data: mockCases,
// 	})
// }

// // @Summary Create a new case
// // @Description Creates a new security case. Requires 'Admin' role.
// // @Tags Cases
// // @Accept json
// // @Produce json
// // @Param request body structs.CreateCaseRequest true "Case Creation Request"
// // @Security ApiKeyAuth
// // @Success 201 {object} structs.SuccessResponse{data=structs.Case} "Case created successfully"
// // @Failure 400 {object} structs.ErrorResponse "Invalid request payload"
// // @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// // @Failure 403 {object} structs.ErrorResponse "Forbidden (insufficient role)"
// // @Failure 500 {object} structs.ErrorResponse "Internal server error"
// // @Router /api/v1/cases [post]
// func (m MockCaseService) CreateCase(c *gin.Context) {
// 	var req structs.CreateCaseRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: "Invalid case data",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	// Get user ID from context (set by auth middleware)
// 	//userID, exists := c.Get("userID")
// 	//if !exists {
// 	//	c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
// 	//		Error:   "unauthorized",
// 	//		Message: "User not authenticated",
// 	//	})
// 	//	return
// 	//}

// 	//case_, err := m.caseService.CreateCase(userID.(string), req)
// 	//if err != nil {
// 	//	c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 	//		Error:   "case_creation_failed",
// 	//		Message: "Could not create case",
// 	//		Details: err.Error(),
// 	//	})
// 	//	return
// 	//}

// 	mockCase := structs.Case{
// 		ID:          "new-case-123",
// 		Title:       req.Title,
// 		Description: req.Description,
// 		Status:      "active",
// 		CreatedBy:   "mock-user-123",
// 	}

// 	c.JSON(http.StatusCreated, structs.SuccessResponse{
// 		Success: true,
// 		Message: "Case created successfully",
// 		//Data:    case_,
// 		Data: mockCase,
// 	})
// }

// // @Summary Get a specific case
// // @Description Retrieves details of a single security case by its ID.
// // @Tags Cases
// // @Accept json
// // @Produce json
// // @Param id path string true "Case ID"
// // @Security ApiKeyAuth
// // @Success 200 {object} structs.SuccessResponse{data=structs.Case} "Case retrieved successfully"
// // @Failure 400 {object} structs.ErrorResponse "Invalid request (e.g., missing case ID)"
// // @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// // @Failure 404 {object} structs.ErrorResponse "Case not found"
// // @Failure 500 {object} structs.ErrorResponse "Internal server error"
// // @Router /api/v1/cases/{id} [get]
// func (m MockCaseService) GetCase(c *gin.Context) {
// 	caseID := c.Param("id")
// 	if caseID == "" {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: "Case ID is required",
// 		})
// 		return
// 	}

// 	//case_, err := m.caseService.GetCase(caseID)
// 	//if err != nil {
// 	//	c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 	//		Error:   "case_fetch_failed",
// 	//		Message: "Could not fetch case",
// 	//		Details: err.Error(),
// 	//	})
// 	//	return
// 	//}

// 	mockCase := structs.Case{
// 		ID:          caseID,
// 		Title:       "Mock Case",
// 		Description: "This is a mock case for testing",
// 		Status:      "active",
// 		CreatedBy:   "user-123",
// 	}

// 	c.JSON(http.StatusOK, structs.SuccessResponse{
// 		Success: true,
// 		Message: "Case retrieved successfully",
// 		//Data:    case_,
// 		Data: mockCase,
// 	})
// }

// // @Summary Update a case
// // @Description Updates the details of an existing security case. Requires 'Admin' role.
// // @Tags Cases
// // @Accept json
// // @Produce json
// // @Param id path string true "Case ID"
// // @Param request body structs.UpdateCaseRequest true "Case Update Request"
// // @Security ApiKeyAuth
// // @Success 200 {object} structs.SuccessResponse "Case updated successfully"
// // @Failure 400 {object} structs.ErrorResponse "Invalid request payload or case ID"
// // @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// // @Failure 403 {object} structs.ErrorResponse "Forbidden (insufficient role)"
// // @Failure 404 {object} structs.ErrorResponse "Case not found"
// // @Failure 500 {object} structs.ErrorResponse "Internal server error"
// // @Router /api/v1/cases/{id} [put]
// func (m MockCaseService) UpdateCase(c *gin.Context) {
// 	caseID := c.Param("id")
// 	if caseID == "" {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: "Case ID is required",
// 		})
// 		return
// 	}

// 	var req structs.UpdateCaseRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: "Invalid case data",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	//err := m.caseService.UpdateCase(caseID, req)
// 	//if err != nil {
// 	//	c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 	//		Error:   "update_failed",
// 	//		Message: "Could not update case",
// 	//		Details: err.Error(),
// 	//	})
// 	//	return
// 	//}

// 	c.JSON(http.StatusOK, structs.SuccessResponse{
// 		Success: true,
// 		Message: "Case updated successfully",
// 	})
// }

// // @Summary Get case collaborators
// // @Description Retrieves a list of users collaborating on a specific case.
// // @Tags Cases
// // @Accept json
// // @Produce json
// // @Param id path string true "Case ID"
// // @Security ApiKeyAuth
// // @Success 200 {object} structs.SuccessResponse{data=[]structs.User} "Collaborators retrieved successfully"
// // @Failure 400 {object} structs.ErrorResponse "Invalid request (e.g., missing case ID)"
// // @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// // @Failure 404 {object} structs.ErrorResponse "Case not found"
// // @Failure 500 {object} structs.ErrorResponse "Internal server error"
// // @Router /api/v1/cases/{id}/collaborators [get]
// func (m MockCaseService) GetCollaborators(c *gin.Context) {
// 	caseID := c.Param("id")
// 	if caseID == "" {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: "Case ID is required",
// 		})
// 		return
// 	}

// 	//collaborators, err := m.caseService.GetCollaborators(caseID)
// 	//if err != nil {
// 	//	c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 	//		Error:   "collaborators_fetch_failed",
// 	//		Message: "Could not fetch collaborators",
// 	//		Details: err.Error(),
// 	//	})
// 	//	return
// 	//}

// 	mockCollaborators := []structs.User{
// 		{
// 			ID:       "collab-1",
// 			Email:    "analyst1@example.com",
// 			FullName: "Analyst One",
// 			Role:     structs.UserRole{Name: "Forensic Analyst"},
// 		},
// 	}

// 	c.JSON(http.StatusOK, structs.SuccessResponse{
// 		Success: true,
// 		Message: "Collaborators retrieved successfully",
// 		//Data:    collaborators,
// 		Data: mockCollaborators,
// 	})
// }

// // @Summary Add a collaborator to a case
// // @Description Adds a user as a collaborator to a specific case. Requires 'Admin' role.
// // @Tags Cases
// // @Accept json
// // @Produce json
// // @Param id path string true "Case ID"
// // @Param request body structs.User true "Collaborator Details"
// // @Security ApiKeyAuth
// // @Success 201 {object} structs.SuccessResponse "Collaborator added successfully"
// // @Failure 400 {object} structs.ErrorResponse "Invalid request payload or case ID"
// // @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// // @Failure 403 {object} structs.ErrorResponse "Forbidden (insufficient role)"
// // @Failure 404 {object} structs.ErrorResponse "Case not found"
// // @Failure 500 {object} structs.ErrorResponse "Internal server error"
// // @Router /api/v1/cases/{id}/collaborators [post]
// func (m MockCaseService) CreateCollaborator(c *gin.Context) {
// 	caseID := c.Param("id")
// 	if caseID == "" {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: "Case ID is required",
// 		})
// 		return
// 	}

// 	var req structs.User
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: "Invalid collaborator data",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	//err := m.caseService.AddCollaborator(caseID, req) //or assignCase
// 	//if err != nil {
// 	//	c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 	//		Error:   "collaborator_creation_failed",
// 	//		Message: "Could not add collaborator",
// 	//		Details: err.Error(),
// 	//	})
// 	//	return
// 	//}

// 	c.JSON(http.StatusCreated, structs.SuccessResponse{
// 		Success: true,
// 		Message: "Collaborator added successfully",
// 	})
// }

// // @Summary Remove a collaborator from a case
// // @Description Removes a user from the list of collaborators on a specific case. Requires 'Admin' role.
// // @Tags Cases
// // @Accept json
// // @Produce json
// // @Param id path string true "Case ID"
// // @Param user path string true "User ID of the collaborator to remove"
// // @Security ApiKeyAuth
// // @Success 200 {object} structs.SuccessResponse "Collaborator removed successfully"
// // @Failure 400 {object} structs.ErrorResponse "Invalid request (e.g., missing IDs)"
// // @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// // @Failure 403 {object} structs.ErrorResponse "Forbidden (insufficient role)"
// // @Failure 404 {object} structs.ErrorResponse "Case or user not found"
// // @Failure 500 {object} structs.ErrorResponse "Internal server error"
// // @Router /api/v1/cases/{id}/collaborators/{user} [delete]
// func (m MockCaseService) RemoveCollaborator(c *gin.Context) {
// 	caseID := c.Param("id")
// 	if caseID == "" {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: "Case ID is required",
// 		})
// 		return
// 	}

// 	userID := c.Param("user")
// 	if userID == "" {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: "User ID is required",
// 		})
// 		return
// 	}

// 	//err := m.caseService.RemoveCollaborator(caseID, userID)
// 	//if err != nil {
// 	//	c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 	//		Error:   "collaborator_removal_failed",
// 	//		Message: "Could not remove collaborator",
// 	//		Details: err.Error(),
// 	//	})
// 	//	return
// 	//}

// 	c.JSON(http.StatusOK, structs.SuccessResponse{
// 		Success: true,
// 		Message: "Collaborator removed successfully",
// 	})
// }

// type MockEvidenceService struct{}

// // @Summary Get evidence for a case
// // @Description Retrieves a list of evidence items associated with a specific security case. Supports filtering.
// // @Tags Evidence
// // @Accept json
// // @Produce json
// // @Param id path string true "Case ID"
// // @Param type query string false "Filter evidence by type (e.g., 'image', 'document', 'application/pdf')"
// // @Param uploaded_by query string false "Filter evidence by the user who uploaded it (User ID)"
// // @Param start_date query string false "Filter evidence uploaded after this date (YYYY-MM-DD)"
// // @Param end_date query string false "Filter evidence uploaded before this date (YYYY-MM-DD)"
// // @Security ApiKeyAuth
// // @Success 200 {object} structs.SuccessResponse{data=[]structs.EvidenceItem} "Evidence retrieved successfully"
// // @Failure 400 {object} structs.ErrorResponse "Invalid request parameters (e.g., missing case ID, invalid query)"
// // @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// // @Failure 404 {object} structs.ErrorResponse "Case not found"
// // @Failure 500 {object} structs.ErrorResponse "Internal server error"
// // @Router /api/v1/cases/{id}/evidence [get]
// func (m MockEvidenceService) GetEvidence(c *gin.Context) {
// 	caseID := c.Param("id")
// 	if caseID == "" {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: "Case ID is required",
// 		})
// 		return
// 	}

// 	var filter structs.EvidenceFilter
// 	if err := c.ShouldBindQuery(&filter); err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: "Invalid query parameters",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	//evidence, err := m.evidenceService.GetEvidence(caseID, filter)
// 	//if err != nil {
// 	//	c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 	//		Error:   "evidence_fetch_failed",
// 	//		Message: "Could not fetch evidence",
// 	//		Details: err.Error(),
// 	//	})
// 	//	return
// 	//}

// 	mockEvidence := []structs.EvidenceItem{
// 		{
// 			ID:     "evidence-1",
// 			CaseID: caseID,
// 			Name:   "suspicious_file.exe",
// 			Type:   "application/octet-stream",
// 			Hash:   "abc123hash",
// 		},
// 	}

// 	c.JSON(http.StatusOK, structs.SuccessResponse{
// 		Success: true,
// 		Message: "Evidence retrieved successfully",
// 		//Data:    evidence,
// 		Data: mockEvidence,
// 	})
// }

// // @Summary Upload evidence to a case
// // @Description Uploads a new evidence file to a specified security case.
// // @Tags Evidence
// // @Accept mpfd
// // @Produce json
// // @Param id path string true "Case ID"
// // @Param file formData file true "Evidence file to upload (max 10GB)"
// // @Param description formData string false "Optional description for the evidence file"
// // @Security ApiKeyAuth
// // @Success 201 {object} structs.SuccessResponse{data=structs.EvidenceItem} "Evidence uploaded successfully"
// // @Failure 400 {object} structs.ErrorResponse "Invalid request (e.g., missing file, invalid case ID, file too large)"
// // @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// // @Failure 403 {object} structs.ErrorResponse "Forbidden (insufficient role)"
// // @Failure 500 {object} structs.ErrorResponse "Internal server error"
// // @Router /api/v1/cases/{id}/evidence [post]
// func (m MockEvidenceService) UploadEvidence(c *gin.Context) {
// 	caseID := c.Param("id")
// 	if caseID == "" {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: "Case ID is required",
// 		})
// 		return
// 	}

// 	// Handle file upload
// 	file, err := c.FormFile("file")
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_file",
// 			Message: "No file uploaded",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	// Validate file size
// 	const maxFileSize = 10 << 30 // 10GB
// 	if file.Size > maxFileSize {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "file_too_large",
// 			Message: "File size exceeds 10GB limit",
// 		})
// 		return
// 	}

// 	//req := structs.UploadEvidenceRequest{
// 	//	Name:        file.Filename,
// 	//	Type:        file.Header.Get("Content-Type"),
// 	//	Description: c.PostForm("description"),
// 	//}

// 	mockEvidence := structs.EvidenceItem{
// 		ID:     "new-evidence-123",
// 		CaseID: caseID,
// 		Name:   file.Filename,
// 		Type:   file.Header.Get("Content-Type"),
// 		Hash:   "mock-hash-123",
// 	}

// 	//evidence, err := m.evidenceService.UploadEvidence(caseID, req)
// 	//if err != nil {
// 	//	c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 	//		Error:   "upload_failed",
// 	//		Message: "Could not upload evidence",
// 	//		Details: err.Error(),
// 	//	})
// 	//	return
// 	//}

// 	c.JSON(http.StatusCreated, structs.SuccessResponse{
// 		Success: true,
// 		Message: "Evidence uploaded successfully",
// 		//Data:    evidence,
// 		Data: mockEvidence,
// 	})
// }

// // @Summary Get a specific evidence item
// // @Description Retrieves details of a single evidence item by its ID within a specific case.
// // @Tags Evidence
// // @Accept json
// // @Produce json
// // @Param id path string true "Case ID"
// // @Param e_id path string true "Evidence Item ID"
// // @Security ApiKeyAuth
// // @Success 200 {object} structs.SuccessResponse{data=structs.EvidenceItem} "Evidence item retrieved successfully"
// // @Failure 400 {object} structs.ErrorResponse "Invalid request (e.g., missing IDs)"
// // @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// // @Failure 404 {object} structs.ErrorResponse "Case or evidence item not found"
// // @Failure 500 {object} structs.ErrorResponse "Internal server error"
// // @Router /api/v1/cases/{id}/evidence/{e_id} [get]
// func (m MockEvidenceService) GetEvidenceItem(c *gin.Context) {
// 	caseID := c.Param("id")
// 	evidenceID := c.Param("e_id")
// 	if caseID == "" || evidenceID == "" {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: "Case ID and Evidence ID are required",
// 		})
// 		return
// 	}

// 	//evidence, err := m.evidenceService.GetEvidenceItem(caseID, evidenceID)
// 	//if err != nil {
// 	//	c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 	//		Error:   "evidence_fetch_failed",
// 	//		Message: "Could not fetch evidence item",
// 	//		Details: err.Error(),
// 	//	})
// 	//	return
// 	//}

// 	mockEvidence := structs.EvidenceItem{
// 		ID:     evidenceID,
// 		CaseID: caseID,
// 		Name:   "evidence_file.pdf",
// 		Type:   "application/pdf",
// 		Hash:   "evidence-hash-456",
// 	}

// 	c.JSON(http.StatusOK, structs.SuccessResponse{
// 		Success: true,
// 		Message: "Evidence item retrieved successfully",
// 		//Data:    evidence,
// 		Data: mockEvidence,
// 	})
// }

// // @Summary Get evidence preview
// // @Description Generates a preview for a specific evidence item.
// // @Tags Evidence
// // @Accept json
// // @Produce json
// // @Param id path string true "Case ID"
// // @Param e_id path string true "Evidence Item ID"
// // @Security ApiKeyAuth
// // @Success 200 {object} structs.SuccessResponse{data=structs.EvidencePreview} "Evidence preview generated successfully"
// // @Failure 400 {object} structs.ErrorResponse "Invalid request (e.g., missing IDs)"
// // @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// // @Failure 404 {object} structs.ErrorResponse "Case or evidence item not found"
// // @Failure 500 {object} structs.ErrorResponse "Internal server error"
// // @Router /api/v1/cases/{id}/evidence/{e_id}/preview [get]
// func (m MockEvidenceService) PreviewEvidence(c *gin.Context) {
// 	caseID := c.Param("id")
// 	evidenceID := c.Param("e_id")
// 	if caseID == "" || evidenceID == "" {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: "Case ID and Evidence ID are required",
// 		})
// 		return
// 	}

// 	mockPreview := structs.EvidencePreview{
// 		ID:         evidenceID,
// 		Name:       "evidence_preview.pdf",
// 		Type:       "application/pdf",
// 		PreviewURL: "/api/v1/evidence/preview/" + evidenceID,
// 	}

// 	//preview, err := m.evidenceService.PreviewEvidence(caseID, evidenceID)
// 	//if err != nil {
// 	//	c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 	//		Error:   "preview_generation_failed",
// 	//		Message: "Could not generate evidence preview",
// 	//		Details: err.Error(),
// 	//	})
// 	//	return
// 	//}

// 	c.JSON(http.StatusOK, structs.SuccessResponse{
// 		Success: true,
// 		Message: "Evidence preview generated successfully",
// 		//Data:    preview,
// 		Data: mockPreview,
// 	})
// }

// type MockUserService struct{}

// // @Summary Get current user's information
// // @Description Retrieves the detailed profile information for the authenticated user.
// // @Tags User
// // @Accept json
// // @Produce json
// // @Security ApiKeyAuth
// // @Success 200 {object} structs.SuccessResponse{data=structs.User} "User information retrieved successfully"
// // @Failure 401 {object} structs.ErrorResponse "Unauthorized (user not authenticated)"
// // @Failure 500 {object} structs.ErrorResponse "Internal server error"
// // @Router /api/v1/user/info [get]
// func (m MockUserService) GetUserInfo(c *gin.Context) {
// 	userID, exists := c.Get("userID")
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
// 			Error:   "unauthorized",
// 			Message: "User not authenticated",
// 		})
// 		return
// 	}

// 	//userInfo, err := m.userService.GetUserInfo(userID.(string))
// 	//if err != nil {
// 	//	c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 	//		Error:   "user_info_failed",
// 	//		Message: "Could not retrieve user information",
// 	//		Details: err.Error(),
// 	//	})
// 	//	return
// 	//}

// 	mockUser := structs.User{
// 		ID:       userID.(string),
// 		Email:    "user@example.com",
// 		FullName: "Mock User",
// 		Role:     structs.UserRole{Name: "Forensic Analyst"},
// 	}

// 	c.JSON(http.StatusOK, structs.SuccessResponse{
// 		Success: true,
// 		Message: "User information retrieved successfully",
// 		//Data:    userInfo,
// 		Data: mockUser,
// 	})
// }

// // @Summary Update current user's information
// // @Description Updates the profile details (e.g., name, email) for the authenticated user.
// // @Tags User
// // @Accept json
// // @Produce json
// // @Param request body structs.UpdateUserInfoRequest true "User Info Update Request"
// // @Security ApiKeyAuth
// // @Success 200 {object} structs.SuccessResponse "User information updated successfully"
// // @Failure 400 {object} structs.ErrorResponse "Invalid request payload"
// // @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// // @Failure 500 {object} structs.ErrorResponse "Internal server error"
// // @Router /api/v1/user/info [put]
// func (m MockUserService) UpdateUserInfo(c *gin.Context) {
// 	_, exists := c.Get("userID") // _ -> userID
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
// 			Error:   "unauthorized",
// 			Message: "User not authenticated",
// 		})
// 		return
// 	}

// 	var req structs.UpdateUserInfoRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: "Invalid user data",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	//err := m.userService.UpdateUserInfo(userID.(string), req)
// 	//if err != nil {
// 	//	c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 	//		Error:   "update_failed",
// 	//		Message: "Could not update user information",
// 	//		Details: err.Error(),
// 	//	})
// 	//	return
// 	//}

// 	c.JSON(http.StatusOK, structs.SuccessResponse{
// 		Success: true,
// 		Message: "User information updated successfully",
// 	})
// }

// // @Summary Get cases assigned to the current user
// // @Description Retrieves a list of security cases that the authenticated user is involved in.
// // @Tags User
// // @Accept json
// // @Produce json
// // @Security ApiKeyAuth
// // @Success 200 {object} structs.SuccessResponse{data=[]structs.Case} "User cases retrieved successfully"
// // @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// // @Failure 500 {object} structs.ErrorResponse "Internal server error"
// // @Router /api/v1/user/cases [get]
// func (m MockUserService) GetUserCases(c *gin.Context) {
// 	userID, exists := c.Get("userID")
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
// 			Error:   "unauthorized",
// 			Message: "User not authenticated",
// 		})
// 		return
// 	}

// 	// Get pagination parameters
// 	//page := c.DefaultQuery("page", "1")
// 	//pageSize := c.DefaultQuery("page_size", "10")

// 	//cases, err := m.userService.GetUserCases(userID.(string)) //, page, pageSize
// 	//if err != nil {
// 	//	c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 	//		Error:   "cases_fetch_failed",
// 	//		Message: "Could not fetch user cases",
// 	//		Details: err.Error(),
// 	//	})
// 	//	return
// 	//}
// 	mockCases := []structs.Case{
// 		{
// 			ID:          "user-case-1",
// 			Title:       "User's Case 1",
// 			Description: "First case assigned to user",
// 			Status:      "active",
// 			CreatedBy:   userID.(string),
// 		},
// 	}

// 	c.JSON(http.StatusOK, structs.SuccessResponse{
// 		Success: true,
// 		Message: "User cases retrieved successfully",
// 		//Data:    cases,
// 		Data: mockCases,
// 	})
// }

// //
// //func NewHandler() *Handler {
// //	return &Handler{
// //		AdminService: DummyAdminService{},
// //		AuthService:  &services.MockAuthService{},
// //		CaseService:  &services.MockCaseService{},
// //		// EvidenceService: &services.MockEvidenceService{},
// //		// UserService:     &services.MockUserService{},
// //	}
// //}

// func NewHandlerWithMocks(
// 	admin AdminServiceInterface,
// 	auth AuthServiceInterface,
// 	caseSvc CaseServiceInterface,
// 	evidence EvidenceServiceInterface,
// 	user UserServiceInterface,
// ) *Handler {
// 	return &Handler{
// 		AdminService:    admin,
// 		AuthService:     auth,
// 		CaseService:     caseSvc,
// 		EvidenceService: evidence,
// 		UserService:     user,
// 	}
// }
