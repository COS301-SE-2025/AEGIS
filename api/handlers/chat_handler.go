package handlers

import (
	"aegis-api/services_/chat"
	"net/http"
	"strconv"

	"aegis-api/services_/auditlog"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChatHandler struct {
	ChatService *chat.ChatService
	auditLogger *auditlog.AuditLogger
}

func NewChatHandler(chatService *chat.ChatService, logger *auditlog.AuditLogger) *ChatHandler {
	return &ChatHandler{ChatService: chatService, auditLogger: logger}
}

// ───── Groups ───────────────────────────────────────────────

func (h *ChatHandler) CreateGroup(c *gin.Context) {
	actor := extractActorFromContext(c)

	var group chat.ChatGroup
	if err := c.ShouldBindJSON(&group); err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "CREATE_GROUP",
			Actor:       actor,
			Target:      auditlog.Target{Type: "group", ID: ""},
			Service:     "chat",
			Status:      "FAILED",
			Description: "Invalid input",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	if err := h.ChatService.Repo().CreateGroup(c.Request.Context(), &group); err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "CREATE_GROUP",
			Actor:       actor,
			Target:      auditlog.Target{Type: "group", ID: ""},
			Service:     "chat",
			Status:      "FAILED",
			Description: "Failed to create group: " + err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create group", "details": err.Error()})
		return
	}

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action:      "CREATE_GROUP",
		Actor:       actor,
		Target:      auditlog.Target{Type: "group", ID: group.ID.Hex()},
		Service:     "chat",
		Status:      "SUCCESS",
		Description: "Group created",
	})
	c.JSON(http.StatusOK, group)
}

func (h *ChatHandler) GetGroupByID(c *gin.Context) {
	actor := extractActorFromContext(c)

	groupID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "GET_GROUP",
			Actor:       actor,
			Target:      auditlog.Target{Type: "group", ID: c.Param("id")},
			Service:     "chat",
			Status:      "FAILED",
			Description: "Invalid group ID",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group ID"})
		return
	}

	group, err := h.ChatService.Repo().GetGroupByID(c.Request.Context(), groupID)
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "GET_GROUP",
			Actor:       actor,
			Target:      auditlog.Target{Type: "group", ID: groupID.Hex()},
			Service:     "chat",
			Status:      "FAILED",
			Description: "Group not found: " + err.Error(),
		})
		c.JSON(http.StatusNotFound, gin.H{"error": "group not found", "details": err.Error()})
		return
	}

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action:      "GET_GROUP",
		Actor:       actor,
		Target:      auditlog.Target{Type: "group", ID: groupID.Hex()},
		Service:     "chat",
		Status:      "SUCCESS",
		Description: "Group retrieved",
	})
	c.JSON(http.StatusOK, group)
}

func (h *ChatHandler) GetUserGroups(c *gin.Context) {
	actor := extractActorFromContext(c)
	email := c.Param("email")

	groups, err := h.ChatService.Repo().GetUserGroups(c.Request.Context(), email)
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "GET_USER_GROUPS",
			Actor:       actor,
			Target:      auditlog.Target{Type: "user_groups", ID: email},
			Service:     "chat",
			Status:      "FAILED",
			Description: "Failed to get user groups: " + err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get groups", "details": err.Error()})
		return
	}

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action:      "GET_USER_GROUPS",
		Actor:       actor,
		Target:      auditlog.Target{Type: "user_groups", ID: email},
		Service:     "chat",
		Status:      "SUCCESS",
		Description: "Retrieved user groups",
	})

	c.JSON(http.StatusOK, groups)
}

func (h *ChatHandler) UpdateGroup(c *gin.Context) {
	actor := extractActorFromContext(c)

	groupID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "UPDATE_GROUP",
			Actor:       actor,
			Target:      auditlog.Target{Type: "group", ID: c.Param("id")},
			Service:     "chat",
			Status:      "FAILED",
			Description: "Invalid group ID",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group ID"})
		return
	}

	var group chat.ChatGroup
	if err := c.ShouldBindJSON(&group); err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "UPDATE_GROUP",
			Actor:       actor,
			Target:      auditlog.Target{Type: "group", ID: groupID.Hex()},
			Service:     "chat",
			Status:      "FAILED",
			Description: "Invalid input body",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}
	group.ID = groupID

	if err := h.ChatService.Repo().UpdateGroup(c.Request.Context(), &group); err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "UPDATE_GROUP",
			Actor:       actor,
			Target:      auditlog.Target{Type: "group", ID: groupID.Hex()},
			Service:     "chat",
			Status:      "FAILED",
			Description: "Failed to update group: " + err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update group", "details": err.Error()})
		return
	}

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action:      "UPDATE_GROUP",
		Actor:       actor,
		Target:      auditlog.Target{Type: "group", ID: groupID.Hex()},
		Service:     "chat",
		Status:      "SUCCESS",
		Description: "Group updated successfully",
	})

	c.JSON(http.StatusOK, group)
}

func (h *ChatHandler) DeleteGroup(c *gin.Context) {
	actor := extractActorFromContext(c)

	groupID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "DELETE_GROUP",
			Actor:       actor,
			Target:      auditlog.Target{Type: "group", ID: c.Param("id")},
			Service:     "chat",
			Status:      "FAILED",
			Description: "Invalid group ID",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group ID"})
		return
	}

	if err := h.ChatService.Repo().DeleteGroup(c.Request.Context(), groupID); err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "DELETE_GROUP",
			Actor:       actor,
			Target:      auditlog.Target{Type: "group", ID: groupID.Hex()},
			Service:     "chat",
			Status:      "FAILED",
			Description: "Failed to delete group: " + err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete group", "details": err.Error()})
		return
	}

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action:      "DELETE_GROUP",
		Actor:       actor,
		Target:      auditlog.Target{Type: "group", ID: groupID.Hex()},
		Service:     "chat",
		Status:      "SUCCESS",
		Description: "Group deleted successfully",
	})

	c.JSON(http.StatusOK, gin.H{"message": "group deleted"})
}

