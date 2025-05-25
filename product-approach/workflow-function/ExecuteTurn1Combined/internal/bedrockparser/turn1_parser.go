package bedrockparser

import (
	"strings"
)

// ParsedTurn1Markdown holds the cleaned Markdown content from Bedrock's Turn 1 response.
type ParsedTurn1Markdown struct {
	AnalysisMarkdown string `json:"analysisMarkdown"`
}

// ParseBedrockResponseAsMarkdown extracts and cleans the Bedrock Turn 1 textual response.
func ParseBedrockResponseAsMarkdown(bedrockTextResponse string) (*ParsedTurn1Markdown, error) {
	if strings.TrimSpace(bedrockTextResponse) == "" {
		// Return empty markdown struct for empty input
		return &ParsedTurn1Markdown{AnalysisMarkdown: ""}, nil
	}

	cleaned := strings.ReplaceAll(bedrockTextResponse, "\r\n", "\n")
	cleaned = strings.TrimSpace(cleaned)

	return &ParsedTurn1Markdown{AnalysisMarkdown: cleaned}, nil
}
