package state

import (
	"fmt"
	"time"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/s3state"
	"workflow-function/shared/schema"
)

// Saver handles saving state data to S3
type Saver struct {
	s3Manager s3state.Manager
	log       logger.Logger
}

// NewSaver creates a new state saver with the given S3 manager and logger
func NewSaver(s3Manager s3state.Manager, log logger.Logger) *Saver {
	return &Saver{
		s3Manager: s3Manager,
		log:       log,
	}
}

// SaveTurn1Prompt saves the Turn 1 prompt and updates the state
func (s *Saver) SaveTurn1Prompt(state *schema.WorkflowState, output *Output) error {
	// Envelope for storing data in S3
	envelope := &s3state.Envelope{
		VerificationID: state.VerificationContext.VerificationId,
		References:     make(map[string]*s3state.Reference),
	}

	// Save prompt data
	if err := s.savePromptData(state, envelope); err != nil {
		return fmt.Errorf("failed to save prompt data: %w", err)
	}

	// Save processing metrics
	if err := s.saveProcessingMetrics(state, envelope); err != nil {
		return fmt.Errorf("failed to save processing metrics: %w", err)
	}

	// Add references to output
	for key, ref := range envelope.References {
		output.References[key] = ref
	}

	// Update output status
	output.Status = state.VerificationContext.Status

	return nil
}

// savePromptData saves the Turn 1 prompt data to S3
func (s *Saver) savePromptData(state *schema.WorkflowState, envelope *s3state.Envelope) error {
	// Check if current prompt has at least one message
	if state.CurrentPrompt == nil || len(state.CurrentPrompt.Messages) == 0 {
		return errors.NewValidationError("Current prompt has no messages", nil)
	}

	// Create prompt data structure
	promptData := &schema.CurrentPrompt{
		Messages:      state.CurrentPrompt.Messages,
		TurnNumber:    state.CurrentPrompt.TurnNumber,
		PromptId:      state.CurrentPrompt.PromptId,
		CreatedAt:     state.CurrentPrompt.CreatedAt,
		PromptVersion: state.CurrentPrompt.PromptVersion,
		IncludeImage:  state.CurrentPrompt.IncludeImage,
	}

	// Save to S3
	if err := s.s3Manager.SaveToEnvelope(envelope, CategoryPrompts, KeyTurn1Prompt, promptData); err != nil {
		return errors.NewInternalError("prompt-save", err)
	}

	s.log.Info("Saved Turn 1 prompt data", map[string]interface{}{
		"promptId":    promptData.PromptId,
		"turnNumber":  promptData.TurnNumber,
		"messageCount": len(promptData.Messages),
	})

	return nil
}

// saveProcessingMetrics saves the Turn 1 processing metrics to S3
func (s *Saver) saveProcessingMetrics(state *schema.WorkflowState, envelope *s3state.Envelope) error {
	// Create metrics data structure
	metrics := map[string]interface{}{
		"verificationId":   state.VerificationContext.VerificationId,
		"verificationType": state.VerificationContext.VerificationType,
		"turnNumber":       state.CurrentPrompt.TurnNumber,
		"promptId":         state.CurrentPrompt.PromptId,
		"createdAt":        state.CurrentPrompt.CreatedAt,
		"status":           state.VerificationContext.Status,
		"timestamp":        time.Now().Format(time.RFC3339),
	}

	// Add image metrics if available
	if state.Images != nil {
		if state.Images.Reference != nil {
			metrics["referenceImageInfo"] = map[string]interface{}{
				"format":          state.Images.Reference.Format,
				"storageMethod":   state.Images.Reference.StorageMethod,
				"base64Generated": state.Images.Reference.Base64Generated,
				"size":            state.Images.Reference.Size,
			}
		}

		if state.Images.Checking != nil {
			metrics["checkingImageInfo"] = map[string]interface{}{
				"format":          state.Images.Checking.Format,
				"storageMethod":   state.Images.Checking.StorageMethod,
				"base64Generated": state.Images.Checking.Base64Generated,
				"size":            state.Images.Checking.Size,
			}
		}
	}

	// Add bedrock config if available
	if state.BedrockConfig != nil {
		metrics["bedrockConfig"] = map[string]interface{}{
			"anthropicVersion": state.BedrockConfig.AnthropicVersion,
			"maxTokens":        state.BedrockConfig.MaxTokens,
		}

		if state.BedrockConfig.Thinking != nil {
			metrics["bedrockConfig"].(map[string]interface{})["thinking"] = map[string]interface{}{
				"type":         state.BedrockConfig.Thinking.Type,
				"budgetTokens": state.BedrockConfig.Thinking.BudgetTokens,
			}
		}
	}

	// Save to S3
	if err := s.s3Manager.SaveToEnvelope(envelope, CategoryProcessing, KeyTurn1Metrics, metrics); err != nil {
		return errors.NewInternalError("metrics-save", err)
	}

	s.log.Info("Saved Turn 1 processing metrics", map[string]interface{}{
		"verificationId": state.VerificationContext.VerificationId,
		"turnNumber":     state.CurrentPrompt.TurnNumber,
	})

	return nil
}