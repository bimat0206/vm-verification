package validator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"shared/promptutils"
)

// Environment variable names for buckets
const (
	ENV_REFERENCE_BUCKET = "REFERENCE_BUCKET"
	ENV_CHECKING_BUCKET  = "CHECKING_BUCKET"
)

// ValidateInput performs comprehensive validation on the Lambda input
func ValidateInput(input *promptutils.Input) error {
	// Basic verification context validation
	if err := validateVerificationContext(input.VerificationContext); err != nil {
		return err
	}
	
	// Type-specific validation
	if input.VerificationContext.VerificationType == "LAYOUT_VS_CHECKING" {
		return validateLayoutVsChecking(input)
	} else if input.VerificationContext.VerificationType == "PREVIOUS_VS_CURRENT" {
		return validatePreviousVsCurrent(input)
	}
	
	return nil
}

// validateVerificationContext checks basic verification context properties
func validateVerificationContext(ctx *promptutils.VerificationContext) error {
	if ctx == nil {
		return &promptutils.ValidationError{
			Field:   "verificationContext",
			Message: "verification context is required",
		}
	}
	
	// Validate verification ID
	if ctx.VerificationID == "" {
		return &promptutils.ValidationError{
			Field:   "verificationId",
			Message: "verification ID is required",
		}
	}
	
	// Validate verification ID format (verif-timestamp or similar pattern)
	if !strings.HasPrefix(ctx.VerificationID, "verif-") {
		return &promptutils.ValidationError{
			Field:   "verificationId",
			Message: "verification ID should start with 'verif-'",
		}
	}
	
	// Validate verification type
	if ctx.VerificationType == "" {
		return &promptutils.ValidationError{
			Field:   "verificationType",
			Message: "verification type is required",
		}
	}
	
	// Ensure verification type is one of the supported types
	if ctx.VerificationType != "LAYOUT_VS_CHECKING" && ctx.VerificationType != "PREVIOUS_VS_CURRENT" {
		return &promptutils.ValidationError{
			Field:   "verificationType",
			Message: "verification type must be LAYOUT_VS_CHECKING or PREVIOUS_VS_CURRENT",
		}
	}
	
	// Validate timestamp if present
	if ctx.VerificationAt != "" {
		if !isValidISO8601(ctx.VerificationAt) {
			return &promptutils.ValidationError{
				Field:   "verificationAt",
				Message: "verification timestamp must be in ISO8601 format",
			}
		}
	}
	
	// Validate vending machine ID if present
	if ctx.VendingMachineID != "" {
		if !isValidVendingMachineID(ctx.VendingMachineID) {
			return &promptutils.ValidationError{
				Field:   "vendingMachineId",
				Message: "invalid vending machine ID format",
			}
		}
	}
	
	return nil
}

