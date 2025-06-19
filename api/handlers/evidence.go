package handlers

import (
	"aegis-api/services/evidence"
	"aegis-api/structs"
	"archive/zip"
	"bytes"
	"log"
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

// @Summary Upload evidence
// @Description Upload new evidence for a case
// @Tags Evidence
// @Accept json
// @Produce json
// @Param request body evidence.UploadEvidenceRequest true "Evidence Upload Request"
// @Success 201 {object} structs.SuccessResponse{data=evidence.Evidence} "Evidence uploaded successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid request payload"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/cases/{id}/evidence [post]
func (e *EvidenceServices) UploadEvidence(c *gin.Context) {
	var req evidence.UploadEvidenceRequest //some fields don't make sense for upload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid evidence data",
			Details: err.Error(),
		})
		return
	}

	caseID := c.Param("id")
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
	req.UploadedBy = userID.(string) //uid or email? request body just says string

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

// @Summary List evidence by case
// @Description Retrieves all evidence items for a specific case
// @Tags Evidence
// @Accept json
// @Produce json
// @Param case_id path string true "Case ID"
// @Success 200 {object} structs.SuccessResponse{data=[]evidence.Evidence} "Evidence retrieved successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid case ID"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/cases/{id}/evidence [get]
func (e *EvidenceServices) ListEvidenceByCase(c *gin.Context) {
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
// @Description Retrieves all evidence items uploaded by a specific user
// @Tags Evidence
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} structs.SuccessResponse{data=[]evidence.Evidence} "Evidence retrieved successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid user ID"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/user/{user_id}/evidence [get]
func (e *EvidenceServices) ListEvidenceByUser(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "User ID is required",
		})
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
// @Description Retrieves a specific evidence item by its ID
// @Tags Evidence
// @Accept json
// @Produce json
// @Param id path string true "Evidence ID"
// @Success 200 {object} structs.SuccessResponse{data=evidence.Evidence} "Evidence retrieved successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid evidence ID"
// @Failure 404 {object} structs.ErrorResponse "Evidence not found"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/cases/{id}/evidence/{e_id} [get]
func (e *EvidenceServices) GetEvidenceByID(c *gin.Context) {
	caseID := c.Param("id")
	evidenceID := c.Param("e_id")
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

// @Summary Download all user evidence files
// @Description Downloads all evidence files uploaded by the current user
// @Tags Evidence
// @Produce application/zip
// @Success 200 {file} binary "ZIP of user evidence files"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/user/me/evidence/download [get]
func (e *EvidenceServices) DownloadEvidenceByUser(c *gin.Context) {

	userID := c.GetString("userID")

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

	for _, file := range files {
		f, err := zipWriter.Create(file.Filename)
		if err != nil {
			continue
		}
		_, _ = f.Write(file.Content)
	}

	err = zipWriter.Close()
	if err != nil {
		return
	}

	c.Header("Content-Disposition", "attachment; filename=\"user_evidence.zip\"")
	c.Data(http.StatusOK, "application/zip", buf.Bytes())
}

// @Summary Get evidence metadata
// @Description Retrieves detailed metadata for a specific evidence item
// @Tags Evidence
// @Accept json
// @Produce json
// @Param id path string true "Evidence ID"
// @Param e_id path string true "Evidence ID"
// @Success 200 {object} structs.SuccessResponse{data=evidence.Metadata} "Metadata retrieved successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid evidence ID"
// @Failure 404 {object} structs.ErrorResponse "Evidence not found"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/cases/{id}/evidence/{e_id}/metadata [get]
func (e *EvidenceServices) GetEvidenceMetadata(c *gin.Context) {
	caseID := c.Param("id")
	evidenceID := c.Param("e_id")

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

	//COME BACK : check if a user has access to the case TO DO (middleware)

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Evidence metadata retrieved successfully",
		Data:    evidenceM,
	})
}

// @Summary Delete evidence
// @Description Deletes a specific evidence item by its id (requires admin privileges)
// @Tags Evidence
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Case ID"
// @Param e_id path string true "Evidence ID"
// @Success 200 {object} structs.SuccessResponse "Evidence deleted successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid evidence ID"
// @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// @Failure 403 {object} structs.ErrorResponse "Forbidden - Admin required"
// @Failure 404 {object} structs.ErrorResponse "Evidence not found"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/cases/{id}/evidence/{e_id} [delete]
func (e *EvidenceServices) DeleteEvidenceByID(c *gin.Context) { //admin only?
	evidenceID := c.Param("e_id")
	if evidenceID == "" {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Evidence ID is required",
		})
		return
	}

	// Get user ID from context for permission checking
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
			Error:   "unauthorized",
			Message: "No authentication provided",
		})
		return
	}

	log.Printf("Admin user %s deleting evidence %s", userID, evidenceID) //log the user deleting the evidence

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
