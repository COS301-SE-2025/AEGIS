package handlers

import (
	timelineai "aegis-api/services_/timeline/timeline_ai"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TimelineAIHandler struct {
	Service timelineai.AIService
}

func NewTimelineAIHandler(service timelineai.AIService) *TimelineAIHandler {
	return &TimelineAIHandler{
		Service: service,
	}
}

// GetEventSuggestions godoc
// @Summary Get AI suggestions for timeline events
// @Description Generate AI suggestions for investigation timeline events
// @Tags timeline, ai
// @Accept json
// @Produce json
// @Param body body timelineai.SuggestionRequest true "Suggestion request"
// @Success 200 {object} timelineai.AIAnalysisResult
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /timeline/ai/suggestions [post]
func (h *TimelineAIHandler) GetEventSuggestions(c *gin.Context) {
	var req timelineai.SuggestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	result, err := h.Service.GetEventSuggestions(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetSeverityRecommendation godoc
// @Summary Get AI severity recommendation
// @Description Get AI recommendation for event severity level
// @Tags timeline, ai
// @Accept json
// @Produce json
// @Param body body map[string]string true "Event description"
// @Success 200 {object} timelineai.SeverityRecommendationDTO
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /timeline/ai/severity [post]
func (h *TimelineAIHandler) GetSeverityRecommendation(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	description, ok := req["description"]
	if !ok || description == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "description is required"})
		return
	}

	severity, confidence, err := h.Service.GetSeverityRecommendation(c.Request.Context(), description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, timelineai.SeverityRecommendationDTO{
		Success:             true,
		RecommendedSeverity: severity,
		Confidence:          confidence,
	})
}

// GetTagSuggestions godoc
// @Summary Get AI tag suggestions
// @Description Get AI suggestions for event tags
// @Tags timeline, ai
// @Accept json
// @Produce json
// @Param body body map[string]string true "Event description"
// @Success 200 {object} timelineai.TagSuggestionDTO
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /timeline/ai/tags [post]
func (h *TimelineAIHandler) GetTagSuggestions(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	description, ok := req["description"]
	if !ok || description == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "description is required"})
		return
	}

	tags, err := h.Service.GetTagSuggestions(c.Request.Context(), description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, timelineai.TagSuggestionDTO{
		Success: true,
		Tags:    tags,
	})
}

// GetNextSteps godoc
// @Summary Get AI next steps suggestions
// @Description Get AI suggestions for next investigation steps
// @Tags timeline, ai
// @Accept json
// @Produce json
// @Param caseID path string true "Case ID"
// @Success 200 {object} timelineai.SuggestionResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /timeline/ai/cases/{caseID}/next-steps [get]
func (h *TimelineAIHandler) GetNextSteps(c *gin.Context) {
	caseID := c.Param("caseID")
	if caseID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "caseID is required"})
		return
	}

	steps, err := h.Service.SuggestNextSteps(c.Request.Context(), caseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, timelineai.SuggestionResponseDTO{
		Success:     true,
		Suggestions: steps,
	})
}

// AnalyzeEvent godoc
// @Summary Analyze event context
// @Description Comprehensive AI analysis of an event
// @Tags timeline, ai
// @Accept json
// @Produce json
// @Param body body map[string]string true "Analysis request"
// @Success 200 {object} timelineai.ContextAnalysisDTO
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /timeline/ai/analyze-event [post]
func (h *TimelineAIHandler) AnalyzeEvent(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	caseID, caseOk := req["case_id"]
	eventText, eventOk := req["event_text"]

	if !caseOk || !eventOk || caseID == "" || eventText == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "case_id and event_text are required"})
		return
	}

	result, err := h.Service.AnalyzeEventContext(c.Request.Context(), caseID, eventText)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to DTO
	dto := timelineai.ContextAnalysisDTO{
		Success:             true,
		RecommendedSeverity: result.RecommendedSeverity,
		SuggestedTags:       result.RecommendedTags,
		ExtractedIOCs:       result.ExtractedIOCs,
		Suggestions:         result.Suggestions,
		Confidence:          result.Confidence,
	}

	c.JSON(http.StatusOK, dto)
}

