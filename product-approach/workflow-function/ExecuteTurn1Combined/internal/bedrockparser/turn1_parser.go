package bedrockparser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"workflow-function/shared/schema"
)

// ParsedTurn1Data holds the structured results from parsing Bedrock's Turn 1 response.
type ParsedTurn1Data struct {
	InitialConfirmation string                  `json:"initialConfirmation"`
	MachineStructure    schema.MachineStructure `json:"machineStructure"`
	ReferenceRowStatus  map[string]string       `json:"referenceRowStatus"`
	ReferenceSummary    string                  `json:"referenceSummary"`
}

// ParseTurn1Response parses the Markdown-like text from Bedrock's Turn 1 analysis
// into a structured ParsedTurn1Data object.
func ParseTurn1Response(text string) (*ParsedTurn1Data, error) {
	if strings.TrimSpace(text) == "" {
		return nil, nil
	}

	result := &ParsedTurn1Data{
		ReferenceRowStatus: make(map[string]string),
	}

	// Normalize line endings
	text = strings.ReplaceAll(text, "\r\n", "\n")

	// Extract Initial Confirmation section
	reInit := regexp.MustCompile(`(?s)\*\*INITIAL CONFIRMATION:\*\*\s*(.*?)\n\s*\*\*ROW STATUS ANALYSIS`)
	if matches := reInit.FindStringSubmatch(text); len(matches) > 1 {
		section := strings.TrimSpace(matches[1])
		result.InitialConfirmation = section

		// Parse machine structure details from the confirmation text
		// Example sentences:
		// "Successfully identified 6 physical rows (A-Top to F-Bottom)."
		// "Successfully identified 7 slots per row (01-Left to 07-Right)."
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

	// Extract Reference Image Summary
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
