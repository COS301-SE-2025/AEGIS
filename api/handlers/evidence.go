package handlers

import (
	"aegis-api/middleware"
	"aegis-api/services/evidence"
	"aegis-api/structs"
	"archive/zip"
	"bytes"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type EvidenceServices struct {
	service *evidence.Service
}

func NewEvidenceHandler(service *evidence.Service) *EvidenceServices {
	return &EvidenceServices{
		service: service,
	}
}

/*
// @Summary Upload evidence
// @Description Upload new evidence to a case. Requires authentication.
// @Tags Evidence
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param case_id path string true "Case ID"
// @Param request body evidence.UploadEvidenceRequest true "Evidence Upload Request"
// @Success 201 {object} structs.SuccessResponse{data=evidence.Evidence} "Evidence uploaded successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid request payload or case ID"
// @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/cases/{case_id}/evidence [post]
func (e *EvidenceServices) UploadEvidence(c *gin.Context) {
	var req evidence.UploadEvidenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid evidence data",
			Details: err.Error(),
		})
		return
	}

	caseID := c.Param("case_id")
	req.CaseID = caseID

	// Get user ID from context
	userID, exists := c.Get("userID") //middleware
	if !exists {
		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
			Error:   "unauthorized",
			Message: "No authentication provided",
		})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "server_error",
			Message: "User ID type assertion failed",
		})
		return
	}

	req.UploadedBy = userIDStr //assuming uploadedby is uid

	newEvidence, err := e.service.UploadEvidence(req)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "invalid case ID") || strings.Contains(err.Error(), "invalid user ID") {
			status = http.StatusBadRequest
		}

		c.JSON(status, structs.ErrorResponse{
			Error:   "upload_failed",
			Message: "Could not upload evidence",
			Details: err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, structs.SuccessResponse{
		Success: true,
		Message: "Evidence uploaded successfully",
		Data:    newEvidence,
	})
}
*/

// @Summary List evidence by case
// @Description Retrieves all evidence items for a specific case. Requires authentication.
// @Tags Evidence
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param case_id path string true "Case ID"
// @Success 200 {object} structs.SuccessResponse{data=[]evidence.Evidence} "Evidence retrieved successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid case ID"
// @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/cases/{case_id}/evidence [get]
func (e *EvidenceServices) ListEvidenceByCaseID(c *gin.Context) {
	caseID := c.Param("case_id")
	if caseID == "" {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Case ID is required",
		})
		return
	}

	evidenceList, err := e.service.ListEvidenceByCase(caseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "fetch_failed",
			Message: "Could not fetch evidence",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Evidence retrieved successfully",
		Data:    evidenceList,
	})
}

// @Summary List evidence by user
// @Description Retrieves all evidence associated with a user. Admins can access any user's evidence, regular users can only access their own.
// @Tags Users, Admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param user_id path string false "User ID (required for admin access to other users)"
// @Success 200 {object} structs.SuccessResponse{data=[]evidence.Evidence} "User evidence retrieved successfully"
// @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// @Failure 403 {object} structs.ErrorResponse "Forbidden"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/users/evidence [get]
// @Router /api/v1/admin/users/{user_id}/evidence [get]
func (e *EvidenceServices) ListEvidenceByUserID(c *gin.Context) {
	userID, ok := middleware.GetTargetUserID(c)
	if !ok {
		return
	}

	evidenceList, err := e.service.ListEvidenceByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "fetch_failed",
			Message: "Could not fetch evidence",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Evidence retrieved successfully",
		Data:    evidenceList,
	})
}

// @Summary Get evidence by ID
// @Description Retrieves a specific evidence item by its ID. Requires authentication.
// @Tags Evidence
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param case_id path string true "Case ID"
// @Param evidence_id path string true "Evidence ID"
// @Success 200 {object} structs.SuccessResponse{data=evidence.Evidence} "Evidence retrieved successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid evidence ID or case ID"
// @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// @Failure 404 {object} structs.ErrorResponse "Evidence not found"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/cases/{case_id}/evidence/{evidence_id} [get]
func (e *EvidenceServices) GetEvidenceByID(c *gin.Context) {
	caseID := c.Param("case_id")
	evidenceID := c.Param("evidence_id")
	if caseID == "" || evidenceID == "" {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Case ID and Evidence ID are required",
		})
		return
	}

	evidenceItem, err := e.service.GetEvidenceByID(evidenceID)
	if err != nil {
		status := http.StatusInternalServerError
		errorType := "fetch_failed"
		if err.Error() == "evidence not found" {
			status = http.StatusNotFound
			errorType = "evidence_not_found"
		} else if strings.Contains(err.Error(), "invalid") {
			status = http.StatusBadRequest
			errorType = "invalid_evidence_id"
		}
		c.JSON(status, structs.ErrorResponse{
			Error:   errorType,
			Message: "Could not fetch evidence",
			Details: err.Error(),
		})
		return
	}

	// Convert UUID to string for comparison with URL parameter
	if evidenceItem.CaseID.String() != caseID {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_path",
			Message: "Evidence does not belong to this case",
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Evidence retrieved successfully",
		Data:    evidenceItem,
	})
}

