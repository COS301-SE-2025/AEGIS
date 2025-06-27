package handlers

import (
	"aegis-api/services_/annotation_threads/messages"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MessageHandler struct {
	service messages.MessageService
}

func NewMessageHandler(service messages.MessageService) *MessageHandler {
	return &MessageHandler{service: service}
}

type sendMessageRequest struct {
	UserID          uuid.UUID   `json:"user_id"`
	Message         string      `json:"message"`
	ParentMessageID *uuid.UUID  `json:"parent_message_id,omitempty"`
	Mentions        []uuid.UUID `json:"mentions,omitempty"`
}

func (h *MessageHandler) SendMessage(c *gin.Context) {
	threadID, err := uuid.Parse(c.Param("threadID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid thread ID"})
		return
	}

	var req sendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	msg, err := h.service.SendMessage(threadID, req.UserID, req.Message, req.ParentMessageID, req.Mentions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, msg)
}

func (h *MessageHandler) GetMessagesByThread(c *gin.Context) {
	threadID, err := uuid.Parse(c.Param("threadID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid thread ID"})
		return
	}

	msgs, err := h.service.GetMessagesByThread(threadID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve messages"})
		return
	}

	c.JSON(http.StatusOK, msgs)
}

type approveRequest struct {
	ApproverID uuid.UUID `json:"approver_id"`
}

func (h *MessageHandler) ApproveMessage(c *gin.Context) {
	messageID, err := uuid.Parse(c.Param("messageID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	var req approveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err = h.service.ApproveMessage(messageID, req.ApproverID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Approval failed", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Approved"})
}

type reactionRequest struct {
	UserID   uuid.UUID `json:"user_id"`
	Reaction string    `json:"reaction"`
}

func (h *MessageHandler) AddReaction(c *gin.Context) {
	messageID, err := uuid.Parse(c.Param("messageID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	var req reactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	err = h.service.AddReaction(messageID, req.UserID, req.Reaction)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add reaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reaction added"})
}

func (h *MessageHandler) RemoveReaction(c *gin.Context) {
	messageID, err := uuid.Parse(c.Param("messageID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	userID, err := uuid.Parse(c.Param("userID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	err = h.service.RemoveReaction(messageID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove reaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reaction removed"})
}