// validateLayoutVsChecking performs validation specific to LAYOUT_VS_CHECKING type
func validateLayoutVsChecking(input *promptutils.Input) error {
	ctx := input.VerificationContext
	
	// Layout ID is required
	if ctx.LayoutID <= 0 {
		return &promptutils.ValidationError{
			Field:   "layoutId",
			Message: "layout ID is required and must be positive",
		}
	}
	
	// LayoutPrefix is required
	if ctx.LayoutPrefix == "" {
		return &promptutils.ValidationError{
			Field:   "layoutPrefix",
			Message: "layout prefix is required",
		}
	}
	
	// Validate layout metadata
	if input.LayoutMetadata == nil {
		return &promptutils.ValidationError{
			Field:   "layoutMetadata",
			Message: "layout metadata is required for LAYOUT_VS_CHECKING",
		}
	}
	
	// Validate machine structure
	if err := validateMachineStructure(input.LayoutMetadata.MachineStructure); err != nil {
		return err
	}
	
	// Get bucket names from environment variables
	referenceBucket := os.Getenv(ENV_REFERENCE_BUCKET)
	checkingBucket := os.Getenv(ENV_CHECKING_BUCKET)
	
	if referenceBucket == "" {
		return &promptutils.ValidationError{
			Field:   ENV_REFERENCE_BUCKET,
			Message: "reference bucket environment variable is not set",
		}
	}
	
	if checkingBucket == "" {
		return &promptutils.ValidationError{
			Field:   ENV_CHECKING_BUCKET,
			Message: "checking bucket environment variable is not set",
		}
	}
	
	// Validate reference image URL if present
	if ctx.ReferenceImageURL != "" {
		if !isValidS3URL(ctx.ReferenceImageURL) {
			return &promptutils.ValidationError{
				Field:   "referenceImageUrl",
				Message: "invalid S3 URL format for reference image",
			}
		}
		
		// Extract bucket from S3 URL and verify it points to reference bucket
		bucket, key, err := extractS3BucketAndKey(ctx.ReferenceImageURL)
		if err != nil {
			return &promptutils.ValidationError{
				Field:   "referenceImageUrl",
				Message: fmt.Sprintf("failed to parse S3 URL: %v", err),
			}
		}
		
		if bucket != referenceBucket {
			return &promptutils.ValidationError{
				Field:   "referenceImageUrl",
				Message: fmt.Sprintf("reference image must be in the reference bucket (%s), found bucket (%s) instead. URL format should be: s3://%s/%s",
					referenceBucket, bucket, referenceBucket, key),
			}
		}
		
		// Validate image format (Bedrock only supports JPEG and PNG)
		if !hasValidImageExtension(ctx.ReferenceImageURL) {
			return &promptutils.ValidationError{
				Field:   "referenceImageUrl",
				Message: "image format must be JPEG or PNG (supported by Bedrock)",
			}
		}
	}
	
	// Validate checking image URL if present
	if ctx.CheckingImageURL != "" {
		if !isValidS3URL(ctx.CheckingImageURL) {
			return &promptutils.ValidationError{
				Field:   "checkingImageUrl",
				Message: "invalid S3 URL format for checking image",
			}
		}
		
		// Extract bucket from S3 URL and verify it points to checking bucket
		bucket, key, err := extractS3BucketAndKey(ctx.CheckingImageURL)
		if err != nil {
			return &promptutils.ValidationError{
				Field:   "checkingImageUrl",
				Message: fmt.Sprintf("failed to parse S3 URL: %v", err),
			}
		}

		if bucket != checkingBucket {
			return &promptutils.ValidationError{
				Field:   "checkingImageUrl",
				Message: fmt.Sprintf("checking image must be in the checking bucket (%s), found bucket (%s) instead. URL format should be: s3://%s/%s",
					checkingBucket, bucket, checkingBucket, key),
			}
		}
		
		// Validate image format (Bedrock only supports JPEG and PNG)
		if !hasValidImageExtension(ctx.CheckingImageURL) {
			return &promptutils.ValidationError{
				Field:   "checkingImageUrl",
				Message: "image format must be JPEG or PNG (supported by Bedrock)",
			}
		}
	}
	
	return nil
}

