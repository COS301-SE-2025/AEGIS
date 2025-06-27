package handlers

// import (
// 	"aegis-api/services_/annotation_threads/messages"
// 	"aegis-api/structs"
// 	"github.com/gin-gonic/gin"
// 	"github.com/google/uuid"
// 	"net/http"
// )

// type MessageHandler struct {
// 	messages *messages.MessageServiceImpl
// }

// func NewMessageService(
// 	messages *messages.MessageServiceImpl,
// ) *MessageHandler {
// 	return &MessageHandler{
// 		messages: messages,
// 	}
// }

// // SendMessage creates a new message in a thread
// // @Summary Send a message in a thread
// // @Description Creates a new message in a specific annotation thread
// // @Tags Thread Messages
// // @Accept json
// // @Produce json
// // @Security ApiKeyAuth
// // @Param case_id path string true "Case ID"
// // @Param thread_id path string true "Thread ID"
// // @Param request body structs.SendMessageRequest true "Send Message Request"
// // @Success 201 {object} structs.SuccessResponse{data=messages.ThreadMessage} "Message sent successfully"
// // @Failure 400 {object} structs.ErrorResponse "Invalid request data"
// // @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// // @Failure 500 {object} structs.ErrorResponse "Internal server error"
// // @Router /api/v1/cases/{case_id}/threads/{thread_id}/messages [post]
// func (t *MessageHandler) SendMessage(c *gin.Context) {
// 	var req structs.SendMessageRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: "Invalid request data",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	threadIDStr := c.Param("thread_id") //from url
// 	threadID, err := uuid.Parse(threadIDStr)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_thread_id",
// 			Message: "Invalid thread ID format",
// 		})
// 		return
// 	}

// 	userIDStr, exists := c.Get("userID") //from middleware
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
// 			Error:   "unauthorized",
// 			Message: "No authentication provided",
// 		})
// 		return
// 	}

// 	userID, err := uuid.Parse(userIDStr.(string))
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_user_id",
// 			Message: "Invalid user ID format",
// 		})
// 		return
// 	}

// 	//call the service function
// 	message, err := t.messages.SendMessage(threadID, userID, req.Message, req.ParentMessageID, req.Mentions)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 			Error:   "message_send_failed",
// 			Message: "Failed to send message",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusCreated, structs.SuccessResponse{
// 		Success: true,
// 		Message: "Message sent successfully",
// 		Data:    message,
// 	})
// }

// // ApproveMessage approves a message
// // @Summary Approve a message
// // @Description Approves a specific message in a thread
// // @Tags Thread Messages
// // @Produce json
// // @Security ApiKeyAuth
// // @Param case_id path string true "Case ID"
// // @Param thread_id path string true "Thread ID"
// // @Param message_id path string true "Message ID"
// // @Success 200 {object} structs.SuccessResponse "Message approved successfully"
// // @Failure 400 {object} structs.ErrorResponse "Invalid message ID"
// // @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// // @Failure 500 {object} structs.ErrorResponse "Internal server error"
// // @Router /api/v1/cases/{case_id}/threads/{thread_id}/messages/{message_id}/approve [put]
// func (t *MessageHandler) ApproveMessage(c *gin.Context) {
// 	messageIDStr := c.Param("message_id") //from url
// 	messageID, err := uuid.Parse(messageIDStr)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_message_id",
// 			Message: "Invalid message ID format",
// 		})
// 		return
// 	}

// 	userIDStr, exists := c.Get("userID") //from middleware
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
// 			Error:   "unauthorized",
// 			Message: "User not authenticated",
// 		})
// 		return
// 	}

// 	approverID, err := uuid.Parse(userIDStr.(string))
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_user_id",
// 			Message: "Invalid user ID format",
// 		})
// 		return
// 	}

// 	//call the service function
// 	err = t.messages.ApproveMessage(messageID, approverID)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 			Error:   "message_approval_failed",
// 			Message: "Failed to approve message",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, structs.SuccessResponse{
// 		Success: true,
// 		Message: "Message approved successfully",
// 	})
// }

