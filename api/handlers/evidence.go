package handlers

import (
	"aegis-api/services/evidence"
	"aegis-api/structs"
	"net/http"

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
// @Router /api/v1/evidence [post]
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

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
			Error:   "unauthorized",
			Message: "No authentication provided",
		})
		return
	}
	req.UploadedBy = userID.(string) //uid or email?

	newEvidence, err := e.service.UploadEvidence(req)
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
// @Router /api/v1/evidence/case/{case_id} [get]
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
// @Router /api/v1/evidence/user/{user_id} [get]
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
// @Router /api/v1/evidence/{id} [get]
func (e *EvidenceServices) GetEvidenceByID(c *gin.Context) {
	evidenceID := c.Param("id")
	if evidenceID == "" {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Evidence ID is required",
		})
		return
	}

	evidenceItem, err := e.service.GetEvidenceByID(evidenceID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "evidence not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, structs.ErrorResponse{
			Error:   "fetch_failed",
			Message: "Could not fetch evidence",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Evidence retrieved successfully",
		Data:    evidenceItem,
	})
}

// @Summary Download evidence file
// @Description Downloads the actual file content for a piece of evidence
// @Tags Evidence
// @Accept json
// @Produce octet-stream
// @Param id path string true "Evidence ID"
// @Success 200 {file} binary "Evidence file downloaded successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid evidence ID"
// @Failure 404 {object} structs.ErrorResponse "Evidence not found"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/evidence/{id}/download [get]
func (e *EvidenceServices) DownloadEvidence(c *gin.Context) {

}

// @Summary Get evidence metadata
// @Description Retrieves detailed metadata for a specific evidence item
// @Tags Evidence
// @Accept json
// @Produce json
// @Param id path string true "Evidence ID"
// @Success 200 {object} structs.SuccessResponse{data=evidence.Metadata} "Metadata retrieved successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid evidence ID"
// @Failure 404 {object} structs.ErrorResponse "Evidence not found"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/evidence/{id}/metadata [get]
func (e *EvidenceServices) GetEvidenceMetadata(c *gin.Context) {
	evidenceID := c.Param("id")
	if evidenceID == "" {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Evidence ID is required",
		})
		return
	}

	evidenceM, err := e.service.GetEvidenceMetadata(evidenceID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "evidence not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, structs.ErrorResponse{
			Error:   "fetch_failed",
			Message: "Could not fetch evidence metadata",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Evidence metadata retrieved successfully",
		Data:    evidenceM.Metadata,
	})
}

// @Summary Delete evidence
// @Description Deletes a specific evidence item (requires appropriate permissions)
// @Tags Evidence
// @Accept json
// @Produce json
// @Param id path string true "Evidence ID"
// @Success 200 {object} structs.SuccessResponse "Evidence deleted successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid evidence ID"
// @Failure 403 {object} structs.ErrorResponse "Permission denied"
// @Failure 404 {object} structs.ErrorResponse "Evidence not found"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/evidence/{id} [delete]
func (e *EvidenceServices) DeleteEvidence(c *gin.Context) {
	evidenceID := c.Param("id")
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

	err := e.service.DeleteEvidenceByID(evidenceID)
	if err != nil {
		status := http.StatusInternalServerError
		switch err.Error() {
		case "evidence not found":
			status = http.StatusNotFound
		case "permission denied":
			status = http.StatusForbidden
		}
		c.JSON(status, structs.ErrorResponse{
			Error:   "deletion_failed",
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
