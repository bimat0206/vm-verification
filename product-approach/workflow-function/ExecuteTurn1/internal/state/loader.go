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
	FileTurn1Prompt      = "turn1-prompt.json"
	
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

	// Create default verification context in case loading fails
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
			state.VerificationContext = verificationContext
			successfulRefs["initialization"] = true
		}
	} else {
		loadErrors = append(loadErrors, fmt.Errorf("missing initialization reference"))
	}

	// Load system prompt with Bedrock configuration
	attemptedRefs["systemPrompt"] = true
	if refs.SystemPrompt != nil {
		currentPrompt, bedrockConfig, err := l.LoadSystemPromptWithConfig(ctx, refs.SystemPrompt)
		if err != nil {
			loadErrors = append(loadErrors, fmt.Errorf("failed to load system prompt: %w", err))
		} else {
			if currentPrompt != nil {
				state.CurrentPrompt = currentPrompt
				successfulRefs["systemPrompt"] = true
			}
			
			if bedrockConfig != nil {
				state.BedrockConfig = bedrockConfig
				successfulRefs["bedrockConfig"] = true
			}
		}
	} else {
		loadErrors = append(loadErrors, fmt.Errorf("missing system prompt reference"))
	}
	
	// Load Turn1 prompt
	attemptedRefs["turn1Prompt"] = true
	if refs.Turn1Prompt != nil {
		turn1Prompt, imageBase64Refs, err := l.LoadTurn1Prompt(ctx, refs.Turn1Prompt)
		if err != nil {
			loadErrors = append(loadErrors, fmt.Errorf("failed to load turn1 prompt: %w", err))
		} else if turn1Prompt != nil {
			// If we got a turn1 prompt, update the current prompt
			state.CurrentPrompt = turn1Prompt
			successfulRefs["turn1Prompt"] = true
			
			// Store the Base64 refs for later
			if imageBase64Refs != nil {
				if imageBase64Refs.Reference != nil {
					refs.ReferenceBase64 = imageBase64Refs.Reference
				}
				if imageBase64Refs.Checking != nil {
					refs.CheckingBase64 = imageBase64Refs.Checking
				}
			}
		}
	}

	// Load images with Base64 references
	attemptedRefs["imageMetadata"] = true
	if refs.ImageMetadata != nil {
		images, err := l.LoadImages(ctx, refs.ImageMetadata, refs)
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

	// Initialize conversation state
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
		
		// Only fail if current prompt is missing
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
// with support for nested verificationContext field
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

	// Validate required fields and set defaults if missing
	if wrapper.VerificationContext.VerificationId == "" {
		l.logger.Warn("VerificationContext missing VerificationId", nil)
		if ref.Key != "" {
			// Try to extract from key
			parts := strings.Split(ref.Key, "/")
			for _, part := range parts {
				if strings.HasPrefix(part, "verif-") {
					wrapper.VerificationContext.VerificationId = part
					l.logger.Info("Extracted VerificationId from key", map[string]interface{}{
						"verificationId": part,
					})
					break
				}
			}
		}
	}
	
	if wrapper.VerificationContext.VerificationAt == "" {
		wrapper.VerificationContext.VerificationAt = schema.FormatISO8601()
	}
	
	if wrapper.VerificationContext.Status == "" {
		wrapper.VerificationContext.Status = schema.StatusVerificationInitialized
	}
	
	return wrapper.VerificationContext, nil
}

