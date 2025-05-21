package state

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	
	"workflow-function/shared/logger"
	"workflow-function/shared/s3state"
	"workflow-function/shared/schema"
	wferrors "workflow-function/shared/errors"
	
	"workflow-function/ExecuteTurn1/internal"
)

// Category constants
const (
	CategoryProcessing = "processing"
	CategoryImages     = "images"
	CategoryPrompts    = "prompts"
	CategoryResponses  = "responses"
)

// File name constants
const (
	// Processing category
	FileInitialization   = "initialization.json"
	FileLayoutMetadata   = "layout-metadata.json"
	FileHistoricalContext = "historical-context.json"
	
	// Images category
	FileImageMetadata    = "metadata.json"
	FileReferenceBase64  = "reference-base64.base64"
	FileCheckingBase64   = "checking-base64.base64"
	
	// Prompts category
	FileSystemPrompt     = "system-prompt.json"
	
	// Responses category
	FileTurn1Response    = "turn1-raw-response.json" // Updated filename
)

// Loader handles loading state from S3 references
type Loader struct {
	stateManager s3state.Manager
	s3Client     *s3.Client
	logger       logger.Logger
	timeout      time.Duration
}

// NewLoader creates a new state loader
func NewLoader(stateManager s3state.Manager, s3Client *s3.Client, logger logger.Logger, timeout time.Duration) *Loader {
	return &Loader{
		stateManager: stateManager,
		s3Client:     s3Client,
		logger:       logger.WithFields(map[string]interface{}{"component": "StateLoader"}),
		timeout:      timeout,
	}
}