// validatePreviousVsCurrent performs validation specific to PREVIOUS_VS_CURRENT type
func validatePreviousVsCurrent(input *promptutils.Input) error {
	ctx := input.VerificationContext
	
	// LayoutID should not be set
	if ctx.LayoutID != 0 {
		return &promptutils.ValidationError{
			Field:   "layoutId",
			Message: "layout ID should not be set for PREVIOUS_VS_CURRENT",
		}
	}
	
	// LayoutPrefix should not be set
	if ctx.LayoutPrefix != "" {
		return &promptutils.ValidationError{
			Field:   "layoutPrefix",
			Message: "layout prefix should not be set for PREVIOUS_VS_CURRENT",
		}
	}
	
	// Get bucket names from environment variables
	checkingBucket := os.Getenv(ENV_CHECKING_BUCKET)
	
	if checkingBucket == "" {
		return &promptutils.ValidationError{
			Field:   ENV_CHECKING_BUCKET,
			Message: "checking bucket environment variable is not set",
		}
	}
	
	// Reference image URL is required and should point to checking bucket
	if ctx.ReferenceImageURL == "" {
		return &promptutils.ValidationError{
			Field:   "referenceImageUrl",
			Message: "reference image URL is required for PREVIOUS_VS_CURRENT",
		}
	}
	
	if !isValidS3URL(ctx.ReferenceImageURL) {
		return &promptutils.ValidationError{
			Field:   "referenceImageUrl",
			Message: "invalid S3 URL format for reference image",
		}
	}
	
	// Extract bucket from S3 URL and verify it points to checking bucket
	bucket, key, err := extractS3BucketAndKey(ctx.ReferenceImageURL)
	if err != nil {
		return &promptutils.ValidationError{
			Field:   "referenceImageUrl",
			Message: fmt.Sprintf("failed to parse S3 URL: %v", err),
		}
	}

	if bucket != checkingBucket {
		return &promptutils.ValidationError{
			Field:   "referenceImageUrl",
			Message: fmt.Sprintf("for PREVIOUS_VS_CURRENT, reference image must be in the checking bucket (%s), found bucket (%s) instead. URL format should be: s3://%s/%s",
				checkingBucket, bucket, checkingBucket, key),
		}
	}
	
	// Validate image format (Bedrock only supports JPEG and PNG)
	if !hasValidImageExtension(ctx.ReferenceImageURL) {
		return &promptutils.ValidationError{
			Field:   "referenceImageUrl",
			Message: "image format must be JPEG or PNG (supported by Bedrock)",
		}
	}
	
	// Validate checking image URL
	if ctx.CheckingImageURL == "" {
		return &promptutils.ValidationError{
			Field:   "checkingImageUrl",
			Message: "checking image URL is required",
		}
	}
	
	if !isValidS3URL(ctx.CheckingImageURL) {
		return &promptutils.ValidationError{
			Field:   "checkingImageUrl",
			Message: "invalid S3 URL format for checking image",
		}
	}
	
	// Extract bucket from S3 URL and verify it points to checking bucket
	bucket, key, err = extractS3BucketAndKey(ctx.CheckingImageURL)
	if err != nil {
		return &promptutils.ValidationError{
			Field:   "checkingImageUrl",
			Message: fmt.Sprintf("failed to parse S3 URL: %v", err),
		}
	}

	if bucket != checkingBucket {
		return &promptutils.ValidationError{
			Field:   "checkingImageUrl",
			Message: fmt.Sprintf("checking image must be in the checking bucket (%s), found bucket (%s) instead. URL format should be: s3://%s/%s",
				checkingBucket, bucket, checkingBucket, key),
		}
	}
	
	// Validate image format (Bedrock only supports JPEG and PNG)
	if !hasValidImageExtension(ctx.CheckingImageURL) {
		return &promptutils.ValidationError{
			Field:   "checkingImageUrl",
			Message: "image format must be JPEG or PNG (supported by Bedrock)",
		}
	}
	
	// Validate historical context if present
	if input.HistoricalContext != nil {
		hCtx := input.HistoricalContext
		
		// Validate previous verification ID
		if hCtx.PreviousVerificationID == "" {
			return &promptutils.ValidationError{
				Field:   "historicalContext.previousVerificationId",
				Message: "previous verification ID is required when historical context is provided",
			}
		}
		
		// Validate previous verification timestamp
		if hCtx.PreviousVerificationAt != "" && !isValidISO8601(hCtx.PreviousVerificationAt) {
			return &promptutils.ValidationError{
				Field:   "historicalContext.previousVerificationAt",
				Message: "previous verification timestamp must be in ISO8601 format",
			}
		}
		
		// Validate machine structure if present
		if hCtx.MachineStructure != nil {
			if err := validateMachineStructure(hCtx.MachineStructure); err != nil {
				return err
			}
		}
	}
	
	return nil
}