// // AddReaction adds a reaction to a message
// // @Summary Add reaction to message
// // @Description Adds a reaction to a specific message
// // @Tags Thread Messages
// // @Accept json
// // @Produce json
// // @Security ApiKeyAuth
// // @Param case_id path string true "Case ID"
// // @Param thread_id path string true "Thread ID"
// // @Param message_id path string true "Message ID"
// // @Param request body structs.AddReactionRequest true "Add Reaction Request"
// // @Success 200 {object} structs.SuccessResponse "Reaction added successfully"
// // @Failure 400 {object} structs.ErrorResponse "Invalid request data"
// // @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// // @Failure 500 {object} structs.ErrorResponse "Internal server error"
// // @Router /api/v1/cases/{case_id}/threads/{thread_id}/messages/{message_id}/reactions [post]
// func (t *MessageHandler) AddReaction(c *gin.Context) {
// 	var req structs.AddReactionRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: "Invalid request data",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	messageIDStr := c.Param("message_id")
// 	messageID, err := uuid.Parse(messageIDStr)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_message_id",
// 			Message: "Invalid message ID format",
// 		})
// 		return
// 	}

// 	userIDStr, exists := c.Get("userID")
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
// 			Error:   "unauthorized",
// 			Message: "User not authenticated",
// 		})
// 		return
// 	}

// 	userID, err := uuid.Parse(userIDStr.(string))
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_user_id",
// 			Message: "Invalid user ID format",
// 		})
// 		return
// 	}

// 	err = t.messages.AddReaction(messageID, userID, req.Reaction)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 			Error:   "reaction_add_failed",
// 			Message: "Failed to add reaction",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, structs.SuccessResponse{
// 		Success: true,
// 		Message: "Reaction added successfully",
// 	})
// }

// // AddMentions adds mentions to a message
// // @Summary Add mentions to message
// // @Description Adds user mentions to a specific message
// // @Tags Thread Messages
// // @Accept json
// // @Produce json
// // @Security ApiKeyAuth
// // @Param case_id path string true "Case ID"
// // @Param thread_id path string true "Thread ID"
// // @Param message_id path string true "Message ID"
// // @Param request body structs.AddMentionsRequest true "Add Mentions Request"
// // @Success 200 {object} structs.SuccessResponse "Mentions added successfully"
// // @Failure 400 {object} structs.ErrorResponse "Invalid request data"
// // @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// // @Failure 500 {object} structs.ErrorResponse "Internal server error"
// // @Router /api/v1/cases/{case_id}/threads/{thread_id}/messages/{message_id}/mentions [post]
// func (t *MessageHandler) AddMentions(c *gin.Context) {
// 	var req structs.AddMentionsRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: "Invalid request data",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	messageIDStr := c.Param("message_id")
// 	messageID, err := uuid.Parse(messageIDStr)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_message_id",
// 			Message: "Invalid message ID format",
// 		})
// 		return
// 	}

// 	err = t.messages.AddMentions(messageID, req.Mentions)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 			Error:   "mentions_add_failed",
// 			Message: "Failed to add mentions",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, structs.SuccessResponse{
// 		Success: true,
// 		Message: "Mentions added successfully",
// 	})
// }

// // GetMessagesByThreadID gets all messages in a thread
// // @Summary Get messages by thread ID
// // @Description Retrieves all messages in a specific annotation thread
// // @Tags Thread Messages
// // @Produce json
// // @Security ApiKeyAuth
// // @Param case_id path string true "Case ID"
// // @Param thread_id path string true "Thread ID"
// // @Success 200 {object} structs.SuccessResponse{data=[]messages.ThreadMessage} "Messages retrieved successfully"
// // @Failure 400 {object} structs.ErrorResponse "Invalid thread ID"
// // @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// // @Failure 500 {object} structs.ErrorResponse "Internal server error"
// // @Router /api/v1/cases/{case_id}/threads/{thread_id}/messages [get]
// func (t *MessageHandler) GetMessagesByThreadID(c *gin.Context) {
// 	threadIDStr := c.Param("thread_id")
// 	threadID, err := uuid.Parse(threadIDStr)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_thread_id",
// 			Message: "Invalid thread ID format",
// 		})
// 		return
// 	}

// 	messages_, err := t.messages.GetMessagesByThread(threadID)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 			Error:   "message_fetch_failed",
// 			Message: "Failed to get messages",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, structs.SuccessResponse{
// 		Success: true,
// 		Message: "Messages retrieved successfully",
// 		Data:    messages_,
// 	})
// }

