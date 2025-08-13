package report

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// DownloadReport handles the request to download a report
func (r *ReportService) DownloadReport(c *gin.Context) {
	reportID := c.Param("reportID")            // Assuming the report ID is passed as a parameter in the URL
	userID := c.MustGet("userID").(string)     // Getting the user ID from the context
	userRole := c.MustGet("userRole").(string) // Getting the user role from the context
	userAgent := c.Request.UserAgent()         // Getting the user agent from the request
	ipAddress := c.ClientIP()                  // Getting the IP address from the request
	email := c.MustGet("email").(string)       // Getting the user's email
	timestamp := time.Now()

	// Log the download action
	err := r.AuditLogger.LogDownloadReport(c.Request.Context(), reportID, userID, userRole, userAgent, ipAddress, email, timestamp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log download event"})
		return
	}

	// Fetch the report from the repository
	report, err := r.CaseReportsRepo.GetByID(c.Request.Context(), reportID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
		return
	}

	// Assume report is a file stored in the server; serve the file for download
	// The report content could be a file, a URL to the file, or base64-encoded data
	c.Header("Content-Type", "application/pdf")                                          // Assuming PDF format
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", report.Name)) // Set the filename
	c.File(report.FilePath)                                                              // Send the file to the client
}
