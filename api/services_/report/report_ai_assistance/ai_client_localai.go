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

// Ensure the Report type exists in the reportshared package:
// If not, add the following to aegis-api/services_/report/shared/shared.go (or the appropriate file):

// package reportshared
//
// type Report struct {
//     ID        string
//     Title     string
//     CreatedAt time.Time
//     // Add other fields as needed
// }

// AIClientLocalAI implements AIClient interface
// GenerateSuggestion generates a draft for a report section using structured metadata
func (c *AIClientLocalAI) GenerateSuggestion(ctx context.Context, input AISuggestionInput) (string, error) {
	prompt := buildAISuggestionPrompt(input)
	log.Printf("[AIClientLocalAI] GenerateSuggestion called. Prompt: %s", prompt)
	result, err := c.callOpenAI(ctx, prompt)
	if err != nil {
		log.Printf("[AIClientLocalAI] Error in GenerateSuggestion: %v", err)
	}
	cleaned := removePromptEcho(result, prompt)
	return cleaned, err
}

// NewAIClientLocalAI returns a new LocalAI AI client
func NewAIClientLocalAI(_ string) *AIClientLocalAI {
	log.Printf("[AIClientLocalAI] Initializing LocalAI client")
	return &AIClientLocalAI{}
}

type AIClientLocalAI struct {
	// Add any necessary fields here
}

// AISuggestionInput provides context for AI suggestion generation
type AISuggestionInput struct {
	Case     *case_creation.Case
	Section  *reportshared.ReportSection
	Timeline []timeline.TimelineEvent
	IOCs     []graphicalmapping.IOC
	Evidence []metadata.Evidence
}

// Helper to build the prompt from structured input
func buildAISuggestionPrompt(input AISuggestionInput) string {
	caseObj := input.Case
	section := input.Section
	timelineEvents := input.Timeline
	iocs := input.IOCs
	evidences := input.Evidence

	sectionTitle := ""
	if section != nil {
		sectionTitle = section.Title
	}

	// Section-specific prompt construction
	switch sectionTitle {
	case "Case Identification":
		return fmt.Sprintf(
			"You are drafting the 'Case Identification' section of a DFIR report.\nCase: %s (%s)\nCreated: %s\nWrite a concise summary introducing the case, its purpose, and key dates.",
			caseObj.Title, caseObj.ID, caseObj.CreatedAt.Format("2006-01-02"))
	case "Evidence Summary":
		evidenceList := ""
		for _, e := range evidences {
			evidenceList += fmt.Sprintf("- %s (%s)\n", e.ID, e.FileType)
		}
		timelineRefs := ""
		for _, ev := range timelineEvents {
			if ev.Description != "" {
				timelineRefs += fmt.Sprintf("- %s\n", ev.Description)
			}
		}
		iocList := ""
		for _, ioc := range iocs {
			iocList += fmt.Sprintf("- %s: %s\n", ioc.Type, ioc.Value)
		}
		return fmt.Sprintf(
			"You are drafting the 'Evidence Summary' section of a DFIR report.\nEvidence collected:\n%sTimeline references:\n%sIndicators of Compromise (IOCs):\n%sWrite a clear, professional summary of the collected evidence suitable for a forensic report.",
			evidenceList, timelineRefs, iocList)
	case "Scope and Objectives":
		return "You are drafting the 'Scope and Objectives' section. Expand the case scope and objectives into formal report objectives."
	case "Tools and Methodologies":
		tools := ""
		for _, ev := range timelineEvents {
			if ev.Description != "" {
				tools += fmt.Sprintf("- %s\n", ev.Description)
			}
		}
		return fmt.Sprintf(
			"You are drafting the 'Tools and Methodologies' section.\nTools used in investigation:\n%sExplain the forensic methodology and tools used.",
			tools)
	case "Findings":
		findings := ""
		for _, ev := range timelineEvents {
			findings += fmt.Sprintf("- %s: %s\n", ev.CreatedAt.Format("2006-01-02 15:04"), ev.Description)
		}
		iocList := ""
		for _, ioc := range iocs {
			iocList += fmt.Sprintf("- %s: %s\n", ioc.Type, ioc.Value)
		}
		return fmt.Sprintf(
			"You are drafting the 'Findings' section.\nKey timeline events:\n%sIndicators of Compromise (IOCs):\n%sSummarize the main findings of the investigation.",
			findings, iocList)
	case "Interpretation and Analysis":
		analysis := ""
		for _, ev := range timelineEvents {
			analysis += fmt.Sprintf("- %s: %s\n", ev.CreatedAt.Format("2006-01-02 15:04"), ev.Description)
		}
		iocList := ""
		for _, ioc := range iocs {
			iocList += fmt.Sprintf("- %s: %s\n", ioc.Type, ioc.Value)
		}
		return fmt.Sprintf(
			"You are drafting the 'Interpretation and Analysis' section.\nTimeline and evidence:\n%sIndicators of Compromise (IOCs):\n%sWrite an analytical narrative interpreting the investigation results.",
			analysis, iocList)
	case "Limitations":
		return "You are drafting the 'Limitations' section. List any gaps in evidence, investigation constraints, or disclaimers relevant to this case."
	case "Conclusion":
		return "You are drafting the 'Conclusion' section. Synthesize the investigation outcome and provide a formal closure."
	case "Appendices":
		return "You are drafting the 'Appendices' section. Auto-generate chain-of-custody logs and evidence tables."
	case "Certifications":
		return "You are drafting the 'Certifications' section. Auto-fill investigator roles and qualifications."
	default:
		return fmt.Sprintf(
			"You are drafting the '%s' section. Use all available context to generate a professional draft. If the user has typed a prompt, expand or rephrase it.",
			sectionTitle)
	}
}

