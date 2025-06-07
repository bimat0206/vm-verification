package parser

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"workflow-function/FinalizeAndStoreResults/internal/models"
)

// ParseTurn2ResponseData parses the turn2 processed response
func ParseTurn2ResponseData(data []byte) (*models.Turn2ParsedData, error) {
	// Try JSON first - check for the expected Turn2ParsedData structure
	var parsed models.Turn2ParsedData
	if err := json.Unmarshal(data, &parsed); err == nil && parsed.VerificationStatus != "" {
		return &parsed, nil
	}

	// Try parsing the actual JSON format from ExecuteTurn2Combined
	var turn2Response struct {
		Discrepancies      []interface{} `json:"discrepancies"`
		VerificationOutcome string       `json:"verificationOutcome"`
		ComparisonSummary  string       `json:"comparisonSummary"`
	}
	if err := json.Unmarshal(data, &turn2Response); err == nil {
		// Convert to Turn2ParsedData structure
		result := models.Turn2ParsedData{
			VerificationStatus: turn2Response.VerificationOutcome,
			InitialConfirmation: turn2Response.ComparisonSummary,
		}

		// Set verification summary based on the outcome
		if turn2Response.VerificationOutcome == "CORRECT" {
			result.VerificationSummary.VerificationOutcome = "CORRECT"
			result.VerificationSummary.OverallAccuracy = "100%"
			result.VerificationSummary.OverallConfidence = "High"
			// Set reasonable defaults for a CORRECT verification
			result.VerificationSummary.TotalPositionsChecked = 1 // At least 1 position was checked
			result.VerificationSummary.CorrectPositions = 1
			result.VerificationSummary.DiscrepantPositions = len(turn2Response.Discrepancies)
		} else {
			result.VerificationSummary.VerificationOutcome = turn2Response.VerificationOutcome
			result.VerificationSummary.DiscrepantPositions = len(turn2Response.Discrepancies)
		}

		return &result, nil
	}

	// Fallback to markdown/text parsing
	content := string(data)
	result := models.Turn2ParsedData{}

	// Extract VERIFICATION SUMMARY section - handle both bullet point and plain text formats
	summaryRe := regexp.MustCompile(`(?s)VERIFICATION SUMMARY:?\n(.*?)(?:\n\n|\n\*\*|$)`)
	matches := summaryRe.FindStringSubmatch(content)
	if len(matches) > 1 {
		summaryLines := strings.Split(strings.TrimSpace(matches[1]), "\n")
		for _, line := range summaryLines {
			// Handle bullet point format: * **Key:** Value
			bulletRe := regexp.MustCompile(`^\*\s*\*\*([^:*]+):\*\*\s*(.+)$`)
			if bulletMatches := bulletRe.FindStringSubmatch(line); len(bulletMatches) == 3 {
				key := strings.TrimSpace(bulletMatches[1])
				value := strings.TrimSpace(bulletMatches[2])
				parseKeyValue(key, value, &result)
				continue
			}

			// Handle plain format: Key: Value
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				parseKeyValue(key, value, &result)
			}
		}
	}

	// Also try to extract individual fields using specific patterns if summary section parsing failed
	if result.VerificationSummary.TotalPositionsChecked == 0 {
		extractIndividualFields(content, &result)
	}

	// Extract INITIAL CONFIRMATION section
	confirmRe := regexp.MustCompile(`(?s)INITIAL CONFIRMATION:?\n(.*?)\n\n`)
	cm := confirmRe.FindStringSubmatch(content)
	if len(cm) > 1 {
		result.InitialConfirmation = strings.TrimSpace(cm[1])
	}

	// Additional patterns to extract verification status if not found in summary
	if result.VerificationStatus == "" {
		// Try to find verification status in other common patterns, including markdown bullet points
		statusPatterns := []string{
			// Markdown bullet point format: * **VERIFICATION STATUS:** CORRECT
			`(?i)\*\s*\*\*verification\s+status:\*\*\s*(CORRECT|INCORRECT|PARTIAL|FAILED)`,
			// Standard colon format: VERIFICATION STATUS: CORRECT
			`(?i)verification\s+status:\s*(CORRECT|INCORRECT|PARTIAL|FAILED)`,
			// Simple status format: STATUS: CORRECT
			`(?i)status:\s*(CORRECT|INCORRECT|PARTIAL|FAILED)`,
			// Outcome format: VERIFICATION OUTCOME: CORRECT
			`(?i)verification\s+outcome:\s*(CORRECT|INCORRECT|PARTIAL|FAILED)`,
			// Simple outcome format: OUTCOME: CORRECT
			`(?i)outcome:\s*(CORRECT|INCORRECT|PARTIAL|FAILED)`,
		}

		for _, pattern := range statusPatterns {
			re := regexp.MustCompile(pattern)
			if matches := re.FindStringSubmatch(content); len(matches) > 1 {
				result.VerificationStatus = strings.ToUpper(matches[1])
				break
			}
		}
	}

	return &result, nil
}

