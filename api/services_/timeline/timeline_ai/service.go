package timelineai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type aiService struct {
	repository AIRepository
	config     *AIModelConfig
	httpClient *http.Client
}

// NewAIService creates a new AI service instance
func NewAIService(repository AIRepository, config *AIModelConfig) AIService {
	return &aiService{
		repository: repository,
		config:     config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// OllamaRequest represents a request to Ollama API
type OllamaRequest struct {
	Model   string                 `json:"model"`
	Prompt  string                 `json:"prompt"`
	Stream  bool                   `json:"stream"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// OllamaResponse represents a response from Ollama API
type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func (s *aiService) GetEventSuggestions(ctx context.Context, req *SuggestionRequest) (*AIAnalysisResult, error) {
	if req.SuggestionType != "completion" && req.SuggestionType != "next_steps" {
		return nil, fmt.Errorf("unsupported suggestion type: %s", req.SuggestionType)
	}

	if !s.config.Enabled {
		// Fallbacks
		var fallback []string
		if req.SuggestionType == "completion" {
			fallback, _ = s.getPatternBasedSuggestions(ctx, req)
		} else {
			fallback = s.getDefaultNextSteps(req)
		}
		return &AIAnalysisResult{
			CaseID:       req.CaseID,
			EventID:      req.EventID,
			AnalysisType: req.SuggestionType,
			InputText:    req.InputText,
			Suggestions:  fallback,
			Confidence:   0.7,
		}, nil
	}

	// Hit /suggestions route
	resp, err := s.callAIService(ctx, http.MethodPost, "suggestions", map[string]interface{}{
		"input_text": req.InputText,
		"case_id":    req.CaseID,
	})
	if err != nil {
		return nil, err
	}

	raw, ok := resp["suggestions"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid AI response format")
	}

	var suggestions []string
	for _, v := range raw {
		if str, ok := v.(string); ok {
			suggestions = append(suggestions, str)
		}
	}

	result := &AIAnalysisResult{
		CaseID:       req.CaseID,
		EventID:      req.EventID,
		AnalysisType: req.SuggestionType,
		InputText:    req.InputText,
		Suggestions:  suggestions,
		Confidence:   0.8,
	}

	if err := s.repository.SaveAnalysis(ctx, result); err != nil {
		return nil, fmt.Errorf("failed to save analysis: %w", err)
	}

	return result, nil
}

func (s *aiService) getCompletionSuggestions(ctx context.Context, req *SuggestionRequest) ([]string, error) {
	if !s.config.Enabled {
		return s.getPatternBasedSuggestions(ctx, req)
	}

	prompt := s.buildCompletionPrompt(req)

	// Pass endpoint + options explicitly
	response, err := s.callAIService(ctx, prompt, "completion", map[string]interface{}{})
	if err != nil {
		return s.getPatternBasedSuggestions(ctx, req)
	}

	// response is a map[string]interface{}
	// You need to extract the actual text field
	respText, ok := response["generated_text"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid response format from AI service")
	}

	return s.parseCompletionResponse(respText), nil
}

func (s *aiService) getNextStepSuggestions(ctx context.Context, req *SuggestionRequest) ([]string, error) {
	if !s.config.Enabled {
		return s.getDefaultNextSteps(req), nil
	}

	prompt := s.buildNextStepsPrompt(req)

	response, err := s.callAIService(ctx, prompt, "next-steps", map[string]interface{}{})
	if err != nil {
		return s.getDefaultNextSteps(req), nil
	}

	respText, ok := response["generated_text"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid response format from AI service")
	}

	return s.parseNextStepsResponse(respText), nil
}

func (s *aiService) GetSeverityRecommendation(ctx context.Context, description string) (string, float64, error) {
	keywordSeverity := ClassifySeverityFromKeywords(description)

	if !s.config.Enabled {
		return keywordSeverity, 0.7, nil
	}

	resp, err := s.callAIService(ctx, http.MethodPost, "severity", map[string]interface{}{
		"description": description,
	})
	if err != nil {
		return keywordSeverity, 0.7, nil
	}

	severity, ok := resp["recommended_severity"].(string)
	if !ok || severity == "" {
		return keywordSeverity, 0.7, nil
	}

	confidence := 0.8
	if c, ok := resp["confidence"].(float64); ok {
		confidence = c
	}

	return strings.ToLower(severity), confidence, nil
}

func (s *aiService) GetTagSuggestions(ctx context.Context, description string) ([]string, error) {
	// First get common tags
	commonTags := GenerateCommonTags(description)

	if !s.config.Enabled {
		return commonTags, nil
	}

	// Call AI /tags endpoint
	response, err := s.callAIService(ctx, http.MethodPost, "tags", map[string]interface{}{
		"description": description,
	})
	if err != nil {
		return commonTags, nil
	}

	// Parse tags from response
	rawTags, ok := response["tags"].([]interface{})
	if !ok {
		return commonTags, nil
	}

	var aiTags []string
	for _, t := range rawTags {
		if str, ok := t.(string); ok {
			aiTags = append(aiTags, str)
		}
	}

	// Combine + deduplicate
	combined := make(map[string]bool)
	var result []string
	for _, tag := range append(commonTags, aiTags...) {
		if !combined[tag] && tag != "" {
			combined[tag] = true
			result = append(result, tag)
		}
	}

	return result, nil
}

func (s *aiService) ExtractIOCs(ctx context.Context, text string) ([]IOCExtraction, error) {
	// Use regex-based extraction
	regexIOCs := ExtractIOCsFromText(text)

	// If AI is enabled, enhance with AI analysis
	if s.config.Enabled {
		aiIOCs, err := s.extractIOCsWithAI(ctx, text)
		if err == nil {
			// Combine and deduplicate
			return s.combineIOCs(regexIOCs, aiIOCs), nil
		}
	}

	return regexIOCs, nil
}

func (s *aiService) extractIOCsWithAI(ctx context.Context, text string) ([]IOCExtraction, error) {
	if !s.config.Enabled {
		return []IOCExtraction{}, nil
	}

	resp, err := s.callAIService(ctx, http.MethodPost, "iocs", map[string]interface{}{
		"text": text,
	})
	if err != nil {
		return nil, err
	}

	raw, ok := resp["iocs"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response from AI IOC extractor")
	}

	var results []IOCExtraction
	for _, v := range raw {
		if m, ok := v.(map[string]interface{}); ok {
			results = append(results, IOCExtraction{
				Type:       fmt.Sprintf("%v", m["type"]),
				Value:      fmt.Sprintf("%v", m["value"]),
				Confidence: m["confidence"].(float64),
			})
		}
	}
	return results, nil
}

func (s *aiService) AnalyzeEventContext(ctx context.Context, caseID string, eventText string) (*AIAnalysisResult, error) {
	result := &AIAnalysisResult{
		CaseID:       caseID,
		AnalysisType: "context_analysis",
		InputText:    eventText,
	}

	// Extract IOCs
	iocs, err := s.ExtractIOCs(ctx, eventText)
	if err == nil {
		result.ExtractedIOCs = iocs
	}

	// Get severity recommendation
	severity, confidence, err := s.GetSeverityRecommendation(ctx, eventText)
	if err == nil {
		result.RecommendedSeverity = severity
		result.Confidence = confidence
	}

	// Get tag suggestions
	tags, err := s.GetTagSuggestions(ctx, eventText)
	if err == nil {
		result.RecommendedTags = tags
	}

	// Save analysis result
	if err := s.repository.SaveAnalysis(ctx, result); err != nil {
		return nil, fmt.Errorf("failed to save context analysis: %w", err)
	}

	return result, nil
}

func (s *aiService) SuggestNextSteps(ctx context.Context, caseID string) ([]string, error) {
	history, err := s.repository.GetAnalysisHistory(ctx, caseID, "")
	if err != nil {
		return s.getDefaultNextSteps(&SuggestionRequest{CaseID: caseID}), nil
	}
	_ = history //ignore value

	if !s.config.Enabled {
		return s.getDefaultNextSteps(&SuggestionRequest{CaseID: caseID}), nil
	}

	// Call AI /cases/{case_id}/next-steps (GET)
	endpoint := fmt.Sprintf("cases/%s/next-steps", caseID)
	response, err := s.callAIService(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return s.getDefaultNextSteps(&SuggestionRequest{CaseID: caseID}), nil
	}

	// Parse steps
	rawSteps, ok := response["suggestions"].([]interface{})
	if !ok {
		return s.getDefaultNextSteps(&SuggestionRequest{CaseID: caseID}), nil
	}

	var steps []string
	for _, sVal := range rawSteps {
		if str, ok := sVal.(string); ok {
			steps = append(steps, str)
		}
	}

	return steps, nil
}

func (s *aiService) AnalyzeCaseProgress(ctx context.Context, caseID string) (*CaseAnalysis, error) {
	history, err := s.repository.GetAnalysisHistory(ctx, caseID, "")
	if err != nil {
		return nil, err
	}

	analysis := &CaseAnalysis{
		CaseID:     caseID,
		AnalyzedAt: time.Now(),
	}

	// Calculate completion score based on common DFIR phases
	requiredPhases := []string{"detection", "analysis", "containment", "eradication", "recovery"}
	completedPhases := s.identifyCompletedPhases(history)

	analysis.CompletionScore = float64(len(completedPhases)) / float64(len(requiredPhases))
	analysis.MissingSteps = s.findMissingSteps(requiredPhases, completedPhases)

	if s.config.Enabled {
		recommendations, err := s.getAIRecommendations(ctx, history)
		if err == nil {
			analysis.RecommendedActions = recommendations
		}
	}

	// Basic risk assessment
	analysis.RiskAssessment = s.assessRisk(history)

	return analysis, nil
}

func (s *aiService) CorrelateEvidence(ctx context.Context, caseID string, eventDescription string) ([]string, error) {
	// This would typically integrate with your evidence management system
	// For now, return empty slice as correlation requires access to evidence data
	return []string{}, nil
}

func (s *aiService) RecordFeedback(ctx context.Context, analysisID string, feedback *AIFeedback) error {
	return s.repository.SaveFeedback(ctx, feedback)
}

func (s *aiService) UpdateModelConfig(ctx context.Context, config *AIModelConfig) error {
	s.config = config
	return nil
}

func (s *aiService) GetModelStatus(ctx context.Context) (*ModelStatus, error) {
	status := &ModelStatus{
		ModelName:   s.config.ModelName,
		LastChecked: time.Now(),
	}

	if !s.config.Enabled {
		status.Status = "offline"
		return status, nil
	}

	start := time.Now()
	resp, err := s.callAIService(ctx, http.MethodGet, "health", nil)
	responseTime := time.Since(start).Milliseconds()

	if err != nil {
		status.Status = "offline"
		status.ErrorMessage = err.Error()
		return status, nil
	}

	// Parse status + model from response
	if v, ok := resp["status"].(string); ok {
		status.Status = v
	}
	if v, ok := resp["model"].(string); ok {
		status.ModelName = v
	}
	status.ResponseTime = responseTime

	return status, nil
}

// Helper methods - all good.
/*
@Param method: HTTP method (GET, POST)
@Param endpoint: API endpoint (e.g., "generate", "health")
@Param requestData: Request payload for POST requests
*/
func (s *aiService) callAIService(ctx context.Context, method string, endpoint string, requestData map[string]interface{}) (map[string]interface{}, error) {
	pythonServiceURL := "http://localhost:5000/api/v1/ai/" + endpoint

	var req *http.Request
	var err error

	if method == http.MethodGet {
		req, err = http.NewRequestWithContext(ctx, method, pythonServiceURL, nil)
	} else {
		jsonBody, err := json.Marshal(requestData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		req, err = http.NewRequestWithContext(ctx, method, pythonServiceURL, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call AI service: %w", err)
	}
	defer resp.Body.Close()

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response, nil
}

func (s *aiService) buildCompletionPrompt(req *SuggestionRequest) string {
	context := ""
	if req.Context != nil && len(req.Context.ExistingEvents) > 0 {
		context = "Recent events in this case:\n"
		for i, event := range req.Context.ExistingEvents {
			if i >= 3 { // Limit context to recent 3 events
				break
			}
			context += fmt.Sprintf("- %s\n", event.Description)
		}
		context += "\n"
	}

	return fmt.Sprintf(`You are a DFIR analyst assistant. Complete this investigation event description.
Provide 3-5 different completion suggestions that are specific and actionable.
Each suggestion should be on a new line starting with "-".

%sPartial description: %s

Completions:`, context, req.InputText)
}

func (s *aiService) buildNextStepsPrompt(req *SuggestionRequest) string {
	context := ""
	if req.Context != nil && len(req.Context.ExistingEvents) > 0 {
		context = "Timeline events in this case:\n"
		for _, event := range req.Context.ExistingEvents {
			context += fmt.Sprintf("- %s [%s]\n", event.Description, event.Severity)
		}
		context += "\n"
	}

	return fmt.Sprintf(`You are a DFIR analyst. Based on the investigation timeline, suggest 3-5 logical next steps.
Focus on evidence collection, analysis, containment, or remediation.
Each suggestion should be on a new line starting with "-".

%sSuggest next investigation steps:`, context)
}

func (s *aiService) parseCompletionResponse(response string) []string {
	lines := strings.Split(response, "\n")
	var suggestions []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "-") {
			suggestion := strings.TrimSpace(strings.TrimPrefix(line, "-"))
			if suggestion != "" {
				suggestions = append(suggestions, suggestion)
			}
		}
	}

	return suggestions
}

func (s *aiService) parseNextStepsResponse(response string) []string {
	return s.parseCompletionResponse(response) // Same parsing logic
}

func (s *aiService) parseTagsResponse(response string) []string {
	// Parse comma-separated tags
	tags := strings.Split(response, ",")
	var result []string

	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		tag = strings.ToLower(tag)
		// Remove any quotes or special characters
		tag = strings.Trim(tag, `"'()[]{}`)
		if tag != "" {
			result = append(result, tag)
		}
	}

	return result
}

func (s *aiService) getPatternBasedSuggestions(ctx context.Context, req *SuggestionRequest) ([]string, error) {
	patterns, err := s.repository.GetSuggestionPatterns(ctx, "default")
	if err != nil {
		return []string{}, err
	}

	inputLower := strings.ToLower(req.InputText)
	var suggestions []string

	// Match patterns based on keywords
	if strings.Contains(inputLower, "malware") || strings.Contains(inputLower, "virus") {
		suggestions = append(suggestions, patterns.Malware[:min(3, len(patterns.Malware))]...)
	} else if strings.Contains(inputLower, "network") || strings.Contains(inputLower, "traffic") {
		suggestions = append(suggestions, patterns.NetworkForensics[:min(3, len(patterns.NetworkForensics))]...)
	} else if strings.Contains(inputLower, "disk") || strings.Contains(inputLower, "file") {
		suggestions = append(suggestions, patterns.DiskForensics[:min(3, len(patterns.DiskForensics))]...)
	} else {
		// Default incident response suggestions
		suggestions = append(suggestions, patterns.IncidentResponse[:min(3, len(patterns.IncidentResponse))]...)
	}

	return suggestions, nil
}

func (s *aiService) getDefaultNextSteps(req *SuggestionRequest) []string {
	return []string{
		"Collect additional evidence from affected systems",
		"Analyze network logs for suspicious activity",
		"Document findings in investigation report",
		"Implement containment measures if needed",
		"Verify system integrity and security posture",
	}
}

func (s *aiService) combineIOCs(regexIOCs, aiIOCs []IOCExtraction) []IOCExtraction {
	seen := make(map[string]IOCExtraction)

	// Add regex IOCs
	for _, ioc := range regexIOCs {
		key := ioc.Type + ":" + ioc.Value
		seen[key] = ioc
	}

	// Add AI IOCs, preferring higher confidence
	for _, ioc := range aiIOCs {
		key := ioc.Type + ":" + ioc.Value
		if existing, exists := seen[key]; !exists || ioc.Confidence > existing.Confidence {
			seen[key] = ioc
		}
	}

	// Convert back to slice
	var result []IOCExtraction
	for _, ioc := range seen {
		result = append(result, ioc)
	}

	return result
}

func (s *aiService) identifyCompletedPhases(history []*AIAnalysisResult) []string {
	var phases []string
	phaseKeywords := map[string][]string{
		"detection":   {"detect", "discover", "alert", "incident"},
		"analysis":    {"analyze", "investigate", "examine", "forensic"},
		"containment": {"contain", "isolate", "quarantine", "block"},
		"eradication": {"remove", "clean", "eradicate", "eliminate"},
		"recovery":    {"restore", "recover", "rebuild", "normal"},
	}

	completed := make(map[string]bool)

	for _, analysis := range history {
		text := strings.ToLower(analysis.InputText)
		for phase, keywords := range phaseKeywords {
			for _, keyword := range keywords {
				if strings.Contains(text, keyword) {
					completed[phase] = true
					break
				}
			}
		}
	}

	for phase := range completed {
		phases = append(phases, phase)
	}

	return phases
}

func (s *aiService) findMissingSteps(required, completed []string) []string {
	completedMap := make(map[string]bool)
	for _, phase := range completed {
		completedMap[phase] = true
	}

	var missing []string
	for _, phase := range required {
		if !completedMap[phase] {
			missing = append(missing, phase)
		}
	}

	return missing
}

func (s *aiService) getAIRecommendations(ctx context.Context, history []*AIAnalysisResult) ([]string, error) {
	if len(history) == 0 {
		return []string{}, nil
	}

	var events []string
	for i, analysis := range history {
		if i >= 10 {
			break
		}
		events = append(events, analysis.InputText)
	}

	resp, err := s.callAIService(ctx, http.MethodPost, "recommendations", map[string]interface{}{
		"events": events,
	})
	if err != nil {
		return []string{}, err
	}

	raw, ok := resp["recommendations"].([]interface{})
	if !ok {
		return []string{}, nil
	}

	var recs []string
	for _, v := range raw {
		if str, ok := v.(string); ok {
			recs = append(recs, str)
		}
	}
	return recs, nil
}

func (s *aiService) assessRisk(history []*AIAnalysisResult) string {
	if len(history) == 0 {
		return "unknown"
	}

	highRiskCount := 0
	criticalCount := 0

	for _, analysis := range history {
		if analysis.RecommendedSeverity == "high" {
			highRiskCount++
		} else if analysis.RecommendedSeverity == "critical" {
			criticalCount++
		}
	}

	if criticalCount > 0 {
		return "critical"
	} else if highRiskCount > len(history)/2 {
		return "high"
	} else if highRiskCount > 0 {
		return "medium"
	}

	return "low"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
