package handlers

import (
	"aegis-api/services_/annotation_threads/messages"
	"aegis-api/services_/auditlog"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type sendMessageRequest struct {
	UserID          uuid.UUID   `json:"user_id"`
	Message         string      `json:"message"`
	ParentMessageID *uuid.UUID  `json:"parent_message_id,omitempty"`
	Mentions        []uuid.UUID `json:"mentions,omitempty"`
}
type approveRequest struct {
	ApproverID uuid.UUID `json:"approver_id"`
}

type reactionRequest struct {
	UserID   uuid.UUID `json:"user_id"`
	Reaction string    `json:"reaction"`
}
type MessageHandler struct {
	service     messages.MessageService
	auditLogger *auditlog.AuditLogger
}

func NewMessageHandler(service messages.MessageService, logger *auditlog.AuditLogger) *MessageHandler {
	return &MessageHandler{service: service, auditLogger: logger}
}

func extractActorFromContext(c *gin.Context) auditlog.Actor {
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")

	uidStr, ok := userID.(string)
	if !ok {
		uidStr = ""
	}
	roleStr, ok := userRole.(string)
	if !ok {
		roleStr = ""
	}

	return auditlog.Actor{
		ID:        uidStr,
		Role:      roleStr,
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}
}

func (h *MessageHandler) SendMessage(c *gin.Context) {
	actor := extractActorFromContext(c)

	threadID, err := uuid.Parse(c.Param("threadID"))
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "SEND_MESSAGE",
			Actor:       actor,
			Target:      auditlog.Target{Type: "thread_message", ID: c.Param("threadID")},
			Service:     "annotation_messages",
			Status:      "FAILED",
			Description: "Invalid thread ID",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid thread ID"})
		return
	}

	var req sendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "SEND_MESSAGE",
			Actor:       actor,
			Target:      auditlog.Target{Type: "thread_message", ID: threadID.String()},
			Service:     "annotation_messages",
			Status:      "FAILED",
			Description: "Invalid request body",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	msg, err := h.service.SendMessage(threadID, req.UserID, req.Message, req.ParentMessageID, req.Mentions)
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "SEND_MESSAGE",
			Actor:  actor,
			Target: auditlog.Target{
				Type:           "thread_message",
				ID:             threadID.String(),
				AdditionalInfo: map[string]string{"content": req.Message},
			},
			Service:     "annotation_messages",
			Status:      "FAILED",
			Description: "Failed to send message: " + err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message", "details": err.Error()})
		return
	}

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "SEND_MESSAGE",
		Actor:  actor,
		Target: auditlog.Target{
			Type:           "thread_message",
			ID:             threadID.String(),
			AdditionalInfo: map[string]string{"message_id": msg.ID.String()},
		},
		Service:     "annotation_messages",
		Status:      "SUCCESS",
		Description: "Message sent successfully",
	})

	c.JSON(http.StatusOK, msg)
}

func (h *MessageHandler) GetMessagesByThread(c *gin.Context) {
	actor := extractActorFromContext(c)
	threadID, err := uuid.Parse(c.Param("threadID"))
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "GET_THREAD_MESSAGES",
			Actor:       actor,
			Target:      auditlog.Target{Type: "thread_message_list", ID: c.Param("threadID")},
			Service:     "annotation_messages",
			Status:      "FAILED",
			Description: "Invalid thread ID",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid thread ID"})
		return
	}

	msgs, err := h.service.GetMessagesByThread(threadID)
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "GET_THREAD_MESSAGES",
			Actor:       actor,
			Target:      auditlog.Target{Type: "thread_message_list", ID: threadID.String()},
			Service:     "annotation_messages",
			Status:      "FAILED",
			Description: "Failed to retrieve messages: " + err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve messages"})
		return
	}

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action:      "GET_THREAD_MESSAGES",
		Actor:       actor,
		Target:      auditlog.Target{Type: "thread_message_list", ID: threadID.String()},
		Service:     "annotation_messages",
		Status:      "SUCCESS",
		Description: fmt.Sprintf("Retrieved %d messages", len(msgs)),
	})

	c.JSON(http.StatusOK, msgs)
}

