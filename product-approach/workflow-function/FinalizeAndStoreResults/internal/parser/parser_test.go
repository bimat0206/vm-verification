package parser

import (
	"testing"
)

func TestParseVerificationStatus(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "verification status in summary section",
			input: `VERIFICATION SUMMARY:
Total positions checked: 10
Correct positions: 8
Discrepant positions: 2
Verification status: INCORRECT

INITIAL CONFIRMATION:
Some confirmation text`,
			expected: "INCORRECT",
		},
		{
			name: "verification outcome pattern",
			input: `Some analysis text here.

Verification outcome: CORRECT

More text here.`,
			expected: "CORRECT",
		},
		{
			name: "status pattern",
			input: `Analysis complete.
Status: PARTIAL
End of analysis.`,
			expected: "PARTIAL",
		},
		{
			name: "no status found",
			input: `Some text without any status information.
Just analysis results.`,
			expected: "",
		},
		{
			name: "case insensitive matching",
			input: `verification Status: correct`,
			expected: "CORRECT",
		},
		{
			name: "markdown bullet point format",
			input: `Some analysis text here.

* **VERIFICATION STATUS:** INCORRECT

More details about the verification.`,
			expected: "INCORRECT",
		},
		{
			name: "markdown bullet point format with correct status",
			input: `Analysis complete.

* **VERIFICATION STATUS:** CORRECT

End of analysis.`,
			expected: "CORRECT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTurn2ResponseData([]byte(tt.input))
			if err != nil {
				t.Fatalf("ParseTurn2ResponseData failed: %v", err)
			}

			if result.VerificationStatus != tt.expected {
				t.Errorf("Expected verificationStatus %q, got %q", tt.expected, result.VerificationStatus)
			}
		})
	}
}

func TestParseJSONInput(t *testing.T) {
	jsonInput := `{
		"VerificationSummary": {
			"total_positions_checked": 5,
			"correct_positions": 5,
			"discrepant_positions": 0
		},
		"InitialConfirmation": "All positions verified",
		"VerificationStatus": "CORRECT"
	}`

	result, err := ParseTurn2ResponseData([]byte(jsonInput))
	if err != nil {
		t.Fatalf("ParseTurn2ResponseData failed: %v", err)
	}

	if result.VerificationStatus != "CORRECT" {
		t.Errorf("Expected verificationStatus 'CORRECT', got %q", result.VerificationStatus)
	}

	if result.InitialConfirmation != "All positions verified" {
		t.Errorf("Expected initialConfirmation 'All positions verified', got %q", result.InitialConfirmation)
	}
}

func TestParseActualTurn2Format(t *testing.T) {
	// Test with the actual format from the turn2-processed-response.md file
	input := `**VENDING MACHINE LAYOUT VERIFICATION REPORT**

**INITIAL CONFIRMATION:**
- Physical rows identified: After analyzing the provided images, I have identified 6 distinct physical rows based on the shelf structure, labeled from A (Top) to F (Bottom).
- Slots per row identified: Through visual inspection, I have determined that each row contains 7 slots, labeled from 01 (Left) to 07 (Right).
- Proceeding with analysis based on this 6x7 physical structure, as determined from the images.

**VERIFICATION SUMMARY:**
* **Total Positions Checked:** 42
* **Correct Positions:** 35
* **Discrepant Positions:** 7
    * Missing Products: 7
    * Incorrect Product Types: 0
    * Unexpected Products: 0
* **Empty Positions in Checking Image:** 7
* **Overall Accuracy:** 83.3% (35/42)
* **Overall Confidence:** 100%
* **VERIFICATION STATUS:** INCORRECT
* **Verification Outcome:** Discrepancies Detected - Entire bottom row (Row F) is missing all "Mì Cung Đình" products`

	result, err := ParseTurn2ResponseData([]byte(input))
	if err != nil {
		t.Fatalf("ParseTurn2ResponseData failed: %v", err)
	}

	// Verify verification status
	if result.VerificationStatus != "INCORRECT" {
		t.Errorf("Expected verificationStatus 'INCORRECT', got %q", result.VerificationStatus)
	}

	// Verify summary fields
	if result.VerificationSummary.TotalPositionsChecked != 42 {
		t.Errorf("Expected TotalPositionsChecked 42, got %d", result.VerificationSummary.TotalPositionsChecked)
	}

	if result.VerificationSummary.CorrectPositions != 35 {
		t.Errorf("Expected CorrectPositions 35, got %d", result.VerificationSummary.CorrectPositions)
	}

	if result.VerificationSummary.DiscrepantPositions != 7 {
		t.Errorf("Expected DiscrepantPositions 7, got %d", result.VerificationSummary.DiscrepantPositions)
	}

	if result.VerificationSummary.DiscrepancyDetails.MissingProducts != 7 {
		t.Errorf("Expected MissingProducts 7, got %d", result.VerificationSummary.DiscrepancyDetails.MissingProducts)
	}

	if result.VerificationSummary.DiscrepancyDetails.IncorrectProductTypes != 0 {
		t.Errorf("Expected IncorrectProductTypes 0, got %d", result.VerificationSummary.DiscrepancyDetails.IncorrectProductTypes)
	}

	if result.VerificationSummary.DiscrepancyDetails.UnexpectedProducts != 0 {
		t.Errorf("Expected UnexpectedProducts 0, got %d", result.VerificationSummary.DiscrepancyDetails.UnexpectedProducts)
	}

	if result.VerificationSummary.EmptyPositionsInCheckingImage != 7 {
		t.Errorf("Expected EmptyPositionsInCheckingImage 7, got %d", result.VerificationSummary.EmptyPositionsInCheckingImage)
	}

	if result.VerificationSummary.OverallAccuracy != "83.3% (35/42)" {
		t.Errorf("Expected OverallAccuracy '83.3%% (35/42)', got %q", result.VerificationSummary.OverallAccuracy)
	}

	if result.VerificationSummary.OverallConfidence != "100%" {
		t.Errorf("Expected OverallConfidence '100%%', got %q", result.VerificationSummary.OverallConfidence)
	}

	expectedOutcome := "Discrepancies Detected - Entire bottom row (Row F) is missing all \"Mì Cung Đình\" products"
	if result.VerificationSummary.VerificationOutcome != expectedOutcome {
		t.Errorf("Expected VerificationOutcome %q, got %q", expectedOutcome, result.VerificationSummary.VerificationOutcome)
	}
}
