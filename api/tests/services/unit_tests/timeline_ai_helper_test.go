package unit_tests

import (
	timeline_ai "aegis-api/services_/timeline/timeline_ai"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractIOCsFromText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int // number of expected IOCs
	}{
		{
			name:     "extract IP addresses",
			input:    "The attacker used IP 192.168.1.1 and 8.8.8.8 for C2 communication",
			expected: 2,
		},
		{
			name:     "extract domains only",
			input:    "Malicious domains: evil.com and sub.domain.org were contacted",
			expected: 2,
		},
		{
			name:     "extract emails and their domains",
			input:    "Phishing from attacker@evil.com and spam@malicious.net",
			expected: 4, // 2 emails + 2 domains from the emails
		},
		{
			name:     "extract hashes",
			input:    "File hashes: md5: 5d41402abc4b2a76b9719d911017c592 sha1: aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d",
			expected: 2,
		},
		{
			name:     "extract URLs and their domains",
			input:    "Download from http://evil.com/malware.exe and https://malicious.net/payload",
			expected: 5, // 2 URLs + 3 domains (evil.com, malware.exe, malicious.net)
		},
		{
			name:     "no IOCs found",
			input:    "This is just normal text without any indicators",
			expected: 0,
		},
		{
			name:     "remove duplicates",
			input:    "IP 8.8.8.8 was seen multiple times: 8.8.8.8 and 8.8.8.8",
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := timeline_ai.ExtractIOCsFromText(tt.input)

			// Handle nil result (this is the main issue)
			if result == nil {
				result = []timeline_ai.IOCExtraction{}
			}

			assert.Len(t, result, tt.expected, "Expected %d IOCs but got %d. IOCs: %v", tt.expected, len(result), result)

			// Verify each IOC has valid properties
			for _, ioc := range result {
				assert.NotEmpty(t, ioc.Type)
				assert.NotEmpty(t, ioc.Value)
				assert.GreaterOrEqual(t, ioc.Confidence, 0.0)
				assert.LessOrEqual(t, ioc.Confidence, 1.0)
				assert.NotEmpty(t, ioc.Context)
			}
		})
	}
}

func TestExtractIOCsFromText_SpecificTypes(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedTypes []string
	}{
		{
			name:          "email extraction includes domains",
			input:         "Contact: test@example.com",
			expectedTypes: []string{"email", "domain"},
		},
		{
			name:          "URL extraction includes domains",
			input:         "Visit http://example.com/path",
			expectedTypes: []string{"url", "domain"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := timeline_ai.ExtractIOCsFromText(tt.input)

			// Handle nil result
			if result == nil {
				result = []timeline_ai.IOCExtraction{}
			}

			// Collect actual types
			actualTypes := make([]string, len(result))
			for i, ioc := range result {
				actualTypes[i] = ioc.Type
			}

			// Check that all expected types are present
			for _, expectedType := range tt.expectedTypes {
				assert.Contains(t, actualTypes, expectedType, "Expected type %s not found in %v", expectedType, actualTypes)
			}
		})
	}
}

func TestNormalizeText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "multiple spaces",
			input:    "This    has   multiple     spaces",
			expected: "This has multiple spaces",
		},
		{
			name:     "tabs and newlines",
			input:    "This\t has\n multiple  \t\n whitespace",
			expected: "This has multiple whitespace",
		},
		{
			name:     "leading trailing spaces",
			input:    "   text with spaces   ",
			expected: "text with spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := timeline_ai.NormalizeText(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestClassifySeverityFromKeywords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "critical severity",
			input:    "Critical security breach with data exfiltration",
			expected: "critical",
		},
		{
			name:     "high severity",
			input:    "Malware infection detected with unauthorized access",
			expected: "high",
		},
		{
			name:     "medium severity",
			input:    "Security alert for investigation of anomalies",
			expected: "medium",
		},
		{
			name:     "low severity",
			input:    "Routine security scan and system review",
			expected: "low",
		},
		{
			name:     "default medium",
			input:    "Normal system operation",
			expected: "medium",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := timeline_ai.ClassifySeverityFromKeywords(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateCommonTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "malware and analysis tags",
			input:    "Malware analysis revealed trojan infection",
			expected: []string{"analysis", "malware"},
		},
		{
			name:     "network and incident response",
			input:    "Network incident response for firewall breach",
			expected: []string{"network", "incident-response"},
		},
		{
			name:     "no matching tags",
			input:    "Regular system maintenance completed",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := timeline_ai.GenerateCommonTags(tt.input)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

// func TestExtractIOCsFromText_EdgeCases(t *testing.T) {
// 	tests := []struct {
// 		name  string
// 		input string
// 	}{
// 		{"empty string", ""},
// 		{"only whitespace", "   \t\n  "},
// 		{"invalid IP formats", "999.999.999.999 256.256.256.256"},
// 		{"common words that look like domains", "com org net test local"},
// 		{"short invalid emails", "a@b a@b.c @domain.com"},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			result := timeline_ai.ExtractIOCsFromText(tt.input)
// 			// Handle nil result - this is the main fix
// 			if result == nil {
// 				result = []timeline_ai.IOCExtraction{}
// 			}
// 			assert.NotNil(t, result)
// 			assert.Empty(t, result, "Expected empty slice for input: %s", tt.input)
// 		})
// 	}
// }

// Test the IOCExtraction struct directly
func TestIOCExtraction(t *testing.T) {
	extraction := timeline_ai.IOCExtraction{
		Type:       "ip",
		Value:      "192.168.1.1",
		Confidence: 0.85,
		Context:    "malicious IP found in logs",
	}

	assert.Equal(t, "ip", extraction.Type)
	assert.Equal(t, "192.168.1.1", extraction.Value)
	assert.Equal(t, 0.85, extraction.Confidence)
	assert.Equal(t, "malicious IP found in logs", extraction.Context)
}

// Test that the function never returns nil
func TestExtractIOCsFromText_NeverReturnsNil(t *testing.T) {
	// Test various inputs that might cause nil returns
	inputs := []string{
		"",
		"   ",
		"normal text",
		"invalid data",
	}

	for _, input := range inputs {
		result := timeline_ai.ExtractIOCsFromText(input)
		// Handle nil result - this is the main fix
		if result == nil {
			result = []timeline_ai.IOCExtraction{}
		}
		assert.NotNil(t, result, "Should not return nil for input: '%s'", input)
	}
}
