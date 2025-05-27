package s3state

import (
	"fmt"
	"strings"
)

// Category constants for organizing state data
const (
	CategoryImages     = "images"
	CategoryPrompts    = "prompts"
	CategoryResponses  = "responses"
	CategoryProcessing = "processing"
)

// AllCategories returns a slice of all valid categories
func AllCategories() []string {
	return []string{
		CategoryImages,
		CategoryPrompts,
		CategoryResponses,
		CategoryProcessing,
	}
}

// IsValidCategory checks if a category is valid
func IsValidCategory(category string) bool {
	switch category {
	case CategoryImages, CategoryPrompts, CategoryResponses, CategoryProcessing:
		return true
	default:
		return false
	}
}

// GenerateKey creates a standardized S3 key for a category
func GenerateKey(verificationID, category, filename string) string {
	return fmt.Sprintf("%s/%s/%s", verificationID, category, filename)
}

// ParseKey extracts verification ID, category, and filename from an S3 key
func ParseKey(key string) (verificationID, category, filename string, err error) {
	parts := strings.Split(key, "/")
	
	if len(parts) < 3 {
		return "", "", "", fmt.Errorf("invalid key format: %s", key)
	}
	
	verificationID = parts[0]
	category = parts[1]
	filename = strings.Join(parts[2:], "/") // Handle nested paths in filename
	
	if !IsValidCategory(category) {
		return "", "", "", fmt.Errorf("invalid category in key: %s", category)
	}
	
	return verificationID, category, filename, nil
}

// GetCategoryFromKey extracts just the category from an S3 key
func GetCategoryFromKey(key string) string {
	parts := strings.Split(key, "/")
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

// GetVerificationIDFromKey extracts just the verification ID from an S3 key
func GetVerificationIDFromKey(key string) string {
	parts := strings.Split(key, "/")
	if len(parts) >= 1 {
		return parts[0]
	}
	return ""
}

// GetFilenameFromKey extracts just the filename from an S3 key
func GetFilenameFromKey(key string) string {
	parts := strings.Split(key, "/")
	if len(parts) >= 3 {
		return strings.Join(parts[2:], "/")
	}
	return ""
}

// CategoryInfo holds basic information about a category
type CategoryInfo struct {
	Name        string
	Description string
	ContentType string
}

// GetCategoryInfo returns basic information about a category
func GetCategoryInfo(category string) (*CategoryInfo, error) {
	switch category {
	case CategoryImages:
		return &CategoryInfo{
			Name:        CategoryImages,
			Description: "Image data (Base64 encoded and metadata)",
			ContentType: "application/json", // Since we store Base64 as JSON
		}, nil
	case CategoryPrompts:
		return &CategoryInfo{
			Name:        CategoryPrompts,
			Description: "AI prompts and conversation data",
			ContentType: "application/json",
		}, nil
	case CategoryResponses:
		return &CategoryInfo{
			Name:        CategoryResponses,
			Description: "AI responses and conversation history",
			ContentType: "application/json",
		}, nil
	case CategoryProcessing:
		return &CategoryInfo{
			Name:        CategoryProcessing,
			Description: "Processed analysis and intermediate results",
			ContentType: "application/json",
		}, nil
	default:
		return nil, fmt.Errorf("unknown category: %s", category)
	}
}

// ValidateKeyStructure checks if an S3 key follows the expected structure
func ValidateKeyStructure(key string) error {
	verificationID, _, filename, err := ParseKey(key)
	if err != nil {
		return err
	}
	
	if verificationID == "" {
		return fmt.Errorf("verification ID cannot be empty")
	}
	
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}
	
	return nil
}

// BuildReferenceKey builds a reference key for the envelope
func BuildReferenceKey(category, filename string) string {
	// Remove file extension and path separators to create a clean reference key
	cleanFilename := strings.TrimSuffix(filename, ".json")
	cleanFilename = strings.ReplaceAll(cleanFilename, "/", "_")
	cleanFilename = strings.ReplaceAll(cleanFilename, "-", "_")
	
	return fmt.Sprintf("%s_%s", category, cleanFilename)
}

// Common filename patterns for each category
const (
	// Images category files
	ImageMetadataFile     = "metadata.json"
	ReferenceBase64File   = "reference-base64.base64"
	CheckingBase64File    = "checking-base64.base64"
	
	// Prompts category files
	SystemPromptFile      = "system-prompt.json"
	Turn1PromptFile       = "turn1-prompt.json"
	Turn2PromptFile       = "turn2-prompt.json"
	
	// Responses category files
	Turn1ResponseFile     = "turn1-raw-response.json"
	Turn2ResponseFile     = "turn2-raw-response.json"
	
	// Processing category files
	InitializationFile    = "initialization.json"
	LayoutMetadataFile    = "layout-metadata.json"
	HistoricalContextFile = "historical-context.json"
	Turn1AnalysisFile     = "turn1-processed-response.json"
	Turn2AnalysisFile     = "turn2-processed-response.json"
	FinalResultsFile      = "final-results.json"
)

// GetStandardFilename returns the standard filename for common file types
func GetStandardFilename(category, fileType string) (string, error) {
	switch category {
	case CategoryImages:
		switch fileType {
		case "metadata":
			return ImageMetadataFile, nil
		case "reference-base64":
			return ReferenceBase64File, nil
		case "checking-base64":
			return CheckingBase64File, nil
		}
	case CategoryPrompts:
		switch fileType {
		case "system":
			return SystemPromptFile, nil
		case "turn1":
			return Turn1PromptFile, nil
		case "turn2":
			return Turn2PromptFile, nil
		}
	case CategoryResponses:
		switch fileType {
		case "turn1":
			return Turn1ResponseFile, nil
		case "turn2":
			return Turn2ResponseFile, nil
		}
	case CategoryProcessing:
		switch fileType {
		case "initialization":
			return InitializationFile, nil
		case "layout-metadata":
			return LayoutMetadataFile, nil
		case "historical-context":
			return HistoricalContextFile, nil
		case "turn1-analysis":
			return Turn1AnalysisFile, nil
		case "turn2-analysis":
			return Turn2AnalysisFile, nil
		case "final-results":
			return FinalResultsFile, nil
		}
	}
	
	// Return error for unknown file type or category
	return "", fmt.Errorf("unknown file type '%s' for category '%s'", fileType, category)
}