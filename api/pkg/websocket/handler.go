package websocket

import (
	"aegis-api/services_/chat"
	"log"

	"github.com/gin-gonic/gin"
)

func RegisterWebSocketRoutes(rg *gin.RouterGroup, manager chat.WebSocketManager) {
	rg.GET("/cases/:caseId", func(c *gin.Context) {
		log.Println("📥 WebSocket route hit")

		// This must be handled inside manager.HandleConnection
		err := manager.HandleConnection(c.Writer, c.Request)
		if err != nil {
			log.Println("❌ Failed to handle WebSocket connection:", err)
		}
	})
}
