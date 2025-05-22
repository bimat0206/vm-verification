// internal/services/schema_integration.go
package services

import (
	"context"
	"fmt"
	"workflow-function/ExecuteTurn1Combined/internal/models"
	"workflow-function/shared/schema"
)

// SchemaIntegratedService demonstrates how to integrate schema validation into services
type SchemaIntegratedService struct {
	bedrockService BedrockService
	s3Service      S3StateManager
}

// NewSchemaIntegratedService creates a new service with schema integration
func NewSchemaIntegratedService(bedrock BedrockService, s3 S3StateManager) *SchemaIntegratedService {
	return &SchemaIntegratedService{
		bedrockService: bedrock,
		s3Service:      s3,
	}
}

// ValidateAndProcessBedrockRequest demonstrates schema validation for Bedrock requests
func (s *SchemaIntegratedService) ValidateAndProcessBedrockRequest(
	ctx context.Context,
	prompt string,
	imageData *schema.ImageData,
) (*schema.TurnResponse, error) {
	
	// Validate image data if provided
	if imageData != nil {
		if errors := schema.ValidateImageData(imageData, true); len(errors) > 0 {
			return nil, fmt.Errorf("image data validation failed: %s", errors.Error())
		}
	}

	// Create Bedrock configuration using schema types
	bedrockConfig := &schema.BedrockConfig{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        4000,
		Temperature:      0.1,
		Thinking: &schema.Thinking{
			Type:         "thinking",
			BudgetTokens: 20000,
		},
	}

	// Validate Bedrock configuration
	if errors := schema.ValidateBedrockConfig(bedrockConfig); len(errors) > 0 {
		return nil, fmt.Errorf("bedrock config validation failed: %s", errors.Error())
	}

	// Build messages using schema helpers
	messages := schema.BuildBedrockMessages(prompt, "reference", imageData)

	// Validate messages
	if errors := schema.ValidateBedrockMessages(messages); len(errors) > 0 {
		return nil, fmt.Errorf("bedrock messages validation failed: %s", errors.Error())
	}

	// Create current prompt structure
	currentPrompt := &schema.CurrentPrompt{
		Text:         prompt,
		TurnNumber:   1,
		IncludeImage: "reference",
		Messages:     messages,
		PromptId:     "turn1-prompt",
		CreatedAt:    schema.FormatISO8601(),
	}

	// Validate current prompt
	if errors := schema.ValidateCurrentPrompt(currentPrompt, true); len(errors) > 0 {
		return nil, fmt.Errorf("current prompt validation failed: %s", errors.Error())
	}

	// Simulate Bedrock response creation with schema types
	turnResponse := &schema.TurnResponse{
		TurnId:    1,
		Timestamp: schema.FormatISO8601(),
		Prompt:    prompt,
		Response: schema.BedrockApiResponse{
			Content:    "Simulated response content",
			Thinking:   "Simulated thinking process",
			StopReason: "end_turn",
			ModelId:    "anthropic.claude-3-sonnet-20240229-v1:0",
			RequestId:  "simulated-request-id",
		},
		LatencyMs: 1500,
		TokenUsage: &schema.TokenUsage{
			InputTokens:    250,
			OutputTokens:   150,
			ThinkingTokens: 75,
			TotalTokens:    475,
		},
		Stage: "reference_analysis",
	}

	return turnResponse, nil
}