// LoadWorkflowState loads the complete workflow state from S3 references
func (l *Loader) LoadWorkflowState(ctx context.Context, refs *internal.StateReferences) (*schema.WorkflowState, error) {
	if refs == nil {
		return nil, wferrors.NewValidationError("StateReferences is nil", nil)
	}

	// Add timeout to context
	ctx, cancel := context.WithTimeout(ctx, l.timeout)
	defer cancel()
	
	l.logger.Info("Loading workflow state from S3 references", map[string]interface{}{
		"verificationId": refs.VerificationId,
	})

	// Initialize empty workflow state with schema version
	state := &schema.WorkflowState{
		SchemaVersion: schema.SchemaVersion,
	}

	// Track which references were attempted and which succeeded
	attemptedRefs := make(map[string]bool)
	successfulRefs := make(map[string]bool)
	var loadErrors []error

	// Create a default empty verification context in case loading fails
	state.VerificationContext = &schema.VerificationContext{
		VerificationId:  refs.VerificationId,
		VerificationAt:  schema.FormatISO8601(),
		Status:          schema.StatusVerificationInitialized,
		VerificationType: schema.VerificationTypeLayoutVsChecking,
	}

	// Load verification context (initialization data)
	attemptedRefs["initialization"] = true
	if refs.Initialization != nil {
		verificationContext, err := l.LoadVerificationContext(ctx, refs.Initialization)
		if err != nil {
			loadErrors = append(loadErrors, fmt.Errorf("failed to load verification context: %w", err))
		} else if verificationContext != nil {
			// Ensure critical fields are present
			if err := l.validateAndFixVerificationContext(verificationContext, refs.VerificationId); err != nil {
				loadErrors = append(loadErrors, fmt.Errorf("invalid verification context: %w", err))
				// Continue with default context
			} else {
				state.VerificationContext = verificationContext
				successfulRefs["initialization"] = true
			}
		}
	} else {
		loadErrors = append(loadErrors, fmt.Errorf("missing initialization reference"))
		// Continue with default context
	}

	// Load system prompt
	attemptedRefs["systemPrompt"] = true
	if refs.SystemPrompt != nil {
		currentPrompt, err := l.LoadCurrentPrompt(ctx, refs.SystemPrompt)
		if err != nil {
			loadErrors = append(loadErrors, fmt.Errorf("failed to load system prompt: %w", err))
		} else if currentPrompt != nil {
			state.CurrentPrompt = currentPrompt
			successfulRefs["systemPrompt"] = true
		}
	} else {
		loadErrors = append(loadErrors, fmt.Errorf("missing system prompt reference"))
	}

	// Create default bedrock config in case loading fails
	state.BedrockConfig = &schema.BedrockConfig{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        4096,
		Temperature:      0.7,
		Thinking: &schema.Thinking{
			Type:         "thinking",
			BudgetTokens: 16000,
		},
	}

	// Load Bedrock config - stored with system prompt
	attemptedRefs["bedrockConfig"] = true
	if refs.SystemPrompt != nil {
		bedrockConfig, err := l.LoadBedrockConfig(ctx, refs.SystemPrompt)
		if err != nil {
			loadErrors = append(loadErrors, fmt.Errorf("failed to load Bedrock config: %w", err))
			// Continue with default config
		} else if bedrockConfig != nil {
			state.BedrockConfig = bedrockConfig
			successfulRefs["bedrockConfig"] = true
		}
	}

	// Load images
	attemptedRefs["imageMetadata"] = true
	if refs.ImageMetadata != nil {
		images, err := l.LoadImages(ctx, refs.ImageMetadata)
		if err != nil {
			loadErrors = append(loadErrors, fmt.Errorf("failed to load images: %w", err))
		} else if images != nil {
			state.Images = images
			successfulRefs["imageMetadata"] = true
		}
	} else {
		l.logger.Warn("No image metadata reference provided", nil)
	}

	// Load use case specific data based on verification type
	if state.VerificationContext != nil {
		if state.VerificationContext.VerificationType == schema.VerificationTypeLayoutVsChecking {
			// UC1: Load layout metadata
			attemptedRefs["layoutMetadata"] = true
			if refs.LayoutMetadata != nil {
				layoutMetadata, err := l.LoadLayoutMetadata(ctx, refs.LayoutMetadata)
				if err != nil {
					l.logger.Warn("Failed to load layout metadata", map[string]interface{}{
						"error": err.Error(),
					})
				} else if layoutMetadata != nil {
					state.LayoutMetadata = layoutMetadata
					successfulRefs["layoutMetadata"] = true
				}
			}
		} else if state.VerificationContext.VerificationType == schema.VerificationTypePreviousVsCurrent {
			// UC2: Load historical context
			attemptedRefs["historicalContext"] = true
			if refs.HistoricalContext != nil {
				historicalContext, err := l.LoadHistoricalContext(ctx, refs.HistoricalContext)
				if err != nil {
					l.logger.Warn("Failed to load historical context", map[string]interface{}{
						"error": err.Error(),
					})
				} else if historicalContext != nil {
					state.HistoricalContext = historicalContext
					successfulRefs["historicalContext"] = true
				}
			}
		}
	}

	// Initialize a new conversation state
	state.ConversationState = &schema.ConversationState{
		CurrentTurn: 0,
		MaxTurns:    2,
		History:     []interface{}{},
	}

	// Load conversation state if available
	attemptedRefs["conversationState"] = true
	if refs.ConversationState != nil {
		conversationState, err := l.LoadConversationState(ctx, refs.ConversationState)
		if err != nil {
			l.logger.Warn("Failed to load conversation state, using default", map[string]interface{}{
				"error": err.Error(),
			})
			// Continue with default state
		} else if conversationState != nil {
			state.ConversationState = conversationState
			successfulRefs["conversationState"] = true
		}
	}

	// Log status of reference loading
	l.logger.Info("Reference loading status", map[string]interface{}{
		"attempted": attemptedRefs,
		"successful": successfulRefs,
		"errorCount": len(loadErrors),
		"verificationId": state.VerificationContext.VerificationId,
	})

	// Check for critical errors
	if len(loadErrors) > 0 {
		l.logger.Error("Failed to load one or more state components", map[string]interface{}{
			"errors": fmt.Sprintf("%v", loadErrors),
			"verificationId": state.VerificationContext.VerificationId,
		})
		
		// Only fail if current prompt is missing as we have defaults for other components
		if state.CurrentPrompt == nil {
			return nil, wferrors.WrapError(fmt.Errorf("Failed to load critical component: CurrentPrompt"), 
				wferrors.ErrorTypeInternal, "state loading failed", false)
		}
		
		// Otherwise, log warnings but continue with partial state
		l.logger.Warn("Some state components failed to load, continuing with partial state", map[string]interface{}{
			"verificationId": state.VerificationContext.VerificationId,
		})
	}

	return state, nil
}

