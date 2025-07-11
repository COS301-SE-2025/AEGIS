package routes

import (
	"aegis-api/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterMessageRoutes(r *gin.RouterGroup, h *handlers.MessageHandler) {
	r.POST("/threads/:threadID/messages", h.SendMessage)
	r.GET("/threads/:threadID/messages", h.GetMessagesByThread)
	r.POST("/messages/:messageID/approve", h.ApproveMessage)
	r.POST("/messages/:messageID/reactions", h.AddReaction)
	r.DELETE("/messages/:messageID/reactions/:userID", h.RemoveReaction)
}
