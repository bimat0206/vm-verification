
package state

import (
	"context"
	"fmt"
	//"strings"
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

	// Initialize empty workflow state
	state := &schema.WorkflowState{
		SchemaVersion: schema.SchemaVersion,
	}

	// Load each component of the state in parallel
	var loadErrors []error

	// Load verification context (initialization data)
	if refs.Initialization != nil {
		verificationContext, err := l.LoadVerificationContext(ctx, refs.Initialization)
		if err != nil {
			loadErrors = append(loadErrors, fmt.Errorf("failed to load verification context: %w", err))
		} else {
			state.VerificationContext = verificationContext
		}
	} else {
		loadErrors = append(loadErrors, fmt.Errorf("missing initialization reference"))
	}

	// Load system prompt
	if refs.SystemPrompt != nil {
		currentPrompt, err := l.LoadCurrentPrompt(ctx, refs.SystemPrompt)
		if err != nil {
			loadErrors = append(loadErrors, fmt.Errorf("failed to load system prompt: %w", err))
		} else {
			state.CurrentPrompt = currentPrompt
		}
	} else {
		loadErrors = append(loadErrors, fmt.Errorf("missing system prompt reference"))
	}

	// Load images
	if refs.ImageMetadata != nil {
		images, err := l.LoadImages(ctx, refs.ImageMetadata)
		if err != nil {
			loadErrors = append(loadErrors, fmt.Errorf("failed to load images: %w", err))
		} else {
			state.Images = images
		}
	}

	// Load Bedrock config - stored with system prompt
	if refs.SystemPrompt != nil {
		bedrockConfig, err := l.LoadBedrockConfig(ctx, refs.SystemPrompt)
		if err != nil {
			loadErrors = append(loadErrors, fmt.Errorf("failed to load Bedrock config: %w", err))
		} else {
			state.BedrockConfig = bedrockConfig
		}
	} else {
		loadErrors = append(loadErrors, fmt.Errorf("missing Bedrock config reference"))
	}

	// Load use case specific data (if available) based on verification type
	if state.VerificationContext != nil {
		if state.VerificationContext.VerificationType == schema.VerificationTypeLayoutVsChecking {
			// UC1: Load layout metadata
			if refs.LayoutMetadata != nil {
				layoutMetadata, err := l.LoadLayoutMetadata(ctx, refs.LayoutMetadata)
				if err != nil {
					l.logger.Warn("Failed to load layout metadata", map[string]interface{}{
						"error": err.Error(),
					})
				} else {
					state.LayoutMetadata = layoutMetadata
				}
			}
		} else if state.VerificationContext.VerificationType == schema.VerificationTypePreviousVsCurrent {
			// UC2: Load historical context
			if refs.HistoricalContext != nil {
				historicalContext, err := l.LoadHistoricalContext(ctx, refs.HistoricalContext)
				if err != nil {
					l.logger.Warn("Failed to load historical context", map[string]interface{}{
						"error": err.Error(),
					})
				} else {
					state.HistoricalContext = historicalContext
				}
			}
		}
	}

	// Load conversation state if available
	if refs.ConversationState != nil {
		conversationState, err := l.LoadConversationState(ctx, refs.ConversationState)
		if err != nil {
			l.logger.Warn("Failed to load conversation state, creating new one", map[string]interface{}{
				"error": err.Error(),
			})
			// Initialize a new conversation state instead of failing
			state.ConversationState = &schema.ConversationState{
				CurrentTurn: 0,
				MaxTurns:    2,
				History:     []interface{}{},
			}
		} else {
			state.ConversationState = conversationState
		}
	} else {
		// Initialize a new conversation state
		state.ConversationState = &schema.ConversationState{
			CurrentTurn: 0,
			MaxTurns:    2,
			History:     []interface{}{},
		}
	}

	// Check for critical errors
	if len(loadErrors) > 0 {
		l.logger.Error("Failed to load one or more state components", map[string]interface{}{
			"errors": fmt.Sprintf("%v", loadErrors),
		})
		
		// Only fail if critical components are missing
		if state.VerificationContext == nil || state.CurrentPrompt == nil || state.BedrockConfig == nil {
			return nil, wferrors.WrapError(fmt.Errorf("Failed to load critical state components"), 
				"state", "state loading failed", false)
		}
		
		// Otherwise, log warnings but continue
		l.logger.Warn("Some non-critical state components failed to load, continuing with partial state", nil)
	}

	l.logger.Info("Workflow state loaded successfully", map[string]interface{}{
		"verificationId": state.VerificationContext.VerificationId,
		"status":         state.VerificationContext.Status,
	})

	return state, nil
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
		return systemPrompt.BedrockConfig, nil
	}

	// Fallback to creating a default config
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

	return &conversationState, nil
}
