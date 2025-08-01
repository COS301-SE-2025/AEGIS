package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type MarkReadRequest struct {
	NotificationIDs []string `json:"notificationIds"`
}

type DeleteRequest struct {
	NotificationIDs []string `json:"notificationIds"`
}

type ArchiveRequest struct {
	NotificationIDs []string `json:"notificationIds"`
}

// GET /api/notifications
func (h *Handler) GetNotifications(c *gin.Context) {
	userID := c.GetString("userID")
	tenantID := c.GetString("tenantID")
	teamID := c.GetString("teamID")

	notifs, err := h.NotificationService.GetNotificationsForUser(tenantID, teamID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve notifications",
		})
		return
	}

	c.JSON(http.StatusOK, notifs)
}

// POST /api/notifications/read
func (h *Handler) MarkNotificationsRead(c *gin.Context) {
	var req MarkReadRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.NotificationIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notificationIds"})
		return
	}

	if err := h.NotificationService.MarkAsRead(req.NotificationIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// DELETE /api/notifications/delete
func (h *Handler) DeleteNotifications(c *gin.Context) {
	var req DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.NotificationIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notificationIds"})
		return
	}

	if err := h.NotificationService.DeleteNotifications(req.NotificationIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// POST /api/notifications/archive
func (h *Handler) ArchiveNotifications(c *gin.Context) {
	var req ArchiveRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.NotificationIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notificationIds"})
		return
	}

	if err := h.NotificationService.ArchiveNotifications(req.NotificationIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to archive notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
