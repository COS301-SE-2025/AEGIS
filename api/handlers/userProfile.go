package handlers

import (
	"aegis-api/services_/auditlog"
	"aegis-api/services_/evidence/upload"
	"aegis-api/services_/user/profile"
	"aegis-api/structs"
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type ProfileHandler struct {
	profileService *profile.ProfileService
	auditLogger    *auditlog.AuditLogger
	ipfsClient     upload.IPFSClientImp
}

func NewProfileHandler(
	profileService *profile.ProfileService,
	auditLogger *auditlog.AuditLogger,
	ipfsClient upload.IPFSClientImp,
) *ProfileHandler {
	return &ProfileHandler{
		profileService: profileService,
		auditLogger:    auditLogger,
		ipfsClient:     ipfsClient,
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
	if err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Error:   "not_found",
			Message: "User profile not found",
		})
		return
	}

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
			Actor:       auditlog.Actor{ID: req.ID, Email: req.Email},
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

	//  Handle base64 image upload
	if req.ImageBase64 != "" {
		imageURL, err := SaveBase64ImageToIPFS(h.ipfsClient, req.ID, req.ImageBase64)
		if err != nil {
			h.auditLogger.Log(c, auditlog.AuditLog{
				Action:      "UPDATE_PROFILE",
				Actor:       auditlog.Actor{ID: req.ID, Email: req.Email},
				Target:      auditlog.Target{Type: "user", ID: req.ID},
				Service:     "profile",
				Status:      "FAILED",
				Description: "Failed to upload profile picture to IPFS: " + err.Error(),
			})
			c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
				Error:   "image_upload_failed",
				Message: "Failed to upload profile picture to IPFS",
			})
			return
		}
		req.ImageURL = imageURL // Set resolved URL for DB update
	}
	fmt.Println("📸 ImageBase64 length:", len(req.ImageBase64))

	err := h.profileService.UpdateProfile(&req)
	status := "SUCCESS"
	if err != nil {
		status = "FAILED"
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "UPDATE_PROFILE",
			Actor:       auditlog.Actor{ID: req.ID, Email: req.Email},
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
	fmt.Println(" Saved ImageURL:", req.ImageURL)

	updatedProfile, err := h.profileService.GetProfile(req.ID)
	if err != nil {
		c.JSON(http.StatusOK, structs.SuccessResponse{
			Success: true,
			Message: "Profile updated but failed to fetch updated data",
			Data:    gin.H{},
		})
		return
	}

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action:      "UPDATE_PROFILE",
		Actor:       auditlog.Actor{ID: req.ID, Email: req.Email},
		Target:      auditlog.Target{Type: "user", ID: req.ID},
		Service:     "profile",
		Status:      status,
		Description: "User profile updated successfully",
	})

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Profile updated successfully",
		Data:    updatedProfile,
	})
}

// SaveBase64ImageToIPFS uploads a base64 image to IPFS and returns the image URL
func SaveBase64ImageToIPFS(ipfsClient upload.IPFSClientImp, userID string, base64Str string) (string, error) {
	if base64Str == "" {
		return "", errors.New("empty image")
	}
	// Strip metadata if present: "data:image/png;base64,..."
	split := strings.SplitN(base64Str, ",", 2)
	if len(split) != 2 {
		return "", errors.New("invalid base64 image format")
	}
	data := split[1]

	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}

	cid, err := ipfsClient.UploadFile(bytes.NewReader(decoded))
	if err != nil {
		return "", err
	}
	ipfsURL := fmt.Sprintf("https://ipfs.io/ipfs/%s", cid)
	return ipfsURL, nil
}