// LoadVerificationContext loads the verification context from initialization.json
func (l *Loader) LoadVerificationContext(ctx context.Context, ref *s3state.Reference) (*schema.VerificationContext, error) {
	if ref == nil {
		return nil, fmt.Errorf("verification context reference is nil")
	}

	// Load as a wrapper structure with the nested "verificationContext" field
	var wrapper struct {
		VerificationContext *schema.VerificationContext `json:"verificationContext"`
		SchemaVersion       string                      `json:"schemaVersion"`
	}
	
	err := l.stateManager.RetrieveJSON(ref, &wrapper)
	if err != nil {
		return nil, err
	}

	// Check if we have a valid verification context in the wrapper
	if wrapper.VerificationContext == nil {
		return nil, fmt.Errorf("initialization.json does not contain verificationContext field")
	}

	return wrapper.VerificationContext, nil
}

// validateAndFixVerificationContext validates and fixes incomplete verification contexts
func (l *Loader) validateAndFixVerificationContext(context *schema.VerificationContext, fallbackId string) error {
	if context == nil {
		return fmt.Errorf("verification context is nil")
	}

	// Track missing fields for diagnostics
	var missingFields []string
	
	// Check and fix required fields
	if context.VerificationId == "" {
		if fallbackId != "" {
			context.VerificationId = fallbackId
			l.logger.Warn("Using fallback verification ID", map[string]interface{}{
				"fallbackId": fallbackId,
			})
		} else {
			missingFields = append(missingFields, "verificationId")
		}
	}
	
	if context.VerificationAt == "" {
		context.VerificationAt = schema.FormatISO8601()
		missingFields = append(missingFields, "verificationAt")
	}
	
	if context.Status == "" {
		context.Status = schema.StatusVerificationInitialized
		missingFields = append(missingFields, "status")
	}
	
	if context.VerificationType == "" {
		context.VerificationType = schema.VerificationTypeLayoutVsChecking
		missingFields = append(missingFields, "verificationType")
	}
	
	// Log missing fields for diagnostics
	if len(missingFields) > 0 {
		l.logger.Warn("Missing required fields in verification context, using defaults", map[string]interface{}{
			"missingFields": missingFields,
			"verificationId": context.VerificationId,
		})
	}
	
	return nil
}

// LoadCurrentPrompt loads the current prompt from the new turn1-prompt.json structure
func (l *Loader) LoadCurrentPrompt(ctx context.Context, ref *s3state.Reference) (*schema.CurrentPrompt, error) {
	if ref == nil {
		return nil, fmt.Errorf("current prompt reference is nil")
	}

	// Check if this is system-prompt.json
	isSystemPrompt := strings.Contains(ref.Key, "system-prompt.json")
	
	// For the system-prompt.json case
	if isSystemPrompt {
		// Load the new schema format
		var systemPrompt struct {
			PromptContent struct {
				SystemMessage   string `json:"systemMessage"`
				TemplateVersion string `json:"templateVersion"`
				PromptType      string `json:"promptType"`
			} `json:"promptContent"`
			VerificationId     string `json:"verificationId"`
			VerificationType   string `json:"verificationType"`
		}
		
		err := l.stateManager.RetrieveJSON(ref, &systemPrompt)
		if err != nil {
			// Try to load as a simple SystemPrompt structure
			var legacySystemPrompt struct {
				Content       string `json:"content"`
				PromptId      string `json:"promptId"`
				PromptVersion string `json:"promptVersion"`
			}
			
			err2 := l.stateManager.RetrieveJSON(ref, &legacySystemPrompt)
			if err2 != nil {
				return nil, fmt.Errorf("failed to load system prompt: %w", err)
			}
			
			// Create a CurrentPrompt from the legacy format
			currentPrompt := &schema.CurrentPrompt{
				Text:          legacySystemPrompt.Content,
				TurnNumber:    1,
				IncludeImage:  "reference",
				PromptId:      legacySystemPrompt.PromptId,
				PromptVersion: legacySystemPrompt.PromptVersion,
			}
			
			l.logger.Warn("Current prompt missing turn number, defaulting to 1", nil)
			return currentPrompt, nil
		}
		
		// Convert to the expected CurrentPrompt structure
		currentPrompt := &schema.CurrentPrompt{
			Text:          systemPrompt.PromptContent.SystemMessage,
			TurnNumber:    1,
			IncludeImage:  "reference",
			PromptId:      fmt.Sprintf("prompt-%s-system", systemPrompt.VerificationId),
			PromptVersion: systemPrompt.PromptContent.TemplateVersion,
			Metadata: map[string]interface{}{
				"promptType": systemPrompt.PromptContent.PromptType,
				"verificationType": systemPrompt.VerificationType,
			},
		}
		
		return currentPrompt, nil
	}
	
	// For turn1-prompt.json
	var newFormatPrompt struct {
		MessageStructure struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
			Role string `json:"role"`
		} `json:"messageStructure"`
		PromptType       string `json:"promptType"`
		VerificationId   string `json:"verificationId"`
		VerificationType string `json:"verificationType"`
		ImageReference   struct {
			ImageType            string `json:"imageType"`
			Base64StorageReference struct {
				Bucket string `json:"bucket"`
				Key    string `json:"key"`
			} `json:"base64StorageReference"`
		} `json:"imageReference"`
		TemplateVersion  string `json:"templateVersion"`
		CreatedAt        string `json:"createdAt"`
	}
	
	err := l.stateManager.RetrieveJSON(ref, &newFormatPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to load prompt: %w", err)
	}
	
	// Extract the text from MessageStructure
	var promptText string
	if len(newFormatPrompt.MessageStructure.Content) > 0 {
		for _, content := range newFormatPrompt.MessageStructure.Content {
			if content.Type == "text" {
				promptText = content.Text
				break
			}
		}
	}
	
	// Set a default include image type if not specified
	imageType := "reference"
	if newFormatPrompt.ImageReference.ImageType != "" {
		imageType = newFormatPrompt.ImageReference.ImageType
	}
	
	// Convert to the expected CurrentPrompt structure
	result := &schema.CurrentPrompt{
		Text:         promptText,
		TurnNumber:   1, // Default for Turn1
		IncludeImage: imageType,
		PromptId:     fmt.Sprintf("prompt-%s-turn1", newFormatPrompt.VerificationId),
		CreatedAt:    newFormatPrompt.CreatedAt,
		PromptVersion: newFormatPrompt.TemplateVersion,
		Metadata: map[string]interface{}{
			"promptType": newFormatPrompt.PromptType,
			"verificationType": newFormatPrompt.VerificationType,
		},
	}
	
	return result, nil
}