// LoadSystemPromptWithConfig loads both system prompt and Bedrock configuration from system-prompt.json
func (l *Loader) LoadSystemPromptWithConfig(ctx context.Context, ref *s3state.Reference) (*schema.CurrentPrompt, *schema.BedrockConfig, error) {
	if ref == nil {
		return nil, nil, fmt.Errorf("system prompt reference is nil")
	}

	// Load the new schema format with separate bedrockConfiguration field
	var systemPromptWrapper struct {
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
		PromptContent struct {
			SystemMessage   string `json:"systemMessage"`
			TemplateVersion string `json:"templateVersion"`
			PromptType      string `json:"promptType"`
		} `json:"promptContent"`
		VerificationId     string `json:"verificationId"`
		VerificationType   string `json:"verificationType"`
		Version            string `json:"version"`
	}
	
	err := l.stateManager.RetrieveJSON(ref, &systemPromptWrapper)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load system prompt: %w", err)
	}
	
	// Create the BedrockConfig
	bedrockConfig := &schema.BedrockConfig{
		AnthropicVersion: systemPromptWrapper.BedrockConfiguration.AnthropicVersion,
		MaxTokens:        systemPromptWrapper.BedrockConfiguration.MaxTokens,
		Temperature:      systemPromptWrapper.BedrockConfiguration.Temperature,
		Thinking: &schema.Thinking{
			Type:         systemPromptWrapper.BedrockConfiguration.Thinking.Type,
			BudgetTokens: systemPromptWrapper.BedrockConfiguration.Thinking.BudgetTokens,
		},
	}
	
	// Set defaults if needed
	if bedrockConfig.AnthropicVersion == "" {
		bedrockConfig.AnthropicVersion = "bedrock-2023-05-31"
	}
	
	if bedrockConfig.MaxTokens <= 0 {
		bedrockConfig.MaxTokens = 4096
	}
	
	if bedrockConfig.Temperature == 0 {
		bedrockConfig.Temperature = 0.7
	}
	
	if bedrockConfig.Thinking.Type == "" {
		bedrockConfig.Thinking.Type = "thinking"
	}
	
	if bedrockConfig.Thinking.BudgetTokens <= 0 {
		bedrockConfig.Thinking.BudgetTokens = 16000
	}
	
	// Create the CurrentPrompt
	currentPrompt := &schema.CurrentPrompt{
		Text:          systemPromptWrapper.PromptContent.SystemMessage,
		TurnNumber:    1, // Default for system prompt
		IncludeImage:  "reference", // Default for system prompt
		PromptId:      fmt.Sprintf("prompt-%s-system", systemPromptWrapper.VerificationId),
		PromptVersion: systemPromptWrapper.PromptContent.TemplateVersion,
		Metadata: map[string]interface{}{
			"promptType":       systemPromptWrapper.PromptContent.PromptType,
			"verificationType": systemPromptWrapper.VerificationType,
			"version":          systemPromptWrapper.Version,
		},
	}

	return currentPrompt, bedrockConfig, nil
}

// ImageBase64References holds references to Base64 encoded image data
type ImageBase64References struct {
	Reference *s3state.Reference
	Checking  *s3state.Reference
}

// LoadTurn1Prompt loads the turn1 prompt and extracts image Base64 references
func (l *Loader) LoadTurn1Prompt(ctx context.Context, ref *s3state.Reference) (*schema.CurrentPrompt, *ImageBase64References, error) {
	if ref == nil {
		return nil, nil, fmt.Errorf("turn1 prompt reference is nil")
	}

	// Load the new schema format with messageStructure
	var promptWrapper struct {
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
	
	err := l.stateManager.RetrieveJSON(ref, &promptWrapper)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load turn1 prompt: %w", err)
	}
	
	// Extract the text from MessageStructure
	var promptText string
	if len(promptWrapper.MessageStructure.Content) > 0 {
		for _, content := range promptWrapper.MessageStructure.Content {
			if content.Type == "text" {
				promptText = content.Text
				break
			}
		}
	}
	
	// Create the CurrentPrompt
	currentPrompt := &schema.CurrentPrompt{
		Text:          promptText,
		TurnNumber:    1,
		IncludeImage:  promptWrapper.ImageReference.ImageType,
		PromptId:      fmt.Sprintf("prompt-%s-turn1", promptWrapper.VerificationId),
		CreatedAt:     promptWrapper.CreatedAt,
		PromptVersion: promptWrapper.TemplateVersion,
		Metadata: map[string]interface{}{
			"promptType":       promptWrapper.PromptType,
			"verificationType": promptWrapper.VerificationType,
		},
	}
	
	// Also create Bedrock-style messages for the API call
	var bedrockMessages []schema.BedrockMessage
	if promptWrapper.MessageStructure.Role != "" && promptText != "" {
		message := schema.BedrockMessage{
			Role: promptWrapper.MessageStructure.Role,
			Content: []schema.BedrockContent{
				{
					Type: "text",
					Text: promptText,
				},
			},
		}
		bedrockMessages = append(bedrockMessages, message)
		currentPrompt.Messages = bedrockMessages
	}
	
	// Extract Base64 references
	var imageBase64Refs *ImageBase64References
	if promptWrapper.ImageReference.Base64StorageReference.Bucket != "" && 
	   promptWrapper.ImageReference.Base64StorageReference.Key != "" {
		
		imageBase64Refs = &ImageBase64References{}
		base64Ref := &s3state.Reference{
			Bucket: promptWrapper.ImageReference.Base64StorageReference.Bucket,
			Key:    promptWrapper.ImageReference.Base64StorageReference.Key,
		}
		
		// Set the appropriate reference based on image type
		if promptWrapper.ImageReference.ImageType == "reference" {
			imageBase64Refs.Reference = base64Ref
		} else if promptWrapper.ImageReference.ImageType == "checking" {
			imageBase64Refs.Checking = base64Ref
		}
	}
	
	return currentPrompt, imageBase64Refs, nil
}

