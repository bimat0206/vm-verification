package state

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"

	wferrors "workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/s3state"
	"workflow-function/shared/schema"

	"workflow-function/ExecuteTurn1/internal"
)

// Saver handles saving state to S3
type Saver struct {
	stateManager s3state.Manager
	s3Client     *s3.Client
	logger       logger.Logger
	timeout      time.Duration
}

// NewSaver creates a new state saver
func NewSaver(stateManager s3state.Manager, s3Client *s3.Client, logger logger.Logger, timeout time.Duration) *Saver {
	return &Saver{
		stateManager: stateManager,
		s3Client:     s3Client,
		logger:       logger.WithFields(map[string]interface{}{"component": "StateSaver"}),
		timeout:      timeout,
	}
}

// generateDatePath creates a S3 path with date partitioning
func generateDatePath(verificationId string) string {
	now := time.Now().UTC()
	return fmt.Sprintf("%04d/%02d/%02d/%s", 
		now.Year(), now.Month(), now.Day(), verificationId)
}

// SaveWorkflowState saves the workflow state to S3 and returns references
func (s *Saver) SaveWorkflowState(ctx context.Context, state *schema.WorkflowState) (*internal.StateReferences, error) {
	if state == nil {
		return nil, wferrors.NewValidationError("WorkflowState is nil", nil)
	}

	// Add timeout to context
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	
	verificationId := state.VerificationContext.VerificationId
	
	s.logger.Info("Saving workflow state to S3", map[string]interface{}{
		"verificationId": verificationId,
		"status":         state.VerificationContext.Status,
	})

	// Generate date-based path
	datePath := generateDatePath(verificationId)

	// Create envelope for references
	envelope := s3state.NewEnvelope(verificationId)
	envelope.SetStatus(state.VerificationContext.Status)

	// Add verification context to summary
	envelope.AddSummary("status", state.VerificationContext.Status)
	envelope.AddSummary("verificationAt", state.VerificationContext.VerificationAt)
	envelope.AddSummary("datePartition", datePath)
	
	// Add turn information to summary if available
	if state.ConversationState != nil {
		envelope.AddSummary("currentTurn", state.ConversationState.CurrentTurn)
		envelope.AddSummary("maxTurns", state.ConversationState.MaxTurns)
	}

	// Save each component of the state
	var saveErrors []error

	// Save verification context as initialization.json
	err := s.stateManager.SaveToEnvelope(envelope, CategoryProcessing, FileInitialization, state.VerificationContext)
	if err != nil {
		saveErrors = append(saveErrors, fmt.Errorf("failed to save initialization data: %w", err))
	}

	// Save use case specific data
	if state.VerificationContext.VerificationType == schema.VerificationTypeLayoutVsChecking {
		// UC1: Save layout metadata
		if state.LayoutMetadata != nil {
			err := s.stateManager.SaveToEnvelope(envelope, CategoryProcessing, FileLayoutMetadata, state.LayoutMetadata)
			if err != nil {
				saveErrors = append(saveErrors, fmt.Errorf("failed to save layout metadata: %w", err))
			}
		}
	} else if state.VerificationContext.VerificationType == schema.VerificationTypePreviousVsCurrent {
		// UC2: Save historical context
		if state.HistoricalContext != nil {
			err := s.stateManager.SaveToEnvelope(envelope, CategoryProcessing, FileHistoricalContext, state.HistoricalContext)
			if err != nil {
				saveErrors = append(saveErrors, fmt.Errorf("failed to save historical context: %w", err))
			}
		}
	}

	// Save current prompt as system-prompt.json
	if state.CurrentPrompt != nil {
		err := s.stateManager.SaveToEnvelope(envelope, CategoryPrompts, FileSystemPrompt, state.CurrentPrompt)
		if err != nil {
			saveErrors = append(saveErrors, fmt.Errorf("failed to save system prompt: %w", err))
		}
	}

	// Save images metadata
	if state.Images != nil {
		err := s.stateManager.SaveToEnvelope(envelope, CategoryImages, FileImageMetadata, state.Images)
		if err != nil {
			saveErrors = append(saveErrors, fmt.Errorf("failed to save images metadata: %w", err))
		}
		
		// Note: Base64 image data should be saved separately with .base64 extension
		// by the FetchImages function. Here we just use the metadata.
	}

	// Save conversation state
	if state.ConversationState != nil {
		// Store in processing/ category since it's part of the workflow state
		err := s.stateManager.SaveToEnvelope(envelope, CategoryProcessing, "conversation-state.json", state.ConversationState)
		if err != nil {
			saveErrors = append(saveErrors, fmt.Errorf("failed to save conversation state: %w", err))
		}
	}

	// Save Turn1 response
	if state.Turn1Response != nil {
		err := s.stateManager.SaveToEnvelope(envelope, CategoryResponses, FileTurn1Response, state.Turn1Response)
		if err != nil {
			saveErrors = append(saveErrors, fmt.Errorf("failed to save Turn1 response: %w", err))
		}
	}

	// Check for errors
	if len(saveErrors) > 0 {
		s.logger.Error("Failed to save one or more state components", map[string]interface{}{
			"errors": fmt.Sprintf("%v", saveErrors),
		})
		return nil, wferrors.WrapError(fmt.Errorf("Failed to save state components: %v", saveErrors), 
			"state", "state saving failed", false)
	}

	// Convert envelope to state references
	stateRefs := &internal.StateReferences{
		VerificationId:     verificationId,
		DatePartition:      datePath,
		Initialization:     envelope.GetReference(fmt.Sprintf("%s_%s", CategoryProcessing, strings.TrimSuffix(FileInitialization, ".json"))),
		SystemPrompt:       envelope.GetReference(fmt.Sprintf("%s_%s", CategoryPrompts, strings.TrimSuffix(FileSystemPrompt, ".json"))),
		ImageMetadata:      envelope.GetReference(fmt.Sprintf("%s_%s", CategoryImages, strings.TrimSuffix(FileImageMetadata, ".json"))),
		ConversationState: envelope.GetReference(fmt.Sprintf("%s_%s", CategoryProcessing, "conversation-state")),
		Turn1Response:     envelope.GetReference(fmt.Sprintf("%s_%s", CategoryResponses, strings.TrimSuffix(FileTurn1Response, ".json"))),
	}
	
	// Add use case specific references
	if state.VerificationContext.VerificationType == schema.VerificationTypeLayoutVsChecking {
		stateRefs.LayoutMetadata = envelope.GetReference(fmt.Sprintf("%s_%s", CategoryProcessing, strings.TrimSuffix(FileLayoutMetadata, ".json")))
	} else if state.VerificationContext.VerificationType == schema.VerificationTypePreviousVsCurrent {
		stateRefs.HistoricalContext = envelope.GetReference(fmt.Sprintf("%s_%s", CategoryProcessing, strings.TrimSuffix(FileHistoricalContext, ".json")))
	}

	s.logger.Info("Workflow state saved successfully", map[string]interface{}{
		"verificationId": verificationId,
		"status":         state.VerificationContext.Status,
		"datePartition":  datePath,
		"references":     len(envelope.References),
	})

	return stateRefs, nil
}