// LoadBedrockConfig loads the Bedrock config from the bedrockConfiguration field in system-prompt.json
func (l *Loader) LoadBedrockConfig(ctx context.Context, ref *s3state.Reference) (*schema.BedrockConfig, error) {
	if ref == nil {
		return nil, fmt.Errorf("Bedrock config reference is nil")
	}

	// Load from the new system-prompt.json structure
	var newSystemPrompt struct {
		BedrockConfiguration struct {
			AnthropicVersion string  `json:"anthropicVersion"`
			MaxTokens        int     `json:"maxTokens"`
			ModelId          string  `json:"modelId,omitempty"`
			Temperature      float64 `json:"temperature,omitempty"`
			Thinking         struct {
				Type         string `json:"type"`
				BudgetTokens int    `json:"budgetTokens"`
			} `json:"thinking,omitempty"`
		} `json:"bedrockConfiguration"`
	}
	
	err := l.stateManager.RetrieveJSON(ref, &newSystemPrompt)
	if err != nil {
		l.logger.Warn("No Bedrock config found in system prompt, using defaults", nil)
		// Fallback to default config
		return &schema.BedrockConfig{
			AnthropicVersion: "bedrock-2023-05-31",
			MaxTokens:        4096,
			Temperature:      0.7,
			Thinking: &schema.Thinking{
				Type:         "thinking",
				BudgetTokens: 16000,
			},
		}, nil
	}

	// Check if we have valid Bedrock configuration
	if newSystemPrompt.BedrockConfiguration.AnthropicVersion == "" {
		l.logger.Warn("No Bedrock config found in system prompt, using defaults", nil)
		return &schema.BedrockConfig{
			AnthropicVersion: "bedrock-2023-05-31",
			MaxTokens:        4096,
			Temperature:      0.7,
			Thinking: &schema.Thinking{
				Type:         "thinking",
				BudgetTokens: 16000,
			},
		}, nil
	}
	
	// Create the BedrockConfig structure
	bedrockConfig := &schema.BedrockConfig{
		AnthropicVersion: newSystemPrompt.BedrockConfiguration.AnthropicVersion,
		MaxTokens:        newSystemPrompt.BedrockConfiguration.MaxTokens,
		Temperature:      newSystemPrompt.BedrockConfiguration.Temperature,
	}
	
	// Ensure we have a valid Thinking configuration
	thinkingType := newSystemPrompt.BedrockConfiguration.Thinking.Type
	if thinkingType == "" {
		thinkingType = "thinking" // Default
	}
	
	budgetTokens := newSystemPrompt.BedrockConfiguration.Thinking.BudgetTokens
	if budgetTokens <= 0 {
		budgetTokens = 16000 // Default
	}
	
	bedrockConfig.Thinking = &schema.Thinking{
		Type:         thinkingType,
		BudgetTokens: budgetTokens,
	}
	
	return bedrockConfig, nil
}

