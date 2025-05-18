package internal

import (
   "fmt"
   "path/filepath"
   "strings"

   "workflow-function/shared/schema"
)

// CreateBedrockImageSource creates a proper Bedrock image source from ImageInfo
func CreateBedrockImageSource(imageInfo *schema.ImageInfo) (*schema.BedrockImageSource, error) {
   if imageInfo == nil {
   	return nil, fmt.Errorf("image info is required")
   }

   // Determine the image format
   format := determineImageFormat(imageInfo)
   if format == "" {
   	return nil, fmt.Errorf("unable to determine image format")
   }

   // Validate format is supported by Bedrock
   if !IsValidBedrockImageFormat(format) {
   	return nil, fmt.Errorf("unsupported image format for Bedrock: %s (supported: jpeg, png)", format)
   }

   // Create image source based on available data
   if imageInfo.Base64Data != "" {
   	// Use Base64 data (preferred for Bedrock)
   	return &schema.BedrockImageSource{
   		Type: "bytes",
   		Bytes: imageInfo.Base64Data,
   	}, nil
   } else if imageInfo.S3Bucket != "" && imageInfo.S3Key != "" {
   	// Use S3 location (fallback)
   	return &schema.BedrockImageSource{
   		Type: "bytes",
   		Bytes:  fmt.Sprintf("s3://%s/%s", imageInfo.S3Bucket, imageInfo.S3Key),
   	}, nil
   }

   return nil, fmt.Errorf("no valid image source found (need Base64 data or S3 location)")
}

// ValidateBedrockImageFormat ensures image format is supported by Bedrock
func ValidateBedrockImageFormat(format string) error {
   if !IsValidBedrockImageFormat(format) {
   	return fmt.Errorf("unsupported image format for Bedrock: %s (supported: jpeg, png)", format)
   }
   return nil
}

// IsValidBedrockImageFormat checks if the format is supported by Bedrock
func IsValidBedrockImageFormat(format string) bool {
   format = strings.ToLower(strings.TrimSpace(format))
   return format == "jpeg" || format == "jpg" || format == "png"
}

// EstimateTokenUsage estimates token usage for prompt text and image
func EstimateTokenUsage(promptText string, imageSize int64) int {
   // Base estimation for text (roughly 4 characters per token)
   textTokens := len(promptText) / 4

   // Image token estimation (simplified calculation)
   // Bedrock typically uses ~85-170 tokens per image depending on size
   var imageTokens int
   if imageSize > 0 {
   	// Rough calculation: 1 token per 20KB of image data
   	imageTokens = int(imageSize / (20 * 1024))
   	
   	// Apply bounds (typical range for Bedrock)
   	if imageTokens < 85 {
   		imageTokens = 85
   	} else if imageTokens > 170 {
   		imageTokens = 170
   	}
   }

   return textTokens + imageTokens
}

// BuildBedrockRequest creates a complete Bedrock request payload
func BuildBedrockRequest(messages []schema.BedrockMessage, config *schema.BedrockConfig) map[string]interface{} {
   request := map[string]interface{}{
   	"messages": messages,
   }

   // Add configuration if provided
   if config != nil {
   	if config.AnthropicVersion != "" {
   		request["anthropic_version"] = config.AnthropicVersion
   	}
   	
   	if config.MaxTokens > 0 {
   		request["max_tokens"] = config.MaxTokens
   	}

   	// Add thinking configuration if enabled
   	if config.Thinking != nil && config.Thinking.Type == "enabled" {
   		thinking := map[string]interface{}{
   			"type": config.Thinking.Type,
   		}
   		if config.Thinking.BudgetTokens > 0 {
   			thinking["budget_tokens"] = config.Thinking.BudgetTokens
   		}
   		request["thinking"] = thinking
   	}
   }

   return request
}

// ValidateBedrockRequest validates a Bedrock request before sending
func ValidateBedrockRequest(request map[string]interface{}) error {
   // Check required fields
   messages, ok := request["messages"]
   if !ok {
   	return fmt.Errorf("messages field is required")
   }

   // Validate messages array
   messagesList, ok := messages.([]schema.BedrockMessage)
   if !ok {
   	return fmt.Errorf("messages must be an array of BedrockMessage")
   }

   if len(messagesList) == 0 {
   	return fmt.Errorf("at least one message is required")
   }

   // Validate each message
   for i, msg := range messagesList {
   	if err := validateBedrockMessage(msg); err != nil {
   		return fmt.Errorf("invalid message at index %d: %w", i, err)
   	}
   }

   // Validate max_tokens if present
   if maxTokens, ok := request["max_tokens"]; ok {
   	if tokens, ok := maxTokens.(int); ok {
   		if tokens <= 0 || tokens > 100000 {
   			return fmt.Errorf("max_tokens must be between 1 and 100000")
   		}
   	}
   }

   return nil
}

// validateBedrockMessage validates a single Bedrock message
func validateBedrockMessage(msg schema.BedrockMessage) error {
   if msg.Role == "" {
   	return fmt.Errorf("message role is required")
   }

   if msg.Role != "user" && msg.Role != "assistant" {
   	return fmt.Errorf("message role must be 'user' or 'assistant'")
   }

   if len(msg.Content) == 0 {
   	return fmt.Errorf("message content is required")
   }

   // Validate content blocks
   for i, content := range msg.Content {
   	if err := validateContentBlock(content); err != nil {
   		return fmt.Errorf("invalid content block at index %d: %w", i, err)
   	}
   }

   return nil
}