// AnalyzeCaseProgress godoc
// @Summary Analyze case progress
// @Description Get AI analysis of case completion and recommendations
// @Tags timeline, ai
// @Accept json
// @Produce json
// @Param caseID path string true "Case ID"
// @Success 200 {object} timelineai.CaseAnalysisDTO
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /timeline/ai/cases/{caseID}/progress [get]
func (h *TimelineAIHandler) AnalyzeCaseProgress(c *gin.Context) {
	caseID := c.Param("caseID")
	if caseID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "caseID is required"})
		return
	}

	analysis, err := h.Service.AnalyzeCaseProgress(c.Request.Context(), caseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to DTO
	dto := timelineai.CaseAnalysisDTO{
		Success:            true,
		CaseID:             analysis.CaseID,
		CompletionScore:    analysis.CompletionScore,
		MissingSteps:       analysis.MissingSteps,
		RecommendedActions: analysis.RecommendedActions,
		RiskAssessment:     analysis.RiskAssessment,
		AnalyzedAt:         analysis.AnalyzedAt,
	}

	c.JSON(http.StatusOK, dto)
}

// CorrelateEvidence godoc
// @Summary Correlate evidence with events
// @Description Get AI suggestions for evidence correlation with timeline events
// @Tags timeline, ai
// @Accept json
// @Produce json
// @Param body body map[string]string true "Correlation request"
// @Success 200 {object} timelineai.EvidenceCorrelationDTO
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /timeline/ai/correlate-evidence [post]
func (h *TimelineAIHandler) CorrelateEvidence(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	caseID, caseOk := req["case_id"]
	eventDescription, eventOk := req["event_description"]

	if !caseOk || !eventOk || caseID == "" || eventDescription == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "case_id and event_description are required"})
		return
	}

	correlatedEvidence, err := h.Service.CorrelateEvidence(c.Request.Context(), caseID, eventDescription)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, timelineai.EvidenceCorrelationDTO{
		Success:            true,
		CorrelatedEvidence: correlatedEvidence,
	})
}

// SubmitFeedback godoc
// @Summary Submit AI feedback
// @Description Submit user feedback on AI suggestions
// @Tags timeline, ai
// @Accept json
// @Produce json
// @Param body body timelineai.AIFeedback true "Feedback data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /timeline/ai/feedback [post]
func (h *TimelineAIHandler) SubmitFeedback(c *gin.Context) {
	var feedback timelineai.AIFeedback
	if err := c.ShouldBindJSON(&feedback); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid feedback payload"})
		return
	}

	// Validate required fields
	if feedback.AnalysisID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "analysis_id is required"})
		return
	}
	if feedback.FeedbackType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "feedback_type is required"})
		return
	}

	err := h.Service.RecordFeedback(c.Request.Context(), feedback.AnalysisID, &feedback)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "feedback recorded successfully"})
}

// GetModelStatus godoc
// @Summary Get AI model status
// @Description Check the status and health of the AI model
// @Tags timeline, ai
// @Accept json
// @Produce json
// @Success 200 {object} timelineai.ModelStatusDTO
// @Failure 500 {object} map[string]string
// @Router /timeline/ai/status [get]
func (h *TimelineAIHandler) GetModelStatus(c *gin.Context) {
	status, err := h.Service.GetModelStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to DTO
	dto := timelineai.ModelStatusDTO{
		Success:      true,
		ModelName:    status.ModelName,
		Status:       status.Status,
		LastChecked:  status.LastChecked,
		ResponseTime: status.ResponseTime,
		ErrorMessage: status.ErrorMessage,
	}

	c.JSON(http.StatusOK, dto)
}

// UpdateModelConfig godoc
// @Summary Update AI model configuration
// @Description Update AI model settings and parameters
// @Tags timeline, ai
// @Accept json
// @Produce json
// @Param body body timelineai.AIModelConfig true "Configuration data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /timeline/ai/config [put]
func (h *TimelineAIHandler) UpdateModelConfig(c *gin.Context) {
	var config timelineai.AIModelConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid configuration payload"})
		return
	}

	err := h.Service.UpdateModelConfig(c.Request.Context(), &config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "configuration updated successfully"})
}

// ExtractIOCs godoc
// @Summary Extract Indicators of Compromise (IOCs)
// @Description Extract IOCs such as IP addresses, domains, file hashes, and emails from text input.
// Supports both regex-based and AI-enhanced extraction.
// @Tags ai, iocs
// @Accept json
// @Produce json
// @Param body body struct{Text string `json:"text"`} true "Text to analyze for IOCs"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /ai/iocs/extract [post]
func (h *TimelineAIHandler) ExtractIOCs(c *gin.Context) {
	var req struct {
		Text string `json:"text" binding:"required"`
	}

	// Parse input
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Call service
	iocs, err := h.Service.ExtractIOCs(c.Request.Context(), req.Text)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to extract IOCs: %v", err),
		})
		return
	}

	// Respond with extracted IOCs
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"iocs":    iocs,
	})
}
