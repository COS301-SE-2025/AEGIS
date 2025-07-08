package routes

import (
	"aegis-api/handlers"
	"aegis-api/services_/annotation_threads/messages"

	"github.com/gin-gonic/gin"
)

func RegisterMessageRoutes(r *gin.RouterGroup, svc messages.MessageService) {
	h := handlers.NewMessageHandler(svc)

	r.POST("/threads/:threadID/messages", h.SendMessage)
	r.GET("/threads/:threadID/messages", h.GetMessagesByThread)
	r.POST("/messages/:messageID/approve", h.ApproveMessage)
	r.POST("/messages/:messageID/reactions", h.AddReaction)
	r.DELETE("/messages/:messageID/reactions/:userID", h.RemoveReaction)
}