// // GetReplies gets all replies to a message
// // @Summary Get message replies
// // @Description Retrieves all replies to a specific message
// // @Tags Thread Messages
// // @Produce json
// // @Security ApiKeyAuth
// // @Param case_id path string true "Case ID"
// // @Param thread_id path string true "Thread ID"
// // @Param message_id path string true "Message ID"
// // @Success 200 {object} structs.SuccessResponse{data=[]messages.ThreadMessage} "Replies retrieved successfully"
// // @Failure 400 {object} structs.ErrorResponse "Invalid message ID"
// // @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// // @Failure 500 {object} structs.ErrorResponse "Internal server error"
// // @Router /api/v1/cases/{case_id}/threads/{thread_id}/messages/{message_id}/replies [get]
// func (t *MessageHandler) GetReplies(c *gin.Context) {
// 	messageIDStr := c.Param("message_id")
// 	messageID, err := uuid.Parse(messageIDStr)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_message_id",
// 			Message: "Invalid message ID format",
// 		})
// 		return
// 	}

// 	replies, err := t.messages.GetReplies(messageID)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 			Error:   "replies_fetch_failed",
// 			Message: "Failed to get replies",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, structs.SuccessResponse{
// 		Success: true,
// 		Message: "Replies retrieved successfully",
// 		Data:    replies,
// 	})
// }

// // GetMessageByID gets a specific message by ID
// // @Summary Get message by ID
// // @Description Retrieves a specific message by its ID
// // @Tags Thread Messages
// // @Produce json
// // @Security ApiKeyAuth
// // @Param case_id path string true "Case ID"
// // @Param thread_id path string true "Thread ID"
// // @Param message_id path string true "Message ID"
// // @Success 200 {object} structs.SuccessResponse{data=messages.ThreadMessage} "Message retrieved successfully"
// // @Failure 400 {object} structs.ErrorResponse "Invalid message ID"
// // @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// // @Failure 404 {object} structs.ErrorResponse "Message not found"
// // @Failure 500 {object} structs.ErrorResponse "Internal server error"
// // @Router /api/v1/cases/{case_id}/threads/{thread_id}/messages/{message_id} [get]
// func (t *MessageHandler) GetMessageByID(c *gin.Context) {
// 	// Get message ID from URL path
// 	messageIDStr := c.Param("message_id")
// 	messageID, err := uuid.Parse(messageIDStr)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_message_id",
// 			Message: "Invalid message ID format",
// 		})
// 		return
// 	}

// 	// Get message by ID
// 	message, err := t.messages.GetMessageByID(messageID)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 			Error:   "message_fetch_failed",
// 			Message: "Failed to get message",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	if message == nil {
// 		c.JSON(http.StatusNotFound, structs.ErrorResponse{
// 			Error:   "message_not_found",
// 			Message: "Message not found",
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, structs.SuccessResponse{
// 		Success: true,
// 		Message: "Message retrieved successfully",
// 		Data:    message,
// 	})
// }

// // RemoveReaction removes a reaction from a message
// // @Summary Remove reaction from message
// // @Description Removes a user's reaction from a specific message
// // @Tags Thread Messages
// // @Produce json
// // @Security ApiKeyAuth
// // @Param case_id path string true "Case ID"
// // @Param thread_id path string true "Thread ID"
// // @Param message_id path string true "Message ID"
// // @Success 200 {object} structs.SuccessResponse "Reaction removed successfully"
// // @Failure 400 {object} structs.ErrorResponse "Invalid message ID"
// // @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// // @Failure 500 {object} structs.ErrorResponse "Internal server error"
// // @Router /api/v1/cases/{case_id}/threads/{thread_id}/messages/{message_id}/reactions [delete]
// func (t *MessageHandler) RemoveReaction(c *gin.Context) {
// 	var req structs.RemoveReactionRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: "Invalid request data",
// 			Details: err.Error(),
// 		})
// 	}

// 	messageIDStr := c.Param("message_id")
// 	messageID, err := uuid.Parse(messageIDStr)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_message_id",
// 			Message: "Invalid message ID format",
// 		})
// 		return
// 	}

// 	userIDStr, exists := c.Get("userID")
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
// 			Error:   "unauthorized",
// 			Message: "User not authenticated",
// 		})
// 		return
// 	}

// 	userID, err := uuid.Parse(userIDStr.(string))
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_user_id",
// 			Message: "Invalid user ID format",
// 		})
// 		return
// 	}

// 	err = t.messages.RemoveReaction(messageID, userID)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 			Error:   "reaction_remove_failed",
// 			Message: "Failed to remove reaction",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, structs.SuccessResponse{
// 		Success: true,
// 		Message: "Reaction removed successfully",
// 	})
// }
