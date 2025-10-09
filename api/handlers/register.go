package handlers

import (
	"aegis-api/services_/admin/delete_user"
	"aegis-api/services_/auditlog"
	"aegis-api/services_/auth/registration"
	"aegis-api/services_/case/ListUsers"
	"aegis-api/services_/user/update_user_role"
	"aegis-api/structs"
	"log"
	"net/http"

	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"

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
	RegisterTenantUser(c *gin.Context)
	RegisterTeamUser(c *gin.Context)
	VerifyEmail(c *gin.Context)
	AcceptTerms(c *gin.Context)
	ListUsers(c *gin.Context)
	ListUsersByTenant(c *gin.Context)
	DeleteUserHandler(c *gin.Context)
	CreateTenant(c *gin.Context)
	CreateTeam(c *gin.Context)
	GetAuditLogs(c *gin.Context)
	ExportAuditLogs(c *gin.Context)
}
type AuthServiceInterface interface {
	LoginHandler(c *gin.Context)
	RequestPasswordReset(c *gin.Context)
	ResetPasswordHandler(c *gin.Context)
	LogoutHandler(c *gin.Context)
	ChangePasswordHandler(c *gin.Context)
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
	GetUserByID(userID uuid.UUID) (*registration.User, error)
}
type CreateTenantRequest struct {
	Name string `json:"name" binding:"required"`
}

type CreateTeamRequest struct {
	TeamName string     `json:"team_name" binding:"required"`
	TenantID *uuid.UUID `json:"tenant_id" binding:"required"`
}
type AdminService struct {
	registrationService *registration.RegistrationService
	listUserService     ListUsers.ListUserService
	userService         *update_user_role.UserService
	userDeleteService   *delete_user.UserDeleteService
	auditLogger         *auditlog.AuditLogger
	auditLogService     auditlog.AuditLogService
}

func NewAdminService(
	regService *registration.RegistrationService,
	listUserSvc ListUsers.ListUserService,
	userService *update_user_role.UserService,
	userDeleteService *delete_user.UserDeleteService,
	auditLogger *auditlog.AuditLogger,
	auditLogService auditlog.AuditLogService,
) *AdminService {
	return &AdminService{
		registrationService: regService,
		listUserService:     listUserSvc,
		userService:         userService,
		userDeleteService:   userDeleteService,
		auditLogger:         auditLogger,
		auditLogService:     auditLogService,
	}
}

// POST /api/v1/register
func (s *AdminService) RegisterUser(c *gin.Context) {
	var req registration.RegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.auditLogger.Log(c, auditlog.AuditLog{
			Action: "REGISTER_USER",
			Actor:  auditlog.Actor{}, // anonymous
			Target: auditlog.Target{
				Type: "user_email",
				ID:   req.Email,
			},
			Service:     "auth",
			Status:      "FAILED",
			Description: "Invalid registration payload",
		})
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}
	// Extract tenant ID from context
	tenantIDVal, exists := c.Get("tenantID")
	if !exists {
		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
			Error:   "unauthorized",
			Message: "Tenant ID missing from token",
		})
		return
	}
	tenantIDStr, ok := tenantIDVal.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
			Error:   "unauthorized",
			Message: "Tenant ID is not a string",
		})
		return
	}
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid tenant ID format",
		})
		return
	}

	// Assign it to the request
	req.TenantID = &tenantID
	teamIDVal, hasTeam := c.Get("teamID")
	if hasTeam {
		if teamIDStr, ok := teamIDVal.(string); ok {
			if teamID, err := uuid.Parse(teamIDStr); err == nil {
				req.TeamID = &teamID
			} else {
				// Optional: log or return error if team ID is invalid
				c.JSON(http.StatusBadRequest, structs.ErrorResponse{
					Error:   "bad_request",
					Message: "Invalid team ID format",
				})
				return
			}
		}
	}

	user, err := s.registrationService.Register(req)
	if err != nil {
		s.auditLogger.Log(c, auditlog.AuditLog{
			Action: "REGISTER_USER",
			Actor:  auditlog.Actor{},
			Target: auditlog.Target{
				Type: "user_email",
				ID:   req.Email,
			},
			Service:     "auth",
			Status:      "FAILED",
			Description: "Registration failed: " + err.Error(),
		})
		log.Printf("Registration error: %v", err)
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "registration_failed",
			Message: err.Error(),
		})
		return
	}

	s.auditLogger.Log(c, auditlog.AuditLog{
		Action: "REGISTER_USER",
		Actor:  auditlog.Actor{}, // or you could pass admin ID if session
		Target: auditlog.Target{
			Type: "user",
			ID:   user.ID.String(),
		},
		Service:     "auth",
		Status:      "SUCCESS",
		Description: "User registered successfully",
	})

	userResp := registration.EntityToResponse(user)

	c.JSON(http.StatusCreated, structs.SuccessResponse{
		Success: true,
		Message: "User registered successfully. Please check your email for a verification link.",
		Data:    userResp,
	})
}