// LoadImages loads image data using metadata and Base64 references from turn1-prompt
// LoadImages loads image data using metadata and Base64 references from turn1-prompt
func (l *Loader) LoadImages(ctx context.Context, ref *s3state.Reference, refs *internal.StateReferences) (*schema.ImageData, error) {
    if ref == nil {
        return nil, fmt.Errorf("image metadata reference is nil")
    }

    // Try to load the images metadata
    var images schema.ImageData
    err := l.stateManager.RetrieveJSON(ref, &images)
    if err != nil {
        return nil, err
    }


// Use the Base64 references if available as before...
    // Use the Base64 references from the prompt if available
    if refs.ReferenceBase64 != nil {
        if images.Reference == nil {
            // Create a new reference image if it doesn't exist
            images.Reference = &schema.ImageInfo{
                Format:          "png", // Default format
                Base64Generated: true,
                Base64S3Bucket:  refs.ReferenceBase64.Bucket,
                Base64S3Key:     refs.ReferenceBase64.Key,
                StorageMethod:   "s3-temporary",
            }
        } else {
            // Update the existing reference image
            images.Reference.Base64S3Bucket = refs.ReferenceBase64.Bucket
            images.Reference.Base64S3Key = refs.ReferenceBase64.Key
            images.Reference.Base64Generated = true
            images.Reference.StorageMethod = "s3-temporary"
            
            // Set format if it's empty
            if images.Reference.Format == "" {
                images.Reference.Format = "png" // Default format
            }
        }
        
        // Also set in legacy field
        images.ReferenceImage = images.Reference
        
        // Set the overall flag
        images.Base64Generated = true
    }
    
    if refs.CheckingBase64 != nil {
        if images.Checking == nil {
            // Create a new checking image if it doesn't exist
            images.Checking = &schema.ImageInfo{
                Format:          "png", // Default format
                Base64Generated: true,
                Base64S3Bucket:  refs.CheckingBase64.Bucket,
                Base64S3Key:     refs.CheckingBase64.Key,
                StorageMethod:   "s3-temporary",
            }
        } else {
            // Update the existing checking image
            images.Checking.Base64S3Bucket = refs.CheckingBase64.Bucket
            images.Checking.Base64S3Key = refs.CheckingBase64.Key
            images.Checking.Base64Generated = true
            images.Checking.StorageMethod = "s3-temporary"
            
            // Set format if it's empty
            if images.Checking.Format == "" {
                images.Checking.Format = "png" // Default format
            }
        }
        
        // Also set in legacy field
        images.CheckingImage = images.Checking
        
        // Set the overall flag
        images.Base64Generated = true
    }
    
    // Make sure we record when the image data was processed
    if images.ProcessedAt == "" {
        images.ProcessedAt = schema.FormatISO8601()
    }

    // Log the state of the images
    l.logger.Info("Loaded images", map[string]interface{}{
        "hasReference": images.Reference != nil || images.ReferenceImage != nil,
        "hasChecking": images.Checking != nil || images.CheckingImage != nil,
        "base64Generated": images.Base64Generated,
    })
    
    // If we have reference images, check the Base64 data is accessible
    if (images.Reference != nil || images.ReferenceImage != nil) && !images.Base64Generated {
        l.logger.Warn("Reference image found but Base64 not generated", map[string]interface{}{
            "referenceBase64": refs.ReferenceBase64 != nil,
            "checkingBase64": refs.CheckingBase64 != nil,
        })
    }

    return &images, nil
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