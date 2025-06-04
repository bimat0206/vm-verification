package schema

import (
	"fmt"
	"strings"
)

// ValidationError represents a schema validation error
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface
func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error for field %s: %s", e.Field, e.Message)
}

// Errors is a collection of validation errors
type Errors []ValidationError

// Error implements the error interface for a collection of errors
func (e Errors) Error() string {
	if len(e) == 0 {
		return ""
	}

	messages := make([]string, len(e))
	for i, err := range e {
		messages[i] = err.Error()
	}
	return strings.Join(messages, "; ")
}

// ValidateVerificationContext validates the VerificationContext struct
func ValidateVerificationContext(ctx *VerificationContext) Errors {
	var errors Errors

	// Required fields
	if ctx.VerificationId == "" {
		errors = append(errors, ValidationError{Field: "verificationId", Message: "required field missing"})
	}
	if ctx.VerificationAt == "" {
		errors = append(errors, ValidationError{Field: "verificationAt", Message: "required field missing"})
	}
	if ctx.Status == "" {
		errors = append(errors, ValidationError{Field: "status", Message: "required field missing"})
	} else {
		// Validate status is a known value
		validStatus := false
		for _, s := range []string{
			StatusVerificationRequested, StatusVerificationInitialized, StatusFetchingImages,
			StatusImagesFetched, StatusPromptPrepared, StatusTurn1PromptReady, StatusTurn1Completed,
			StatusTurn1Processed, StatusTurn2PromptReady, StatusTurn2Completed, StatusTurn2Processed,
			StatusResultsFinalized, StatusResultsStored, StatusCompleted,
			StatusInitializationFailed, StatusHistoricalFetchFailed, StatusImageFetchFailed,
			StatusBedrockProcessingFailed, StatusVerificationFailed,
		} {
			if ctx.Status == s {
				validStatus = true
				break
			}
		}
		if !validStatus {
			errors = append(errors, ValidationError{Field: "status", Message: "invalid status value"})
		}
	}
	if ctx.VerificationType == "" {
		errors = append(errors, ValidationError{Field: "verificationType", Message: "required field missing"})
	} else {
		// Validate verification type
		if ctx.VerificationType != VerificationTypeLayoutVsChecking &&
			ctx.VerificationType != VerificationTypePreviousVsCurrent {
			errors = append(errors, ValidationError{
				Field:   "verificationType",
				Message: "must be either LAYOUT_VS_CHECKING or PREVIOUS_VS_CURRENT",
			})
		}
	}
	if ctx.ReferenceImageUrl == "" {
		errors = append(errors, ValidationError{Field: "referenceImageUrl", Message: "required field missing"})
	}
	if ctx.CheckingImageUrl == "" {
		errors = append(errors, ValidationError{Field: "checkingImageUrl", Message: "required field missing"})
	}
	if ctx.VendingMachineId == "" {
		errors = append(errors, ValidationError{Field: "vendingMachineId", Message: "required field missing"})
	}

	// Verification type-specific validations
	if ctx.VerificationType == VerificationTypeLayoutVsChecking {
		if ctx.LayoutId == 0 {
			errors = append(errors, ValidationError{
				Field:   "layoutId",
				Message: "required for LAYOUT_VS_CHECKING verification type",
			})
		}
		if ctx.LayoutPrefix == "" {
			errors = append(errors, ValidationError{
				Field:   "layoutPrefix",
				Message: "required for LAYOUT_VS_CHECKING verification type",
			})
		}
	}

	return errors
}

// ValidateWorkflowState validates the complete workflow state
func ValidateWorkflowState(state *WorkflowState) Errors {
	var errors Errors

	// Check schema version - v2.0.0
	if state.SchemaVersion == "" {
		state.SchemaVersion = SchemaVersion
	} else if state.SchemaVersion != SchemaVersion {
		errors = append(errors, ValidationError{
			Field: "schemaVersion",
			Message: fmt.Sprintf("unsupported schema version: %s (supported: %s)",
				state.SchemaVersion, SchemaVersion),
		})
	}

	// Always validate verification context
	if state.VerificationContext == nil {
		errors = append(errors, ValidationError{Field: "verificationContext", Message: "required field missing"})
	} else {
		ctxErrors := ValidateVerificationContext(state.VerificationContext)
		errors = append(errors, ctxErrors...)
	}

	// Add additional validations for other fields based on the current state
	// For example, if in TURN1_COMPLETED, validate Turn1Response exists
	if state.VerificationContext != nil {
		status := state.VerificationContext.Status
		switch status {
		case StatusTurn1Completed, StatusTurn1Processed:
			if state.Turn1Response == nil || len(state.Turn1Response) == 0 {
				errors = append(errors, ValidationError{
					Field:   "turn1Response",
					Message: fmt.Sprintf("required when status is %s", status),
				})
			}
		case StatusTurn2Completed, StatusTurn2Processed:
			if state.Turn2Response == nil || len(state.Turn2Response) == 0 {
				errors = append(errors, ValidationError{
					Field:   "turn2Response",
					Message: fmt.Sprintf("required when status is %s", status),
				})
			}
		case StatusResultsFinalized, StatusResultsStored, StatusCompleted:
			if state.FinalResults == nil {
				errors = append(errors, ValidationError{
					Field:   "finalResults",
					Message: fmt.Sprintf("required when status is %s", status),
				})
			}
		}
	}

	return errors
}

