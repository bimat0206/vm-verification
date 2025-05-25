package bedrockparser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"workflow-function/shared/schema"
)

// ParsedTurn1Markdown holds the cleaned Markdown content from Bedrock's Turn 1 response.
type ParsedTurn1Markdown struct {
	AnalysisMarkdown string `json:"analysisMarkdown"`
}

// ParsedTurn1Data holds structured fields extracted from the Turn 1 Markdown.
// This mirrors the earlier parser behaviour so dependent code continues to
// compile, even though the Markdown-centric approach may not require these
// fields at runtime.
type ParsedTurn1Data struct {
	InitialConfirmation string                  `json:"initialConfirmation"`
	MachineStructure    schema.MachineStructure `json:"machineStructure"`
	ReferenceRowStatus  map[string]string       `json:"referenceRowStatus"`
	ReferenceSummary    string                  `json:"referenceSummary"`
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

// ParseTurn1Response converts the markdown formatted analysis into a structured
// ParsedTurn1Data. If the input text is empty, nil is returned to signal no
// actionable content was found.
func ParseTurn1Response(text string) (*ParsedTurn1Data, error) {
	if strings.TrimSpace(text) == "" {
		return nil, nil
	}

	result := &ParsedTurn1Data{
		ReferenceRowStatus: make(map[string]string),
	}

	// Normalize line endings for consistent regex processing
	text = strings.ReplaceAll(text, "\r\n", "\n")

	// Extract INITIAL CONFIRMATION section
	reInit := regexp.MustCompile(`(?s)\*\*INITIAL CONFIRMATION:\*\*\s*(.*?)\n\s*\*\*ROW STATUS ANALYSIS`)
	if matches := reInit.FindStringSubmatch(text); len(matches) > 1 {
		section := strings.TrimSpace(matches[1])
		result.InitialConfirmation = section

		rowRe := regexp.MustCompile(`(?i)(\d+)\s+physical\s+rows\s*\((\w)-Top\s+to\s+(\w)-Bottom\)`)
		colRe := regexp.MustCompile(`(?i)(\d+)\s+slots\s+per\s+row\s*\(01-Left\s+to\s+(\d+)\-Right\)`)
		if rm := rowRe.FindStringSubmatch(section); len(rm) == 4 {
			rowCount := rm[1]
			startRow := rm[2]
			endRow := rm[3]
			if n, err := strconv.Atoi(rowCount); err == nil {
				result.MachineStructure.RowCount = n
				result.MachineStructure.RowOrder = buildRowOrder(startRow, endRow)
			}
		}
		if cm := colRe.FindStringSubmatch(section); len(cm) == 3 {
			colCount := cm[1]
			endCol := cm[2]
			if n, err := strconv.Atoi(colCount); err == nil {
				result.MachineStructure.ColumnsPerRow = n
				result.MachineStructure.ColumnOrder = buildColumnOrder(n, endCol)
			}
		}
	}

	// Extract Reference Row Status section
	reRows := regexp.MustCompile(`(?s)\*\*ROW STATUS ANALYSIS \(Reference Image\):\*\*\n(.*?)\n\s*\*\*REFERENCE IMAGE SUMMARY`)
	if matches := reRows.FindStringSubmatch(text); len(matches) > 1 {
		rowsBlock := strings.TrimSpace(matches[1])
		lines := strings.Split(rowsBlock, "\n")
		rowRe := regexp.MustCompile(`\*\s*\*\*Row\s+(\w+)\s*(?:\(.*?\))?:\*\*\s*(.*)`)
		for _, l := range lines {
			l = strings.TrimSpace(l)
			if l == "" {
				continue
			}
			if rm := rowRe.FindStringSubmatch(l); len(rm) == 3 {
				rowLabel := rm[1]
				desc := strings.TrimSpace(rm[2])
				result.ReferenceRowStatus[rowLabel] = desc
			}
		}
	}

	// Extract Reference Image Summary section
	reSummary := regexp.MustCompile(`(?s)\*\*REFERENCE IMAGE SUMMARY:\*\*\s*(.*?)\n\s*(?:\*\*|$)`)
	if matches := reSummary.FindStringSubmatch(text); len(matches) > 1 {
		result.ReferenceSummary = strings.TrimSpace(matches[1])
	}

	return result, nil
}

func buildRowOrder(start, end string) []string {
	startRune := []rune(start)[0]
	endRune := []rune(end)[0]
	if endRune < startRune {
		return nil
	}
	var rows []string
	for r := startRune; r <= endRune; r++ {
		rows = append(rows, string(r))
	}
	return rows
}

func buildColumnOrder(count int, endCol string) []string {
	var cols []string
	for i := 1; i <= count; i++ {
		cols = append(cols, fmt.Sprintf("%02d", i))
	}
	return cols
}
