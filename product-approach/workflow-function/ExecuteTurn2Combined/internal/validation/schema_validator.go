// internal/validation/schema_validator.go
package validation

import (
	"fmt"
	"workflow-function/ExecuteTurn2Combined/internal/models"
	"workflow-function/shared/schema"
)

// SchemaValidator provides validation methods using the schema package
type SchemaValidator struct{}

// NewSchemaValidator creates a new schema validator instance
func NewSchemaValidator() *SchemaValidator {
	return &SchemaValidator{}
}

// ValidateTurn2Request validates the Turn2Request using schema validation
func (v *SchemaValidator) ValidateTurn2Request(req *models.Turn2Request) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	// Validate VerificationID
	if req.VerificationID == "" {
		return schema.ValidationError{Field: "verificationId", Message: "required field missing"}
	}

	// Convert local VerificationContext to schema format for validation
	schemaContext := convertToSchemaVerificationContextTurn2(req.VerificationID, &req.VerificationContext)
	if errors := schema.ValidateVerificationContext(schemaContext); len(errors) > 0 {
		return fmt.Errorf("verification context validation failed: %s", errors.Error())
	}

	// Validate S3 references for Turn2
	if err := v.validateTurn2S3Refs(&req.S3Refs); err != nil {
		return fmt.Errorf("s3 references validation failed: %w", err)
	}

	return nil
}

// Legacy ValidateRequest for compatibility (Turn1)
func (v *SchemaValidator) ValidateRequest(req interface{}) error {
	return fmt.Errorf("legacy Turn1 validation not supported in Turn2 validator")
}

// ValidateTurn2Response validates the Turn2Response
func (v *SchemaValidator) ValidateTurn2Response(resp *models.Turn2Response) error {
	if resp == nil {
		return fmt.Errorf("response cannot be nil")
	}

	// Validate S3 references
	if err := v.validateTurn2S3ResponseRefs(&resp.S3Refs); err != nil {
		return fmt.Errorf("response s3 references validation failed: %w", err)
	}

	// Validate status
	if resp.Status == "" {
		return schema.ValidationError{Field: "status", Message: "required field missing"}
	}

	// Validate summary
	if err := v.validateSummary(&resp.Summary); err != nil {
		return fmt.Errorf("summary validation failed: %w", err)
	}

	// Validate discrepancies
	if err := v.validateDiscrepancies(resp.Discrepancies); err != nil {
		return fmt.Errorf("discrepancies validation failed: %w", err)
	}

	// Validate verification outcome
	if resp.VerificationOutcome == "" {
		return schema.ValidationError{Field: "verificationOutcome", Message: "required field missing"}
	}

	return nil
}

// Legacy ValidateResponse for compatibility (Turn1)
func (v *SchemaValidator) ValidateResponse(resp interface{}) error {
	return fmt.Errorf("legacy Turn1 validation not supported in Turn2 validator")
}

// ValidateWorkflowState validates a complete workflow state
func (v *SchemaValidator) ValidateWorkflowState(state *schema.WorkflowState) error {
	if errors := schema.ValidateWorkflowState(state); len(errors) > 0 {
		return fmt.Errorf("workflow state validation failed: %s", errors.Error())
	}
	return nil
}

// ValidateBedrockMessages validates Bedrock message format
func (v *SchemaValidator) ValidateBedrockMessages(messages []schema.BedrockMessage) error {
	if errors := schema.ValidateBedrockMessages(messages); len(errors) > 0 {
		return fmt.Errorf("bedrock messages validation failed: %s", errors.Error())
	}
	return nil
}

// ValidateImageData validates image data structure
func (v *SchemaValidator) ValidateImageData(images *schema.ImageData, requireBase64 bool) error {
	if errors := schema.ValidateImageData(images, requireBase64); len(errors) > 0 {
		return fmt.Errorf("image data validation failed: %s", errors.Error())
	}
	return nil
}

// Private helper methods for Turn2

func (v *SchemaValidator) validateTurn2S3Refs(refs *models.Turn2RequestS3Refs) error {
	if refs == nil {
		return fmt.Errorf("s3 references cannot be nil")
	}

	// Validate prompts
	if err := v.validateS3Reference(&refs.Prompts.System, "prompts.system"); err != nil {
		return err
	}

	// Validate images
	if err := v.validateS3Reference(&refs.Images.CheckingBase64, "images.checkingBase64"); err != nil {
		return err
	}

	// Validate Turn1 references
	if err := v.validateS3Reference(&refs.Turn1.ProcessedResponse, "turn1.processedResponse"); err != nil {
		return err
	}
	if err := v.validateS3Reference(&refs.Turn1.RawResponse, "turn1.rawResponse"); err != nil {
		return err
	}

	return nil
}

