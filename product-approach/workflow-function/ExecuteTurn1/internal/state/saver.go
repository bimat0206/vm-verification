package state

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	
	"workflow-function/shared/logger"
	"workflow-function/shared/s3state"
	"workflow-function/shared/schema"
	wferrors "workflow-function/shared/errors"
	
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

	// Create envelope for references
	envelope := s3state.NewEnvelope(verificationId)
	envelope.SetStatus(state.VerificationContext.Status)

	// Add verification context to summary
	envelope.AddSummary("status", state.VerificationContext.Status)
	envelope.AddSummary("verificationAt", state.VerificationContext.VerificationAt)
	
	// Add turn information to summary if available
	if state.ConversationState != nil {
		envelope.AddSummary("currentTurn", state.ConversationState.CurrentTurn)
		envelope.AddSummary("maxTurns", state.ConversationState.MaxTurns)
	}

	// Save each component of the state
	var saveErrors []error

	// Save verification context
	if state.VerificationContext != nil {
		err := s.stateManager.SaveToEnvelope(envelope, "contexts", "verification-context.json", state.VerificationContext)
		if err != nil {
			saveErrors = append(saveErrors, fmt.Errorf("failed to save verification context: %w", err))
		}
	}

	// Save current prompt
	if state.CurrentPrompt != nil {
		err := s.stateManager.SaveToEnvelope(envelope, "prompts", "system-prompt.json", state.CurrentPrompt)
		if err != nil {
			saveErrors = append(saveErrors, fmt.Errorf("failed to save current prompt: %w", err))
		}
	}

	// Save images
	if state.Images != nil {
		err := s.stateManager.SaveToEnvelope(envelope, "images", "image-data.json", state.Images)
		if err != nil {
			saveErrors = append(saveErrors, fmt.Errorf("failed to save images: %w", err))
		}
	}

	// Save Bedrock config
	if state.BedrockConfig != nil {
		err := s.stateManager.SaveToEnvelope(envelope, "configs", "bedrock-config.json", state.BedrockConfig)
		if err != nil {
			saveErrors = append(saveErrors, fmt.Errorf("failed to save Bedrock config: %w", err))
		}
	}

	// Save conversation state
	if state.ConversationState != nil {
		err := s.stateManager.SaveToEnvelope(envelope, "conversations", "conversation-state.json", state.ConversationState)
		if err != nil {
			saveErrors = append(saveErrors, fmt.Errorf("failed to save conversation state: %w", err))
		}
	}

	// Save Turn1 response if available
	if state.Turn1Response != nil {
		err := s.stateManager.SaveToEnvelope(envelope, "responses", "turn1-response.json", state.Turn1Response)
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
		VerificationContext: envelope.GetReference("contexts_verification-context"),
		SystemPrompt:        envelope.GetReference("prompts_system-prompt"),
		Images:              envelope.GetReference("images_image-data"),
		BedrockConfig:       envelope.GetReference("configs_bedrock-config"),
		ConversationState:   envelope.GetReference("conversations_conversation-state"),
		Turn1Response:       envelope.GetReference("responses_turn1-response"),
	}

	s.logger.Info("Workflow state saved successfully", map[string]interface{}{
		"verificationId": verificationId,
		"status":         state.VerificationContext.Status,
		"references":     len(envelope.References),
	})

	return stateRefs, nil
}

// SaveTurn1Response saves just the Turn1 response and returns updated references
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

	// Store the turn response directly
	ref, err := s.stateManager.StoreJSON("responses", 
		fmt.Sprintf("%s/turn1-response.json", verificationId), turnResponse)
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

// SaveThinkingContent extracts and saves thinking content separately
func (s *Saver) SaveThinkingContent(ctx context.Context, verificationId string, thinking string) (*s3state.Reference, error) {
	if thinking == "" {
		s.logger.Info("No thinking content to save", nil)
		return nil, nil
	}

	// Add timeout to context
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	
	s.logger.Info("Saving thinking content to S3", map[string]interface{}{
		"verificationId":  verificationId,
		"thinkingLength": len(thinking),
	})

	// Create a wrapper object for the thinking content
	thinkingObj := map[string]interface{}{
		"verificationId": verificationId,
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
		"content":        thinking,
		"length":         len(thinking),
	}

	// Store the thinking content
	ref, err := s.stateManager.StoreJSON("thinking", 
		fmt.Sprintf("%s/turn1-thinking.json", verificationId), thinkingObj)
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