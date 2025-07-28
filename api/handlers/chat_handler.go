package handlers

import (
	"aegis-api/services_/chat"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

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
	actor := auditlog.MakeActor(c)

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

	// 🛡️ Validate that CaseID is provided
	if group.CaseID == "" {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "CREATE_GROUP",
			Actor:       actor,
			Target:      auditlog.Target{Type: "group", ID: ""},
			Service:     "chat",
			Status:      "FAILED",
			Description: "Missing case ID for group creation",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "case ID is required to create a group"})
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
	actor := auditlog.MakeActor(c)

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
	actor := auditlog.MakeActor(c)
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
	// Ensure non-nil slice
	if groups == nil {
		groups = []*chat.ChatGroup{}
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
	actor := auditlog.MakeActor(c)

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
	actor := auditlog.MakeActor(c)

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
	actor := auditlog.MakeActor(c)

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
	actor := auditlog.MakeActor(c)

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
	actor := auditlog.MakeActor(c)
	email, _ := c.Get("email")
	fullNameVal, _ := c.Get("fullName")
	fullName := fullNameVal.(string)

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

	var req struct {
		Content  string `json:"content"`
		File     string `json:"file"`     // base64 encoded
		FileName string `json:"fileName"` // e.g. "photo.png"
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "SEND_GROUP_MESSAGE",
			Actor:       actor,
			Target:      auditlog.Target{Type: "group", ID: groupID.Hex()},
			Service:     "chat",
			Status:      "FAILED",
			Description: "Invalid input JSON",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	var attachments []*chat.Attachment
	messageType := "text"

	if req.File != "" && req.FileName != "" {
		data, err := base64.StdEncoding.DecodeString(req.File)
		if err != nil {
			h.auditLogger.Log(c, auditlog.AuditLog{
				Action:      "SEND_GROUP_MESSAGE",
				Actor:       actor,
				Target:      auditlog.Target{Type: "group", ID: groupID.Hex()},
				Service:     "chat",
				Status:      "FAILED",
				Description: "Invalid base64 file encoding",
			})
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid base64 file"})
			return
		}
		contentType := http.DetectContentType(data[:512])

		ext := filepath.Ext(req.FileName)
		switch ext {
		case ".docx":
			contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
		case ".pptx":
			contentType = "application/vnd.openxmlformats-officedocument.presentationml.presentation"
		case ".xlsx":
			contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		}

		fallbackType := mime.TypeByExtension(ext)
		if contentType == "application/octet-stream" && fallbackType != "" {
			contentType = fallbackType
		}
		fileSize := int64(len(data))

		result, err := h.ChatService.IPFSUploader().UploadBytes(
			c.Request.Context(),
			data,
			fmt.Sprintf("%s-%s", primitive.NewObjectID().Hex(), req.FileName),
		)
		if err != nil {
			h.auditLogger.Log(c, auditlog.AuditLog{
				Action:      "SEND_GROUP_MESSAGE",
				Actor:       actor,
				Target:      auditlog.Target{Type: "group", ID: groupID.Hex()},
				Service:     "chat",
				Status:      "FAILED",
				Description: "IPFS upload failed: " + err.Error(),
			})
			c.JSON(http.StatusInternalServerError, gin.H{"error": "file upload failed", "details": err.Error()})
			return
		}

		fileUrl := h.ChatService.IPFSUploader().GetFileURL(result.Hash)

		attachments = []*chat.Attachment{
			{
				ID:       primitive.NewObjectID().Hex(),
				FileName: req.FileName,
				FileType: contentType,
				FileSize: fileSize,
				URL:      fileUrl,
				Hash:     result.Hash,
			},
		}
		messageType = "file"
	}

	msg := &chat.Message{
		GroupID:       groupID,
		SenderEmail:   email.(string),
		SenderName:    fullName, // if you have `Name` on actor
		Content:       req.Content,
		MessageType:   messageType,
		Attachments:   attachments,
		Status:        chat.MessageStatus{Sent: time.Now()},
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		IsDeleted:     false,
		AttachmentURL: "", // keep empty unless you want a direct link outside Attachments
	}

	if err := h.ChatService.Repo().CreateMessage(c.Request.Context(), msg); err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "SEND_GROUP_MESSAGE",
			Actor:       actor,
			Target:      auditlog.Target{Type: "group", ID: groupID.Hex()},
			Service:     "chat",
			Status:      "FAILED",
			Description: "Failed to save message: " + err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save message", "details": err.Error()})
		return
	}

	_ = h.ChatService.WsManager().BroadcastToGroup(groupID.Hex(), chat.WebSocketMessage{
		Type:      "NEW_MESSAGE", // ✅ correct message type
		GroupID:   groupID.Hex(),
		Payload:   msg, // ✅ fixed field name
		Timestamp: time.Now(),
	})

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action:      "SEND_GROUP_MESSAGE",
		Actor:       actor,
		Target:      auditlog.Target{Type: "group_message", ID: msg.ID},
		Service:     "chat",
		Status:      "SUCCESS",
		Description: "Message sent",
	})

	c.JSON(http.StatusOK, msg)
}
func (h *ChatHandler) GetGroupsByCaseID(c *gin.Context) {
	actor := auditlog.MakeActor(c)

	caseIDHex := c.Param("caseId")
	caseID, err := primitive.ObjectIDFromHex(caseIDHex)
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "GET_GROUPS_BY_CASE",
			Actor:       actor,
			Target:      auditlog.Target{Type: "case", ID: caseIDHex},
			Service:     "chat",
			Status:      "FAILED",
			Description: "Invalid case ID",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case ID"})
		return
	}

	groups, err := h.ChatService.Repo().GetGroupsByCaseID(c.Request.Context(), caseID)
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "GET_GROUPS_BY_CASE",
			Actor:       actor,
			Target:      auditlog.Target{Type: "case", ID: caseID.Hex()},
			Service:     "chat",
			Status:      "FAILED",
			Description: "Failed to retrieve groups: " + err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve groups", "details": err.Error()})
		return
	}

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action:      "GET_GROUPS_BY_CASE",
		Actor:       actor,
		Target:      auditlog.Target{Type: "case", ID: caseID.Hex()},
		Service:     "chat",
		Status:      "SUCCESS",
		Description: "Retrieved groups by case ID",
	})
	c.JSON(http.StatusOK, groups)
}

func (h *ChatHandler) GetMessages(c *gin.Context) {
	actor := auditlog.MakeActor(c)

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
	// Ensure we return [] not null
	if messages == nil {
		messages = []*chat.Message{} // update to actual message type if needed
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

// UpdateGroupImage handles updating a group's image.
func (h *ChatHandler) UpdateGroupImage(c *gin.Context) {
	extractActorFromContext(c)

	groupID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group ID"})
		return
	}

	file, err := c.FormFile("group_url")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer src.Close()

	// Read file into []byte
	data, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	// Upload to IPFS or your storage
	result, err := h.ChatService.IPFSUploader().UploadBytes(
		c.Request.Context(),
		data,
		file.Filename,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "image upload failed", "details": err.Error()})
		return
	}
	imageURL := h.ChatService.IPFSUploader().GetFileURL(result.Hash)

	// Update group image in DB
	if err := h.ChatService.Repo().UpdateGroupImage(c.Request.Context(), groupID, imageURL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update group image", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"group_url": imageURL})
}
