package handlers

import (
	annotationthreads "aegis-api/services_/annotation_threads/threads"
	"aegis-api/structs"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

type ThreadHandler struct {
	threads *annotationthreads.Annotationthreadservice
}

func NewThread(
	threads *annotationthreads.Annotationthreadservice,
) *ThreadHandler {
	return &ThreadHandler{
		threads: threads,
	}
}

// CreateThread creates a new annotation thread
// @Summary Create a new annotation thread
// @Description Creates a new annotation thread for a specific case and file
// @Tags Annotation Threads
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param case_id path string true "Case ID"
// @Param request body structs.CreateThreadRequest true "Thread Creation Request"
// @Success 201 {object} structs.SuccessResponse{data=annotationthreads.AnnotationThread} "Thread created successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid request data"
// @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/cases/{case_id}/threads [post]
func (t *ThreadHandler) CreateThread(c *gin.Context) {
	var req structs.CreateThreadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request data",
			Details: err.Error(),
		})
		return
	}

	caseIDStr := c.Param("case_id") //from url
	caseID, err := uuid.Parse(caseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_case_id",
			Message: "Invalid case ID format",
		})
	}

	userIDStr, exists := c.Get("userID") //from middleware
	if !exists {
		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
		})
		return
	}

	//call the service
	newThread, err := t.threads.CreateThread(caseID, req.FileID, userID, req.Title, req.Tags, annotationthreads.ThreadPriority(req.Priority))

	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "thread_creation_failed",
			Message: "Failed to create thread",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, structs.SuccessResponse{
		Success: true,
		Message: "Thread created successfully",
		Data:    newThread,
	})
}

// AddParticipant adds a participant to a thread
// @Summary Add participant to thread
// @Description Adds a user as a participant to an annotation thread
// @Tags Annotation Threads
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param case_id path string true "Case ID"
// @Param thread_id path string true "Thread ID"
// @Param request body structs.AddParticipantRequest true "Add Participant Request"
// @Success 200 {object} structs.SuccessResponse "Participant added successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid request data"
// @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/cases/{case_id}/threads/{thread_id}/participants [post]
func (t *ThreadHandler) AddParticipant(c *gin.Context) {
	var req structs.AddParticipantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request data",
			Details: err.Error(),
		})
	}

	threadIDStr := c.Param("thread_id")
	threadID, err := uuid.Parse(threadIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_thread_id",
			Message: "Invalid thread ID format",
		})
	}

	err = t.threads.AddParticipant(threadID, req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "participant_add_failed",
			Message: "Failed to add participant",
			Details: err.Error(),
		})
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Participant added successfully",
	})
}

// GetThreadsByFileID gets threads by file ID
// @Summary Get threads by file ID
// @Description Retrieves all annotation threads associated with a specific file
// @Tags Annotation Threads
// @Produce json
// @Security ApiKeyAuth
// @Param case_id path string true "Case ID"
// @Param file_id path string true "File ID"
// @Success 200 {object} structs.SuccessResponse{data=[]annotationthreads.AnnotationThread} "Threads retrieved successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid file ID"
// @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/cases/{case_id}/threads/by-file/{file_id} [get]
func (t *ThreadHandler) GetThreadsByFileID(c *gin.Context) {
	fileIDStr := c.Param("file_id")
	fileID, err := uuid.Parse(fileIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_file_id",
			Message: "Invalid file ID format",
		})
		return
	}

	threads, err := t.threads.GetThreadsByFile(fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "thread_fetch_failed",
			Message: "Failed to get threads",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Threads retrieved successfully",
		Data:    threads,
	})
}

// GetThreadsByCaseID gets threads by case ID
// @Summary Get threads by case ID
// @Description Retrieves all annotation threads associated with a specific case
// @Tags Annotation Threads
// @Produce json
// @Security ApiKeyAuth
// @Param case_id path string true "Case ID"
// @Success 200 {object} structs.SuccessResponse{data=[]annotationthreads.AnnotationThread} "Threads retrieved successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid case ID"
// @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/cases/{case_id}/threads [get]
func (t *ThreadHandler) GetThreadsByCaseID(c *gin.Context) {
	caseIDStr := c.Param("case_id")
	caseID, err := uuid.Parse(caseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_case_id",
			Message: "Invalid case ID format",
		})
		return
	}

	threads, err := t.threads.GetThreadsByCase(caseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "thread_fetch_failed",
			Message: "Failed to get threads",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Threads retrieved successfully",
		Data:    threads,
	})
}