// CreateWorkflowState demonstrates creating a standardized workflow state
func (s *SchemaIntegratedService) CreateWorkflowState(
	verificationID string,
	verificationType string,
	imageData *schema.ImageData,
) (*schema.WorkflowState, error) {
	
	// Create verification context using schema types
	verificationContext := &schema.VerificationContext{
		VerificationId:   verificationID,
		VerificationAt:   schema.FormatISO8601(),
		Status:           schema.StatusTurn1PromptReady,
		VerificationType: verificationType,
		ReferenceImageUrl: "s3://bucket/reference-image.jpg",
		CheckingImageUrl:  "s3://bucket/checking-image.jpg",
		VendingMachineId:  "VM001",
		RequestMetadata: &schema.RequestMetadata{
			RequestId:         verificationID,
			RequestTimestamp:  schema.FormatISO8601(),
			ProcessingStarted: schema.FormatISO8601(),
		},
		TurnTimestamps: &schema.TurnTimestamps{
			Initialized:  schema.FormatISO8601(),
			Turn1Started: schema.FormatISO8601(),
		},
		NotificationEnabled: true,
	}

	// For layout verification, add layout-specific fields
	if verificationType == schema.VerificationTypeLayoutVsChecking {
		verificationContext.LayoutId = 12345
		verificationContext.LayoutPrefix = "layout-v1"
	}

	// For previous vs current, add historical context
	if verificationType == schema.VerificationTypePreviousVsCurrent {
		verificationContext.PreviousVerificationId = "prev-verification-123"
	}

	// Validate verification context
	if errors := schema.ValidateVerificationContext(verificationContext); len(errors) > 0 {
		return nil, fmt.Errorf("verification context validation failed: %s", errors.Error())
	}

	// Create system prompt
	systemPrompt := &schema.SystemPrompt{
		Content:       "You are a vending machine verification assistant...",
		PromptId:      "system-prompt-v1",
		PromptVersion: "1.0.0",
	}

	// Create Bedrock configuration
	bedrockConfig := &schema.BedrockConfig{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        4000,
		Temperature:      0.1,
	}

	// Create workflow state
	workflowState := &schema.WorkflowState{
		SchemaVersion:       schema.SchemaVersion,
		VerificationContext: verificationContext,
		Images:              imageData,
		SystemPrompt:        systemPrompt,
		BedrockConfig:       bedrockConfig,
	}

	// Validate complete workflow state
	if errors := schema.ValidateWorkflowState(workflowState); len(errors) > 0 {
		return nil, fmt.Errorf("workflow state validation failed: %s", errors.Error())
	}

	return workflowState, nil
}

// ConvertLegacyToSchema demonstrates converting legacy types to schema types
func (s *SchemaIntegratedService) ConvertLegacyToSchema(legacyRequest *models.Turn1Request) (*schema.VerificationContext, error) {
	// Convert legacy verification context to schema format
	schemaContext := &schema.VerificationContext{
		VerificationId:    legacyRequest.VerificationID,
		VerificationAt:    schema.FormatISO8601(),
		Status:            schema.StatusTurn1PromptReady,
		VerificationType:  legacyRequest.VerificationContext.VerificationType,
		VendingMachineId:  legacyRequest.VerificationContext.VendingMachineId,
		ReferenceImageUrl: legacyRequest.S3Refs.Images.ReferenceBase64.Key,
		CheckingImageUrl:  "checking-image-url", // Would be extracted from context
		RequestMetadata: &schema.RequestMetadata{
			RequestId:         legacyRequest.VerificationID,
			RequestTimestamp:  schema.FormatISO8601(),
			ProcessingStarted: schema.FormatISO8601(),
		},
		NotificationEnabled: true,
	}

	// Note: LayoutMetadata and HistoricalContext are stored at WorkflowState level in schema
	// These will be handled when creating the full WorkflowState

	// Validate converted context
	if errors := schema.ValidateVerificationContext(schemaContext); len(errors) > 0 {
		return nil, fmt.Errorf("converted verification context validation failed: %s", errors.Error())
	}

	return schemaContext, nil
}

// CreateErrorInfoWithSchema demonstrates creating standardized error information
func (s *SchemaIntegratedService) CreateErrorInfoWithSchema(code, message string, details map[string]interface{}) *schema.ErrorInfo {
	return &schema.ErrorInfo{
		Code:      code,
		Message:   message,
		Details:   details,
		Timestamp: schema.FormatISO8601(),
	}
}

// ValidateImageForBedrock demonstrates image validation for Bedrock API
func (s *SchemaIntegratedService) ValidateImageForBedrock(imageInfo *schema.ImageInfo) error {
	// Validate image info with Base64 requirements for Bedrock
	if errors := schema.ValidateImageInfo(imageInfo, true); len(errors) > 0 {
		return fmt.Errorf("image validation for Bedrock failed: %s", errors.Error())
	}

	// Additional Bedrock-specific validations
	if imageInfo.GetBase64SizeEstimate() > schema.BedrockMaxImageSize {
		return fmt.Errorf("image size %d bytes exceeds Bedrock limit of %d bytes", 
			imageInfo.GetBase64SizeEstimate(), schema.BedrockMaxImageSize)
	}

	return nil
}