func (v *SchemaValidator) validateTurn2S3ResponseRefs(refs *models.Turn2ResponseS3Refs) error {
	if refs == nil {
		return fmt.Errorf("s3 response references cannot be nil")
	}

	if err := v.validateS3Reference(&refs.RawResponse, "rawResponse"); err != nil {
		return err
	}

	if err := v.validateS3Reference(&refs.ProcessedResponse, "processedResponse"); err != nil {
		return err
	}

	return nil
}

func (v *SchemaValidator) validateDiscrepancies(discrepancies []models.Discrepancy) error {
	for i, d := range discrepancies {
		if d.Item == "" {
			return schema.ValidationError{Field: fmt.Sprintf("discrepancies[%d].item", i), Message: "required field missing"}
		}
		if d.Type == "" {
			return schema.ValidationError{Field: fmt.Sprintf("discrepancies[%d].type", i), Message: "required field missing"}
		}
		// Validate type is one of the expected values
		validTypes := []string{"MISSING", "MISPLACED", "INCORRECT_PRODUCT"}
		isValidType := false
		for _, validType := range validTypes {
			if d.Type == validType {
				isValidType = true
				break
			}
		}
		if !isValidType {
			return schema.ValidationError{Field: fmt.Sprintf("discrepancies[%d].type", i), Message: "invalid discrepancy type"}
		}
	}
	return nil
}

func (v *SchemaValidator) validateS3Reference(ref *models.S3Reference, fieldName string) error {
	if ref == nil {
		return fmt.Errorf("%s reference cannot be nil", fieldName)
	}

	if ref.Bucket == "" {
		return schema.ValidationError{Field: fieldName + ".bucket", Message: "required field missing"}
	}

	if ref.Key == "" {
		return schema.ValidationError{Field: fieldName + ".key", Message: "required field missing"}
	}

	return nil
}

func (v *SchemaValidator) validateSummary(summary *models.Summary) error {
	if summary == nil {
		return fmt.Errorf("summary cannot be nil")
	}

	if summary.AnalysisStage == "" {
		return schema.ValidationError{Field: "analysisStage", Message: "required field missing"}
	}

	if summary.BedrockRequestID == "" {
		return schema.ValidationError{Field: "bedrockRequestId", Message: "required field missing"}
	}

	return nil
}

// ValidateTokenUsage validates token usage structure
func (v *SchemaValidator) ValidateTokenUsage(tokenUsage *models.TokenUsage) error {
	if tokenUsage == nil {
		return fmt.Errorf("token usage cannot be nil")
	}

	if tokenUsage.InputTokens < 0 {
		return schema.ValidationError{Field: "inputTokens", Message: "must be non-negative"}
	}

	if tokenUsage.OutputTokens < 0 {
		return schema.ValidationError{Field: "outputTokens", Message: "must be non-negative"}
	}

	if tokenUsage.TotalTokens != tokenUsage.InputTokens+tokenUsage.OutputTokens+tokenUsage.ThinkingTokens {
		return schema.ValidationError{Field: "totalTokens", Message: "must equal sum of input, output, and thinking tokens"}
	}

	return nil
}

// GetSchemaVersion returns the current schema version
func (v *SchemaValidator) GetSchemaVersion() string {
	return schema.SchemaVersion
}

// CreateErrorInfo creates a standardized error info structure
func (v *SchemaValidator) CreateErrorInfo(code, message string, details map[string]interface{}) *schema.ErrorInfo {
	return &schema.ErrorInfo{
		Code:      code,
		Message:   message,
		Details:   details,
		Timestamp: schema.FormatISO8601(),
	}
}

// convertToSchemaVerificationContextTurn2 converts local VerificationContext to schema format for Turn2
func convertToSchemaVerificationContextTurn2(verificationID string, localCtx *models.VerificationContext) *schema.VerificationContext {
	schemaCtx := &schema.VerificationContext{
		VerificationId:    verificationID,
		VerificationAt:    schema.FormatISO8601(),
		Status:            schema.StatusTurn2PromptReady, // Default status for Turn2
		VerificationType:  localCtx.VerificationType,
		VendingMachineId:  localCtx.VendingMachineId,
		LayoutId:          localCtx.LayoutId,
		LayoutPrefix:      localCtx.LayoutPrefix,
		ReferenceImageUrl: "reference-image-url", // Would be set from context
		CheckingImageUrl:  "checking-image-url",  // Would be set from context
		RequestMetadata: &schema.RequestMetadata{
			RequestId:         verificationID,
			RequestTimestamp:  schema.FormatISO8601(),
			ProcessingStarted: schema.FormatISO8601(),
		},
	}

	// Note: LayoutMetadata and HistoricalContext are handled at WorkflowState level
	// They are not part of the VerificationContext in the schema

	return schemaCtx
}

// Legacy convertToSchemaVerificationContext for compatibility (Turn1)
func convertToSchemaVerificationContext(verificationID string, localCtx *models.VerificationContext) *schema.VerificationContext {
	// This is kept for compatibility but should not be used for Turn2
	return convertToSchemaVerificationContextTurn2(verificationID, localCtx)
}