// GET /api/v1/auth/verify?token=xyz
// GET /api/v1/auth/verify?token=xyz
func (s *AdminService) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		s.auditLogger.Log(c, auditlog.AuditLog{
			Action: "VERIFY_EMAIL",
			Actor:  auditlog.Actor{},
			Target: auditlog.Target{
				Type: "email_verification",
				ID:   "",
			},
			Service:     "auth",
			Status:      "FAILED",
			Description: "Missing verification token",
		})
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "missing_token",
			Message: "Verification token is required",
		})
		return
	}

	err := s.registrationService.VerifyUser(token)
	if err != nil {
		s.auditLogger.Log(c, auditlog.AuditLog{
			Action: "VERIFY_EMAIL",
			Actor:  auditlog.Actor{},
			Target: auditlog.Target{
				Type: "email_verification",
				ID:   token,
			},
			Service: "auth",
			Status:  "FAILED",
			Description: "Verification failed: " + (func() string {
				if err != nil {
					return err.Error()
				}
				return "unknown error"
			})(),
		})
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "verification_failed",
			Message: err.Error(),
		})
		return
	}

	s.auditLogger.Log(c, auditlog.AuditLog{
		Action: "VERIFY_EMAIL",
		Actor:  auditlog.Actor{},
		Target: auditlog.Target{
			Type: "email_verification",
			ID:   token,
		},
		Service:     "auth",
		Status:      "SUCCESS",
		Description: "Email verified successfully",
	})

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Email verified successfully",
	})
}

// POST /api/v1/auth/accept-terms
func (s *AdminService) AcceptTerms(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		s.auditLogger.Log(c, auditlog.AuditLog{
			Action: "ACCEPT_TERMS",
			Actor:  auditlog.Actor{},
			Target: auditlog.Target{
				Type: "terms_acceptance",
				ID:   "",
			},
			Service:     "auth",
			Status:      "FAILED",
			Description: "Missing or invalid token: " + err.Error(),
		})

		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Token is required",
		})
		return
	}

	err := s.registrationService.AcceptTerms(req.Token)
	if err != nil {
		s.auditLogger.Log(c, auditlog.AuditLog{
			Action: "ACCEPT_TERMS",
			Actor:  auditlog.Actor{},
			Target: auditlog.Target{
				Type: "terms_acceptance",
				ID:   req.Token,
			},
			Service: "auth",
			Status:  "FAILED",
			Description: "Terms acceptance failed: " + (func() string {
				if err != nil {
					return err.Error()
				}
				return "unknown error"
			})(),
		})

		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "accept_terms_failed",
			Message: err.Error(),
		})
		return
	}

	s.auditLogger.Log(c, auditlog.AuditLog{
		Action: "ACCEPT_TERMS",
		Actor:  auditlog.Actor{},
		Target: auditlog.Target{
			Type: "terms_acceptance",
			ID:   req.Token,
		},
		Service:     "auth",
		Status:      "SUCCESS",
		Description: "Terms and conditions accepted successfully",
	})

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Terms accepted successfully",
	})
}

func (s *AdminService) CreateTenant(c *gin.Context) {
	var req CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	tenant, err := s.registrationService.CreateTenant(req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "create_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, structs.SuccessResponse{
		Success: true,
		Message: "Tenant created successfully",
		Data:    tenant,
	})
}

