package processors

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
	"workflow-function/shared/s3state"
	
	"workflow-function/PrepareSystemPrompt/internal/config"
	"workflow-function/PrepareSystemPrompt/internal/models"
)

// ValidationProcessor handles input validation
type ValidationProcessor struct {
	config *config.Config
	logger logger.Logger
}

// NewValidationProcessor creates a new validation processor
func NewValidationProcessor(cfg *config.Config, log logger.Logger) *ValidationProcessor {
	return &ValidationProcessor{
		config: cfg,
		logger: log,
	}
}

// ValidateInput validates the input structure
func (p *ValidationProcessor) ValidateInput(input *models.Input) error {
	if input == nil {
		return fmt.Errorf("input is nil")
	}
	
	// For S3 reference inputs, validate the reference structure
	if input.Type == models.InputTypeS3Reference {
		if input.S3Envelope == nil {
			return fmt.Errorf("S3 envelope is nil for S3_REFERENCE input type")
		}
		
		if input.S3Envelope.VerificationID == "" {
			return fmt.Errorf("verification ID is required in S3 envelope")
		}
		
		if input.S3Envelope.References == nil || len(input.S3Envelope.References) == 0 {
			return fmt.Errorf("S3 references are required in envelope")
		}
		
		// Validate required references exist
		processingRef := findReferenceByCategory(input.S3Envelope.References, "processing")
		if processingRef == nil {
			return fmt.Errorf("processing reference not found in S3 envelope")
		}
		
		return nil
	}
	
	// For direct JSON inputs, validate the verification context
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
	switch input.VerificationContext.VerificationType {
	case schema.VerificationTypeLayoutVsChecking:
		return p.validateLayoutVsChecking(input)
	case schema.VerificationTypePreviousVsCurrent:
		return p.validatePreviousVsCurrent(input)
	default:
		return fmt.Errorf("unsupported verification type: %s", input.VerificationContext.VerificationType)
	}
}

// validateLayoutVsChecking validates LAYOUT_VS_CHECKING inputs
func (p *ValidationProcessor) validateLayoutVsChecking(input *models.Input) error {
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
	ms, err := models.ExtractMachineStructure(input.LayoutMetadata)
	if err != nil {
		return schema.ValidationError{Field: "layoutMetadata", Message: fmt.Sprintf("failed to extract machine structure: %v", err)}
	}
	
	if ms == nil {
		return schema.ValidationError{Field: "machineStructure", Message: "machine structure is required"}
	}
	
	if err := p.validateMachineStructure(ms); err != nil {
		return err
	}
	
	// Validate reference image URL if present
	if ctx.ReferenceImageUrl != "" {
		if err := p.validateImageURL(ctx.ReferenceImageUrl, p.config.ReferenceBucket); err != nil {
			return schema.ValidationError{Field: "referenceImageUrl", Message: err.Error()}
		}
	}
	
	// Validate checking image URL if present
	if ctx.CheckingImageUrl != "" {
		if err := p.validateImageURL(ctx.CheckingImageUrl, p.config.CheckingBucket); err != nil {
			return schema.ValidationError{Field: "checkingImageUrl", Message: err.Error()}
		}
	}
	
	return nil
}

// validatePreviousVsCurrent validates PREVIOUS_VS_CURRENT inputs
func (p *ValidationProcessor) validatePreviousVsCurrent(input *models.Input) error {
	ctx := input.VerificationContext
	
	// LayoutID should not be set
	if ctx.LayoutId != 0 {
		return schema.ValidationError{Field: "layoutId", Message: "layout ID should not be set for PREVIOUS_VS_CURRENT"}
	}
	
	// LayoutPrefix should not be set
	if ctx.LayoutPrefix != "" {
		return schema.ValidationError{Field: "layoutPrefix", Message: "layout prefix should not be set for PREVIOUS_VS_CURRENT"}
	}
	
	// Reference image URL is required and should point to checking bucket
	if ctx.ReferenceImageUrl == "" {
		return schema.ValidationError{Field: "referenceImageUrl", Message: "reference image URL is required for PREVIOUS_VS_CURRENT"}
	}
	
	if err := p.validateImageURL(ctx.ReferenceImageUrl, p.config.CheckingBucket); err != nil {
		return schema.ValidationError{Field: "referenceImageUrl", Message: err.Error()}
	}
	
	// Validate checking image URL
	if ctx.CheckingImageUrl == "" {
		return schema.ValidationError{Field: "checkingImageUrl", Message: "checking image URL is required"}
	}
	
	if err := p.validateImageURL(ctx.CheckingImageUrl, p.config.CheckingBucket); err != nil {
		return schema.ValidationError{Field: "checkingImageUrl", Message: err.Error()}
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
			msData, err := models.ExtractMachineStructure(map[string]interface{}{"machineStructure": msVal})
			if err == nil && msData != nil {
				if err := p.validateMachineStructure(msData); err != nil {
					return err
				}
			}
		}
	}
	
	return nil
}