// validateContentBlock validates a content block
func validateContentBlock(content schema.BedrockContent) error {
   if content.Type == "" {
   	return fmt.Errorf("content block type is required")
   }

   switch content.Type {
   case "text":
   	if content.Text == "" {
   		return fmt.Errorf("text content cannot be empty")
   	}
   case "image":
   	if content.Image == nil {
   		return fmt.Errorf("image content is required for image block")
   	}
   	if err := validateImageBlock(*content.Image); err != nil {
   		return fmt.Errorf("invalid image block: %w", err)
   	}
   default:
   	return fmt.Errorf("unsupported content block type: %s", content.Type)
   }

   return nil
}

// validateImageBlock validates an image content block
func validateImageBlock(image schema.BedrockImageData) error {
   if image.Format == "" {
   	return fmt.Errorf("image format is required")
   }

   if !IsValidBedrockImageFormat(image.Format) {
   	return fmt.Errorf("unsupported image format: %s", image.Format)
   }

   // BedrockImageSource is a struct, not a pointer, so we can't check if it's nil

   // Validate based on source type
   if image.Source.Type == "bytes" {
   	if image.Source.Bytes == "" {
   		return fmt.Errorf("bytes data is required for bytes source")
   	}
   } else if image.Source.Type == "s3" {
   	if image.Source.Bytes == "" {
   		return fmt.Errorf("S3 URI is required for S3 source")
   	}
   	if !strings.HasPrefix(image.Source.Bytes, "s3://") {
   		return fmt.Errorf("invalid S3 URI format: %s", image.Source.Bytes)
   	}
   } else {
   	return fmt.Errorf("unsupported image source type: %s", image.Source.Type)
   }

   return nil
}

// Helper functions

// determineImageFormat determines the image format from various sources
func determineImageFormat(imageInfo *schema.ImageInfo) string {
   // First, check if format is already set
   if imageInfo.Format != "" {
   	return normalizeImageFormat(imageInfo.Format)
   }

   // Try to determine from content type
   if imageInfo.ContentType != "" {
   	format := extractFormatFromContentType(imageInfo.ContentType)
   	if format != "" {
   		return format
   	}
   }

   // Try to determine from S3 key extension
   if imageInfo.S3Key != "" {
   	format := getImageFormatFromKey(imageInfo.S3Key)
   	if format != "" {
   		return format
   	}
   }

   // Try to determine from URL extension
   if imageInfo.URL != "" {
   	format := getImageFormatFromKey(imageInfo.URL)
   	if format != "" {
   		return format
   	}
   }

   return ""
}

// normalizeImageFormat normalizes image format string
func normalizeImageFormat(format string) string {
   format = strings.ToLower(strings.TrimSpace(format))
   
   // Normalize common variations
   switch format {
   case "jpg":
   	return "jpeg"
   case "jpeg", "png":
   	return format
   default:
   	return format
   }
}

// extractFormatFromContentType extracts image format from content type
func extractFormatFromContentType(contentType string) string {
   contentType = strings.ToLower(contentType)
   
   if strings.Contains(contentType, "jpeg") || strings.Contains(contentType, "jpg") {
   	return "jpeg"
   }
   if strings.Contains(contentType, "png") {
   	return "png"
   }
   
   return ""
}

// getImageFormatFromKey extracts the image format from an S3 key or URL
func getImageFormatFromKey(key string) string {
   ext := strings.ToLower(filepath.Ext(key))
   if ext == "" {
   	return ""
   }
   
   // Remove leading dot
   ext = ext[1:]
   
   // Check if format is supported
   supportedFormats := map[string]string{
   	"jpeg": "jpeg",
   	"jpg":  "jpeg",
   	"png":  "png",
   }
   
   if format, ok := supportedFormats[ext]; ok {
   	return format
   }
   
   return ""
}

// CreateDefaultBedrockConfig creates a default Bedrock configuration
func CreateDefaultBedrockConfig() *schema.BedrockConfig {
   return &schema.BedrockConfig{
   	AnthropicVersion: "bedrock-2023-05-31",
   	MaxTokens:        24000,
   	Thinking: &schema.Thinking{
   		Type:         "enabled",
   		BudgetTokens: 16000,
   	},
   }
}

// MergeBedrockConfigs merges two Bedrock configurations, with override taking precedence
func MergeBedrockConfigs(base, override *schema.BedrockConfig) *schema.BedrockConfig {
   if base == nil && override == nil {
   	return CreateDefaultBedrockConfig()
   }
   
   if base == nil {
   	return override
   }
   
   if override == nil {
   	return base
   }
   
   result := &schema.BedrockConfig{
   	AnthropicVersion: base.AnthropicVersion,
   	MaxTokens:        base.MaxTokens,
   }
   
   // Override values if present
   if override.AnthropicVersion != "" {
   	result.AnthropicVersion = override.AnthropicVersion
   }
   
   if override.MaxTokens > 0 {
   	result.MaxTokens = override.MaxTokens
   }
   
   // Merge thinking configuration
   if base.Thinking != nil || override.Thinking != nil {
   	result.Thinking = &schema.Thinking{}
   	
   	if base.Thinking != nil {
   		result.Thinking.Type = base.Thinking.Type
   		result.Thinking.BudgetTokens = base.Thinking.BudgetTokens
   	}
   	
   	if override.Thinking != nil {
   		if override.Thinking.Type != "" {
   			result.Thinking.Type = override.Thinking.Type
   		}
   		if override.Thinking.BudgetTokens > 0 {
   			result.Thinking.BudgetTokens = override.Thinking.BudgetTokens
   		}
   	}
   }
   
   return result
}