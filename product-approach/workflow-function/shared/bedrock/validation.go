package bedrock

import (
	"fmt"
	"strings"
	"regexp"
)

// Validation constants
const (
	// Expected image included value for Turn 1
	ExpectedImageIncluded = "reference"
)

var (
	// Supported image formats for Bedrock Converse API
	supportedConverseFormats = []string{"png", "jpeg"}
	
	// Expanded supported image formats for our system (will be converted as needed)
	supportedInputFormats = []string{"png", "jpeg", "jpg"}
	
	// Regex for validating base64 data
	base64Regex = regexp.MustCompile(`^[A-Za-z0-9+/]*={0,2}$`)
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

	// Validate system prompt if provided
	if request.System != "" {
		if len(request.System) > 32000 {
			return fmt.Errorf("system prompt too long: %d characters (max 32000)", len(request.System))
		}
	}

	// Validate guardrail config if provided
	if request.GuardrailConfig != nil {
		if request.GuardrailConfig.GuardrailIdentifier == "" {
			return fmt.Errorf("guardrail identifier cannot be empty")
		}
	}

	// Validate temperature and thinking mode compatibility
	if err := ValidateTemperatureThinkingCompatibility(request); err != nil {
		return err
	}

	return nil
}

// ValidateTemperatureThinkingCompatibility validates that temperature and thinking mode are compatible
func ValidateTemperatureThinkingCompatibility(request *ConverseRequest) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	// Check if temperature is set to 1.0
	if request.InferenceConfig.Temperature != nil && *request.InferenceConfig.Temperature >= 1.0 {
		// When temperature is 1.0, thinking mode must be enabled
		thinkingEnabled := false

		// Check if thinking is enabled via structured thinking field
		if request.Thinking != nil && len(request.Thinking) > 0 {
			if thinkingType, ok := request.Thinking["type"]; ok {
				if typeStr, ok := thinkingType.(string); ok && typeStr == "enabled" {
					thinkingEnabled = true
				}
			}
		}

		// Check if thinking is enabled via legacy reasoning fields
		if !thinkingEnabled {
			reasoning := request.InferenceConfig.Reasoning
			if reasoning == "" {
				reasoning = request.Reasoning
			}
			if reasoning == "enable" || reasoning == "enabled" {
				thinkingEnabled = true
			}
		}

		if !thinkingEnabled {
			return fmt.Errorf("temperature may only be set to 1 when thinking is enabled. Please set thinking type to 'enabled' or reduce temperature below 1.0")
		}
	}

	return nil
}

// ValidateMessageWrapper validates a message wrapper
func ValidateMessageWrapper(msg MessageWrapper) error {
	if msg.Role != "user" && msg.Role != "assistant" && msg.Role != "system" {
		return fmt.Errorf("invalid role: %s (must be 'user', 'assistant', or 'system')", msg.Role)
	}

	if len(msg.Content) == 0 {
		return fmt.Errorf("content cannot be empty for role: %s", msg.Role)
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
			return fmt.Errorf("text content cannot be empty for text content block")
		}
	case "image":
		if content.Image == nil {
			return fmt.Errorf("image field cannot be nil for image content block")
		}
		if err := ValidateImageBlock(content.Image); err != nil {
			return fmt.Errorf("invalid image block: %w", err)
		}
	default:
		return fmt.Errorf("unsupported content type: %s (must be 'text' or 'image')", content.Type)
	}

	return nil
}

// ValidateImageBlock validates an image block
func ValidateImageBlock(img *ImageBlock) error {
	if img == nil {
		return fmt.Errorf("image block cannot be nil")
	}

	// Normalize format
	format := strings.ToLower(img.Format)
	
	// Validate format for Converse API
	if !isValidConverseFormat(format) {
		return fmt.Errorf("invalid image format for Converse API: %s (must be 'png' or 'jpeg')", format)
	}

	// Validate source
	if err := ValidateImageSource(img.Source); err != nil {
		return fmt.Errorf("invalid image source: %w", err)
	}

	return nil
}

// ValidateImageSource validates an image source
func ValidateImageSource(source ImageSource) error {
	// Bedrock Converse API only supports bytes source type
	if source.Type != "" && source.Type != "bytes" {
		return fmt.Errorf("invalid source type: %s (must be 'bytes' or empty)", source.Type)
	}

	// Validate bytes field is present
	if source.Bytes == "" {
		return fmt.Errorf("bytes field cannot be empty for image source")
	}

	// Basic validation of base64 format (could be enhanced)
	if len(source.Bytes) > 10*1024*1024 { // 10MB limit
		return fmt.Errorf("image data too large: %d bytes (max 10MB)", len(source.Bytes))
	}

	// Optional basic validation of base64 encoding
	// Note: This is a simplified check that may need to be adjusted based on how
	// base64 data is formatted in your application
	sampleToCheck := source.Bytes
	if len(sampleToCheck) > 100 {
		sampleToCheck = sampleToCheck[:100] // Check just the beginning
	}
	
	if !base64Regex.MatchString(sampleToCheck) {
		return fmt.Errorf("bytes field does not appear to be valid base64 data")
	}

	return nil
}

