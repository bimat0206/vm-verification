package bedrock

import (
	"fmt"
	//"regexp"
	"strings"
)

// Validation constants
const (
	// Expected turn numbers
	ExpectedTurn1Number = 1
	ExpectedTurn2Number = 2

	// Expected image included value for Turn 1
	ExpectedImageIncluded = "reference"
	
	// Analysis stage identifiers
	AnalysisStageTurn1 = "TURN1"
	AnalysisStageTurn2 = "TURN2"
)

var (
	// Supported image formats
	supportedImageFormats = []string{"png", "jpeg", "jpg", "gif", "webp"}
)

// ValidateConverseRequest validates a Converse API request
func ValidateConverseRequest(request *ConverseRequest) error {
	if request == nil {
		return fmt.Errorf("converse request cannot be nil")
	}

	if request.ModelId == "" {
		return fmt.Errorf("model ID cannot be empty")
	}

	if len(request.Messages) == 0 {
		return fmt.Errorf("messages cannot be empty")
	}

	// Validate messages
	for i, msg := range request.Messages {
		if err := ValidateMessageWrapper(msg); err != nil {
			return fmt.Errorf("invalid message at index %d: %w", i, err)
		}
	}

	// Validate inference config
	if request.InferenceConfig.MaxTokens <= 0 {
		return fmt.Errorf("max tokens must be positive")
	}

	return nil
}

// ValidateMessageWrapper validates a message wrapper
func ValidateMessageWrapper(msg MessageWrapper) error {
	if msg.Role != "user" && msg.Role != "assistant" && msg.Role != "system" {
		return fmt.Errorf("invalid role: %s", msg.Role)
	}

	if len(msg.Content) == 0 {
		return fmt.Errorf("content cannot be empty")
	}

	// Validate content blocks
	for i, content := range msg.Content {
		if err := ValidateContentBlock(content); err != nil {
			return fmt.Errorf("invalid content block at index %d: %w", i, err)
		}
	}

	return nil
}

// ValidateContentBlock validates a content block
func ValidateContentBlock(content ContentBlock) error {
	switch content.Type {
	case "text":
		if content.Text == "" {
			return fmt.Errorf("text content cannot be empty")
		}
	case "image":
		if content.Image == nil {
			return fmt.Errorf("image content cannot be nil")
		}
		if err := ValidateImageBlock(content.Image); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported content type: %s", content.Type)
	}

	return nil
}

// ValidateImageBlock validates an image block
func ValidateImageBlock(img *ImageBlock) error {
	if img == nil {
		return fmt.Errorf("image block cannot be nil")
	}

	// Validate format
	if !isValidImageFormat(img.Format) {
		return fmt.Errorf("invalid image format: %s", img.Format)
	}

	// Validate source
	if err := ValidateImageSource(img.Source); err != nil {
		return err
	}

	return nil
}

// ValidateImageSource validates an image source
func ValidateImageSource(source ImageSource) error {
	// Check the type of source
	switch source.Type {
	case "bytes":
		// For bytes type, validate the base64 data
		if source.Bytes == "" {
			return fmt.Errorf("base64 data cannot be empty for bytes type")
		}
		// Additional base64 validation could be added here if needed
	case "":
		// If type is not specified, check if we can infer it
		if source.Bytes != "" {
			// Infer bytes type
		} else {
			return fmt.Errorf("image source must have bytes data")
		}
	default:
		return fmt.Errorf("unsupported image source type: %s", source.Type)
	}

	return nil
}

// Helper functions

// isValidImageFormat checks if the image format is supported
func isValidImageFormat(format string) bool {
	format = strings.ToLower(format)
	for _, supported := range supportedImageFormats {
		if format == supported {
			return true
		}
	}
	return false
}

// ExtractAndValidateCurrentPrompt extracts the current prompt from the input
func ExtractAndValidateCurrentPrompt(messages []MessageWrapper) error {
	if len(messages) == 0 {
		return fmt.Errorf("no messages in current prompt")
	}

	// Validate the first (and should be only) message
	if err := ValidateMessageWrapper(messages[0]); err != nil {
		return err
	}

	return nil
}

// ValidateTurn1Response validates a Turn 1 response
func ValidateTurn1Response(turn1Response *Turn1Response) error {
	if turn1Response == nil {
		return fmt.Errorf("turn 1 response cannot be nil")
	}

	if turn1Response.TurnID != ExpectedTurn1Number {
		return fmt.Errorf("unexpected turn ID: got %d, expected %d", turn1Response.TurnID, ExpectedTurn1Number)
	}

	if turn1Response.AnalysisStage != AnalysisStageTurn1 {
		return fmt.Errorf("unexpected analysis stage: got %s, expected %s", turn1Response.AnalysisStage, AnalysisStageTurn1)
	}

	if turn1Response.Response.Content == "" {
		return fmt.Errorf("response content cannot be empty")
	}

	return nil
}

// ValidateTurn2Response validates a Turn 2 response
func ValidateTurn2Response(turn2Response *Turn2Response) error {
	if turn2Response == nil {
		return fmt.Errorf("turn 2 response cannot be nil")
	}

	if turn2Response.TurnID != ExpectedTurn2Number {
		return fmt.Errorf("unexpected turn ID: got %d, expected %d", turn2Response.TurnID, ExpectedTurn2Number)
	}

	if turn2Response.AnalysisStage != AnalysisStageTurn2 {
		return fmt.Errorf("unexpected analysis stage: got %s, expected %s", turn2Response.AnalysisStage, AnalysisStageTurn2)
	}

	if turn2Response.Response.Content == "" {
		return fmt.Errorf("response content cannot be empty")
	}

	if turn2Response.PreviousTurn == nil {
		return fmt.Errorf("previous turn cannot be nil for turn 2 response")
	}

	// Validate the previous turn 1 response
	if err := ValidateTurn1Response(turn2Response.PreviousTurn); err != nil {
		return fmt.Errorf("invalid previous turn: %w", err)
	}

	return nil
}