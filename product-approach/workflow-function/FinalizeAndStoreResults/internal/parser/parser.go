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
	// Try JSON first
	var parsed models.Turn2ParsedData
	if err := json.Unmarshal(data, &parsed); err == nil && parsed.VerificationStatus != "" {
		return &parsed, nil
	}

	// Fallback to markdown/text parsing
	content := string(data)
	result := models.Turn2ParsedData{}

	// Extract VERIFICATION SUMMARY section
	summaryRe := regexp.MustCompile(`(?s)VERIFICATION SUMMARY:?\n(.*?)\n\n`)
	matches := summaryRe.FindStringSubmatch(content)
	if len(matches) > 1 {
		summaryLines := strings.Split(strings.TrimSpace(matches[1]), "\n")
		for _, line := range summaryLines {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
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
	}

	// Extract INITIAL CONFIRMATION section
	confirmRe := regexp.MustCompile(`(?s)INITIAL CONFIRMATION:?\n(.*?)\n\n`)
	cm := confirmRe.FindStringSubmatch(content)
	if len(cm) > 1 {
		result.InitialConfirmation = strings.TrimSpace(cm[1])
	}

	return &result, nil
}