// @Summary Download evidence by user
// @Description Downloads a specific evidence file associated with a user. Admins can access any user's evidence, regular users can only access their own.
// @Tags Users, Admin
// @Accept json
// @Produce octet-stream
// @Security ApiKeyAuth
// @Param user_id path string false "User ID (required for admin access to other users)"
// @Param evidence_id path string true "Evidence ID"
// @Success 200 {file} binary "Evidence file"
// @Failure 400 {object} structs.ErrorResponse "Invalid request"
// @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// @Failure 403 {object} structs.ErrorResponse "Forbidden"
// @Failure 404 {object} structs.ErrorResponse "Evidence not found"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/users/evidence/{evidence_id} [get]
// @Router /api/v1/admin/users/{user_id}/evidence/{evidence_id} [get]
func (e *EvidenceServices) DownloadEvidenceByUserID(c *gin.Context) {

	userID, ok := middleware.GetTargetUserID(c)
	if !ok {
		return
	}

	files, err := e.service.DownloadEvidenceByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "download_failed",
			Message: "Could not download evidence file",
			Details: err.Error(),
		})
		return
	}

	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	//for _, file := range files {
	//	f, err := zipWriter.Create(file.Filename)
	//	if err != nil {
	//		continue
	//	}
	//	_, _ = f.Write(file.Content)
	//}

	for _, file := range files {
		f, err := zipWriter.Create(file.Filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
				Error:   "zip_creation_failed",
				Message: "Failed to create zip file entry",
				Details: err.Error(),
			})
			return
		}
		_, err = f.Write(file.Content)
		if err != nil {
			c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
				Error:   "zip_write_failed",
				Message: "Failed to write content to zip file",
				Details: err.Error(),
			})
			return
		}
	}

	err = zipWriter.Close()
	//if err != nil {
	//	return
	//}
	if err = zipWriter.Close(); err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "zip_close_failed",
			Message: "Failed to finalize zip file",
			Details: err.Error(),
		})
		return
	}

	c.Header("Content-Disposition", "attachment; filename=\"user_evidence.zip\"")
	c.Data(http.StatusOK, "application/zip", buf.Bytes())
}

// @Summary Get evidence metadata
// @Description Retrieves detailed metadata for a specific evidence item. Requires authentication.
// @Tags Evidence
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param case_id path string true "Case ID"
// @Param evidence_id path string true "Evidence ID"
// @Success 200 {object} structs.SuccessResponse{data=evidence.Metadata} "Metadata retrieved successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid evidence ID or case ID"
// @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// @Failure 404 {object} structs.ErrorResponse "Evidence not found"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/cases/{case_id}/evidence/{evidence_id}/metadata [get]
func (e *EvidenceServices) GetEvidenceMetadata(c *gin.Context) {
	caseID := c.Param("case_id")
	evidenceID := c.Param("evidence_id")

	evidenceM, err := e.service.GetEvidenceMetadata(evidenceID)
	if err != nil {
		status := http.StatusInternalServerError
		errorType := "fetch_failed"

		switch {
		case strings.Contains(err.Error(), "not found"):
			status = http.StatusNotFound
			errorType = "evidence_not_found"
		case strings.Contains(err.Error(), "invalid"):
			status = http.StatusBadRequest
			errorType = "invalid_evidence_id"
		}

		c.JSON(status, structs.ErrorResponse{
			Error:   errorType,
			Message: "Could not fetch evidence metadata",
			Details: err.Error(),
		})
		return
	}

	if evidenceM.CaseID != caseID {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_path",
			Message: "Evidence does not belong to this case",
		})
		return
	}

	//COME BACK: check if a user has access to the case TO DO

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Evidence metadata retrieved successfully",
		Data:    evidenceM,
	})
}

// @Summary Delete evidence
// @Description Deletes a specific evidence item by its ID. Only administrators can perform this action.
// @Tags Evidence
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param case_id path string true "Case ID"
// @Param evidence_id path string true "Evidence ID"
// @Success 200 {object} structs.SuccessResponse "Evidence deleted successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid evidence ID or case ID"
// @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// @Failure 403 {object} structs.ErrorResponse "Forbidden - Admin role required"
// @Failure 404 {object} structs.ErrorResponse "Evidence not found"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/cases/{case_id}/evidence/{evidence_id} [delete]
func (e *EvidenceServices) DeleteEvidenceByID(c *gin.Context) { //admin only?
	evidenceID := c.Param("evidence_id")
	if evidenceID == "" {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Evidence ID is required",
		})
		return
	}

	// Get user ID from context for permission checking
	_, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
			Error:   "unauthorized",
			Message: "No authentication provided",
		})
		return
	}

	//log.Printf("Admin user %s deleting evidence %s", userID, evidenceID) //log the user deleting the evidence

	err := e.service.DeleteEvidenceByID(evidenceID) //why doesn't it take the userid to know who initiated the deletion?
	if err != nil {
		status := http.StatusInternalServerError
		errorType := "deletion_failed"
		switch err.Error() {
		case "evidence not found":
			status = http.StatusNotFound
		case "permission denied":
			status = http.StatusForbidden
		}
		c.JSON(status, structs.ErrorResponse{
			Error:   errorType,
			Message: "Could not delete evidence",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Evidence deleted successfully",
	})
}
