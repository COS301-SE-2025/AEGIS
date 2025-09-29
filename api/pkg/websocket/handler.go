// package websocket

// import (
// 	chatModels "aegis-api/pkg/chatModels"
// 	"log"

// 	"github.com/gin-gonic/gin"
// )

// func RegisterWebSocketRoutes(rg *gin.RouterGroup, manager chatModels.WebSocketManager) {
// 	rg.GET("/cases/:caseId", func(c *gin.Context) {
// 		log.Println("📥 WebSocket route hit")

// 		// This must be handled inside manager.HandleConnection
// 		err := manager.HandleConnection(c.Writer, c.Request)
// 		if err != nil {
// 			log.Println("❌ Failed to handle WebSocket connection:", err)
// 		}
// 	})
// }

package websocket

import (
	chatModels "aegis-api/pkg/chatModels"
	"log"

	"github.com/gin-gonic/gin"
)

func RegisterWebSocketRoutes(rg *gin.RouterGroup, manager chatModels.WebSocketManager) {
	rg.GET("/cases/:caseId", func(c *gin.Context) {
		log.Println("📥 WebSocket route hit")

		// Extract caseId from URL parameter
		caseId := c.Param("caseId")
		log.Printf("📥 Case ID: %s", caseId)

		// Extract user info from context (set by your auth middleware)
		userID := c.GetString("userID")
		tenantID := c.GetString("tenantID")
		email := c.GetString("email")

		log.Printf("📥 User: %s, Tenant: %s, Email: %s", userID, tenantID, email)

		if userID == "" || tenantID == "" {
			log.Println("❌ WebSocket auth failed: missing user context")
			c.JSON(401, gin.H{"error": "authentication required"})
			return
		}

		// Pass the Gin context to the manager
		err := manager.HandleConnection(c.Writer, c.Request, c)
		if err != nil {
			log.Println("❌ Failed to handle WebSocket connection:", err)
		}
	})
}