// parseKeyValue parses a key-value pair and assigns it to the appropriate field in the result
func parseKeyValue(key, value string, result *models.Turn2ParsedData) {
	switch strings.ToLower(key) {
	case "total positions checked":
		fmt.Sscan(value, &result.VerificationSummary.TotalPositionsChecked)
	case "correct positions":
		fmt.Sscan(value, &result.VerificationSummary.CorrectPositions)
	case "discrepant positions":
		fmt.Sscan(value, &result.VerificationSummary.DiscrepantPositions)
	case "missing products":
		fmt.Sscan(value, &result.VerificationSummary.DiscrepancyDetails.MissingProducts)
	case "incorrect product types":
		fmt.Sscan(value, &result.VerificationSummary.DiscrepancyDetails.IncorrectProductTypes)
	case "unexpected products":
		fmt.Sscan(value, &result.VerificationSummary.DiscrepancyDetails.UnexpectedProducts)
	case "empty positions in checking image":
		fmt.Sscan(value, &result.VerificationSummary.EmptyPositionsInCheckingImage)
	case "overall accuracy":
		result.VerificationSummary.OverallAccuracy = value
	case "overall confidence":
		result.VerificationSummary.OverallConfidence = value
	case "verification outcome":
		result.VerificationSummary.VerificationOutcome = value
	case "verification status":
		result.VerificationStatus = value
	}
}

// extractIndividualFields tries to extract fields using individual regex patterns
func extractIndividualFields(content string, result *models.Turn2ParsedData) {
	// Individual field patterns for bullet point format
	patterns := map[string]*regexp.Regexp{
		"total_positions":    regexp.MustCompile(`(?i)\*\s*\*\*Total Positions Checked:\*\*\s*(\d+)`),
		"correct_positions":  regexp.MustCompile(`(?i)\*\s*\*\*Correct Positions:\*\*\s*(\d+)`),
		"discrepant_positions": regexp.MustCompile(`(?i)\*\s*\*\*Discrepant Positions:\*\*\s*(\d+)`),
		"missing_products":   regexp.MustCompile(`(?i)\*\s*Missing Products:\s*(\d+)`),
		"incorrect_types":    regexp.MustCompile(`(?i)\*\s*Incorrect Product Types:\s*(\d+)`),
		"unexpected_products": regexp.MustCompile(`(?i)\*\s*Unexpected Products:\s*(\d+)`),
		"empty_positions":    regexp.MustCompile(`(?i)\*\s*\*\*Empty Positions in Checking Image:\*\*\s*(\d+)`),
		"overall_accuracy":   regexp.MustCompile(`(?i)\*\s*\*\*Overall Accuracy:\*\*\s*([^*\n]+)`),
		"overall_confidence": regexp.MustCompile(`(?i)\*\s*\*\*Overall Confidence:\*\*\s*([^*\n]+)`),
		"verification_outcome": regexp.MustCompile(`(?i)\*\s*\*\*Verification Outcome:\*\*\s*([^*\n]+)`),
	}

	for field, pattern := range patterns {
		if matches := pattern.FindStringSubmatch(content); len(matches) > 1 {
			value := strings.TrimSpace(matches[1])
			switch field {
			case "total_positions":
				fmt.Sscan(value, &result.VerificationSummary.TotalPositionsChecked)
			case "correct_positions":
				fmt.Sscan(value, &result.VerificationSummary.CorrectPositions)
			case "discrepant_positions":
				fmt.Sscan(value, &result.VerificationSummary.DiscrepantPositions)
			case "missing_products":
				fmt.Sscan(value, &result.VerificationSummary.DiscrepancyDetails.MissingProducts)
			case "incorrect_types":
				fmt.Sscan(value, &result.VerificationSummary.DiscrepancyDetails.IncorrectProductTypes)
			case "unexpected_products":
				fmt.Sscan(value, &result.VerificationSummary.DiscrepancyDetails.UnexpectedProducts)
			case "empty_positions":
				fmt.Sscan(value, &result.VerificationSummary.EmptyPositionsInCheckingImage)
			case "overall_accuracy":
				result.VerificationSummary.OverallAccuracy = value
			case "overall_confidence":
				result.VerificationSummary.OverallConfidence = value
			case "verification_outcome":
				result.VerificationSummary.VerificationOutcome = value
			}
		}
	}
}