// ValidateImageInfo validates image information including S3 storage for Base64 data
func ValidateImageInfo(img *ImageInfo, requireBase64 bool) Errors {
	var errors Errors

	if img == nil {
		errors = append(errors, ValidationError{Field: "imageInfo", Message: "cannot be nil"})
		return errors
	}

	// Basic validations
	if img.URL == "" {
		errors = append(errors, ValidationError{Field: "url", Message: "required"})
	}
	if img.S3Key == "" {
		errors = append(errors, ValidationError{Field: "s3Key", Message: "required"})
	}
	if img.S3Bucket == "" {
		errors = append(errors, ValidationError{Field: "s3Bucket", Message: "required"})
	}

	// Base64 validation when required
	if requireBase64 {
		// Check for S3 storage of Base64 data
		if !img.HasBase64Data() {
			errors = append(errors, ValidationError{Field: "base64S3Key", Message: "S3 key for Base64 data required for Bedrock API"})
		}
		if img.Format == "" {
			errors = append(errors, ValidationError{Field: "format", Message: "required for Base64 data"})
		}

		// Validate supported formats
		supportedFormats := []string{"png", "jpeg", "jpg"}
		formatSupported := false
		for _, fmt := range supportedFormats {
			if img.Format == fmt {
				formatSupported = true
				break
			}
		}
		if !formatSupported {
			errors = append(errors, ValidationError{
				Field:   "format",
				Message: fmt.Sprintf("unsupported format: %s (supported: %v)", img.Format, supportedFormats),
			})
		}

		// Validate Base64 size
		if err := img.ValidateBase64Size(); err != nil {
			errors = append(errors, ValidationError{
				Field:   "base64Size",
				Message: err.Error(),
			})
		}
	}

	return errors
}

// ValidateImageData validates the complete ImageData structure
func ValidateImageData(images *ImageData, requireBase64 bool) Errors {
	var errors Errors

	if images == nil {
		errors = append(errors, ValidationError{Field: "images", Message: "cannot be nil"})
		return errors
	}

	// Validate reference image
	if images.Reference == nil {
		errors = append(errors, ValidationError{Field: "reference", Message: "required"})
	} else {
		refErrors := ValidateImageInfo(images.Reference, requireBase64)
		for _, err := range refErrors {
			errors = append(errors, ValidationError{
				Field:   "reference." + err.Field,
				Message: err.Message,
			})
		}
	}

	// Validate checking image
	if images.Checking == nil {
		errors = append(errors, ValidationError{Field: "checking", Message: "required"})
	} else {
		checkErrors := ValidateImageInfo(images.Checking, requireBase64)
		for _, err := range checkErrors {
			errors = append(errors, ValidationError{
				Field:   "checking." + err.Field,
				Message: err.Message,
			})
		}
	}

	// Validate Base64 generation flag consistency
	if requireBase64 && !images.Base64Generated {
		errors = append(errors, ValidationError{
			Field:   "base64Generated",
			Message: "must be true when Base64 data is required",
		})
	}

	return errors
}

