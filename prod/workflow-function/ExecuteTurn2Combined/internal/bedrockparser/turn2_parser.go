package bedrockparser

import (
	"regexp"
	"strings"

	"workflow-function/ExecuteTurn2Combined/internal/models"
)

// ParsedTurn2Markdown holds the cleaned Markdown content from Bedrock's Turn 2 response.
type ParsedTurn2Markdown struct {
	ComparisonMarkdown string `json:"comparisonMarkdown"`
}

// ParsedTurn2Data holds structured fields extracted from the Turn 2 Markdown.
type ParsedTurn2Data struct {
	Discrepancies       []models.Discrepancy `json:"discrepancies"`
	VerificationOutcome string               `json:"verificationOutcome"`
	ComparisonSummary   string               `json:"comparisonSummary"`
}

// ParseTurn2BedrockResponseAsMarkdown extracts and cleans the Bedrock Turn 2 textual response.
func ParseTurn2BedrockResponseAsMarkdown(bedrockTextResponse string) (*ParsedTurn2Markdown, error) {
	if strings.TrimSpace(bedrockTextResponse) == "" {
		// Return empty markdown struct for empty input
		return &ParsedTurn2Markdown{ComparisonMarkdown: ""}, nil
	}

	cleaned := strings.ReplaceAll(bedrockTextResponse, "\r\n", "\n")
	cleaned = strings.TrimSpace(cleaned)

	return &ParsedTurn2Markdown{ComparisonMarkdown: cleaned}, nil
}

// ParseTurn2Response converts the markdown formatted analysis into a structured
// ParsedTurn2Data. If the input text is empty, nil is returned to signal no
// actionable content was found. If structured parsing fails, it provides defaults.
func ParseTurn2Response(text string) (*ParsedTurn2Data, error) {
	if strings.TrimSpace(text) == "" {
		return nil, nil
	}

	result := &ParsedTurn2Data{
		Discrepancies:     []models.Discrepancy{},
		ComparisonSummary: "",
	}

	// Normalize line endings for consistent regex processing
	text = strings.ReplaceAll(text, "\r\n", "\n")

	// Extract verification outcome
	outcomeRe := regexp.MustCompile(`(?i)verification\s+outcome:\s*(CORRECT|INCORRECT)`)
	if matches := outcomeRe.FindStringSubmatch(text); len(matches) > 1 {
		result.VerificationOutcome = strings.ToUpper(matches[1])
	}

	// Extract discrepancies section
	discrepanciesRe := regexp.MustCompile(`(?s)discrepancies:(.*?)(?:\n\n|\n\*\*|$)`)
	if matches := discrepanciesRe.FindStringSubmatch(text); len(matches) > 1 {
		discrepanciesBlock := strings.TrimSpace(matches[1])
		lines := strings.Split(discrepanciesBlock, "\n")

		// Process each line as a potential discrepancy
		discrepancyRe := regexp.MustCompile(`(?i)\s*[-*]\s*([^:]+):\s*expected\s+in\s+([^,]+),\s*(?:found\s+in\s+([^,]+)|not\s+found)`)
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			if matches := discrepancyRe.FindStringSubmatch(line); len(matches) > 2 {
				item := strings.TrimSpace(matches[1])
				expected := strings.TrimSpace(matches[2])
				found := ""
				discrepancyType := "MISSING"
				severity := "MEDIUM"

				if len(matches) > 3 && matches[3] != "" {
					found = strings.TrimSpace(matches[3])
					discrepancyType = "MISPLACED"
					severity = "MEDIUM"
				} else {
					severity = "HIGH"
				}

				result.Discrepancies = append(result.Discrepancies, models.Discrepancy{
					Item:     item,
					Expected: expected,
					Found:    found,
					Type:     discrepancyType,
					Severity: severity,
				})
			}
		}
	}

	// Extract comparison summary
	summaryRe := regexp.MustCompile(`(?s)\*\*COMPARISON SUMMARY:\*\*\s*(.*?)(?:\n\n|\n\*\*|$)`)
	if matches := summaryRe.FindStringSubmatch(text); len(matches) > 1 {
		result.ComparisonSummary = strings.TrimSpace(matches[1])
	}

	// If no structured data was found, provide defaults based on content analysis
	if result.VerificationOutcome == "" && result.ComparisonSummary == "" && len(result.Discrepancies) == 0 {
		// Analyze the text for common patterns to infer outcome
		lowerText := strings.ToLower(text)
		if strings.Contains(lowerText, "all") && (strings.Contains(lowerText, "filled") || strings.Contains(lowerText, "products")) {
			result.VerificationOutcome = "CORRECT"
			result.ComparisonSummary = "Analysis indicates all positions are properly filled with expected products."
		} else if strings.Contains(lowerText, "discrepanc") || strings.Contains(lowerText, "missing") || strings.Contains(lowerText, "incorrect") {
			result.VerificationOutcome = "INCORRECT"
			result.ComparisonSummary = "Analysis indicates potential discrepancies in product placement."
		} else {
			// Default to CORRECT if no clear issues are mentioned
			result.VerificationOutcome = "CORRECT"
			result.ComparisonSummary = "Initial analysis completed. No specific discrepancies identified in the provided response."
		}
	}

	return result, nil
}
