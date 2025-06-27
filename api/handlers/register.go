package handlers

import (
	"aegis-api/services/delete_user"
	"aegis-api/services_/auth/registration"
	"aegis-api/services_/case/ListUsers"
	"aegis-api/services_/user/update_user_role"
	"aegis-api/structs"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// //"aegis-core/services"
// "aegis-api/services/delete_user"
// evidence "aegis-api/services/evidence/upload"
// "aegis-api/services/registration"
// "aegis-api/services/reset_password"
// auth "aegis-api/services_/auth/login"
// "aegis-api/services_/case/ListUsers"
// "aegis-api/services_/case/case_assign"
// "aegis-api/services_/case/case_creation"
// "aegis-api/services_/user/update_user_role"
type AdminServiceInterface interface {
	RegisterUser(c *gin.Context)
	VerifyEmail(c *gin.Context)
}
type AuthServiceInterface interface {
	LoginHandler(c *gin.Context)
	RequestPasswordReset(c *gin.Context)
	ResetPasswordHandler(c *gin.Context)
}

// type CaseServiceInterface interface {
// 	CreateCase(c *gin.Context)
// 	AssignCase(c *gin.Context)
// 	GetCaseByID(c *gin.Context)
// }

type EvidenceServiceInterface interface {
	UploadEvidence(c *gin.Context)
	GetEvidenceByID(c *gin.Context)
	DeleteEvidence(c *gin.Context)
}

type UserServiceInterface interface {
	GetUserInfo(c *gin.Context)
	UpdateUserInfo(c *gin.Context)
	GetUserCases(c *gin.Context)
}

type AdminService struct {
	registrationService *registration.RegistrationService
	listUserService     ListUsers.ListUserService
	userService         *update_user_role.UserService
	userDeleteService   *delete_user.UserDeleteService
}

func NewAdminService(
	regService *registration.RegistrationService,
	listUserSvc ListUsers.ListUserService,
	userService *update_user_role.UserService,
	userDeleteService *delete_user.UserDeleteService,
) *AdminService {
	return &AdminService{
		registrationService: regService,
		listUserService:     listUserSvc,
		userService:         userService,
		userDeleteService:   userDeleteService,
	}
}

// POST /api/v1/register
func (s *AdminService) RegisterUser(c *gin.Context) {
	var req registration.RegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	user, err := s.registrationService.Register(req)
	if err != nil {
		log.Printf("Registration error: %v", err)
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "registration_failed",
			Message: err.Error(),
		})
		return
	}

	userResp := registration.EntityToResponse(user)

	c.JSON(http.StatusCreated, structs.SuccessResponse{
		Success: true,
		Message: "User registered successfully. Please check your email for a verification link.",
		Data:    userResp,
	})
}

// GET /api/v1/auth/verify?token=xyz
func (s *AdminService) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "missing_token",
			Message: "Verification token is required",
		})
		return
	}

	err := s.registrationService.VerifyUser(token)
	if err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "verification_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Email verified successfully",
	})
}
