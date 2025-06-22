package websocket

import (
	"github.com/gin-gonic/gin"
)

func RegisterWebSocketRoutes(r *gin.Engine, hub *Hub) {
	r.GET("/ws/cases/:caseID", func(c *gin.Context) {
		caseID := c.Param("caseID")
		userID := c.GetString("userID") // Requires middleware to set
		ServeWS(hub, c.Writer, c.Request, userID, caseID)
	})
}