func (h *ChatHandler) AddMemberToGroup(c *gin.Context) {
	actor := extractActorFromContext(c)

	groupID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "ADD_GROUP_MEMBER",
			Actor:       actor,
			Target:      auditlog.Target{Type: "group", ID: c.Param("id")},
			Service:     "chat",
			Status:      "FAILED",
			Description: "Invalid group ID",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group ID"})
		return
	}

	var member chat.Member
	if err := c.ShouldBindJSON(&member); err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "ADD_GROUP_MEMBER",
			Actor:       actor,
			Target:      auditlog.Target{Type: "group", ID: groupID.Hex()},
			Service:     "chat",
			Status:      "FAILED",
			Description: "Invalid input body",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	if err := h.ChatService.Repo().AddMemberToGroup(c.Request.Context(), groupID, &member); err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "ADD_GROUP_MEMBER",
			Actor:       actor,
			Target:      auditlog.Target{Type: "group", ID: groupID.Hex()},
			Service:     "chat",
			Status:      "FAILED",
			Description: "Failed to add member: " + err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add member", "details": err.Error()})
		return
	}

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action:      "ADD_GROUP_MEMBER",
		Actor:       actor,
		Target:      auditlog.Target{Type: "group", ID: groupID.Hex()},
		Service:     "chat",
		Status:      "SUCCESS",
		Description: "Member added to group",
	})

	c.JSON(http.StatusOK, gin.H{"message": "member added"})
}

func (h *ChatHandler) RemoveMemberFromGroup(c *gin.Context) {
	actor := extractActorFromContext(c)

	groupID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "REMOVE_GROUP_MEMBER",
			Actor:       actor,
			Target:      auditlog.Target{Type: "group", ID: c.Param("id")},
			Service:     "chat",
			Status:      "FAILED",
			Description: "Invalid group ID",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group ID"})
		return
	}

	email := c.Param("email")
	if err := h.ChatService.Repo().RemoveMemberFromGroup(c.Request.Context(), groupID, email); err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "REMOVE_GROUP_MEMBER",
			Actor:       actor,
			Target:      auditlog.Target{Type: "group", ID: groupID.Hex()},
			Service:     "chat",
			Status:      "FAILED",
			Description: "Failed to remove member: " + err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove member", "details": err.Error()})
		return
	}

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action:      "REMOVE_GROUP_MEMBER",
		Actor:       actor,
		Target:      auditlog.Target{Type: "group", ID: groupID.Hex()},
		Service:     "chat",
		Status:      "SUCCESS",
		Description: "Member removed from group",
	})

	c.JSON(http.StatusOK, gin.H{"message": "member removed"})
}

// ───── Messages ─────────────────────────────────────────────

func (h *ChatHandler) SendMessage(c *gin.Context) {
	actor := extractActorFromContext(c)

	groupID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "SEND_GROUP_MESSAGE",
			Actor:       actor,
			Target:      auditlog.Target{Type: "group", ID: c.Param("id")},
			Service:     "chat",
			Status:      "FAILED",
			Description: "Invalid group ID",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group ID"})
		return
	}

	var msg chat.Message
	if err := c.ShouldBindJSON(&msg); err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "SEND_GROUP_MESSAGE",
			Actor:       actor,
			Target:      auditlog.Target{Type: "group", ID: groupID.Hex()},
			Service:     "chat",
			Status:      "FAILED",
			Description: "Invalid message input",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}
	msg.GroupID = groupID

	if err := h.ChatService.Repo().CreateMessage(c.Request.Context(), &msg); err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "SEND_GROUP_MESSAGE",
			Actor:       actor,
			Target:      auditlog.Target{Type: "group", ID: groupID.Hex()},
			Service:     "chat",
			Status:      "FAILED",
			Description: "Failed to create message: " + err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create message", "details": err.Error()})
		return
	}

	_ = h.ChatService.WsManager().BroadcastToGroup(groupID.Hex(), msg)

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action:      "SEND_GROUP_MESSAGE",
		Actor:       actor,
		Target:      auditlog.Target{Type: "group_message", ID: msg.ID.Hex()},
		Service:     "chat",
		Status:      "SUCCESS",
		Description: "Message sent",
	})
	c.JSON(http.StatusOK, msg)
}

func (h *ChatHandler) GetMessages(c *gin.Context) {
	actor := extractActorFromContext(c)

	groupID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "GET_GROUP_MESSAGES",
			Actor:       actor,
			Target:      auditlog.Target{Type: "group", ID: c.Param("id")},
			Service:     "chat",
			Status:      "FAILED",
			Description: "Invalid group ID",
		})
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
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "GET_GROUP_MESSAGES",
			Actor:       actor,
			Target:      auditlog.Target{Type: "group", ID: groupID.Hex()},
			Service:     "chat",
			Status:      "FAILED",
			Description: "Failed to get messages: " + err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get messages", "details": err.Error()})
		return
	}

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action:      "GET_GROUP_MESSAGES",
		Actor:       actor,
		Target:      auditlog.Target{Type: "group", ID: groupID.Hex()},
		Service:     "chat",
		Status:      "SUCCESS",
		Description: "Retrieved group messages",
	})

	c.JSON(http.StatusOK, messages)
}