// GetThreadParticipants gets participants of a thread
// @Summary Get thread participants
// @Description Retrieves all participants of a specific annotation thread
// @Tags Annotation Threads
// @Produce json
// @Security ApiKeyAuth
// @Param case_id path string true "Case ID"
// @Param thread_id path string true "Thread ID"
// @Success 200 {object} structs.SuccessResponse{data=[]annotationthreads.ThreadParticipant} "Participants retrieved successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid thread ID"
// @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/cases/{case_id}/threads/{thread_id}/participants [get]
func (t *ThreadHandler) GetThreadParticipants(c *gin.Context) {
	threadIDStr := c.Param("thread_id")
	threadID, err := uuid.Parse(threadIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_thread_id",
			Message: "Invalid thread ID format",
		})
		return
	}

	participants, err := t.threads.GetThreadParticipants(threadID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "participant_fetch_failed",
			Message: "Failed to get participants",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Participants retrieved successfully",
		Data:    participants,
	})
}

// GetThreadByID gets a thread by ID
// @Summary Get thread by ID
// @Description Retrieves a specific annotation thread by its ID
// @Tags Annotation Threads
// @Produce json
// @Security ApiKeyAuth
// @Param case_id path string true "Case ID"
// @Param thread_id path string true "Thread ID"
// @Success 200 {object} structs.SuccessResponse{data=annotationthreads.AnnotationThread} "Thread retrieved successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid thread ID"
// @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// @Failure 404 {object} structs.ErrorResponse "Thread not found"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/cases/{case_id}/threads/{thread_id} [get]
func (t *ThreadHandler) GetThreadByID(c *gin.Context) {
	threadIDStr := c.Param("thread_id")
	threadID, err := uuid.Parse(threadIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_thread_id",
			Message: "Invalid thread ID format",
		})
		return
	}

	thread, err := t.threads.GetThreadByID(threadID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "thread_fetch_failed",
			Message: "Failed to get thread",
			Details: err.Error(),
		})
		return
	}

	if thread == nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Error:   "thread_not_found",
			Message: "Thread not found",
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Thread retrieved successfully",
		Data:    thread,
	})
}

// GetUserByID gets a user by ID
// @Summary Get user by ID
// @Description Retrieves a specific user by their ID
// @Tags Users
// @Produce json
// @Security ApiKeyAuth
// @Param user_id path string true "User ID"
// @Success 200 {object} structs.SuccessResponse{data=annotationthreads.User} "User retrieved successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid user ID"
// @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// @Failure 404 {object} structs.ErrorResponse "User not found"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/users/{user_id} [get]
func (t *ThreadHandler) GetUserByID(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
		})
		return
	}

	user, err := t.threads.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "user_fetch_failed",
			Message: "Failed to get user",
			Details: err.Error(),
		})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Error:   "user_not_found",
			Message: "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "User retrieved successfully",
		Data:    user,
	})
}

// UpdateThreadStatus updates the status of a thread
// @Summary Update thread status
// @Description Updates the status of a specific annotation thread
// @Tags Annotation Threads
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param case_id path string true "Case ID"
// @Param thread_id path string true "Thread ID"
// @Param request body structs.UpdateThreadStatusRequest true "Update Thread Status Request"
// @Success 200 {object} structs.SuccessResponse "Thread status updated successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid request data"
// @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/cases/{case_id}/threads/{thread_id}/status [put]
func (t *ThreadHandler) UpdateThreadStatus(c *gin.Context) {
	var req structs.UpdateThreadStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request data",
			Details: err.Error(),
		})
		return
	}

	threadIDStr := c.Param("thread_id")
	threadID, err := uuid.Parse(threadIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_thread_id",
			Message: "Invalid thread ID format",
		})
		return
	}

	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
		})
		return
	}

	err = t.threads.UpdateThreadStatus(threadID, annotationthreads.ThreadStatus(req.Status), userID)
	if err != nil {
		if err.Error() == "only lead investigators can update thread status" {
			c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
				Error:   "permission_denied",
				Message: "Only lead investigators can update thread status",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "status_update_failed",
			Message: "Failed to update thread status",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Thread status updated successfully",
	})
}

// UpdateThreadPriority updates the priority of a thread
// @Summary Update thread priority
// @Description Updates the priority of a specific annotation thread
// @Tags Annotation Threads
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param case_id path string true "Case ID"
// @Param thread_id path string true "Thread ID"
// @Param request body structs.UpdateThreadPriorityRequest true "Update Thread Priority Request"
// @Success 200 {object} structs.SuccessResponse "Thread priority updated successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid request data"
// @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/cases/{case_id}/threads/{thread_id}/priority [put]
func (t *ThreadHandler) UpdateThreadPriority(c *gin.Context) {
	var req structs.UpdateThreadPriorityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request data",
			Details: err.Error(),
		})
	}

	threadIDStr := c.Param("thread_id")
	threadID, err := uuid.Parse(threadIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_thread_id",
			Message: "Invalid thread ID format",
		})
	}

	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
		})
		return
	}

	err = t.threads.UpdateThreadPriority(threadID, annotationthreads.ThreadPriority(req.Priority), userID)
	if err != nil {
		if err.Error() == "only lead investigators can update thread priority" {
			c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
				Error:   "permission_denied",
				Message: "Only lead investigators can update thread priority",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "priority_update_failed",
			Message: "Failed to update thread priority",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Thread priority updated successfully",
	})
}