// Remove prompt echo from AI response
func removePromptEcho(result, prompt string) string {
	// Remove exact match at start
	if len(result) > len(prompt) && result[:len(prompt)] == prompt {
		result = result[len(prompt):]
	}
	// Remove prompt if found anywhere
	if idx := indexOf(result, prompt); idx >= 0 {
		result = result[idx+len(prompt):]
	}

	promptLines := SplitLines(prompt)
	resultLines := SplitLines(result)
	cleanedLines := []string{}

	// Expanded list of generic/repetitive/boilerplate patterns
	genericPatterns := []string{
		"You: your professional forensic analyst",
		"You: our professional forensic analyst",
		"You: your expert forensic analyst",
		"You: our expert forensic analyst",
		"You: your professional forensic analyst.",
		"You: our professional forensic analyst.",
		"You: your expert forensic analyst.",
		"You: our expert forensic analyst.",
		"Summary: what we found",
		"Other: some",
		"Case Number:",
		"Case Name:",
		"Case Index:",
		"Case Definition:",
		"Case Text:",
		"You are a professional Digital Forensics and Incident Response (DFIR) report writer.",
		"Generate clear, concise, and relevant forensic report content for the following section:",
		"You are drafting the",
		"Write a clear, professional summary of the collected evidence suitable for a forensic report.",
		"Write a concise summary introducing the case, its purpose, and key dates.",
		"Expand the case scope and objectives into formal report objectives.",
		"Explain the forensic methodology and tools used.",
		"Summarize the main findings of the investigation.",
		"Write an analytical narrative interpreting the investigation results.",
		"List any gaps in evidence, investigation constraints, or disclaimers relevant to this case.",
		"Synthesize the investigation outcome and provide a formal closure.",
		"Auto-generate chain-of-custody logs and evidence tables.",
		"Auto-fill investigator roles and qualifications.",
		"Use all available context to generate a professional draft.",
		"If the user has typed a prompt, expand or rephrase it.",
		"Provide relevant references or citations for the",
		"Summarize the following evidence and timeline for a forensic report:",
		"Based on the following analysis, suggest next steps and recommendations:",
		"Refine the following report section based on feedback:",
		"Section:",
		"Feedback:",
		"Case:",
		"Analysis Summary:",
		"Timeline references:",
		"Indicators of Compromise (IOCs):",
		"Evidence collected:",
		"Tools used in investigation:",
		"Key timeline events:",
		"Timeline and evidence:",
		"Investigator roles and qualifications:",
		"Chain-of-custody logs:",
		"Evidence tables:",
		"DFIR report",
		"DFIR",
		"Forensic report",
		"Forensic",
		"Report section:",
		"Section title:",
		"Appendices:",
		"Certifications:",
		"Limitations:",
		"Conclusion:",
		"Scope and Objectives:",
		"Findings:",
		"Interpretation and Analysis:",
		"Case Identification:",
		"Evidence Summary:",
		"Tools and Methodologies:",
		// Common verbose/generic openers
		"In conclusion",
		"This report",
		"The investigation",
		"It is important to note",
		"The following",
		"The purpose of this section",
		"The scope of this report",
		"The findings of this report",
		"The analysis presented",
		"The evidence collected",
		"The tools and methodologies",
		"The limitations of this investigation",
		"The appendices include",
		"The certifications provided",
		"The case identification",
		"The evidence summary",
		"The scope and objectives",
		"The findings",
		"The interpretation and analysis",
		"The conclusion",
		"The limitations",
		"The appendices",
		"The certifications",
		// New patterns to catch repeated/boilerplate output
		"You are writing a technical report.",
		"You are writing a technical report",
		"You are writing an analysis.",
		"You are writing an analysis",
		"Include a demonstration of the analysis.",
		"Include a demonstration of the analysis",
		"Include a demonstration of your technical report.",
		"Include a demonstration of your technical report",
		"Include an example of your technical report.",
		"Include an example of your technical report",
		// Catch lines that start with these phrases
		"You are writing",
		"Include a demonstration",
		"Include an example",
	}

	sectionTitle := ""
	if len(promptLines) > 0 {
		for _, pLine := range promptLines {
			if len(pLine) > 0 && (indexOf(pLine, "section of a DFIR report") >= 0 || indexOf(pLine, "You are drafting the '") >= 0) {
				sectionTitle = pLine
				break
			}
		}
	}
	if len(promptLines) > 0 {
		for _, pLine := range promptLines {
			if len(pLine) > 0 && (indexOf(pLine, "section of a DFIR report") >= 0 || indexOf(pLine, "You are drafting the '") >= 0) {
				sectionTitle = pLine
				break
			}
		}
	}

	for _, line := range resultLines {
		trimmedLine := TrimWhitespace(line)
		if trimmedLine == "" {
			continue // skip empty lines
		}
		skip := false

		// Remove lines that match any line in the prompt or section title
		for _, pLine := range promptLines {
			trimmedPrompt := TrimWhitespace(pLine)
			if trimmedLine == trimmedPrompt || (len(trimmedPrompt) > 0 && trimmedLine == TrimWhitespace(trimmedPrompt[:min(40, len(trimmedPrompt))])) {
				skip = true
				break
			}
		}
		if sectionTitle != "" && (trimmedLine == sectionTitle || indexOf(trimmedLine, sectionTitle) == 0) {
			skip = true
		}

		// Remove lines that match any generic/boilerplate pattern or start with them
		if !skip {
			for _, pattern := range genericPatterns {
				if len(trimmedLine) >= len(pattern) && trimmedLine[:len(pattern)] == pattern {
					skip = true
					break
				}
				// Also skip if line starts with pattern (for verbose openers)
				if len(trimmedLine) > len(pattern) && trimmedLine[:len(pattern)] == pattern {
					skip = true
					break
				}
			}
		}

		if !skip {
			cleanedLines = append(cleanedLines, line)
		}
	}

	return joinLines(cleanedLines)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func joinLines(lines []string) string {
	out := ""
	for _, l := range lines {
		if len(out) > 0 {
			out += "\n"
		}
		out += l
	}
	return out
}

func findPromptFragment(result, prompt string) int {
	// Try to find a fragment of the prompt in the result
	frag := prompt
	if len(frag) > 100 {
		frag = frag[:100]
	}
	return indexOf(result, frag)
}

func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// RefineSuggestion refines an existing suggestion using user feedback
func (c *AIClientLocalAI) RefineSuggestion(ctx context.Context, existing string, feedback string) (string, error) {
	prompt := fmt.Sprintf("Refine the following report section based on feedback:\n\nSection:\n%s\n\nFeedback:\n%s", existing, feedback)
	log.Printf("[AIClientLocalAI] RefineSuggestion called. Prompt: %s", prompt)
	result, err := c.callOpenAI(ctx, prompt)
	if err != nil {
		log.Printf("[AIClientLocalAI] Error in RefineSuggestion: %v", err)
	}
	return result, err
}

// SummarizeEvidence generates a summary from evidence, IOCs, and timeline
func (c *AIClientLocalAI) SummarizeEvidence(ctx context.Context, evidence []metadata.Evidence, iocs []graphicalmapping.IOC, timeline []timeline.TimelineEvent) (string, error) {
	// Simple summary stub using available fields
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
	prompt := fmt.Sprintf("Summarize the following evidence and timeline for a forensic report:\nEvidence:\n%sTimeline:\n%s", evidenceList, timelineRefs)
	log.Printf("[AIClientLocalAI] SummarizeEvidence called. Prompt: %s", prompt)
	result, err := c.callOpenAI(ctx, prompt)
	if err != nil {
		log.Printf("[AIClientLocalAI] Error in SummarizeEvidence: %v", err)
	}
	return result, err
}

// GenerateRecommendations provides next steps or mitigation strategies
func (c *AIClientLocalAI) GenerateRecommendations(ctx context.Context, caseData *case_creation.Case, analysisSummary string) (string, error) {
	prompt := fmt.Sprintf("Based on the following analysis, suggest next steps and recommendations:\nCase: %+v\nAnalysis Summary:\n%s", caseData, analysisSummary)
	log.Printf("[AIClientLocalAI] GenerateRecommendations called. Prompt: %s", prompt)
	result, err := c.callOpenAI(ctx, prompt)
	if err != nil {
		log.Printf("[AIClientLocalAI] Error in GenerateRecommendations: %v", err)
	}
	return result, err
}

// EvaluateFeedback logs user feedback or adjusts AI prompts
func (c *AIClientLocalAI) EvaluateFeedback(ctx context.Context, suggestionID string, feedback string) error {
	log.Printf("Feedback received for suggestion %s: %s", suggestionID, feedback)
	// Optional: store in DB or use to adjust future prompts
	return nil
}

// GenerateSectionReferences provides citations or reference links for a section
func (c *AIClientLocalAI) GenerateSectionReferences(ctx context.Context, sectionName string, report *reportshared.Report) ([]string, error) {
	prompt := fmt.Sprintf("Provide relevant references or citations for the '%s' section of a forensic report for case %+v", sectionName, report)
	log.Printf("[AIClientLocalAI] GenerateSectionReferences called. Prompt: %s", prompt)
	refs, err := c.callOpenAI(ctx, prompt)
	if err != nil {
		log.Printf("[AIClientLocalAI] Error in GenerateSectionReferences: %v", err)
		return nil, err
	}
	// Simple stub: split by newlines
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

// ---- Internal helpers ----

func (c *AIClientLocalAI) callOpenAI(ctx context.Context, prompt string) (string, error) {
	log.Printf("[AIClientLocalAI] callOpenAI called. LocalAI endpoint, Prompt: %s", prompt)
	start := time.Now()
	// Prepare request body
	reqBody, err := json.Marshal(map[string]string{"prompt": prompt})
	if err != nil {
		return "", err
	}
	// Make HTTP POST request to localai
	httpReq, err := http.NewRequestWithContext(ctx, "POST", "http://localai:5000/generate", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(httpReq)
	elapsed := time.Since(start)
	if err != nil {
		log.Printf("[AIClientLocalAI] Error from localai: %v (elapsed: %s)", err, elapsed)
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("[AIClientLocalAI] localai returned status: %d", resp.StatusCode)
		return "", fmt.Errorf("localai error: status %d", resp.StatusCode)
	}
	var respData struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return "", err
	}
	log.Printf("[AIClientLocalAI] localai response received in %s", elapsed)
	log.Printf("[AIClientLocalAI] Response: %s", respData.Text)
	return respData.Text, nil
}