// isValidConverseFormat checks if the image format is supported by Converse API
func isValidConverseFormat(format string) bool {
	format = strings.ToLower(format)
	
	// Special case for jpg (Converse API expects 'jpeg')
	if format == "jpg" {
		format = "jpeg"
	}
	
	for _, supported := range supportedConverseFormats {
		if format == supported {
			return true
		}
	}
	return false
}

// isValidInputFormat checks if the image format is supported for input
func isValidInputFormat(format string) bool {
	format = strings.ToLower(format)
	for _, supported := range supportedInputFormats {
		if format == supported {
			return true
		}
	}
	return false
}

// NormalizeImageFormat converts formats like 'jpg' to 'jpeg' for API compatibility
func NormalizeImageFormat(format string) string {
	format = strings.ToLower(format)
	if format == "jpg" {
		return "jpeg"
	}
	return format
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

	// Validate BedrockMetadata
	if turn1Response.BedrockMetadata.ModelID == "" {
		return fmt.Errorf("model ID cannot be empty in BedrockMetadata")
	}

	if turn1Response.BedrockMetadata.APIType != APITypeConverse {
		return fmt.Errorf("API type must be %s", APITypeConverse)
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

	// Validate BedrockMetadata
	if turn2Response.BedrockMetadata.ModelID == "" {
		return fmt.Errorf("model ID cannot be empty in BedrockMetadata")
	}

	if turn2Response.BedrockMetadata.APIType != APITypeConverse {
		return fmt.Errorf("API type must be %s", APITypeConverse)
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

// ValidateImageData validates image data for various operations
func ValidateImageData(format string, data string) error {
	// Normalize format
	format = NormalizeImageFormat(format)
	
	// Validate format
	if !isValidConverseFormat(format) {
		return fmt.Errorf("invalid image format: %s (supported formats: png, jpeg)", format)
	}
	
	// Validate data
	if data == "" {
		return fmt.Errorf("image data cannot be empty")
	}
	
	// Simple base64 validation
	if len(data) < 100 {
		return fmt.Errorf("image data suspiciously small: %d bytes", len(data))
	}
	
	// More thorough validation could be added here
	
	return nil
}

// ValidateModelID validates a model ID for Bedrock
func ValidateModelID(modelID string) error {
	if modelID == "" {
		return fmt.Errorf("model ID cannot be empty")
	}
	
	// Check for common model ID prefixes
	if !strings.HasPrefix(modelID, "anthropic.claude") && 
	   !strings.HasPrefix(modelID, "amazon.titan") &&
	   !strings.HasPrefix(modelID, "stability.") &&
	   !strings.HasPrefix(modelID, "ai21.") &&
	   !strings.HasPrefix(modelID, "meta.") {
		// Not a recognized prefix, but we won't fail validation
		// Just warn about it
		fmt.Printf("Warning: Model ID '%s' doesn't match common Bedrock model ID patterns\n", modelID)
	}
	
	return nil
}

// ValidateBase64Data performs basic validation of base64 encoded data
func ValidateBase64Data(data string) error {
	if data == "" {
		return fmt.Errorf("base64 data cannot be empty")
	}
	
	// Sample test on first part of the string
	sampleSize := 100
	if len(data) < sampleSize {
		sampleSize = len(data)
	}
	
	sample := data[:sampleSize]
	if !base64Regex.MatchString(sample) {
		return fmt.Errorf("data does not appear to be valid base64")
	}
	
	return nil
}

// ValidateRequestCompleteness checks if a request has all required elements
func ValidateRequestCompleteness(request *ConverseRequest) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}
	
	// Check model ID
	if err := ValidateModelID(request.ModelId); err != nil {
		return err
	}
	
	// Check messages
	if len(request.Messages) == 0 {
		return fmt.Errorf("request must contain at least one message")
	}
	
	// Check inference config
	if request.InferenceConfig.MaxTokens <= 0 {
		return fmt.Errorf("maxTokens must be a positive integer")
	}
	
	// Validate first message (user message)
	firstMsg := request.Messages[0]
	if firstMsg.Role != "user" {
		return fmt.Errorf("first message must have role 'user', got '%s'", firstMsg.Role)
	}
	
	if len(firstMsg.Content) == 0 {
		return fmt.Errorf("first message must have content")
	}
	
	return nil
}
