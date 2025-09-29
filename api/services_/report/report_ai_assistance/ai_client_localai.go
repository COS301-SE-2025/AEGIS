package report_ai_assistance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	graphicalmapping "aegis-api/services_/GraphicalMapping"
	"aegis-api/services_/case/case_creation"
	"aegis-api/services_/evidence/metadata"
	reportshared "aegis-api/services_/report/shared"
	"aegis-api/services_/timeline"
)

// AIClientLocalAI implements AIClient interface
type AIClientLocalAI struct{}

// NewAIClientLocalAI returns a new LocalAI AI client
func NewAIClientLocalAI(_ string) *AIClientLocalAI {
	log.Printf("[AIClientLocalAI] Initializing LocalAI client (Flan-T5 mode)")
	return &AIClientLocalAI{}
}

// AISuggestionInput provides context for AI suggestion generation
type AISuggestionInput struct {
	Case     *case_creation.Case
	Section  *reportshared.ReportSection
	Timeline []timeline.TimelineEvent
	IOCs     []graphicalmapping.IOC
	Evidence []metadata.Evidence
}

// GenerateSuggestion generates a draft for a report section
func (c *AIClientLocalAI) GenerateSuggestion(ctx context.Context, input AISuggestionInput) (string, error) {
	prompt := buildAISuggestionPrompt(input)
	log.Printf("[AIClientLocalAI] GenerateSuggestion called. Prompt:\n%s", prompt)
	result, err := c.callOpenAI(ctx, prompt, 350)
	if err != nil {
		log.Printf("[AIClientLocalAI] Error in GenerateSuggestion: %v", err)
	}
	cleaned := removePromptEcho(result, prompt)
	return cleaned, err
}

// RefineSuggestion refines an existing suggestion using user feedback
func (c *AIClientLocalAI) RefineSuggestion(ctx context.Context, existing string, feedback string) (string, error) {
	prompt := fmt.Sprintf("Refine the following forensic report section based on feedback:\n\nSection:\n%s\n\nFeedback:\n%s\n\nRewrite the section accordingly.", existing, feedback)
	log.Printf("[AIClientLocalAI] RefineSuggestion called. Prompt:\n%s", prompt)
	return c.callOpenAI(ctx, prompt, 300)
}

// SummarizeEvidence generates a summary from evidence, IOCs, and timeline
func (c *AIClientLocalAI) SummarizeEvidence(ctx context.Context, evidence []metadata.Evidence, iocs []graphicalmapping.IOC, timeline []timeline.TimelineEvent) (string, error) {
	evidenceList := ""
	for _, e := range evidence {
		evidenceList += fmt.Sprintf("- %s (%s)\n", e.ID, e.FileType)
	}
	timelineRefs := ""
	for _, ev := range timeline {
		if ev.Description != "" {
			timelineRefs += fmt.Sprintf("- %s\n", ev.Description)
		}
	}
	iocList := ""
	for _, ioc := range iocs {
		iocList += fmt.Sprintf("- %s: %s\n", ioc.Type, ioc.Value)
	}
	prompt := fmt.Sprintf(
		"Summarize the following evidence and timeline for inclusion in a DFIR report.\n\nEvidence:\n%s\nTimeline:\n%s\nIOCs:\n%s\n\nWrite a concise, professional forensic summary.",
		evidenceList, timelineRefs, iocList)
	log.Printf("[AIClientLocalAI] SummarizeEvidence called. Prompt:\n%s", prompt)
	return c.callOpenAI(ctx, prompt, 250)
}

// GenerateRecommendations provides next steps or mitigation strategies
func (c *AIClientLocalAI) GenerateRecommendations(ctx context.Context, caseData *case_creation.Case, analysisSummary string) (string, error) {
	prompt := fmt.Sprintf(
		"Based on the following forensic case and analysis, suggest next steps and recommendations.\n\nCase: %+v\n\nAnalysis Summary:\n%s\n\nProvide actionable, concise recommendations.",
		caseData, analysisSummary)
	log.Printf("[AIClientLocalAI] GenerateRecommendations called. Prompt:\n%s", prompt)
	return c.callOpenAI(ctx, prompt, 200)
}

// EvaluateFeedback logs user feedback
func (c *AIClientLocalAI) EvaluateFeedback(ctx context.Context, suggestionID string, feedback string) error {
	log.Printf("[AIClientLocalAI] Feedback received for suggestion %s: %s", suggestionID, feedback)
	return nil
}