// ValidateBedrockMessages validates Bedrock message format according to schema v2.0.0
func ValidateBedrockMessages(messages []BedrockMessage) Errors {
	var errors Errors

	if len(messages) == 0 {
		errors = append(errors, ValidationError{Field: "messages", Message: "at least one message required"})
		return errors
	}

	for i, msg := range messages {
		if msg.Role == "" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("messages[%d].role", i),
				Message: "required",
			})
		} else if msg.Role != "user" && msg.Role != "assistant" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("messages[%d].role", i),
				Message: "must be 'user' or 'assistant'",
			})
		}

		if len(msg.Content) == 0 {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("messages[%d].content", i),
				Message: "at least one content item required",
			})
		}

		hasText := false
		for j, content := range msg.Content {
			if content.Type == "" {
				errors = append(errors, ValidationError{
					Field:   fmt.Sprintf("messages[%d].content[%d].type", i, j),
					Message: "required",
				})
			} else if content.Type != "text" && content.Type != "image" {
				errors = append(errors, ValidationError{
					Field:   fmt.Sprintf("messages[%d].content[%d].type", i, j),
					Message: "must be 'text' or 'image'",
				})
			}

			if content.Type == "text" {
				hasText = true
				if content.Text == "" {
					errors = append(errors, ValidationError{
						Field:   fmt.Sprintf("messages[%d].content[%d].text", i, j),
						Message: "required for text content",
					})
				}
			}

			if content.Type == "image" {
				if content.Image == nil {
					errors = append(errors, ValidationError{
						Field:   fmt.Sprintf("messages[%d].content[%d].image", i, j),
						Message: "required for image content",
					})
				} else {
					if content.Image.Format == "" {
						errors = append(errors, ValidationError{
							Field:   fmt.Sprintf("messages[%d].content[%d].image.format", i, j),
							Message: "required",
						})
					}
					// Validate the updated v2.0.0 format with new field names
					if content.Image.Source.Type != "base64" {
						errors = append(errors, ValidationError{
							Field:   fmt.Sprintf("messages[%d].content[%d].image.source.type", i, j),
							Message: "must be 'base64'",
						})
					}

					if content.Image.Source.Media_type == "" {
						errors = append(errors, ValidationError{
							Field:   fmt.Sprintf("messages[%d].content[%d].image.source.media_type", i, j),
							Message: "media type is required",
						})
					} else {
						validMediaTypes := []string{"image/png", "image/jpeg", "image/jpg"}
						isValidMediaType := false
						for _, validType := range validMediaTypes {
							if content.Image.Source.Media_type == validType {
								isValidMediaType = true
								break
							}
						}
						if !isValidMediaType {
							errors = append(errors, ValidationError{
								Field:   fmt.Sprintf("messages[%d].content[%d].image.source.media_type", i, j),
								Message: fmt.Sprintf("must be one of %v", validMediaTypes),
							})
						}
					}

					if content.Image.Source.Data == "" {
						errors = append(errors, ValidationError{
							Field:   fmt.Sprintf("messages[%d].content[%d].image.source.data", i, j),
							Message: "Base64 image data required",
						})
					}
				}
			}
		}

		// Ensure each message has at least one text content
		if !hasText {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("messages[%d]", i),
				Message: "must contain at least one text content",
			})
		}
	}

	return errors
}

// ValidateCurrentPrompt validates the CurrentPrompt structure
func ValidateCurrentPrompt(prompt *CurrentPrompt, requireMessages bool) Errors {
	var errors Errors

	if prompt == nil {
		errors = append(errors, ValidationError{Field: "currentPrompt", Message: "cannot be nil"})
		return errors
	}

	if prompt.TurnNumber <= 0 {
		errors = append(errors, ValidationError{Field: "turnNumber", Message: "must be greater than 0"})
	}

	if prompt.IncludeImage == "" {
		errors = append(errors, ValidationError{Field: "includeImage", Message: "required"})
	} else {
		validIncludeImage := false
		for _, val := range []string{"reference", "checking", "both", "none"} {
			if prompt.IncludeImage == val {
				validIncludeImage = true
				break
			}
		}
		if !validIncludeImage {
			errors = append(errors, ValidationError{
				Field:   "includeImage",
				Message: "must be one of: reference, checking, both, none",
			})
		}
	}

	// Validate messages if required (for Bedrock API calls)
	if requireMessages {
		if len(prompt.Messages) == 0 {
			errors = append(errors, ValidationError{Field: "messages", Message: "required for Bedrock API"})
		} else {
			msgErrors := ValidateBedrockMessages(prompt.Messages)
			errors = append(errors, msgErrors...)
		}
	}

	// Backward compatibility: validate text prompt if messages are not present
	if len(prompt.Messages) == 0 && prompt.Text == "" {
		errors = append(errors, ValidationError{Field: "text", Message: "required when messages are not provided"})
	}

	return errors
}

