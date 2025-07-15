package handlers

import (
	"aegis-api/services_/auditlog"
	"aegis-api/services_/user/profile"
	"aegis-api/structs"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProfileHandler struct {
	profileService *profile.ProfileService
	auditLogger    *auditlog.AuditLogger
}

func NewProfileHandler(
	profileService *profile.ProfileService,
	auditLogger *auditlog.AuditLogger,
) *ProfileHandler {
	return &ProfileHandler{
		profileService: profileService,
		auditLogger:    auditLogger,
	}
}

func (h *ProfileHandler) GetProfileHandler(c *gin.Context) {
	userID := c.Param("userID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "userID is required",
		})
		return
	}

	profileData, err := h.profileService.GetProfile(userID)
	status := "SUCCESS"
	if err != nil {
		status = "FAILED"
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "GET_PROFILE",
			Actor:       auditlog.Actor{ID: userID},
			Target:      auditlog.Target{Type: "user", ID: userID},
			Service:     "profile",
			Status:      status,
			Description: "Failed to retrieve user profile",
		})
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Error:   "not_found",
			Message: "User profile not found",
		})
		return
	}

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action:      "GET_PROFILE",
		Actor:       auditlog.Actor{ID: userID},
		Target:      auditlog.Target{Type: "user", ID: userID},
		Service:     "profile",
		Status:      status,
		Description: "User profile retrieved successfully",
	})

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Profile retrieved successfully",
		Data:    profileData,
	})
}

func (h *ProfileHandler) UpdateProfileHandler(c *gin.Context) {
	var req profile.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "UPDATE_PROFILE",
			Actor:       auditlog.Actor{ID: req.ID},
			Target:      auditlog.Target{Type: "user", ID: req.ID},
			Service:     "profile",
			Status:      "FAILED",
			Description: "Invalid update profile request payload",
		})
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	err := h.profileService.UpdateProfile(&req)
	status := "SUCCESS"
	if err != nil {
		status = "FAILED"
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "UPDATE_PROFILE",
			Actor:       auditlog.Actor{ID: req.ID},
			Target:      auditlog.Target{Type: "user", ID: req.ID},
			Service:     "profile",
			Status:      status,
			Description: "Failed to update user profile: " + err.Error(),
		})
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
		return
	}

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action:      "UPDATE_PROFILE",
		Actor:       auditlog.Actor{ID: req.ID},
		Target:      auditlog.Target{Type: "user", ID: req.ID},
		Service:     "profile",
		Status:      status,
		Description: "User profile updated successfully",
	})

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Profile updated successfully",
	})
}
