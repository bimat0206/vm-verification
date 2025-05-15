package internal

import (
	"fmt"
	"workflow-function/shared/bedrock"
	"workflow-function/shared/errors"
	"workflow-function/shared/schema"
)

// ExecuteTurn1Input represents the input to ExecuteTurn1 function
type ExecuteTurn1Input struct {
	VerificationContext schema.VerificationContext `json:"verificationContext"`
	CurrentPrompt       CurrentPromptWrapper       `json:"currentPrompt"`
	BedrockConfig       schema.BedrockConfig       `json:"bedrockConfig"`
	Images              *schema.Images             `json:"images,omitempty"`
	LayoutMetadata      *schema.LayoutMetadata     `json:"layoutMetadata,omitempty"`
	SystemPrompt        *SystemPromptWrapper       `json:"systemPrompt,omitempty"`
	HistoricalContext   map[string]interface{}     `json:"historicalContext,omitempty"`
	ConversationState   map[string]interface{}     `json:"conversationState,omitempty"`
}

// CurrentPromptWrapper wraps the nested currentPrompt structure
type CurrentPromptWrapper struct {
	CurrentPrompt CurrentPrompt            `json:"currentPrompt"`
	Messages      []bedrock.MessageWrapper `json:"messages,omitempty"`
	TurnNumber    int                      `json:"turnNumber,omitempty"`
	PromptID      string                   `json:"promptId,omitempty"`
	CreatedAt     string                   `json:"createdAt,omitempty"`
	PromptVersion string                   `json:"promptVersion,omitempty"`
	ImageIncluded string                   `json:"imageIncluded,omitempty"`
}

// CurrentPrompt represents the prompt structure for Turn 1
type CurrentPrompt struct {
	Messages      []bedrock.MessageWrapper `json:"messages"`
	TurnNumber    int                      `json:"turnNumber"`
	PromptID      string                   `json:"promptId"`
	CreatedAt     string                   `json:"createdAt"`
	PromptVersion string                   `json:"promptVersion"`
	ImageIncluded string                   `json:"imageIncluded"`
}

// SystemPromptWrapper wraps the nested systemPrompt structure
type SystemPromptWrapper struct {
	SystemPrompt SystemPrompt `json:"systemPrompt"`
}

// SystemPrompt represents the system prompt data
type SystemPrompt struct {
	Content       string `json:"content"`
	PromptID      string `json:"promptId"`
	CreatedAt     string `json:"createdAt"`
	PromptVersion string `json:"promptVersion"`
}

// ExtractAndValidateCurrentPrompt extracts the current prompt from the input
// This is a compatibility function to bridge between the ExecuteTurn1 input structure
// and the shared bedrock validation functions
func ExtractAndValidateCurrentPrompt(input *ExecuteTurn1Input) (*CurrentPrompt, error) {
	if input == nil {
		return nil, errors.NewValidationError("Input cannot be nil", nil)
	}

	// Extract current prompt
	currentPrompt := &input.CurrentPrompt.CurrentPrompt
	if currentPrompt == nil {
		return nil, errors.NewMissingFieldError("currentPrompt")
	}

	// Validate turn number
	if currentPrompt.TurnNumber != 1 {
		return nil, errors.NewInvalidFieldError("turnNumber", currentPrompt.TurnNumber, "1")
	}

	// Validate image included
	if currentPrompt.ImageIncluded != "reference" {
		return nil, errors.NewInvalidFieldError("imageIncluded", currentPrompt.ImageIncluded, "reference")
	}

	// Validate messages
	if len(currentPrompt.Messages) == 0 {
		return nil, errors.NewValidationError("Messages cannot be empty", nil)
	}

	// Validate the messages using shared bedrock validation
	for i, msg := range currentPrompt.Messages {
		if err := bedrock.ValidateMessageWrapper(msg); err != nil {
			return nil, errors.NewValidationError(
				fmt.Sprintf("Invalid message at index %d: %v", i, err),
				map[string]interface{}{"index": i},
			)
		}
	}

	// Additional validation for prompt metadata
	if currentPrompt.PromptID == "" {
		return nil, errors.NewMissingFieldError("promptId")
	}

	if currentPrompt.PromptVersion == "" {
		return nil, errors.NewMissingFieldError("promptVersion")
	}

	return currentPrompt, nil
}

// ExtractBucketOwner extracts the bucket owner from the input
func ExtractBucketOwner(input *ExecuteTurn1Input) string {
	// Try to get from Images
	if input.Images != nil && input.Images.BucketOwner != "" {
		return input.Images.BucketOwner
	}

	// Try to get from reference image metadata
	if input.Images != nil && input.Images.ReferenceImageMeta.BucketOwner != "" {
		return input.Images.ReferenceImageMeta.BucketOwner
	}

	// Try to get from checking image metadata
	if input.Images != nil && input.Images.CheckingImageMeta.BucketOwner != "" {
		return input.Images.CheckingImageMeta.BucketOwner
	}

	// Default value
	return ""
}