// ValidateBedrockConfig validates the Bedrock configuration
func ValidateBedrockConfig(config *BedrockConfig) Errors {
	var errors Errors

	if config == nil {
		errors = append(errors, ValidationError{Field: "bedrockConfig", Message: "cannot be nil"})
		return errors
	}

	if config.AnthropicVersion == "" {
		errors = append(errors, ValidationError{Field: "anthropicVersion", Message: "required"})
	}

	if config.MaxTokens <= 0 {
		errors = append(errors, ValidationError{Field: "maxTokens", Message: "must be greater than 0"})
	}

	// Validate thinking configuration if present
	if config.Thinking != nil {
		if config.Thinking.Type == "" {
			errors = append(errors, ValidationError{Field: "thinking.type", Message: "required when thinking is enabled"})
		}
		if config.Thinking.BudgetTokens <= 0 {
			errors = append(errors, ValidationError{Field: "thinking.budgetTokens", Message: "must be greater than 0"})
		}
	}

	return errors
}

// Add these validation functions for new structures

// ValidateStatusHistoryEntry validates status history entries
func ValidateStatusHistoryEntry(entry *StatusHistoryEntry) Errors {
	var errors Errors

	if entry == nil {
		errors = append(errors, ValidationError{Field: "statusHistoryEntry", Message: "cannot be nil"})
		return errors
	}

	if entry.Status == "" {
		errors = append(errors, ValidationError{Field: "status", Message: "required"})
	}

	if entry.Timestamp == "" {
		errors = append(errors, ValidationError{Field: "timestamp", Message: "required"})
	}

	if entry.FunctionName == "" {
		errors = append(errors, ValidationError{Field: "functionName", Message: "required"})
	}

	return errors
}

// ValidateProcessingMetrics validates processing metrics
func ValidateProcessingMetrics(metrics *ProcessingMetrics) Errors {
	var errors Errors

	if metrics == nil {
		return errors // Optional field
	}

	// Add specific validations for metrics
	if metrics.WorkflowTotal != nil {
		if metrics.WorkflowTotal.TotalTimeMs < 0 {
			errors = append(errors, ValidationError{Field: "workflowTotal.totalTimeMs", Message: "cannot be negative"})
		}
	}

	return errors
}

// ValidateErrorTracking validates error tracking structure
func ValidateErrorTracking(tracking *ErrorTracking) Errors {
	var errors Errors

	if tracking == nil {
		return errors // Optional field
	}

	if tracking.RecoveryAttempts < 0 {
		errors = append(errors, ValidationError{Field: "recoveryAttempts", Message: "cannot be negative"})
	}

	return errors
}

// ValidateTemplateProcessor validates template processor structure
func ValidateTemplateProcessor(processor *TemplateProcessor) Errors {
	var errors Errors

	if processor == nil {
		return errors // Optional field
	}

	if processor.Template != nil {
		if processor.Template.TemplateId == "" {
			errors = append(errors, ValidationError{Field: "template.templateId", Message: "required"})
		}
		if processor.Template.Content == "" {
			errors = append(errors, ValidationError{Field: "template.content", Message: "required"})
		}
	}

	if processor.ProcessingTime < 0 {
		errors = append(errors, ValidationError{Field: "processingTime", Message: "cannot be negative"})
	}

	return errors
}

// ValidateCombinedTurnResponse validates combined turn response structure
func ValidateCombinedTurnResponse(response *CombinedTurnResponse) Errors {
	var errors Errors

	if response == nil {
		errors = append(errors, ValidationError{Field: "combinedTurnResponse", Message: "cannot be nil"})
		return errors
	}

	// Validate embedded TurnResponse
	if response.TurnResponse != nil {
		if response.TurnResponse.TurnId <= 0 {
			errors = append(errors, ValidationError{Field: "turnId", Message: "must be greater than 0"})
		}
	}

	// Validate processing stages
	for i, stage := range response.ProcessingStages {
		if stage.StageName == "" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("processingStages[%d].stageName", i),
				Message: "required",
			})
		}
		if stage.Duration < 0 {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("processingStages[%d].duration", i),
				Message: "cannot be negative",
			})
		}
	}

	return errors
}

// ValidateConversationTracker validates conversation tracker structure
func ValidateConversationTracker(tracker *ConversationTracker) Errors {
	var errors Errors

	if tracker == nil {
		return errors // Optional field
	}

	if tracker.ConversationId == "" {
		errors = append(errors, ValidationError{Field: "conversationId", Message: "required"})
	}

	if tracker.CurrentTurn < 0 {
		errors = append(errors, ValidationError{Field: "currentTurn", Message: "cannot be negative"})
	}

	if tracker.MaxTurns <= 0 {
		errors = append(errors, ValidationError{Field: "maxTurns", Message: "must be greater than 0"})
	}

	if tracker.CurrentTurn > tracker.MaxTurns {
		errors = append(errors, ValidationError{Field: "currentTurn", Message: "cannot exceed maxTurns"})
	}

	if tracker.Turn1ProcessedPath != "" && !strings.HasPrefix(tracker.Turn1ProcessedPath, "s3://") {
		errors = append(errors, ValidationError{Field: "turn1ProcessedPath", Message: "must start with s3://"})
	}

	if tracker.Turn2ProcessedPath != "" && !strings.HasPrefix(tracker.Turn2ProcessedPath, "s3://") {
		errors = append(errors, ValidationError{Field: "turn2ProcessedPath", Message: "must start with s3://"})
	}

	return errors
}