// validateMachineStructure checks machine structure validity
func validateMachineStructure(ms *promptutils.MachineStructure) error {
	if ms == nil {
		return &promptutils.ValidationError{
			Field:   "machineStructure",
			Message: "machine structure is required",
		}
	}
	
	// Validate row count
	if ms.RowCount <= 0 {
		return &promptutils.ValidationError{
			Field:   "machineStructure.rowCount",
			Message: "row count must be positive",
		}
	}
	
	// Validate column count
	if ms.ColumnsPerRow <= 0 {
		return &promptutils.ValidationError{
			Field:   "machineStructure.columnsPerRow",
			Message: "columns per row must be positive",
		}
	}
	
	// Validate row order array
	if len(ms.RowOrder) == 0 {
		return &promptutils.ValidationError{
			Field:   "machineStructure.rowOrder",
			Message: "row order array cannot be empty",
		}
	}
	
	if len(ms.RowOrder) != ms.RowCount {
		return &promptutils.ValidationError{
			Field:   "machineStructure.rowOrder",
			Message: fmt.Sprintf("row order array length (%d) does not match row count (%d)", 
				len(ms.RowOrder), ms.RowCount),
		}
	}
	
	// Validate column order array
	if len(ms.ColumnOrder) == 0 {
		return &promptutils.ValidationError{
			Field:   "machineStructure.columnOrder",
			Message: "column order array cannot be empty",
		}
	}
	
	if len(ms.ColumnOrder) != ms.ColumnsPerRow {
		return &promptutils.ValidationError{
			Field:   "machineStructure.columnOrder",
			Message: fmt.Sprintf("column order array length (%d) does not match columns per row (%d)", 
				len(ms.ColumnOrder), ms.ColumnsPerRow),
		}
	}
	
	return nil
}

// hasValidImageExtension checks if an S3 URL has a valid image extension (JPEG or PNG only for Bedrock)
func hasValidImageExtension(url string) bool {
	ext := strings.ToLower(filepath.Ext(url))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png"
}

// isValidISO8601 checks if a string is a valid ISO8601 timestamp
func isValidISO8601(timestamp string) bool {
	// This is a simplified pattern that matches common ISO8601 formats
	pattern := `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(.\d+)?(Z|[+-]\d{2}:\d{2})?$`
	match, _ := regexp.MatchString(pattern, timestamp)
	return match
}

// isValidS3URL checks if a string is a valid S3 URL
func isValidS3URL(url string) bool {
	// This pattern matches s3://bucket-name/path/to/object format
	// Allow alphanumeric, dots, hyphens, spaces, underscores, and other common characters in the path
	pattern := `^s3://[\w.\-]+/[\w.\-/ _+%()]+$`
	match, _ := regexp.MatchString(pattern, url)
	return match
}

// extractS3BucketAndKey extracts bucket and key from an S3 URL
func extractS3BucketAndKey(s3URL string) (string, string, error) {
	// Clean URL first
	s3URL = strings.TrimSpace(s3URL)
	
	// Ensure s3:// prefix
	if !strings.HasPrefix(s3URL, "s3://") {
		return "", "", fmt.Errorf("not a valid S3 URL: %s", s3URL)
	}
	
	// Remove prefix
	s3Path := strings.TrimPrefix(s3URL, "s3://")
	
	// Split into bucket and key
	parts := strings.SplitN(s3Path, "/", 2)
	if len(parts) < 2 {
		return parts[0], "", nil // No key, just bucket
	}
	
	return parts[0], parts[1], nil
}

// isValidVendingMachineID checks if a string is a valid vending machine ID
func isValidVendingMachineID(id string) bool {
	// This pattern matches VM-XXXX format
	pattern := `^VM-\d+$`
	match, _ := regexp.MatchString(pattern, id)
	return match
}