func (s *AdminService) CreateTeam(c *gin.Context) {
	var req CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	team, err := s.registrationService.CreateTeam(req.TeamName, req.TenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "create_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, structs.SuccessResponse{
		Success: true,
		Message: "Team created successfully",
		Data:    team,
	})
}
func (s *AdminService) RegisterTenantUser(c *gin.Context) {
	var req registration.RegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}
	user, err := s.registrationService.RegisterTenantUser(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "registration_failed",
			Message: err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, structs.SuccessResponse{
		Success: true,
		Message: "Tenant and user registered successfully.",
		Data:    registration.EntityToResponse(user),
	})
}

func (s *AdminService) RegisterTeamUser(c *gin.Context) {
	var req registration.RegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}
	tenantIDVal, exists := c.Get("tenantID")
	if !exists {
		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
			Error:   "unauthorized",
			Message: "Tenant ID missing from token",
		})
		return
	}

	tenantIDStr, ok := tenantIDVal.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
			Error:   "unauthorized",
			Message: "Tenant ID is not a string",
		})
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid tenant ID format",
		})
		return
	}
	req.TenantID = &tenantID
	user, err := s.registrationService.RegisterTeamUser(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "registration_failed",
			Message: err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, structs.SuccessResponse{
		Success: true,
		Message: "User registered successfully",
		Data:    user,
	})
}

