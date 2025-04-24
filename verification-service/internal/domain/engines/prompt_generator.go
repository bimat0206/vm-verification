package engines

import (
	"context"
	"fmt"

	"verification-service/internal/domain/models"
)

// PromptGenerator implements the domain.PromptGenerator interface
type PromptGenerator struct{}

// NewPromptGenerator creates a new prompt generator
func NewPromptGenerator() *PromptGenerator {
	return &PromptGenerator{}
}

// GenerateSystemPrompt creates the base system prompt for the verification process
func (g *PromptGenerator) GenerateSystemPrompt(
	ctx context.Context,
	verificationContext models.VerificationContext,
	layoutMetadata map[string]interface{},
) (string, error) {
	// Create a comprehensive system prompt for the verification task
	systemPrompt := `You are an AI assistant specialized in analyzing vending machine product placement. Your task is to perform a two-turn analysis:

Turn 1: Analyze the reference layout image to understand what products SHOULD be in each position of the vending machine.
Turn 2: Analyze the checking image and compare it to your understanding of the reference layout to identify discrepancies.

Focus on these specific aspects:
1. Position discrepancies: Check if products are in the correct position (row/column)
2. Product identity: Verify if the correct product is placed in each position
3. Quantity: Check if the expected number of products is present in each position
4. Label visibility: Ensure product labels are visible and properly oriented

For your final output in Turn 2, provide your analysis in structured JSON format with the following fields:
- discrepancies: Array of detected discrepancies
- totalDiscrepancies: Total number of discrepancies found
- severity: Overall assessment (Low, Medium, High)

For each discrepancy, include:
- position: The position (e.g., "A1", "B3")
- expected: What should be in this position according to the reference
- found: What was actually found in the checking image
- issue: The type of discrepancy (Position, Identity, Quantity, Visibility)
- confidence: Your confidence level (0-100)
- evidence: Brief description of visual evidence
- verificationResult: "CORRECT" or "INCORRECT"`

	return systemPrompt, nil
}

// GenerateTurn1Prompt creates the prompt for reference layout analysis
func (g *PromptGenerator) GenerateTurn1Prompt(
	ctx context.Context,
	verificationContext models.VerificationContext,
	layoutMetadata map[string]interface{}, 
	referenceImageB64 string,
) (string, error) {
	// Extract machine structure from metadata if available
	machineStructureDesc := "The vending machine has multiple rows and columns."
	if meta, ok := layoutMetadata["machineStructure"].(map[string]interface{}); ok {
		if rowCount, ok := meta["rowCount"].(float64); ok {
			if colCount, ok := meta["columnsPerRow"].(float64); ok {
				machineStructureDesc = fmt.Sprintf("The vending machine has %d rows (A-%s) and %d columns per row.",
					int(rowCount),
					string(rune('A'+int(rowCount)-1)),
					int(colCount))
			}
		}
	}

	// Create prompt for Turn 1
	prompt := fmt.Sprintf(`The FIRST image provided ALWAYS depicts the Reference Layout (the expected state). 

TASK: Analyze the reference layout image for vending machine ID: %s.

%s

Your goal for this first turn is to:
1. Identify all products visible in the reference image
2. Map them to their correct positions (row and column)
3. Create a structured understanding of what SHOULD be in each position
4. Note any empty positions in the reference layout

Do NOT analyze any checking images yet - that will be in Turn 2.`, 
		verificationContext.VendingMachineID,
		machineStructureDesc)

	return prompt, nil
}

// GenerateTurn2Prompt creates the prompt for checking image comparison
func (g *PromptGenerator) GenerateTurn2Prompt(
	ctx context.Context,
	verificationContext models.VerificationContext,
	layoutMetadata map[string]interface{},
	checkingImageB64 string,
	referenceAnalysis *models.ReferenceAnalysis,
) (string, error) {
	// Create prompt for Turn 2
	prompt := fmt.Sprintf(`The SECOND image provided ALWAYS depicts the Checking Image (the current state to be verified).

TASK: Compare the checking image with your understanding of the reference layout for vending machine ID: %s.

Based on your analysis of the reference layout in Turn 1, now:
1. Identify what products are actually present in the checking image
2. Compare them with what SHOULD be there according to the reference layout
3. Detect any discrepancies in positioning, product identity, quantity, or label visibility
4. Pay special attention to empty positions that should be filled
5. Generate structured JSON output of all discrepancies found

Remember to output your final analysis as JSON with the format specified in the system prompt.`, 
		verificationContext.VendingMachineID)

	return prompt, nil
}