// GenerateSectionReferences provides citations or reference links for a section
func (c *AIClientLocalAI) GenerateSectionReferences(ctx context.Context, sectionName string, report *reportshared.Report) ([]string, error) {
	prompt := fmt.Sprintf(
		"Provide relevant references or citations for the '%s' section of this forensic report:\nCase: %+v",
		sectionName, report)
	log.Printf("[AIClientLocalAI] GenerateSectionReferences called. Prompt:\n%s", prompt)

	refs, err := c.callOpenAI(ctx, prompt, 200)
	if err != nil {
		return nil, err
	}

	lines := SplitLines(refs)
	var out []string
	for _, l := range lines {
		trimmed := TrimWhitespace(l)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out, nil
}

// ----------------- Helpers -----------------

// Flan-T5 optimized prompt builder
func buildAISuggestionPrompt(input AISuggestionInput) string {
	caseObj := input.Case
	section := input.Section
	sectionTitle := ""
	if section != nil {
		sectionTitle = section.Title
	}

	switch sectionTitle {
	case "Case Identification":
		return fmt.Sprintf(
			"Write the 'Case Identification' section of a DFIR report.\n\nCase: %s (%s)\nCreated: %s\n\nSummarize the case purpose and key dates.",
			caseObj.Title, caseObj.ID, caseObj.CreatedAt.Format("2006-01-02"))
	case "Evidence Summary":
		evidenceList := ""
		for _, e := range input.Evidence {
			evidenceList += fmt.Sprintf("- %s (%s)\n", e.ID, e.FileType)
		}
		timelineRefs := ""
		for _, ev := range input.Timeline {
			if ev.Description != "" {
				timelineRefs += fmt.Sprintf("- %s\n", ev.Description)
			}
		}
		iocList := ""
		for _, ioc := range input.IOCs {
			iocList += fmt.Sprintf("- %s: %s\n", ioc.Type, ioc.Value)
		}
		return fmt.Sprintf(
			"Write the 'Evidence Summary' section of a DFIR report.\n\nEvidence:\n%s\nTimeline:\n%s\nIOCs:\n%s\n\nProduce a clear, professional summary.",
			evidenceList, timelineRefs, iocList)
	case "Scope and Objectives":
		return "Write the 'Scope and Objectives' section of a DFIR report. Expand the scope into clear, professional objectives."
	case "Tools and Methodologies":
		tools := ""
		for _, ev := range input.Timeline {
			if ev.Description != "" {
				tools += fmt.Sprintf("- %s\n", ev.Description)
			}
		}
		return fmt.Sprintf(
			"Write the 'Tools and Methodologies' section of a DFIR report.\n\nTools:\n%s\n\nExplain the methodology and tools used.",
			tools)
	case "Findings":
		findings := ""
		for _, ev := range input.Timeline {
			findings += fmt.Sprintf("- %s: %s\n", ev.CreatedAt.Format("2006-01-02 15:04"), ev.Description)
		}
		iocList := ""
		for _, ioc := range input.IOCs {
			iocList += fmt.Sprintf("- %s: %s\n", ioc.Type, ioc.Value)
		}
		return fmt.Sprintf(
			"Write the 'Findings' section of a DFIR report.\n\nTimeline:\n%s\nIOCs:\n%s\n\nSummarize the investigation findings.",
			findings, iocList)
	case "Interpretation and Analysis":
		analysis := ""
		for _, ev := range input.Timeline {
			analysis += fmt.Sprintf("- %s: %s\n", ev.CreatedAt.Format("2006-01-02 15:04"), ev.Description)
		}
		iocList := ""
		for _, ioc := range input.IOCs {
			iocList += fmt.Sprintf("- %s: %s\n", ioc.Type, ioc.Value)
		}
		return fmt.Sprintf(
			"Write the 'Interpretation and Analysis' section of a DFIR report.\n\nTimeline and Evidence:\n%s\nIOCs:\n%s\n\nProvide an analytical narrative.",
			analysis, iocList)
	case "Limitations":
		return "Write the 'Limitations' section of a DFIR report. List any evidence gaps, constraints, or disclaimers."
	case "Conclusion":
		return "Write the 'Conclusion' section of a DFIR report. Summarize the overall outcome and provide closure."
	case "Appendices":
		return "Write the 'Appendices' section of a DFIR report. Include chain-of-custody logs and evidence tables."
	case "Certifications":
		return "Write the 'Certifications' section of a DFIR report. List investigator roles and qualifications."
	default:
		return fmt.Sprintf(
			"Write the '%s' section of a DFIR report. Use all available context to create a professional draft.",
			sectionTitle)
	}
}

// removePromptEcho trims the prompt text if it is echoed back in the result.
func removePromptEcho(result, prompt string) string {
	if len(result) >= len(prompt) && result[:len(prompt)] == prompt {
		return result[len(prompt):]
	}
	return result
}

// Call Flask LocalAI service
func (c *AIClientLocalAI) callOpenAI(ctx context.Context, prompt string, maxLength int) (string, error) {
	start := time.Now()
	reqBody, err := json.Marshal(map[string]interface{}{
		"prompt":     prompt,
		"max_length": maxLength,
	})
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "http://localai:5000/generate", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	elapsed := time.Since(start)
	if err != nil {
		log.Printf("[AIClientLocalAI] Error from LocalAI: %v (elapsed: %s)", err, elapsed)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("LocalAI error: status %d", resp.StatusCode)
	}

	var respData struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return "", err
	}

	log.Printf("[AIClientLocalAI] LocalAI response received in %s", elapsed)
	return respData.Text, nil
}