// Update ValidateVerificationContext to include new fields validation
func ValidateVerificationContextEnhanced(ctx *VerificationContext) Errors {
	// Start with existing validation
	errors := ValidateVerificationContext(ctx)

	if ctx == nil {
		return errors
	}

	// Validate new fields
	if len(ctx.StatusHistory) > 0 {
		for i, entry := range ctx.StatusHistory {
			entryErrors := ValidateStatusHistoryEntry(&entry)
			for _, err := range entryErrors {
				errors = append(errors, ValidationError{
					Field:   fmt.Sprintf("statusHistory[%d].%s", i, err.Field),
					Message: err.Message,
				})
			}
		}
	}

	if ctx.ProcessingMetrics != nil {
		metricErrors := ValidateProcessingMetrics(ctx.ProcessingMetrics)
		for _, err := range metricErrors {
			errors = append(errors, ValidationError{
				Field:   "processingMetrics." + err.Field,
				Message: err.Message,
			})
		}
	}

	if ctx.ErrorTracking != nil {
		trackingErrors := ValidateErrorTracking(ctx.ErrorTracking)
		for _, err := range trackingErrors {
			errors = append(errors, ValidationError{
				Field:   "errorTracking." + err.Field,
				Message: err.Message,
			})
		}
	}

	return errors
}

// ADD: New validation functions
func ValidateConversationHistory(ch *ConversationHistory) Errors {
	var errors Errors

	if ch == nil {
		errors = append(errors, ValidationError{Field: "conversationHistory", Message: "cannot be nil"})
		return errors
	}

	if ch.VerificationId == "" {
		errors = append(errors, ValidationError{Field: "verificationId", Message: "required"})
	}

	if ch.ConversationAt == "" {
		errors = append(errors, ValidationError{Field: "conversationAt", Message: "required"})
	}

	if ch.CurrentTurn < 0 || ch.CurrentTurn > ch.MaxTurns {
		errors = append(errors, ValidationError{Field: "currentTurn", Message: "invalid turn number"})
	}

	// Validate turn status
	validTurnStatuses := []string{TurnStatusActive, TurnStatusCompleted, TurnStatusFailed}
	if !contains(validTurnStatuses, ch.TurnStatus) {
		errors = append(errors, ValidationError{Field: "turnStatus", Message: "invalid turn status"})
	}

	return errors
}

func ValidateVerificationResults(vr *VerificationResults) Errors {
	var errors Errors

	if vr == nil {
		errors = append(errors, ValidationError{Field: "verificationResults", Message: "cannot be nil"})
		return errors
	}

	if vr.VerificationId == "" {
		errors = append(errors, ValidationError{Field: "verificationId", Message: "required"})
	}

	if vr.VerificationAt == "" {
		errors = append(errors, ValidationError{Field: "verificationAt", Message: "required"})
	}

	// Validate verification type
	validTypes := []string{VerificationTypeLayoutVsChecking, VerificationTypePreviousVsCurrent}
	if !contains(validTypes, vr.VerificationType) {
		errors = append(errors, ValidationError{Field: "verificationType", Message: "invalid verification type"})
	}

	// Validate verification status
	validStatuses := []string{VerificationStatusCorrect, VerificationStatusIncorrect, VerificationStatusPartial, VerificationStatusFailed}
	if !contains(validStatuses, vr.VerificationStatus) {
		errors = append(errors, ValidationError{Field: "verificationStatus", Message: "invalid verification status"})
	}

	// Type-specific validations
	if vr.VerificationType == VerificationTypeLayoutVsChecking {
		if vr.LayoutId == 0 {
			errors = append(errors, ValidationError{Field: "layoutId", Message: "required for layout verification"})
		}
		if vr.LayoutPrefix == "" {
			errors = append(errors, ValidationError{Field: "layoutPrefix", Message: "required for layout verification"})
		}
	}

	if vr.VerificationType == VerificationTypePreviousVsCurrent {
		if vr.PreviousVerificationId == "" {
			errors = append(errors, ValidationError{Field: "previousVerificationId", Message: "required for temporal verification"})
		}
	}

	return errors
}

// Helper function
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
