package internal

import (
   "fmt"
   "net/url"
   "path/filepath"
   "regexp"
   "strings"

   "workflow-function/shared/schema"
)

// ValidationError represents a structured validation error
type ValidationError struct {
   Field   string
   Message string
   Details map[string]interface{}
}

func (e ValidationError) Error() string {
   if len(e.Details) > 0 {
   	return fmt.Sprintf("validation error [%s]: %s (details: %v)", e.Field, e.Message, e.Details)
   }
   return fmt.Sprintf("validation error [%s]: %s", e.Field, e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string, details map[string]interface{}) *ValidationError {
   return &ValidationError{
   	Field:   field,
   	Message: message,
   	Details: details,
   }
}

// ValidateInputComplete performs comprehensive validation on the Lambda input
func ValidateInputComplete(input *Input) error {
   // Basic structure validation
   if err := validateBasicStructure(input); err != nil {
   	return err
   }
   
   // Turn-specific validation
   if err := validateTurnSpecific(input); err != nil {
   	return err
   }
   
   // Verification type validation
   if err := validateVerificationType(input.VerificationContext.VerificationType); err != nil {
   	return err
   }
   
   // Type-specific validation
   switch input.VerificationContext.VerificationType {
   case schema.VerificationTypeLayoutVsChecking:
   	return validateLayoutVsChecking(input)
   case schema.VerificationTypePreviousVsCurrent:
   	return validatePreviousVsCurrent(input)
   default:
   	return NewValidationError("verificationType", 
   		"unsupported verification type", 
   		map[string]interface{}{"type": input.VerificationContext.VerificationType})
   }
}

// validateBasicStructure validates the basic structure of the input
func validateBasicStructure(input *Input) error {
   if input == nil {
   	return NewValidationError("input", "input cannot be nil", nil)
   }
   
   if input.VerificationContext == nil {
   	return NewValidationError("verificationContext", "verification context is required", nil)
   }
   
   // Validate verification context fields
   if err := validateVerificationContext(input.VerificationContext); err != nil {
   	return err
   }
   
   return nil
}

// validateVerificationContext validates the verification context fields
func validateVerificationContext(ctx *VerificationContext) error {
   // Verification ID
   if ctx.VerificationID == "" {
   	return NewValidationError("verificationId", "verification ID is required", nil)
   }
   
   // Basic format check for verification ID
   if len(ctx.VerificationID) < 3 {
   	return NewValidationError("verificationId", "verification ID too short", 
   		map[string]interface{}{"length": len(ctx.VerificationID)})
   }
   
   // Verification type
   if ctx.VerificationType == "" {
   	return NewValidationError("verificationType", "verification type is required", nil)
   }
   
   // Reference image URL
   if ctx.ReferenceImageURL == "" {
   	return NewValidationError("referenceImageUrl", "reference image URL is required", nil)
   }
   
   if err := validateS3URL(ctx.ReferenceImageURL); err != nil {
   	return NewValidationError("referenceImageUrl", "invalid S3 URL format", 
   		map[string]interface{}{"url": ctx.ReferenceImageURL, "error": err.Error()})
   }
   
   // Optional field validation
   if ctx.VerificationAt != "" {
   	if err := validateISO8601Timestamp(ctx.VerificationAt); err != nil {
   		return NewValidationError("verificationAt", "invalid timestamp format", 
   			map[string]interface{}{"timestamp": ctx.VerificationAt})
   	}
   }
   
   if ctx.VendingMachineID != "" {
   	if err := validateVendingMachineID(ctx.VendingMachineID); err != nil {
   		return NewValidationError("vendingMachineId", "invalid vending machine ID format", 
   			map[string]interface{}{"id": ctx.VendingMachineID})
   	}
   }
   
   return nil
}

// validateTurnSpecific validates turn-specific requirements
func validateTurnSpecific(input *Input) error {
   // Turn number must be 1
   if input.TurnNumber != 1 {
   	return NewValidationError("turnNumber", "turn number must be 1 for PrepareTurn1Prompt", 
   		map[string]interface{}{"actual": input.TurnNumber, "expected": 1})
   }
   
   // Include image must be "reference"
   if input.IncludeImage != "reference" {
   	return NewValidationError("includeImage", "include image must be 'reference' for Turn 1", 
   		map[string]interface{}{"actual": input.IncludeImage, "expected": "reference"})
   }
   
   return nil
}

// validateVerificationType validates the verification type value
func validateVerificationType(verificationType string) error {
   validTypes := []string{
   	schema.VerificationTypeLayoutVsChecking,
   	schema.VerificationTypePreviousVsCurrent,
   }
   
   for _, validType := range validTypes {
   	if verificationType == validType {
   		return nil
   	}
   }
   
   return NewValidationError("verificationType", "invalid verification type", 
   	map[string]interface{}{
   		"provided": verificationType,
   		"valid":    validTypes,
   	})
}

// validateLayoutVsChecking validates layout vs checking specific requirements
func validateLayoutVsChecking(input *Input) error {
   ctx := input.VerificationContext
   
   // Layout ID is required
   if ctx.LayoutID <= 0 {
   	return NewValidationError("layoutId", "layout ID must be positive for LAYOUT_VS_CHECKING", 
   		map[string]interface{}{"actual": ctx.LayoutID})
   }
   
   // Layout prefix is required
   if ctx.LayoutPrefix == "" {
   	return NewValidationError("layoutPrefix", "layout prefix is required for LAYOUT_VS_CHECKING", nil)
   }
   
   // Validate layout prefix format (basic alphanumeric check)
   if !isValidLayoutPrefix(ctx.LayoutPrefix) {
   	return NewValidationError("layoutPrefix", "invalid layout prefix format", 
   		map[string]interface{}{"prefix": ctx.LayoutPrefix})
   }
   
   // Check reference image URL bucket
   if err := validateS3URLBucket(ctx.ReferenceImageURL, "REFERENCE_BUCKET"); err != nil {
   	return err
   }
   
   // Validate layout metadata if present
   if input.LayoutMetadata != nil {
   	if err := validateLayoutMetadata(input.LayoutMetadata); err != nil {
   		return err
   	}
   }
   
   return nil
}

// validatePreviousVsCurrent validates previous vs current specific requirements
func validatePreviousVsCurrent(input *Input) error {
   ctx := input.VerificationContext
   
   // Check reference image URL bucket (should be in checking bucket for this type)
   if err := validateS3URLBucket(ctx.ReferenceImageURL, "CHECKING_BUCKET"); err != nil {
   	return err
   }
   
   // Validate historical context if present
   if input.HistoricalContext != nil {
   	if err := validateHistoricalContext(input.HistoricalContext); err != nil {
   		return err
   	}
   }
   
   return nil
}

// validateLayoutMetadata validates layout metadata structure
func validateLayoutMetadata(metadata *LayoutMetadata) error {
   if metadata.MachineStructure != nil {
   	if err := validateMachineStructure(metadata.MachineStructure); err != nil {
   		return err
   	}
   }
   
   // Validate product position map if present
   if metadata.ProductPositionMap != nil {
   	if err := validateProductPositionMap(metadata.ProductPositionMap); err != nil {
   		return err
   	}
   }
   
   return nil
}

// validateMachineStructure validates machine structure fields
func validateMachineStructure(ms *MachineStructure) error {
   // Row count validation
   if ms.RowCount <= 0 {
   	return NewValidationError("machineStructure.rowCount", "row count must be positive", 
   		map[string]interface{}{"actual": ms.RowCount})
   }
   
   if ms.RowCount > 26 { // Reasonable upper limit
   	return NewValidationError("machineStructure.rowCount", "row count exceeds reasonable limit", 
   		map[string]interface{}{"actual": ms.RowCount, "limit": 26})
   }
   
   // Column count validation
   if ms.ColumnsPerRow <= 0 {
   	return NewValidationError("machineStructure.columnsPerRow", "columns per row must be positive", 
   		map[string]interface{}{"actual": ms.ColumnsPerRow})
   }
   
   if ms.ColumnsPerRow > 50 { // Reasonable upper limit
   	return NewValidationError("machineStructure.columnsPerRow", "columns per row exceeds reasonable limit", 
   		map[string]interface{}{"actual": ms.ColumnsPerRow, "limit": 50})
   }
   
   // Row order validation
   if len(ms.RowOrder) != ms.RowCount {
   	return NewValidationError("machineStructure.rowOrder", "row order length must match row count", 
   		map[string]interface{}{
   			"rowOrderLength": len(ms.RowOrder),
   			"rowCount":       ms.RowCount,
   		})
   }
   
   // Column order validation
   if len(ms.ColumnOrder) != ms.ColumnsPerRow {
   	return NewValidationError("machineStructure.columnOrder", "column order length must match columns per row", 
   		map[string]interface{}{
   			"columnOrderLength": len(ms.ColumnOrder),
   			"columnsPerRow":     ms.ColumnsPerRow,
   		})
   }
   
   // Check for duplicates in row order
   if hasDuplicates(ms.RowOrder) {
   	return NewValidationError("machineStructure.rowOrder", "row order contains duplicates", 
   		map[string]interface{}{"rowOrder": ms.RowOrder})
   }
   
   // Check for duplicates in column order
   if hasDuplicates(ms.ColumnOrder) {
   	return NewValidationError("machineStructure.columnOrder", "column order contains duplicates", 
   		map[string]interface{}{"columnOrder": ms.ColumnOrder})
   }
   
   return nil
}

// validateProductPositionMap validates product position map
func validateProductPositionMap(positionMap map[string]ProductInfo) error {
   for position, info := range positionMap {
   	// Validate position format (basic check)
   	if !isValidPosition(position) {
   		return NewValidationError("productPositionMap.position", "invalid position format", 
   			map[string]interface{}{"position": position})
   	}
   	
   	// Validate product info
   	if info.ProductID <= 0 {
   		return NewValidationError("productPositionMap.productId", "product ID must be positive", 
   			map[string]interface{}{"position": position, "productId": info.ProductID})
   	}
   	
   	if info.ProductName == "" {
   		return NewValidationError("productPositionMap.productName", "product name is required", 
   			map[string]interface{}{"position": position})
   	}
   }
   
   return nil
}

// validateHistoricalContext validates historical context fields
func validateHistoricalContext(hCtx *HistoricalContext) error {
   // Previous verification ID (required if historical context is provided)
   if hCtx.PreviousVerificationID == "" {
   	return NewValidationError("historicalContext.previousVerificationId", 
   		"previous verification ID is required when historical context is provided", nil)
   }
   
   // Previous verification timestamp validation
   if hCtx.PreviousVerificationAt != "" {
   	if err := validateISO8601Timestamp(hCtx.PreviousVerificationAt); err != nil {
   		return NewValidationError("historicalContext.previousVerificationAt", 
   			"invalid timestamp format", 
   			map[string]interface{}{"timestamp": hCtx.PreviousVerificationAt})
   	}
   }
   
   // Hours since last verification validation
   if hCtx.HoursSinceLastVerification < 0 {
   	return NewValidationError("historicalContext.hoursSinceLastVerification", 
   		"hours since last verification cannot be negative", 
   		map[string]interface{}{"hours": hCtx.HoursSinceLastVerification})
   }
   
   // Validate machine structure if present
   if hCtx.MachineStructure != nil {
   	if err := validateMachineStructure(hCtx.MachineStructure); err != nil {
   		return err
   	}
   }
   
   return nil
}

// validateS3URL validates S3 URL format
func validateS3URL(s3URL string) error {
   if s3URL == "" {
   	return fmt.Errorf("S3 URL cannot be empty")
   }
   
   // Parse URL
   parsedURL, err := url.Parse(s3URL)
   if err != nil {
   	return fmt.Errorf("invalid URL format: %w", err)
   }
   
   // Check scheme
   if parsedURL.Scheme != "s3" {
   	return fmt.Errorf("URL must use s3:// scheme, got: %s", parsedURL.Scheme)
   }
   
   // Check bucket name
   if parsedURL.Host == "" {
   	return fmt.Errorf("S3 URL missing bucket name")
   }
   
   // Check key (path)
   if parsedURL.Path == "" || parsedURL.Path == "/" {
   	return fmt.Errorf("S3 URL missing object key")
   }
   
   // Validate file extension for images
   ext := strings.ToLower(filepath.Ext(parsedURL.Path))
   if !isValidImageExtension(ext) {
   	return fmt.Errorf("unsupported image format: %s (supported: .jpg, .jpeg, .png)", ext)
   }
   
   return nil
}

// validateS3URLBucket validates that S3 URL uses the expected bucket
func validateS3URLBucket(s3URL, bucketEnvVar string) error {
   expectedBucket := GetEnvWithDefault(bucketEnvVar, "")
   if expectedBucket == "" {
   	return NewValidationError(bucketEnvVar, "environment variable not set", nil)
   }
   
   bucket, _, err := ExtractS3BucketAndKey(s3URL)
   if err != nil {
   	return NewValidationError("s3Url", "failed to parse S3 URL", 
   		map[string]interface{}{"url": s3URL, "error": err.Error()})
   }
   
   if bucket != expectedBucket {
   	return NewValidationError("s3Url", "S3 URL uses incorrect bucket", 
   		map[string]interface{}{
   			"expectedBucket": expectedBucket,
   			"actualBucket":   bucket,
   			"url":            s3URL,
   		})
   }
   
   return nil
}

// validateISO8601Timestamp validates ISO8601 timestamp format
func validateISO8601Timestamp(timestamp string) error {
   // RFC3339 is a subset of ISO8601 and is what Go uses
   pattern := `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(.\d+)?(Z|[+-]\d{2}:\d{2})$`
   matched, err := regexp.MatchString(pattern, timestamp)
   if err != nil {
   	return fmt.Errorf("regex error: %w", err)
   }
   
   if !matched {
   	return fmt.Errorf("timestamp must be in ISO8601/RFC3339 format (e.g., 2025-01-01T12:00:00Z)")
   }
   
   return nil
}

// validateVendingMachineID validates vending machine ID format
func validateVendingMachineID(id string) error {
   // Allow flexible format - just check it's not empty and reasonable length
   if len(id) < 2 || len(id) > 50 {
   	return fmt.Errorf("vending machine ID must be between 2 and 50 characters")
   }
   
   // Basic format check - alphanumeric with hyphens/underscores
   pattern := `^[a-zA-Z0-9\-_]+$`
   matched, err := regexp.MatchString(pattern, id)
   if err != nil {
   	return fmt.Errorf("regex error: %w", err)
   }
   
   if !matched {
   	return fmt.Errorf("vending machine ID contains invalid characters (allowed: letters, numbers, hyphens, underscores)")
   }
   
   return nil
}

// isValidLayoutPrefix validates layout prefix format
func isValidLayoutPrefix(prefix string) bool {
   if len(prefix) < 1 || len(prefix) > 20 {
   	return false
   }
   
   // Allow alphanumeric characters
   pattern := `^[a-zA-Z0-9]+$`
   matched, _ := regexp.MatchString(pattern, prefix)
   return matched
}

// isValidPosition validates position format (e.g., "A01", "B12")
func isValidPosition(position string) bool {
   if len(position) < 2 || len(position) > 5 {
   	return false
   }
   
   // Basic pattern: letter(s) followed by number(s)
   pattern := `^[A-Z]+[0-9]+$`
   matched, _ := regexp.MatchString(pattern, strings.ToUpper(position))
   return matched
}

// isValidImageExtension checks if file extension is valid for images
func isValidImageExtension(ext string) bool {
   ext = strings.ToLower(ext)
   validExtensions := []string{".jpg", ".jpeg", ".png"}
   
   for _, validExt := range validExtensions {
   	if ext == validExt {
   		return true
   	}
   }
   
   return false
}

// hasDuplicates checks if string slice has duplicate values
func hasDuplicates(slice []string) bool {
   seen := make(map[string]bool)
   for _, item := range slice {
   	if seen[item] {
   		return true
   	}
   	seen[item] = true
   }
   return false
}

// Helper validation functions that can be used independently

// ValidatePositiveInt validates that an integer is positive
func ValidatePositiveInt(value int, fieldName string) error {
   if value <= 0 {
   	return NewValidationError(fieldName, "must be positive", 
   		map[string]interface{}{"value": value})
   }
   return nil
}

// ValidateStringNotEmpty validates that a string is not empty
func ValidateStringNotEmpty(value, fieldName string) error {
   if value == "" {
   	return NewValidationError(fieldName, "cannot be empty", nil)
   }
   return nil
}

// ValidateStringLength validates string length is within bounds
func ValidateStringLength(value, fieldName string, minLen, maxLen int) error {
   if len(value) < minLen || len(value) > maxLen {
   	return NewValidationError(fieldName, "length out of bounds", 
   		map[string]interface{}{
   			"length": len(value),
   			"min":    minLen,
   			"max":    maxLen,
   		})
   }
   return nil
}

// ValidateSliceLength validates slice length is within bounds
func ValidateSliceLength(slice []string, fieldName string, expectedLen int) error {
   if len(slice) != expectedLen {
   	return NewValidationError(fieldName, "incorrect length", 
   		map[string]interface{}{
   			"actual":   len(slice),
   			"expected": expectedLen,
   		})
   }
   return nil
}