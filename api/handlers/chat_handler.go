package handlers

import (
	"aegis-api/services_/chat"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChatHandler struct {
	ChatService *chat.ChatService
}

func NewChatHandler(chatService *chat.ChatService) *ChatHandler {
	return &ChatHandler{ChatService: chatService}
}

// ───── Groups ───────────────────────────────────────────────

func (h *ChatHandler) CreateGroup(c *gin.Context) {
	var group chat.ChatGroup
	if err := c.ShouldBindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}
	if err := h.ChatService.Repo().CreateGroup(c.Request.Context(), &group); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create group", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, group)
}

func (h *ChatHandler) GetGroupByID(c *gin.Context) {
	groupID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group ID"})
		return
	}
	group, err := h.ChatService.Repo().GetGroupByID(c.Request.Context(), groupID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "group not found", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, group)
}

func (h *ChatHandler) GetUserGroups(c *gin.Context) {
	email := c.Param("email")
	groups, err := h.ChatService.Repo().GetUserGroups(c.Request.Context(), email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get groups", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, groups)
}

func (h *ChatHandler) UpdateGroup(c *gin.Context) {
	groupID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group ID"})
		return
	}

	var group chat.ChatGroup
	if err := c.ShouldBindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}
	group.ID = groupID

	if err := h.ChatService.Repo().UpdateGroup(c.Request.Context(), &group); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update group", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, group)
}

func (h *ChatHandler) DeleteGroup(c *gin.Context) {
	groupID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group ID"})
		return
	}
	if err := h.ChatService.Repo().DeleteGroup(c.Request.Context(), groupID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete group", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "group deleted"})
}

func (h *ChatHandler) AddMemberToGroup(c *gin.Context) {
	groupID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group ID"})
		return
	}
	var member chat.Member
	if err := c.ShouldBindJSON(&member); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}
	if err := h.ChatService.Repo().AddMemberToGroup(c.Request.Context(), groupID, &member); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add member", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "member added"})
}

func (h *ChatHandler) RemoveMemberFromGroup(c *gin.Context) {
	groupID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group ID"})
		return
	}
	email := c.Param("email")
	if err := h.ChatService.Repo().RemoveMemberFromGroup(c.Request.Context(), groupID, email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove member", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "member removed"})
}

// ───── Messages ─────────────────────────────────────────────

func (h *ChatHandler) SendMessage(c *gin.Context) {
	groupID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group ID"})
		return
	}
	var msg chat.Message
	if err := c.ShouldBindJSON(&msg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}
	msg.GroupID = groupID
	if err := h.ChatService.Repo().CreateMessage(c.Request.Context(), &msg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create message", "details": err.Error()})
		return
	}
	_ = h.ChatService.WsManager().BroadcastToGroup(groupID.Hex(), msg)
	c.JSON(http.StatusOK, msg)
}

func (h *ChatHandler) GetMessages(c *gin.Context) {
	groupID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group ID"})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	var before *primitive.ObjectID
	if beforeStr := c.Query("before"); beforeStr != "" {
		objID, _ := primitive.ObjectIDFromHex(beforeStr)
		before = &objID
	}
	messages, err := h.ChatService.Repo().GetMessages(c.Request.Context(), groupID, limit, before)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get messages", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, messages)
}
