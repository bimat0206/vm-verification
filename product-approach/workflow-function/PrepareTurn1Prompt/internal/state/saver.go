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
// Modified to accept input and ensure reference preservation
func (s *Saver) SaveTurn1Prompt(state *schema.WorkflowState, input *Input, output *Output) error {
	// Create a new envelope for storing data in S3
	envelope := &s3state.Envelope{
		VerificationID: state.VerificationContext.VerificationId,
		Status:         state.VerificationContext.Status,
		References:     make(map[string]*s3state.Reference),
	}

	// Preserve all existing references from input if not already using the output
	if input != nil && input.References != nil {
		// Use our CopyReferences helper to preserve existing references
		CopyReferences(envelope.References, input.References)

		s.log.Info("Preserved existing references from input", map[string]interface{}{
			"verificationId": state.VerificationContext.VerificationId,
			"referenceCount": len(input.References),
		})
	}

	// Save prompt data and update references
	if err := s.savePromptData(state, envelope); err != nil {
		return fmt.Errorf("failed to save prompt data: %w", err)
	}

	// Save processing metrics and update references
	if err := s.saveProcessingMetrics(state, envelope); err != nil {
		return fmt.Errorf("failed to save processing metrics: %w", err)
	}

	// Update output with all references including preserved ones
	CopyReferences(output.References, envelope.References)

	// Update output status
	output.Status = state.VerificationContext.Status

	inputCount := 0
	if input != nil && input.References != nil {
		inputCount = len(input.References)
	}
	s.log.Info("Successfully saved Turn 1 prompt with reference accumulation", map[string]interface{}{
		"verificationId":  state.VerificationContext.VerificationId,
		"status":          state.VerificationContext.Status,
		"totalReferences": len(output.References),
		"newReferences":   len(envelope.References) - inputCount,
	})

	return nil
}