// LoadLayoutMetadata loads the layout metadata from S3
func (l *Loader) LoadLayoutMetadata(ctx context.Context, ref *s3state.Reference) (map[string]interface{}, error) {
	if ref == nil {
		return nil, fmt.Errorf("layout metadata reference is nil")
	}

	var layoutMetadata map[string]interface{}
	err := l.stateManager.RetrieveJSON(ref, &layoutMetadata)
	if err != nil {
		return nil, err
	}

	// Check for machine structure inconsistency
	if machineStructure, ok := layoutMetadata["machineStructure"].(map[string]interface{}); ok {
		if columnsPerRow, ok := machineStructure["columnsPerRow"].(float64); ok {
			l.logger.Info("Layout metadata specifies machine structure", map[string]interface{}{
				"columnsPerRow": int(columnsPerRow),
			})
		}
	}

	return layoutMetadata, nil
}

// LoadHistoricalContext loads the historical context from S3
func (l *Loader) LoadHistoricalContext(ctx context.Context, ref *s3state.Reference) (map[string]interface{}, error) {
	if ref == nil {
		return nil, fmt.Errorf("historical context reference is nil")
	}

	var historicalContext map[string]interface{}
	err := l.stateManager.RetrieveJSON(ref, &historicalContext)
	if err != nil {
		return nil, err
	}

	return historicalContext, nil
}

// LoadImages loads the images from S3 with the new schema format
func (l *Loader) LoadImages(ctx context.Context, ref *s3state.Reference) (*schema.ImageData, error) {
	if ref == nil {
		return nil, fmt.Errorf("images reference is nil")
	}

	var images schema.ImageData
	err := l.stateManager.RetrieveJSON(ref, &images)
	if err != nil {
		return nil, err
	}

	// Ensure consistency between old and new field names
	// If we have Reference but not ReferenceImage, copy it
	if images.Reference != nil && images.ReferenceImage == nil {
		images.ReferenceImage = images.Reference
	}
	
	// If we have ReferenceImage but not Reference, copy it
	if images.ReferenceImage != nil && images.Reference == nil {
		images.Reference = images.ReferenceImage
	}
	
	// Same for Checking fields
	if images.Checking != nil && images.CheckingImage == nil {
		images.CheckingImage = images.Checking
	}
	
	if images.CheckingImage != nil && images.Checking == nil {
		images.Checking = images.CheckingImage
	}

	// Set processed timestamp if missing
	if images.ProcessedAt == "" {
		images.ProcessedAt = schema.FormatISO8601()
	}

	return &images, nil
}

// LoadConversationState loads the conversation state from S3
func (l *Loader) LoadConversationState(ctx context.Context, ref *s3state.Reference) (*schema.ConversationState, error) {
	if ref == nil {
		return nil, fmt.Errorf("conversation state reference is nil")
	}

	var conversationState schema.ConversationState
	err := l.stateManager.RetrieveJSON(ref, &conversationState)
	if err != nil {
		return nil, err
	}

	// Ensure we have valid values
	if conversationState.MaxTurns <= 0 {
		conversationState.MaxTurns = 2
	}
	
	if conversationState.History == nil {
		conversationState.History = []interface{}{}
	}

	return &conversationState, nil
}

// getVerificationIdFromKey extracts verification ID from an S3 key if possible
func (l *Loader) getVerificationIdFromKey(key string) string {
	// Try to extract from common patterns
	
	// Pattern: YYYY/MM/DD/verif-ID/...
	parts := strings.Split(key, "/")
	if len(parts) >= 4 {
		if strings.HasPrefix(parts[3], "verif-") {
			return parts[3]
		}
	}
	
	// Pattern: verif-ID/category/...
	if len(parts) >= 2 && strings.HasPrefix(parts[0], "verif-") {
		return parts[0]
	}
	
	return ""
}

// SaveTurn1Response saves the Turn1 response with the new schema format
func (l *Loader) SaveTurn1Response(ctx context.Context, state *schema.WorkflowState, turnResponse *schema.TurnResponse) *s3state.Reference {
	// Turn1 response now includes thinking content in the same file
	// This is a helper method that's not used directly in this file, but I'm including it for completeness
	
	return nil // Implementation would be in saver.go
}