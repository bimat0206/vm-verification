package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	//"strings"

	"workflow-function/shared/s3utils"
	"workflow-function/shared/schema"
	"workflow-function/shared/schema/validation"
)

// Environment variable names for buckets
const (
	ENV_REFERENCE_BUCKET = "REFERENCE_BUCKET"
	ENV_CHECKING_BUCKET  = "CHECKING_BUCKET"
)

// ValidateInput performs comprehensive validation on the Lambda input
func ValidateInput(input *Input) error {
	if input == nil {
		return fmt.Errorf("input is nil")
	}
	
	if input.VerificationContext == nil {
		return fmt.Errorf("verification context is required")
	}
	
	// Validate verification context using shared schema validation
	errs := schema.ValidateVerificationContext(input.VerificationContext)
	if len(errs) > 0 {
		// Return the first error
		return errs[0]
	}
	
	// Type-specific validation
	if input.VerificationContext.VerificationType == schema.VerificationTypeLayoutVsChecking {
		return validateLayoutVsChecking(input)
	} else if input.VerificationContext.VerificationType == schema.VerificationTypePreviousVsCurrent {
		return validatePreviousVsCurrent(input)
	}
	
	return nil
}

// validateLayoutVsChecking performs validation specific to LAYOUT_VS_CHECKING type
func validateLayoutVsChecking(input *Input) error {
	ctx := input.VerificationContext
	
	// Layout ID is required
	if ctx.LayoutId <= 0 {
		return schema.ValidationError{Field: "layoutId", Message: "layout ID is required and must be positive"}
	}
	
	// LayoutPrefix is required
	if ctx.LayoutPrefix == "" {
		return schema.ValidationError{Field: "layoutPrefix", Message: "layout prefix is required"}
	}
	
	// Validate layout metadata
	if input.LayoutMetadata == nil {
		return schema.ValidationError{Field: "layoutMetadata", Message: "layout metadata is required for LAYOUT_VS_CHECKING"}
	}
	
	// Extract and validate machine structure
	ms, err := ExtractMachineStructure(input.LayoutMetadata)
	if err != nil {
		return schema.ValidationError{Field: "layoutMetadata", Message: fmt.Sprintf("failed to extract machine structure: %v", err)}
	}
	
	if ms == nil {
		return schema.ValidationError{Field: "machineStructure", Message: "machine structure is required"}
	}
	
	if err := validateMachineStructure(ms); err != nil {
		return err
	}
	
	// Get bucket names from environment variables
	referenceBucket := os.Getenv(ENV_REFERENCE_BUCKET)
	checkingBucket := os.Getenv(ENV_CHECKING_BUCKET)
	
	if referenceBucket == "" {
		return schema.ValidationError{Field: ENV_REFERENCE_BUCKET, Message: "reference bucket environment variable is not set"}
	}
	
	if checkingBucket == "" {
		return schema.ValidationError{Field: ENV_CHECKING_BUCKET, Message: "checking bucket environment variable is not set"}
	}
	
	// Validate reference image URL if present
	if ctx.ReferenceImageUrl != "" {
		// Create a simple S3Utils instance for validation
		s3Util := &s3utils.S3Utils{}
		
		// Parse S3 URL using shared S3Utils
		parsed, err := s3Util.ParseS3URL(ctx.ReferenceImageUrl)
		if err != nil {
			return schema.ValidationError{Field: "referenceImageUrl", Message: fmt.Sprintf("invalid S3 URL: %v", err)}
		}
		
		if parsed.Bucket != referenceBucket {
			return schema.ValidationError{Field: "referenceImageUrl", 
				Message: fmt.Sprintf("reference image must be in the reference bucket (%s), found bucket (%s) instead", 
					referenceBucket, parsed.Bucket)}
		}
		
		// Validate image format
		if !s3Util.IsValidImageExtension(filepath.Ext(parsed.Key)) {
			return schema.ValidationError{Field: "referenceImageUrl", 
				Message: "image format must be JPEG or PNG (supported by Bedrock)"}
		}
	}
	
	// Validate checking image URL if present
	if ctx.CheckingImageUrl != "" {
		// Create a simple S3Utils instance for validation
		s3Util := &s3utils.S3Utils{}
		
		// Parse S3 URL using shared S3Utils
		parsed, err := s3Util.ParseS3URL(ctx.CheckingImageUrl)
		if err != nil {
			return schema.ValidationError{Field: "checkingImageUrl", Message: fmt.Sprintf("invalid S3 URL: %v", err)}
		}
		
		if parsed.Bucket != checkingBucket {
			return schema.ValidationError{Field: "checkingImageUrl", 
				Message: fmt.Sprintf("checking image must be in the checking bucket (%s), found bucket (%s) instead", 
					checkingBucket, parsed.Bucket)}
		}
		
		// Validate image format
		if !s3Util.IsValidImageExtension(filepath.Ext(parsed.Key)) {
			return schema.ValidationError{Field: "checkingImageUrl", 
				Message: "image format must be JPEG or PNG (supported by Bedrock)"}
		}
	}
	
	return nil
}

