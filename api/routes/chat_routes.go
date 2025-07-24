package routes

import (
	"aegis-api/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterChatRoutes(router *gin.RouterGroup, handler *handlers.ChatHandler) {
	chat := router.Group("/chat")
	{
		// Groups
		chat.POST("/groups", handler.CreateGroup)
		chat.GET("/groups/:id", handler.GetGroupByID)
		chat.GET("/groups/user/:email", handler.GetUserGroups)
		chat.PUT("/groups/:id", handler.UpdateGroup)
		chat.DELETE("/groups/:id", handler.DeleteGroup)
		chat.POST("/groups/:id/members", handler.AddMemberToGroup)
		chat.DELETE("/groups/:id/members/:email", handler.RemoveMemberFromGroup)
		chat.GET("/groups/case/:caseId", handler.GetGroupsByCaseID)
		chat.PUT("/groups/:id/image", handler.UpdateGroupImage)
		// Messages
		chat.POST("/groups/:id/messages", handler.SendMessage)
		chat.GET("/groups/:id/messages", handler.GetMessages)
	}
}
