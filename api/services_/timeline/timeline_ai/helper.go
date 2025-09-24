package timelineai

import (
	"regexp"
	"strings"
	"unicode"
)

// IOCPatterns contains regex patterns for different IOC types
var IOCPatterns = map[string]*regexp.Regexp{
	"ip":     regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`),
	"domain": regexp.MustCompile(`\b[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*\.[a-zA-Z]{2,}\b`),
	"email":  regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`),
	"md5":    regexp.MustCompile(`\b[a-fA-F0-9]{32}\b`),
	"sha1":   regexp.MustCompile(`\b[a-fA-F0-9]{40}\b`),
	"sha256": regexp.MustCompile(`\b[a-fA-F0-9]{64}\b`),
	"url":    regexp.MustCompile(`https?://[^\s<>"{}|\\^` + "`" + `\[\]]+`),
}

// ExtractIOCsFromText extracts indicators of compromise from text using regex
func ExtractIOCsFromText(text string) []IOCExtraction {
	var extractions []IOCExtraction

	for iocType, pattern := range IOCPatterns {
		matches := pattern.FindAllString(text, -1)
		for _, match := range matches {
			if isValidIOC(iocType, match) {
				extraction := IOCExtraction{
					Type:       iocType,
					Value:      match,
					Confidence: calculateIOCConfidence(iocType, match, text),
					Context:    extractContext(text, match, 50),
				}
				extractions = append(extractions, extraction)
			}
		}
	}

	return removeDuplicateIOCs(extractions)
}

// isValidIOC performs additional validation on extracted IOCs
func isValidIOC(iocType, value string) bool {
	switch iocType {
	case "ip":
		return isValidIP(value)
	case "domain":
		return isValidDomain(value)
	case "email":
		return len(value) > 5 && strings.Contains(value, "@")
	default:
		return len(value) > 0
	}
}

// isValidIP validates IP addresses
func isValidIP(ip string) bool {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return false
	}

	for _, part := range parts {
		if len(part) == 0 || len(part) > 3 {
			return false
		}

		num := 0
		for _, r := range part {
			if !unicode.IsDigit(r) {
				return false
			}
			num = num*10 + int(r-'0')
		}

		if num > 255 {
			return false
		}
	}

	return true
}

// isValidDomain performs basic domain validation
func isValidDomain(domain string) bool {
	if len(domain) == 0 || len(domain) > 255 {
		return false
	}

	// Skip common words that might match domain pattern
	commonWords := []string{"com", "org", "net", "edu", "local", "test"}
	for _, word := range commonWords {
		if strings.ToLower(domain) == word {
			return false
		}
	}

	return strings.Contains(domain, ".")
}

// calculateIOCConfidence calculates confidence score for an IOC
func calculateIOCConfidence(iocType, value, context string) float64 {
	baseConfidence := 0.7

	// Adjust confidence based on context
	contextLower := strings.ToLower(context)

	// Higher confidence if found in security-related context
	securityKeywords := []string{"malware", "suspicious", "infected", "attack", "compromise", "threat"}
	for _, keyword := range securityKeywords {
		if strings.Contains(contextLower, keyword) {
			baseConfidence += 0.15
			break
		}
	}

	// Adjust based on IOC type
	switch iocType {
	case "md5", "sha1", "sha256":
		baseConfidence += 0.2 // Hashes are typically high confidence
	case "ip":
		if isPrivateIP(value) {
			baseConfidence -= 0.2 // Private IPs less likely to be external threats
		}
	}

	// Cap confidence at 1.0
	if baseConfidence > 1.0 {
		baseConfidence = 1.0
	}

	return baseConfidence
}

// isPrivateIP checks if an IP address is in private ranges
func isPrivateIP(ip string) bool {
	privateRanges := []string{
		"10.", "192.168.", "172.16.", "172.17.", "172.18.", "172.19.",
		"172.20.", "172.21.", "172.22.", "172.23.", "172.24.", "172.25.",
		"172.26.", "172.27.", "172.28.", "172.29.", "172.30.", "172.31.",
		"127.", "169.254.",
	}

	for _, prefix := range privateRanges {
		if strings.HasPrefix(ip, prefix) {
			return true
		}
	}

	return false
}

// extractContext extracts surrounding text for context
func extractContext(text, match string, contextLength int) string {
	index := strings.Index(text, match)
	if index == -1 {
		return ""
	}

	start := index - contextLength
	if start < 0 {
		start = 0
	}

	end := index + len(match) + contextLength
	if end > len(text) {
		end = len(text)
	}

	return strings.TrimSpace(text[start:end])
}

// removeDuplicateIOCs removes duplicate IOC extractions
func removeDuplicateIOCs(extractions []IOCExtraction) []IOCExtraction {
	seen := make(map[string]bool)
	var result []IOCExtraction

	for _, extraction := range extractions {
		key := extraction.Type + ":" + extraction.Value
		if !seen[key] {
			seen[key] = true
			result = append(result, extraction)
		}
	}

	return result
}

// NormalizeText normalizes text for better AI processing
func NormalizeText(text string) string {
	// Remove excessive whitespace
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")

	// Trim whitespace
	text = strings.TrimSpace(text)

	return text
}

// ClassifySeverityFromKeywords attempts to classify severity based on keywords
func ClassifySeverityFromKeywords(text string) string {
	textLower := strings.ToLower(text)

	criticalKeywords := []string{"critical", "severe", "breach", "ransomware", "data theft", "exfiltration"}
	highKeywords := []string{"malware", "infection", "compromise", "unauthorized access", "suspicious activity"}
	mediumKeywords := []string{"alert", "warning", "anomaly", "investigation", "analysis"}
	lowKeywords := []string{"scan", "check", "review", "update", "maintenance"}

	for _, keyword := range criticalKeywords {
		if strings.Contains(textLower, keyword) {
			return "critical"
		}
	}

	for _, keyword := range highKeywords {
		if strings.Contains(textLower, keyword) {
			return "high"
		}
	}

	for _, keyword := range mediumKeywords {
		if strings.Contains(textLower, keyword) {
			return "medium"
		}
	}

	for _, keyword := range lowKeywords {
		if strings.Contains(textLower, keyword) {
			return "low"
		}
	}

	return "medium" // Default
}

// GenerateCommonTags generates common tags based on text content
func GenerateCommonTags(text string) []string {
	textLower := strings.ToLower(text)
	var tags []string

	tagMap := map[string][]string{
		"analysis":          {"analysis", "analyze", "investigation", "examine"},
		"malware":           {"malware", "virus", "trojan", "ransomware", "backdoor"},
		"network":           {"network", "traffic", "connection", "packet", "firewall"},
		"forensics":         {"forensics", "evidence", "artifact", "timeline"},
		"containment":       {"contain", "isolate", "quarantine", "block"},
		"incident-response": {"incident", "response", "emergency", "breach"},
		"memory-analysis":   {"memory", "ram", "dump", "process"},
		"disk-analysis":     {"disk", "file", "registry", "filesystem"},
		"ioc":               {"ioc", "indicator", "hash", "signature"},
		"threat-hunting":    {"hunt", "threat", "suspicious", "anomaly"},
	}

	for tag, keywords := range tagMap {
		for _, keyword := range keywords {
			if strings.Contains(textLower, keyword) {
				tags = append(tags, tag)
				break
			}
		}
	}

	return tags
}