// validatePreviousVsCurrent performs validation specific to PREVIOUS_VS_CURRENT type
func validatePreviousVsCurrent(input *Input) error {
	ctx := input.VerificationContext
	
	// LayoutID should not be set
	if ctx.LayoutId != 0 {
		return schema.ValidationError{Field: "layoutId", Message: "layout ID should not be set for PREVIOUS_VS_CURRENT"}
	}
	
	// LayoutPrefix should not be set
	if ctx.LayoutPrefix != "" {
		return schema.ValidationError{Field: "layoutPrefix", Message: "layout prefix should not be set for PREVIOUS_VS_CURRENT"}
	}
	
	// Get bucket names from environment variables
	checkingBucket := os.Getenv(ENV_CHECKING_BUCKET)
	
	if checkingBucket == "" {
		return schema.ValidationError{Field: ENV_CHECKING_BUCKET, Message: "checking bucket environment variable is not set"}
	}
	
	// Reference image URL is required and should point to checking bucket
	if ctx.ReferenceImageUrl == "" {
		return schema.ValidationError{Field: "referenceImageUrl", Message: "reference image URL is required for PREVIOUS_VS_CURRENT"}
	}
	
	// Create a simple S3Utils instance for validation
	s3Util := &s3utils.S3Utils{}
	
	// Parse S3 URL using shared S3Utils
	parsed, err := s3Util.ParseS3URL(ctx.ReferenceImageUrl)
	if err != nil {
		return schema.ValidationError{Field: "referenceImageUrl", Message: fmt.Sprintf("invalid S3 URL: %v", err)}
	}
	
	if parsed.Bucket != checkingBucket {
		return schema.ValidationError{Field: "referenceImageUrl", 
			Message: fmt.Sprintf("for PREVIOUS_VS_CURRENT, reference image must be in the checking bucket (%s), found bucket (%s) instead", 
				checkingBucket, parsed.Bucket)}
	}
	
	// Validate image format
	if !s3Util.IsValidImageExtension(filepath.Ext(parsed.Key)) {
		return schema.ValidationError{Field: "referenceImageUrl", 
			Message: "image format must be JPEG or PNG (supported by Bedrock)"}
	}
	
	// Validate checking image URL
	if ctx.CheckingImageUrl == "" {
		return schema.ValidationError{Field: "checkingImageUrl", Message: "checking image URL is required"}
	}
	
	// Parse S3 URL using shared S3Utils
	parsed2, err := s3Util.ParseS3URL(ctx.CheckingImageUrl)
	if err != nil {
		return schema.ValidationError{Field: "checkingImageUrl", Message: fmt.Sprintf("invalid S3 URL: %v", err)}
	}
	
	if parsed2.Bucket != checkingBucket {
		return schema.ValidationError{Field: "checkingImageUrl", 
			Message: fmt.Sprintf("checking image must be in the checking bucket (%s), found bucket (%s) instead", 
				checkingBucket, parsed2.Bucket)}
	}
	
	// Validate image format
	if !s3Util.IsValidImageExtension(filepath.Ext(parsed2.Key)) {
		return schema.ValidationError{Field: "checkingImageUrl", 
			Message: "image format must be JPEG or PNG (supported by Bedrock)"}
	}
	
	// Validate historical context if present
	if input.HistoricalContext != nil {
		var prevVerificationId string
		var prevVerificationAt string
		
		// Extract previous verification ID
		if idVal, ok := input.HistoricalContext["previousVerificationId"]; ok {
			if idStr, ok := idVal.(string); ok {
				prevVerificationId = idStr
			}
		}
		
		if prevVerificationId == "" {
			return schema.ValidationError{Field: "historicalContext.previousVerificationId", 
				Message: "previous verification ID is required when historical context is provided"}
		}
		
		// Extract and validate previous verification timestamp
		if atVal, ok := input.HistoricalContext["previousVerificationAt"]; ok {
			if atStr, ok := atVal.(string); ok {
				prevVerificationAt = atStr
				if !isValidISO8601(prevVerificationAt) {
					return schema.ValidationError{Field: "historicalContext.previousVerificationAt", 
						Message: "previous verification timestamp must be in ISO8601 format"}
				}
			}
		}
		
		// Extract and validate machine structure if present
		if msVal, ok := input.HistoricalContext["machineStructure"]; ok {
			// Convert to JSON and extract machine structure
			msData, err := ExtractMachineStructure(map[string]interface{}{"machineStructure": msVal})
			if err == nil && msData != nil {
				if err := validateMachineStructure(msData); err != nil {
					return err
				}
			}
		}
	}
	
	return nil
}

// validateMachineStructure checks machine structure validity
func validateMachineStructure(ms *MachineStructure) error {
	if ms == nil {
		return validation.NewValidationError("machineStructure", "machine structure is required")
	}
	
	// Validate row count
	if ms.RowCount <= 0 {
		return validation.NewValidationError("machineStructure.rowCount", "row count must be positive")
	}
	
	// Validate column count
	if ms.ColumnsPerRow <= 0 {
		return validation.NewValidationError("machineStructure.columnsPerRow", "columns per row must be positive")
	}
	
	// Validate row order array
	if len(ms.RowOrder) == 0 {
		return validation.NewValidationError("machineStructure.rowOrder", "row order array cannot be empty")
	}
	
	if len(ms.RowOrder) != ms.RowCount {
		return validation.NewValidationError("machineStructure.rowOrder",
			fmt.Sprintf("row order array length (%d) does not match row count (%d)", 
				len(ms.RowOrder), ms.RowCount))
	}
	
	// Validate column order array
	if len(ms.ColumnOrder) == 0 {
		return validation.NewValidationError("machineStructure.columnOrder", "column order array cannot be empty")
	}
	
	if len(ms.ColumnOrder) != ms.ColumnsPerRow {
		return validation.NewValidationError("machineStructure.columnOrder",
			fmt.Sprintf("column order array length (%d) does not match columns per row (%d)", 
				len(ms.ColumnOrder), ms.ColumnsPerRow))
	}
	
	return nil
}

// isValidISO8601 checks if a string is a valid ISO8601 timestamp
func isValidISO8601(timestamp string) bool {
	// This is a simplified pattern that matches common ISO8601 formats
	pattern := `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(.\d+)?(Z|[+-]\d{2}:\d{2})?$`
	match, _ := regexp.MatchString(pattern, timestamp)
	return match
}
