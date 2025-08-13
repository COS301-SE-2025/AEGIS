package report

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// DeleteReport handles the request to delete a report
func (r *ReportService) DeleteReport(c *gin.Context) {
	reportID := c.Param("reportID")        // Assuming the report ID is passed as a parameter in the URL
	userID := c.MustGet("userID").(string) // Getting the user ID from the context
	userRole := c.MustGet("role").(string) // Getting the user role from the context
	userAgent := c.Request.UserAgent()     // Getting the user agent from the request
	ipAddress := c.ClientIP()              // Getting the IP address from the request
	email := c.MustGet("email").(string)   // Getting the user's email
	timestamp := time.Now()

	// Log the delete action
	err := r.AuditLogger.LogDeleteReport(c.Request.Context(), reportID, userID, userRole, userAgent, ipAddress, email, timestamp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log delete event"})
		return
	}

	// Delete the report from the repository
	err = r.CaseReportsRepo.DeleteReportByID(c, reportID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{"message": "Report deleted successfully"})
}