// SaveThinkingContent saves the thinking content separately and returns a reference
func (s *Saver) SaveThinkingContent(ctx context.Context, verificationId string, thinkingContent string) (*s3state.Reference, error) {
	if thinkingContent == "" {
		return nil, wferrors.NewValidationError("ThinkingContent is empty", nil)
	}

	// Add timeout to context
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	
	s.logger.Info("Saving thinking content to S3", map[string]interface{}{
		"verificationId": verificationId,
	})

	// Generate date-based path
	datePath := generateDatePath(verificationId)
	
	// Store the thinking content in the responses category
	ref, err := s.stateManager.StoreJSON(CategoryResponses, 
		fmt.Sprintf("%s/turn1-thinking.json", datePath), map[string]interface{}{
			"verificationId": verificationId,
			"thinking": thinkingContent,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	if err != nil {
		s.logger.Error("Failed to save thinking content", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, wferrors.WrapError(err, "state", "failed to save thinking content", false)
	}

	s.logger.Info("Thinking content saved successfully", map[string]interface{}{
		"reference": ref.String(),
	})

	return ref, nil
}

// SaveTurn1Response saves the Turn1 response with thinking content included
func (s *Saver) SaveTurn1Response(ctx context.Context, verificationId string, turnResponse *schema.TurnResponse) (*s3state.Reference, error) {
    if turnResponse == nil {
        return nil, wferrors.NewValidationError("TurnResponse is nil", nil)
    }

    // Add timeout to context
    ctx, cancel := context.WithTimeout(ctx, s.timeout)
    defer cancel()
    
    s.logger.Info("Saving Turn1 response to S3", map[string]interface{}{
        "verificationId": verificationId,
        "turnId":         turnResponse.TurnId,
    })

    // Generate date-based path
    datePath := generateDatePath(verificationId)
    
    // Create the Turn1 response object in the new format with thinking content included
    turn1ResponseWithThinking := map[string]interface{}{
        "verificationId":   verificationId,
        "turnId":           turnResponse.TurnId,
        "analysisStage":    "REFERENCE_ANALYSIS",
        "verificationType": "LAYOUT_VS_CHECKING", // This should be from the actual verification context
        "response": map[string]interface{}{
            "content": []map[string]interface{}{
                {
                    "type": "text",
                    "text": turnResponse.Response.Content,
                },
            },
        },
        "tokenUsage": map[string]interface{}{
            "input":    turnResponse.TokenUsage.InputTokens,
            "output":   turnResponse.TokenUsage.OutputTokens,
            "thinking": turnResponse.TokenUsage.ThinkingTokens,
            "total":    turnResponse.TokenUsage.TotalTokens,
        },
        "latencyMs":  turnResponse.LatencyMs,
        "timestamp":  turnResponse.Timestamp,
        "status":     "SUCCESS",
    }
    
    // Add thinking content if available
    if turnResponse.Metadata != nil && turnResponse.Metadata["thinking"] != nil {
        if thinkingContent, ok := turnResponse.Metadata["thinking"].(string); ok && thinkingContent != "" {
            // Add thinking to the response object
            thinking := turnResponse.Metadata["thinking"].(string)
            responseMap := turn1ResponseWithThinking["response"].(map[string]interface{})
            responseMap["thinking"] = thinking
        }
    }
    
    // Add bedrock metadata if available
    if turnResponse.Metadata != nil {
        turn1ResponseWithThinking["bedrockMetadata"] = map[string]interface{}{
            "modelId":    turnResponse.Metadata["modelId"],
            "requestId":  "req-" + verificationId,
            "stopReason": turnResponse.Response.StopReason,
        }
    }
    
    // Store the turn response in the responses category
    ref, err := s.stateManager.StoreJSON(CategoryResponses, 
        fmt.Sprintf("%s/%s", datePath, FileTurn1Response), turn1ResponseWithThinking)
    if err != nil {
        s.logger.Error("Failed to save Turn1 response", map[string]interface{}{
            "error": err.Error(),
        })
        return nil, wferrors.WrapError(err, "state", "failed to save Turn1 response", false)
    }

    s.logger.Info("Turn1 response saved successfully", map[string]interface{}{
        "reference": ref.String(),
    })

    return ref, nil
}