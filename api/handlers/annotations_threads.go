package handlers

import (
	annotationthreads "aegis-api/services_/annotation_threads/threads"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AnnotationThreadHandler struct {
	service annotationthreads.AnnotationThreadService
}

func NewAnnotationThreadHandler(service annotationthreads.AnnotationThreadService) *AnnotationThreadHandler {
	return &AnnotationThreadHandler{service: service}
}

type createThreadRequest struct {
	CaseID   string   `json:"case_id" binding:"required,uuid"`
	FileID   string   `json:"file_id" binding:"required,uuid"`
	UserID   string   `json:"user_id" binding:"required,uuid"`
	Title    string   `json:"title" binding:"required"`
	Tags     []string `json:"tags"`
	Priority string   `json:"priority"`
}

func (h *AnnotationThreadHandler) CreateThread(c *gin.Context) {
	var req createThreadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	caseID, _ := uuid.Parse(req.CaseID)
	fileID, _ := uuid.Parse(req.FileID)
	userID, _ := uuid.Parse(req.UserID)
	priority := annotationthreads.ThreadPriority(req.Priority)
	if priority == "" {
		priority = annotationthreads.PriorityMedium
	}

	thread, err := h.service.CreateThread(caseID, fileID, userID, req.Title, req.Tags, priority)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, thread)
}

func (h *AnnotationThreadHandler) GetThreadsByFile(c *gin.Context) {
	fileID, err := uuid.Parse(c.Param("fileID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file ID"})
		return
	}

	threads, err := h.service.GetThreadsByFile(fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, threads)
}

func (h *AnnotationThreadHandler) GetThreadsByCase(c *gin.Context) {
	caseID, err := uuid.Parse(c.Param("caseID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case ID"})
		return
	}

	threads, err := h.service.GetThreadsByCase(caseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, threads)
}

func (h *AnnotationThreadHandler) GetThreadByID(c *gin.Context) {
	threadID, err := uuid.Parse(c.Param("threadID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid thread ID"})
		return
	}

	thread, err := h.service.GetThreadByID(threadID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, thread)
}

type updateStatusRequest struct {
	Status string `json:"status" binding:"required"`
	UserID string `json:"user_id" binding:"required,uuid"`
}

func (h *AnnotationThreadHandler) UpdateThreadStatus(c *gin.Context) {
	var req updateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	threadID, _ := uuid.Parse(c.Param("threadID"))
	userID, _ := uuid.Parse(req.UserID)

	err := h.service.UpdateThreadStatus(threadID, annotationthreads.ThreadStatus(req.Status), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "status updated"})
}

type updatePriorityRequest struct {
	Priority string `json:"priority" binding:"required"`
	UserID   string `json:"user_id" binding:"required,uuid"`
}

func (h *AnnotationThreadHandler) UpdateThreadPriority(c *gin.Context) {
	var req updatePriorityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	threadID, _ := uuid.Parse(c.Param("threadID"))
	userID, _ := uuid.Parse(req.UserID)

	err := h.service.UpdateThreadPriority(threadID, annotationthreads.ThreadPriority(req.Priority), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "priority updated"})
}

type addParticipantRequest struct {
	UserID string `json:"user_id" binding:"required,uuid"`
}

func (h *AnnotationThreadHandler) AddParticipant(c *gin.Context) {
	var req addParticipantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	threadID, _ := uuid.Parse(c.Param("threadID"))
	userID, _ := uuid.Parse(req.UserID)

	err := h.service.AddParticipant(threadID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "participant added"})
}

func (h *AnnotationThreadHandler) GetThreadParticipants(c *gin.Context) {
	threadID, err := uuid.Parse(c.Param("threadID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid thread ID"})
		return
	}

	participants, err := h.service.GetThreadParticipants(threadID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, participants)
}
