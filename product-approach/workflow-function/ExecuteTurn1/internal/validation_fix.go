package internal

import (
	"fmt"
)

// ExtractAndValidateCurrentPrompt extracts and validates the current prompt
func ExtractAndValidateCurrentPrompt(input *ExecuteTurn1Input) (*CurrentPrompt, error) {
	// Handle the nested currentPrompt structure
	if input.CurrentPrompt.CurrentPrompt.Messages != nil {
		return ValidateExtractedCurrentPrompt(&input.CurrentPrompt.CurrentPrompt)
	}
	
	// Fallback: try to use the outer structure directly
	if len(input.CurrentPrompt.Messages) > 0 {
		prompt := CurrentPrompt{
			Messages:      input.CurrentPrompt.Messages,
			TurnNumber:    input.CurrentPrompt.TurnNumber,
			PromptID:      input.CurrentPrompt.PromptID,
			CreatedAt:     input.CurrentPrompt.CreatedAt,
			PromptVersion: input.CurrentPrompt.PromptVersion,
			ImageIncluded: input.CurrentPrompt.ImageIncluded,
		}
		return ValidateExtractedCurrentPrompt(&prompt)
	}
	
	return nil, NewValidationError("No valid currentPrompt found in input", nil)
}

// ValidateExtractedCurrentPrompt validates an extracted CurrentPrompt
func ValidateExtractedCurrentPrompt(prompt *CurrentPrompt) (*CurrentPrompt, error) {
	// Validate turn number
	if prompt.TurnNumber != ExpectedTurnNumber {
		return nil, NewInvalidFieldError("turnNumber", prompt.TurnNumber, fmt.Sprintf("%d", ExpectedTurnNumber))
	}

	// Validate image included
	if prompt.ImageIncluded != ExpectedImageIncluded {
		return nil, NewInvalidFieldError("imageIncluded", prompt.ImageIncluded, ExpectedImageIncluded)
	}

	// Validate messages
	if len(prompt.Messages) == 0 {
		return nil, NewValidationError("Messages cannot be empty", nil)
	}

	// Validate the first (and should be only) message
	if err := ValidateBedrockMessage(&prompt.Messages[0]); err != nil {
		return nil, err
	}

	// Additional validation for prompt metadata
	if prompt.PromptID == "" {
		return nil, NewMissingFieldError("promptId")
	}

	if prompt.PromptVersion == "" {
		return nil, NewMissingFieldError("promptVersion")
	}

	return prompt, nil
}

// ExtractBucketOwner extracts bucket owner from the input
func ExtractBucketOwner(input *ExecuteTurn1Input) string {
	if input.Images != nil && input.Images.BucketOwner != "" {
		return input.Images.BucketOwner
	}
	
	if input.Images != nil && input.Images.ReferenceImageMeta.BucketOwner != "" {
		return input.Images.ReferenceImageMeta.BucketOwner
	}
	
	// Default bucket owner
	return "defaultBucketOwner"
}