// savePromptData saves the Turn 1 prompt data to S3
// Enhanced to properly handle Base64 storage references
func (s *Saver) savePromptData(state *schema.WorkflowState, envelope *s3state.Envelope) error {
	// Check if current prompt has at least one message
	if state.CurrentPrompt == nil || len(state.CurrentPrompt.Messages) == 0 {
		return errors.NewValidationError("Current prompt has no messages", nil)
	}

	// Get reference image if available
	var refImage *schema.ImageInfo
	if state.Images != nil {
		refImage = state.Images.GetReference()
	}

	// Create content array for message structure according to schema
	content := make([]map[string]interface{}, 0)

	// First add text content from the first message
	if len(state.CurrentPrompt.Messages) > 0 && len(state.CurrentPrompt.Messages[0].Content) > 0 {
		textContent := map[string]interface{}{
			"type": "text",
			"text": state.CurrentPrompt.Messages[0].Content[0].Text,
		}
		content = append(content, textContent)
	}

	// Create message structure according to schema
	messageStructure := map[string]interface{}{
		"role":    "user",
		"content": content,
	}

	// Create contextual instructions for better documentation
	contextualInstructions := map[string]interface{}{
		"analysisObjective": "Analyze reference image in detail",
	}

	// Add machine structure if available
	if state.VerificationContext.VerificationType == schema.VerificationTypeLayoutVsChecking && state.LayoutMetadata != nil {
		if machineStructure, ok := state.LayoutMetadata["machineStructure"].(map[string]interface{}); ok {
			contextualInstructions["machineStructure"] = machineStructure
		}
	} else if state.VerificationContext.VerificationType == schema.VerificationTypePreviousVsCurrent && state.HistoricalContext != nil {
		if machineStructure, ok := state.HistoricalContext["machineStructure"].(map[string]interface{}); ok {
			contextualInstructions["machineStructure"] = machineStructure
		}
	}

	// Add use case specific guidance
	if state.VerificationContext.VerificationType == schema.VerificationTypeLayoutVsChecking {
		contextualInstructions["useCaseSpecificGuidance"] = "Layout validation against reference planogram"
	} else if state.VerificationContext.VerificationType == schema.VerificationTypePreviousVsCurrent {
		contextualInstructions["useCaseSpecificGuidance"] = "Baseline establishment from previous state"
	}

	// Create image reference object with full Base64 storage reference
	imageReference := map[string]interface{}{
		"imageType": "reference",
	}

	// Add source URL and Base64 reference information
	if refImage != nil && refImage.URL != "" {
		imageReference["sourceUrl"] = refImage.URL

		// Add explicit base64StorageReference for downstream functions
		if refImage.StorageMethod == schema.StorageMethodS3Temporary &&
			refImage.Base64S3Bucket != "" && refImage.GetBase64S3Key() != "" {
			imageReference["base64StorageReference"] = map[string]interface{}{
				"bucket": refImage.Base64S3Bucket,
				"key":    refImage.GetBase64S3Key(),
			}

			s.log.Info("Added Base64 storage reference to prompt", map[string]interface{}{
				"bucket": refImage.Base64S3Bucket,
				"key":    refImage.GetBase64S3Key(),
			})
		} else {
			s.log.Warn("Missing Base64 storage reference for image", map[string]interface{}{
				"url":             refImage.URL,
				"storageMethod":   refImage.StorageMethod,
				"base64Generated": refImage.Base64Generated,
			})
		}
	}

	// Create generation metadata
	generationMetadata := map[string]interface{}{
		"promptSource":   "TEMPLATE_BASED",
		"contextSources": []string{"INITIALIZATION", "IMAGE_METADATA"},
	}

	// Add layout metadata if available
	if state.LayoutMetadata != nil {
		generationMetadata["contextSources"] = append(
			generationMetadata["contextSources"].([]string),
			"LAYOUT_METADATA",
		)
	}

	// Add historical context if available
	if state.HistoricalContext != nil {
		generationMetadata["contextSources"] = append(
			generationMetadata["contextSources"].([]string),
			"HISTORICAL_CONTEXT",
		)
	}

	// Create Turn 1 prompt data structure according to schema
	promptData := map[string]interface{}{
		"verificationId":         state.VerificationContext.VerificationId,
		"promptType":             "TURN1",
		"verificationType":       state.VerificationContext.VerificationType,
		"messageStructure":       messageStructure,
		"contextualInstructions": contextualInstructions,
		"imageReference":         imageReference,
		"templateVersion":        state.CurrentPrompt.PromptVersion,
		"createdAt":              state.CurrentPrompt.CreatedAt,
		"generationMetadata":     generationMetadata,
	}

	// Save to S3 and update envelope references
	// Let the S3Manager handle the verificationId prefix in the path
	if err := s.s3Manager.SaveToEnvelope(envelope, CategoryPrompts, KeyTurn1Prompt, promptData); err != nil {
		return errors.NewInternalError("prompt-save", err)
	}

	// Log successful save with reference information
	s.log.Info("Saved Turn 1 prompt data with proper image references", map[string]interface{}{
		"verificationId":     state.VerificationContext.VerificationId,
		"promptType":         "TURN1",
		"createdAt":          state.CurrentPrompt.CreatedAt,
		"hasBase64Reference": imageReference["base64StorageReference"] != nil,
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

	// Add image metrics if available, including Base64 reference information
	if state.Images != nil {
		if state.Images.Reference != nil {
			refImage := state.Images.Reference
			metrics["referenceImageInfo"] = map[string]interface{}{
				"format":          refImage.Format,
				"storageMethod":   refImage.StorageMethod,
				"base64Generated": refImage.Base64Generated,
				"size":            refImage.Size,
				"base64S3Bucket":  refImage.Base64S3Bucket,
				"base64S3Key":     refImage.GetBase64S3Key(),
			}
		}

		if state.Images.Checking != nil {
			checkImage := state.Images.Checking
			metrics["checkingImageInfo"] = map[string]interface{}{
				"format":          checkImage.Format,
				"storageMethod":   checkImage.StorageMethod,
				"base64Generated": checkImage.Base64Generated,
				"size":            checkImage.Size,
				"base64S3Bucket":  checkImage.Base64S3Bucket,
				"base64S3Key":     checkImage.GetBase64S3Key(),
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

	// Save to S3 and update envelope references
	// Let the S3Manager handle the verificationId prefix in the path
	if err := s.s3Manager.SaveToEnvelope(envelope, CategoryProcessing, KeyTurn1Metrics, metrics); err != nil {
		return errors.NewInternalError("metrics-save", err)
	}

	s.log.Info("Saved Turn 1 processing metrics", map[string]interface{}{
		"verificationId": state.VerificationContext.VerificationId,
		"turnNumber":     state.CurrentPrompt.TurnNumber,
		"base64RefsIncluded": state.Images != nil && state.Images.Reference != nil &&
			state.Images.Reference.Base64S3Bucket != "" && state.Images.Reference.GetBase64S3Key() != "",
	})

	return nil
}
