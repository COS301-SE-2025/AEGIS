package routes

// import (
// 	"aegis-api/pkg/websocket"

// 	"github.com/gin-gonic/gin"
// )

// // RegisterWebSocketRoutes sets up the WebSocket route and injects the hub
// func RegisterWebSocketRoutes(r *gin.Engine, hub *websocket.Hub) {
// 	r.GET("/ws/cases/:caseID", func(c *gin.Context) {
// 		caseID := c.Param("caseID")

// 		userID, ok := c.Get("userID")
// 		if !ok {
// 			c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
// 			return
// 		}

// 		websocket.ServeWS(hub, c.Writer, c.Request, userID.(string), caseID)
// 	})
// }