func (h *MessageHandler) ApproveMessage(c *gin.Context) {
	actor := extractActorFromContext(c)
	messageID, err := uuid.Parse(c.Param("messageID"))
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "APPROVE_MESSAGE",
			Actor:       actor,
			Target:      auditlog.Target{Type: "message_approval", ID: c.Param("messageID")},
			Service:     "annotation_messages",
			Status:      "FAILED",
			Description: "Invalid message ID",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	var req approveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "APPROVE_MESSAGE",
			Actor:       actor,
			Target:      auditlog.Target{Type: "message_approval", ID: messageID.String()},
			Service:     "annotation_messages",
			Status:      "FAILED",
			Description: "Invalid request body",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err = h.service.ApproveMessage(messageID, req.ApproverID)
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "APPROVE_MESSAGE",
			Actor:       actor,
			Target:      auditlog.Target{Type: "message_approval", ID: messageID.String()},
			Service:     "annotation_messages",
			Status:      "FAILED",
			Description: "Approval failed: " + err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Approval failed", "details": err.Error()})
		return
	}

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action:      "APPROVE_MESSAGE",
		Actor:       actor,
		Target:      auditlog.Target{Type: "message_approval", ID: messageID.String()},
		Service:     "annotation_messages",
		Status:      "SUCCESS",
		Description: "Message approved successfully",
	})

	c.JSON(http.StatusOK, gin.H{"message": "Approved"})
}

func (h *MessageHandler) AddReaction(c *gin.Context) {
	actor := extractActorFromContext(c)
	messageID, err := uuid.Parse(c.Param("messageID"))
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "ADD_REACTION",
			Actor:       actor,
			Target:      auditlog.Target{Type: "message_reaction", ID: c.Param("messageID")},
			Service:     "annotation_messages",
			Status:      "FAILED",
			Description: "Invalid message ID",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	var req reactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "ADD_REACTION",
			Actor:       actor,
			Target:      auditlog.Target{Type: "message_reaction", ID: messageID.String()},
			Service:     "annotation_messages",
			Status:      "FAILED",
			Description: "Invalid input",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	err = h.service.AddReaction(messageID, req.UserID, req.Reaction)
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "ADD_REACTION",
			Actor:       actor,
			Target:      auditlog.Target{Type: "message_reaction", ID: messageID.String()},
			Service:     "annotation_messages",
			Status:      "FAILED",
			Description: "Failed to add reaction: " + err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add reaction"})
		return
	}

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action:      "ADD_REACTION",
		Actor:       actor,
		Target:      auditlog.Target{Type: "message_reaction", ID: messageID.String()},
		Service:     "annotation_messages",
		Status:      "SUCCESS",
		Description: "Reaction added successfully",
	})

	c.JSON(http.StatusOK, gin.H{"message": "Reaction added"})
}

func (h *MessageHandler) RemoveReaction(c *gin.Context) {
	actor := extractActorFromContext(c)
	messageID, err := uuid.Parse(c.Param("messageID"))
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "REMOVE_REACTION",
			Actor:       actor,
			Target:      auditlog.Target{Type: "message_reaction", ID: c.Param("messageID")},
			Service:     "annotation_messages",
			Status:      "FAILED",
			Description: "Invalid message ID",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	userID, err := uuid.Parse(c.Param("userID"))
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "REMOVE_REACTION",
			Actor:       actor,
			Target:      auditlog.Target{Type: "message_reaction", ID: messageID.String()},
			Service:     "annotation_messages",
			Status:      "FAILED",
			Description: "Invalid user ID",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	err = h.service.RemoveReaction(messageID, userID)
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "REMOVE_REACTION",
			Actor:       actor,
			Target:      auditlog.Target{Type: "message_reaction", ID: messageID.String()},
			Service:     "annotation_messages",
			Status:      "FAILED",
			Description: "Failed to remove reaction: " + err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove reaction"})
		return
	}

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action:      "REMOVE_REACTION",
		Actor:       actor,
		Target:      auditlog.Target{Type: "message_reaction", ID: messageID.String()},
		Service:     "annotation_messages",
		Status:      "SUCCESS",
		Description: "Reaction removed successfully",
	})

	c.JSON(http.StatusOK, gin.H{"message": "Reaction removed"})
}