// Replace the ExportAuditLogs function with corrected field names:
func (s *AdminService) ExportAuditLogs(c *gin.Context) {
	ctx := c.Request.Context()

	// Get user context for audit logging with proper nil checks
	userID, _ := c.Get("userID")
	userEmail, _ := c.Get("userEmail")
	userRole, _ := c.Get("userRole")

	// Safely convert to strings with nil checks
	var userIDStr, userEmailStr, userRoleStr string

	if userID != nil {
		if id, ok := userID.(string); ok {
			userIDStr = id
		}
	}

	if userEmail != nil {
		if email, ok := userEmail.(string); ok {
			userEmailStr = email
		}
	}

	if userRole != nil {
		if role, ok := userRole.(string); ok {
			userRoleStr = role
		}
	}

	// Parse query parameters (same as GetAuditLogs)
	filter := auditlog.AuditLogFilter{
		Status:  c.DefaultQuery("status", "ALL"),
		Action:  c.Query("action"),
		Service: c.Query("service"),
		Limit:   1000, // Higher limit for export
	}

	// Parse limit if provided
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			if limit > 10000 { // Reasonable max for export
				limit = 10000
			}
			filter.Limit = limit
		}
	}

	// Get audit logs
	logs, err := s.auditLogService.GetAuditLogs(ctx, filter)
	if err != nil {
		// Log export failure with safe string values - use correct field names
		s.auditLogger.Log(c, auditlog.AuditLog{
			Action: "EXPORT_AUDIT_LOGS",
			Actor: auditlog.Actor{
				ID:    userIDStr,
				Email: userEmailStr,
				Role:  userRoleStr,
				// Remove IP and UserAgent if they don't exist in the struct
			},
			Target: auditlog.Target{
				Type: "audit_logs",
				ID:   "export_failed",
			},
			Service:     "admin",
			Status:      "FAILED",
			Description: "Failed to export audit logs",
			Metadata: map[string]string{
				"timestamp":  time.Now().Format(time.RFC3339),
				"error":      err.Error(),
				"filter":     fmt.Sprintf("%+v", filter),
				"ip_address": c.ClientIP(),
				"user_agent": c.GetHeader("User-Agent"),
			},
		})

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve audit logs for export"})
		return
	}

	// Generate CSV content
	var csvBuffer bytes.Buffer
	writer := csv.NewWriter(&csvBuffer)

	// Write CSV header
	header := []string{
		"ID",
		"Timestamp",
		"Action",
		"Actor_ID",
		"Actor_Email",
		"Actor_Role",
		"Target_Type",
		"Target_ID",
		"Service",
		"Status",
		"Description",
		"Metadata",
	}

	if err := writer.Write(header); err != nil {
		s.auditLogger.Log(c, auditlog.AuditLog{
			Action: "EXPORT_AUDIT_LOGS",
			Actor: auditlog.Actor{
				ID:    userIDStr,
				Email: userEmailStr,
				Role:  userRoleStr,
			},
			Target: auditlog.Target{
				Type: "audit_logs",
				ID:   "csv_write_failed",
			},
			Service:     "admin",
			Status:      "FAILED",
			Description: "Failed to write CSV header",
			Metadata: map[string]string{
				"timestamp":  time.Now().Format(time.RFC3339),
				"error":      err.Error(),
				"ip_address": c.ClientIP(),
				"user_agent": c.GetHeader("User-Agent"),
			},
		})

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate CSV"})
		return
	}

	// Write log data
	for _, log := range logs {
		// Convert metadata to JSON string
		metadataStr := ""
		if log.Metadata != nil {
			if metadataBytes, err := json.Marshal(log.Metadata); err == nil {
				metadataStr = string(metadataBytes)
			}
		}

		// Get UserAgent from metadata if it's stored there
		if log.Metadata != nil {
			if _, exists := log.Metadata["user_agent"]; exists {
				// user_agent exists in metadata
			}
		}

		record := []string{
			log.ID,
			log.Timestamp.Format(time.RFC3339),
			log.Action,
			log.Actor.ID,
			log.Actor.Email,
			log.Actor.Role,
			log.Target.Type,
			log.Target.ID,
			log.Service,
			log.Status,
			log.Description,
			metadataStr,
		}

		if err := writer.Write(record); err != nil {
			s.auditLogger.Log(c, auditlog.AuditLog{
				Action: "EXPORT_AUDIT_LOGS",
				Actor: auditlog.Actor{
					ID:    userIDStr,
					Email: userEmailStr,
					Role:  userRoleStr,
				},
				Target: auditlog.Target{
					Type: "audit_logs",
					ID:   "csv_write_failed",
				},
				Service:     "admin",
				Status:      "FAILED",
				Description: "Failed to write CSV record",
				Metadata: map[string]string{
					"timestamp":  time.Now().Format(time.RFC3339),
					"error":      err.Error(),
					"record_id":  log.ID,
					"ip_address": c.ClientIP(),
					"user_agent": c.GetHeader("User-Agent"),
				},
			})

			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate CSV"})
			return
		}
	}

	writer.Flush()

	if err := writer.Error(); err != nil {
		s.auditLogger.Log(c, auditlog.AuditLog{
			Action: "EXPORT_AUDIT_LOGS",
			Actor: auditlog.Actor{
				ID:    userIDStr,
				Email: userEmailStr,
				Role:  userRoleStr,
			},
			Target: auditlog.Target{
				Type: "audit_logs",
				ID:   "csv_flush_failed",
			},
			Service:     "admin",
			Status:      "FAILED",
			Description: "Failed to flush CSV writer",
			Metadata: map[string]string{
				"timestamp":  time.Now().Format(time.RFC3339),
				"error":      err.Error(),
				"ip_address": c.ClientIP(),
				"user_agent": c.GetHeader("User-Agent"),
			},
		})

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate CSV"})
		return
	}

	// Log successful export
	s.auditLogger.Log(c, auditlog.AuditLog{
		Action: "EXPORT_AUDIT_LOGS",
		Actor: auditlog.Actor{
			ID:    userIDStr,
			Email: userEmailStr,
			Role:  userRoleStr,
		},
		Target: auditlog.Target{
			Type: "audit_logs",
			ID:   "export_success",
		},
		Service:     "admin",
		Status:      "SUCCESS",
		Description: fmt.Sprintf("Successfully exported %d audit logs to CSV", len(logs)),
		Metadata: map[string]string{
			"timestamp":    time.Now().Format(time.RFC3339),
			"record_count": strconv.Itoa(len(logs)),
			"filter":       fmt.Sprintf("%+v", filter),
			"file_size":    strconv.Itoa(csvBuffer.Len()),
			"ip_address":   c.ClientIP(),
			"user_agent":   c.GetHeader("User-Agent"),
		},
	})

	// Generate filename with timestamp
	filename := fmt.Sprintf("audit_logs_%s.csv", time.Now().Format("2006-01-02_15-04-05"))

	// Set response headers for file download
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Length", strconv.Itoa(csvBuffer.Len()))

	// Send CSV data
	c.Data(http.StatusOK, "text/csv", csvBuffer.Bytes())
}
