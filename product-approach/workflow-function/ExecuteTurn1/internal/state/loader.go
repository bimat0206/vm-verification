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
	FileTurn1Response    = "turn1-response.json"
	FileTurn1Thinking    = "turn1-thinking.json"
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
	// This ensures we always have at least a minimal structure
	state.VerificationContext = &schema.VerificationContext{
		VerificationId:  refs.VerificationId,
		VerificationAt:  schema.FormatISO8601(),
		Status:          schema.StatusVerificationInitialized,
		VerificationType: schema.VerificationTypeLayoutVsChecking, // Default type
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

	// Load use case specific data (if available) based on verification type
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

	// Final validation of the constructed state
	if state.VerificationContext.VerificationId == "" {
		if refs.VerificationId != "" {
			state.VerificationContext.VerificationId = refs.VerificationId
		} else {
			l.logger.Warn("No verification ID found, using timestamp", nil)
			state.VerificationContext.VerificationId = fmt.Sprintf("verif-%s", time.Now().Format("20060102150405"))
		}
	}

	l.logger.Info("Workflow state loaded successfully", map[string]interface{}{
		"verificationId": state.VerificationContext.VerificationId,
		"status":         state.VerificationContext.Status,
		"componentsLoaded": successfulRefs,
	})

	return state, nil
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

// LoadVerificationContext loads the verification context (initialization data) from S3
func (l *Loader) LoadVerificationContext(ctx context.Context, ref *s3state.Reference) (*schema.VerificationContext, error) {
	if ref == nil {
		return nil, fmt.Errorf("verification context reference is nil")
	}

	var verificationContext schema.VerificationContext
	err := l.stateManager.RetrieveJSON(ref, &verificationContext)
	if err != nil {
		return nil, err
	}

	return &verificationContext, nil
}

// LoadLayoutMetadata loads the layout metadata from S3 (UC1)
func (l *Loader) LoadLayoutMetadata(ctx context.Context, ref *s3state.Reference) (map[string]interface{}, error) {
	if ref == nil {
		return nil, fmt.Errorf("layout metadata reference is nil")
	}

	var layoutMetadata map[string]interface{}
	err := l.stateManager.RetrieveJSON(ref, &layoutMetadata)
	if err != nil {
		return nil, err
	}

	return layoutMetadata, nil
}

// LoadHistoricalContext loads the historical context from S3 (UC2)
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

// LoadCurrentPrompt loads the current prompt from S3
func (l *Loader) LoadCurrentPrompt(ctx context.Context, ref *s3state.Reference) (*schema.CurrentPrompt, error) {
	if ref == nil {
		return nil, fmt.Errorf("current prompt reference is nil")
	}

	var currentPrompt schema.CurrentPrompt
	err := l.stateManager.RetrieveJSON(ref, &currentPrompt)
	if err != nil {
		return nil, err
	}

	// Validate the loaded prompt has required fields
	if currentPrompt.TurnNumber == 0 {
		currentPrompt.TurnNumber = 1
		l.logger.Warn("Current prompt missing turn number, defaulting to 1", nil)
	}
	
	if currentPrompt.IncludeImage == "" {
		currentPrompt.IncludeImage = "reference"
		l.logger.Warn("Current prompt missing includeImage, defaulting to 'reference'", nil)
	}
	
	// If we have neither text nor messages, log a warning
	if currentPrompt.Text == "" && (currentPrompt.Messages == nil || len(currentPrompt.Messages) == 0) {
		l.logger.Warn("Current prompt missing both text and messages", nil)
	}

	return &currentPrompt, nil
}

// LoadImages loads the images from S3
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

// LoadBedrockConfig loads the Bedrock config from S3
func (l *Loader) LoadBedrockConfig(ctx context.Context, ref *s3state.Reference) (*schema.BedrockConfig, error) {
	if ref == nil {
		return nil, fmt.Errorf("Bedrock config reference is nil")
	}

	// Try to load from metadata in system prompt first
	var systemPrompt schema.SystemPrompt
	err := l.stateManager.RetrieveJSON(ref, &systemPrompt)
	if err != nil {
		return nil, err
	}

	// Extract Bedrock config from system prompt metadata
	if systemPrompt.BedrockConfig != nil {
		// Validate and set defaults if needed
		if systemPrompt.BedrockConfig.AnthropicVersion == "" {
			systemPrompt.BedrockConfig.AnthropicVersion = "bedrock-2023-05-31"
		}
		
		if systemPrompt.BedrockConfig.MaxTokens <= 0 {
			systemPrompt.BedrockConfig.MaxTokens = 4096
		}
		
		if systemPrompt.BedrockConfig.Thinking == nil {
			systemPrompt.BedrockConfig.Thinking = &schema.Thinking{
				Type:         "thinking",
				BudgetTokens: 16000,
			}
		}
		
		return systemPrompt.BedrockConfig, nil
	}

	// Fallback to creating a default config
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