// validateMachineStructure validates machine structure
func (p *ValidationProcessor) validateMachineStructure(ms *models.MachineStructure) error {
	if ms == nil {
		return schema.ValidationError{Field: "machineStructure", Message: "machine structure is required"}
	}
	
	// Validate row count
	if ms.RowCount <= 0 {
		return schema.ValidationError{Field: "machineStructure.rowCount", Message: "row count must be positive"}
	}
	
	// Validate column count
	if ms.ColumnsPerRow <= 0 {
		return schema.ValidationError{Field: "machineStructure.columnsPerRow", Message: "columns per row must be positive"}
	}
	
	// Validate row order array
	if len(ms.RowOrder) == 0 {
		return schema.ValidationError{Field: "machineStructure.rowOrder", Message: "row order array cannot be empty"}
	}
	
	if len(ms.RowOrder) != ms.RowCount {
		return schema.ValidationError{Field: "machineStructure.rowOrder",
			Message: fmt.Sprintf("row order array length (%d) does not match row count (%d)", 
				len(ms.RowOrder), ms.RowCount)}
	}
	
	// Validate column order array
	if len(ms.ColumnOrder) == 0 {
		return schema.ValidationError{Field: "machineStructure.columnOrder", Message: "column order array cannot be empty"}
	}
	
	if len(ms.ColumnOrder) != ms.ColumnsPerRow {
		return schema.ValidationError{Field: "machineStructure.columnOrder",
			Message: fmt.Sprintf("column order array length (%d) does not match columns per row (%d)", 
				len(ms.ColumnOrder), ms.ColumnsPerRow)}
	}
	
	return nil
}

// validateImageURL validates an S3 image URL
func (p *ValidationProcessor) validateImageURL(imageURL, expectedBucket string) error {
	// Simple validation - we don't have access to s3utils in this implementation
	// Check that the URL is an S3 URL
	if !isS3URL(imageURL) {
		return fmt.Errorf("invalid S3 URL format: %s", imageURL)
	}
	
	// Extract bucket and key
	bucket, key, err := parseS3URL(imageURL)
	if err != nil {
		return fmt.Errorf("failed to parse S3 URL: %w", err)
	}
	
	// Validate bucket
	if bucket != expectedBucket {
		return fmt.Errorf("image must be in the %s bucket, found %s instead", expectedBucket, bucket)
	}
	
	// Validate image format
	ext := filepath.Ext(key)
	if !isValidImageExtension(ext) {
		return fmt.Errorf("image format must be JPEG or PNG (supported by Bedrock)")
	}
	
	return nil
}

// isS3URL checks if a URL is an S3 URL
func isS3URL(url string) bool {
	return regexp.MustCompile(`^s3://[^/]+/.+$`).MatchString(url)
}

// parseS3URL extracts bucket and key from an S3 URL
func parseS3URL(url string) (string, string, error) {
	if !isS3URL(url) {
		return "", "", fmt.Errorf("not an S3 URL: %s", url)
	}
	
	// Remove s3:// prefix
	path := url[5:]
	
	// Split into bucket and key
	parts := regexp.MustCompile(`^([^/]+)/(.+)$`).FindStringSubmatch(path)
	if len(parts) != 3 {
		return "", "", fmt.Errorf("invalid S3 URL format: %s", url)
	}
	
	return parts[1], parts[2], nil
}

// isValidImageExtension checks if a file extension is a valid image format
func isValidImageExtension(ext string) bool {
	ext = filepath.Ext(ext)
	ext = regexp.MustCompile(`^\.`).ReplaceAllString(ext, "")
	ext = regexp.MustCompile(`\?.*$`).ReplaceAllString(ext, "")
	ext = regexp.MustCompile(`#.*$`).ReplaceAllString(ext, "")
	
	switch ext {
	case "jpg", "jpeg", "png":
		return true
	default:
		return false
	}
}

// isValidISO8601 checks if a string is a valid ISO8601 timestamp
func isValidISO8601(timestamp string) bool {
	// This is a simplified pattern that matches common ISO8601 formats
	pattern := `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(.\d+)?(Z|[+-]\d{2}:\d{2})?$`
	match, _ := regexp.MatchString(pattern, timestamp)
	return match
}

// findReferenceByCategory finds a reference by category in a map of references
func findReferenceByCategory(refs map[string]*s3state.Reference, category string) *s3state.Reference {
	for _, ref := range refs {
		if ref != nil && strings.Contains(ref.Key, category) {
			return ref
		}
	}
	return nil
}