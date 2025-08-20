package routes

import (
	"aegis-api/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterThreadRoutes(r *gin.RouterGroup, h *handlers.AnnotationThreadHandler) {
	r.POST("/threads", h.CreateThread)
	r.GET("/threads/file/:fileID", h.GetThreadsByFile)
	r.GET("/threads/case/:caseID", h.GetThreadsByCase)
	r.GET("/threads/:threadID", h.GetThreadByID)
	r.PATCH("/threads/:threadID/status", h.UpdateThreadStatus)
	r.PATCH("/threads/:threadID/priority", h.UpdateThreadPriority)
	r.POST("/threads/:threadID/participants", h.AddParticipant)
	r.GET("/threads/:threadID/participants", h.GetThreadParticipants)
}