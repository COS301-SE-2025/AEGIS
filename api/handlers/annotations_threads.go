package handlers

import (
	annotationthreads "aegis-api/services_/annotation_threads/threads"
	"aegis-api/services_/auditlog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AnnotationThreadHandler struct {
	service     annotationthreads.AnnotationThreadService
	auditLogger auditlog.AuditLogger
}

func NewAnnotationThreadHandler(service annotationthreads.AnnotationThreadService, auditLogger auditlog.AuditLogger) *AnnotationThreadHandler {
	return &AnnotationThreadHandler{
		service:     service,
		auditLogger: auditLogger,
	}
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
	log := auditlog.AuditLog{
		Action: "CREATE_THREAD",
		Actor:  auditlog.MakeActor(c),
		Target: auditlog.Target{
			Type: "thread",
			ID:   "", // will fill below
			AdditionalInfo: map[string]string{
				"case_id": req.CaseID,
				"file_id": req.FileID,
			},
		},
		Service:     "annotation_threads",
		Description: "Create annotation thread",
		Metadata:    map[string]string{"title": req.Title, "priority": string(priority)},
	}

	if err != nil {
		log.Status = "FAILURE"
		log.Description += ": " + err.Error()
		h.auditLogger.Log(c, log)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Status = "SUCCESS"
	log.Target.ID = thread.ID.String()
	h.auditLogger.Log(c, log)
	c.JSON(http.StatusCreated, thread)
}

func (h *AnnotationThreadHandler) GetThreadsByFile(c *gin.Context) {

	fileID, err := uuid.Parse(c.Param("fileID"))
	log := auditlog.AuditLog{
		Action: "GET_THREADS_BY_FILE",
		Actor:  auditlog.MakeActor(c),
		Target: auditlog.Target{
			Type: "file",
			ID:   c.Param("fileID"),
		},
		Service:     "annotation_threads",
		Description: "Retrieve threads by file",
		Metadata:    map[string]string{},
	}

	if err != nil {
		log.Status = "FAILURE"
		log.Description += ": invalid UUID"
		h.auditLogger.Log(c, log)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file ID"})
		return
	}

	threads, err := h.service.GetThreadsByFile(fileID)
	if err != nil {
		log.Status = "FAILURE"
		log.Description += ": " + err.Error()
		h.auditLogger.Log(c, log)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Status = "SUCCESS"
	h.auditLogger.Log(c, log)
	c.JSON(http.StatusOK, threads)
}

func (h *AnnotationThreadHandler) GetThreadsByCase(c *gin.Context) {

	caseID, err := uuid.Parse(c.Param("caseID"))
	log := auditlog.AuditLog{
		Action: "GET_THREADS_BY_CASE",
		Actor:  auditlog.MakeActor(c),
		Target: auditlog.Target{
			Type: "case",
			ID:   c.Param("caseID"),
		},
		Service:     "annotation_threads",
		Description: "Retrieve threads by case",
		Metadata:    map[string]string{},
	}

	if err != nil {
		log.Status = "FAILURE"
		log.Description += ": invalid UUID"
		h.auditLogger.Log(c, log)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case ID"})
		return
	}

	threads, err := h.service.GetThreadsByCase(caseID)
	if err != nil {
		log.Status = "FAILURE"
		log.Description += ": " + err.Error()
		h.auditLogger.Log(c, log)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Status = "SUCCESS"
	h.auditLogger.Log(c, log)
	c.JSON(http.StatusOK, threads)
}

func (h *AnnotationThreadHandler) GetThreadByID(c *gin.Context) {

	threadID, err := uuid.Parse(c.Param("threadID"))
	log := auditlog.AuditLog{
		Action: "GET_THREAD_BY_ID",
		Actor:  auditlog.MakeActor(c),
		Target: auditlog.Target{
			Type: "thread",
			ID:   c.Param("threadID"),
		},
		Service:     "annotation_threads",
		Description: "Retrieve single thread by ID",
		Metadata:    map[string]string{},
	}

	if err != nil {
		log.Status = "FAILURE"
		log.Description += ": invalid UUID"
		h.auditLogger.Log(c, log)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid thread ID"})
		return
	}

	thread, err := h.service.GetThreadByID(threadID)
	if err != nil {
		log.Status = "FAILURE"
		log.Description += ": " + err.Error()
		h.auditLogger.Log(c, log)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Status = "SUCCESS"
	h.auditLogger.Log(c, log)
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
	log := auditlog.AuditLog{
		Action: "UPDATE_THREAD_STATUS",
		Actor:  auditlog.MakeActor(c),
		Target: auditlog.Target{
			Type: "thread",
			ID:   threadID.String(),
		},
		Service:     "annotation_threads",
		Description: "Update annotation thread status",
		Metadata:    map[string]string{"new_status": req.Status},
	}

	if err != nil {
		log.Status = "FAILURE"
		log.Description += ": " + err.Error()
		h.auditLogger.Log(c, log)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Status = "SUCCESS"
	h.auditLogger.Log(c, log)
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
	log := auditlog.AuditLog{
		Action: "UPDATE_THREAD_PRIORITY",
		Actor:  auditlog.MakeActor(c),
		Target: auditlog.Target{
			Type: "thread",
			ID:   threadID.String(),
		},
		Service:     "annotation_threads",
		Description: "Update annotation thread priority",
		Metadata:    map[string]string{"new_priority": req.Priority},
	}

	if err != nil {
		log.Status = "FAILURE"
		log.Description += ": " + err.Error()
		h.auditLogger.Log(c, log)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Status = "SUCCESS"
	h.auditLogger.Log(c, log)
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
	log := auditlog.AuditLog{
		Action: "ADD_THREAD_PARTICIPANT",
		Actor:  auditlog.MakeActor(c),
		Target: auditlog.Target{
			Type: "thread",
			ID:   threadID.String(),
		},
		Service:     "annotation_threads",
		Description: "Add participant to annotation thread",
		Metadata:    map[string]string{"participant_id": req.UserID},
	}

	if err != nil {
		log.Status = "FAILURE"
		log.Description += ": " + err.Error()
		h.auditLogger.Log(c, log)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Status = "SUCCESS"
	h.auditLogger.Log(c, log)
	c.JSON(http.StatusOK, gin.H{"message": "participant added"})
}

func (h *AnnotationThreadHandler) GetThreadParticipants(c *gin.Context) {

	threadID, err := uuid.Parse(c.Param("threadID"))
	log := auditlog.AuditLog{
		Action: "GET_THREAD_PARTICIPANTS",
		Actor:  auditlog.MakeActor(c),
		Target: auditlog.Target{
			Type: "thread",
			ID:   c.Param("threadID"),
		},
		Service:     "annotation_threads",
		Description: "Retrieve participants of thread",
		Metadata:    map[string]string{},
	}

	if err != nil {
		log.Status = "FAILURE"
		log.Description += ": invalid UUID"
		h.auditLogger.Log(c, log)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid thread ID"})
		return
	}

	participants, err := h.service.GetThreadParticipants(threadID)
	if err != nil {
		log.Status = "FAILURE"
		log.Description += ": " + err.Error()
		h.auditLogger.Log(c, log)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Status = "SUCCESS"
	h.auditLogger.Log(c, log)
	c.JSON(http.StatusOK, participants)
}
