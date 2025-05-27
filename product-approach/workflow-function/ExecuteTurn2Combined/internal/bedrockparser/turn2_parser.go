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
	Discrepancies      []models.Discrepancy `json:"discrepancies"`
	VerificationOutcome string               `json:"verificationOutcome"`
	ComparisonSummary  string               `json:"comparisonSummary"`
}

// ParseBedrockResponseAsMarkdown extracts and cleans the Bedrock Turn 2 textual response.
func ParseBedrockResponseAsMarkdown(bedrockTextResponse string) (*ParsedTurn2Markdown, error) {
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
// actionable content was found.
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
				
				if len(matches) > 3 && matches[3] != "" {
					found = strings.TrimSpace(matches[3])
					discrepancyType = "MISPLACED"
				}
				
				result.Discrepancies = append(result.Discrepancies, models.Discrepancy{
					Item:     item,
					Expected: expected,
					Found:    found,
					Type:     discrepancyType,
				})
			}
		}
	}

	// Extract comparison summary
	summaryRe := regexp.MustCompile(`(?s)\*\*COMPARISON SUMMARY:\*\*\s*(.*?)(?:\n\n|\n\*\*|$)`)
	if matches := summaryRe.FindStringSubmatch(text); len(matches) > 1 {
		result.ComparisonSummary = strings.TrimSpace(matches[1])
	}

	return result, nil
}