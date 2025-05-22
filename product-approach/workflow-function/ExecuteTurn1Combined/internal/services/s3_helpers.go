// internal/services/s3_helpers.go - S3 UTILITY FUNCTIONS
package services

import (
	"fmt"
	"strings"
	"time"

	"workflow-function/ExecuteTurn1Combined/internal/models"
	"workflow-function/shared/errors"
	"workflow-function/shared/schema"
)

// S3Helpers provides utility functions for S3 operations
type S3Helpers struct{}

// NewS3Helpers creates S3 utility functions
func NewS3Helpers() *S3Helpers {
	return &S3Helpers{}
}

// GenerateS3Key generates standardized S3 keys
func (h *S3Helpers) GenerateS3Key(category, verificationID, filename string) string {
	return fmt.Sprintf("%s/%s/%s", category, verificationID, filename)
}

// GenerateTimestampedKey generates keys with timestamps
func (h *S3Helpers) GenerateTimestampedKey(category, verificationID, filename string) string {
	timestamp := time.Now().UTC().Format("20060102-150405")
	name := strings.TrimSuffix(filename, ".json")
	return fmt.Sprintf("%s/%s/%s-%s.json", category, verificationID, name, timestamp)
}

// ValidateS3Reference validates S3 reference structure
func (h *S3Helpers) ValidateS3Reference(ref models.S3Reference, operation string) error {
	if ref.Bucket == "" {
		return errors.NewValidationError(
			"S3 bucket cannot be empty",
			map[string]interface{}{
				"operation": operation,
				"key":       ref.Key,
			})
	}
	
	if ref.Key == "" {
		return errors.NewValidationError(
			"S3 key cannot be empty",
			map[string]interface{}{
				"operation": operation,
				"bucket":    ref.Bucket,
			})
	}
	
	return nil
}

// GetCategoryFromKey extracts category from S3 key
func (h *S3Helpers) GetCategoryFromKey(key string) string {
	parts := strings.Split(key, "/")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// GetVerificationIDFromKey extracts verification ID from S3 key
func (h *S3Helpers) GetVerificationIDFromKey(key string) string {
	parts := strings.Split(key, "/")
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

// GetFilenameFromKey extracts filename from S3 key
func (h *S3Helpers) GetFilenameFromKey(key string) string {
	parts := strings.Split(key, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

// CreateS3ReferenceCollection creates organized S3 reference collections
func (h *S3Helpers) CreateS3ReferenceCollection(verificationID string) *S3ReferenceCollection {
	return &S3ReferenceCollection{
		VerificationID: verificationID,
		References:     make(map[string]models.S3Reference),
		Categories:     make(map[string][]string),
	}
}

// S3ReferenceCollection organizes S3 references by category
type S3ReferenceCollection struct {
	VerificationID string                         `json:"verificationId"`
	References     map[string]models.S3Reference `json:"references"`
	Categories     map[string][]string           `json:"categories"`
	CreatedAt      string                        `json:"createdAt"`
	UpdatedAt      string                        `json:"updatedAt"`
}

// AddReference adds a reference to the collection
func (c *S3ReferenceCollection) AddReference(name string, ref models.S3Reference) {
	c.References[name] = ref
	
	category := c.getCategoryFromKey(ref.Key)
	if category != "" {
		if c.Categories[category] == nil {
			c.Categories[category] = make([]string, 0)
		}
		c.Categories[category] = append(c.Categories[category], name)
	}
	
	c.UpdatedAt = schema.FormatISO8601()
	if c.CreatedAt == "" {
		c.CreatedAt = c.UpdatedAt
	}
}

// GetReference gets a reference by name
func (c *S3ReferenceCollection) GetReference(name string) (models.S3Reference, bool) {
	ref, exists := c.References[name]
	return ref, exists
}

// GetReferencesByCategory gets all references in a category
func (c *S3ReferenceCollection) GetReferencesByCategory(category string) map[string]models.S3Reference {
	result := make(map[string]models.S3Reference)
	
	if names, exists := c.Categories[category]; exists {
		for _, name := range names {
			if ref, exists := c.References[name]; exists {
				result[name] = ref
			}
		}
	}
	
	return result
}

// ListCategories lists all categories with reference counts
func (c *S3ReferenceCollection) ListCategories() map[string]int {
	result := make(map[string]int)
	
	for category, names := range c.Categories {
		result[category] = len(names)
	}
	
	return result
}

// GetSummary returns a summary of the collection
func (c *S3ReferenceCollection) GetSummary() map[string]interface{} {
	return map[string]interface{}{
		"verification_id":    c.VerificationID,
		"total_references":   len(c.References),
		"categories":         c.ListCategories(),
		"created_at":         c.CreatedAt,
		"updated_at":         c.UpdatedAt,
	}
}

// getCategoryFromKey extracts category from S3 key
func (c *S3ReferenceCollection) getCategoryFromKey(key string) string {
	parts := strings.Split(key, "/")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// S3KeyBuilder helps build standardized S3 keys
type S3KeyBuilder struct {
	category       string
	verificationID string
	filename       string
	timestamp      bool
}

// NewS3KeyBuilder creates a new S3 key builder
func NewS3KeyBuilder() *S3KeyBuilder {
	return &S3KeyBuilder{}
}

// WithCategory sets the category
func (b *S3KeyBuilder) WithCategory(category string) *S3KeyBuilder {
	b.category = category
	return b
}

// WithVerificationID sets the verification ID
func (b *S3KeyBuilder) WithVerificationID(verificationID string) *S3KeyBuilder {
	b.verificationID = verificationID
	return b
}

// WithFilename sets the filename
func (b *S3KeyBuilder) WithFilename(filename string) *S3KeyBuilder {
	b.filename = filename
	return b
}

// WithTimestamp adds timestamp to the key
func (b *S3KeyBuilder) WithTimestamp(enabled bool) *S3KeyBuilder {
	b.timestamp = enabled
	return b
}

// Build creates the S3 key
func (b *S3KeyBuilder) Build() string {
	if b.category == "" || b.verificationID == "" || b.filename == "" {
		return ""
	}
	
	filename := b.filename
	if b.timestamp {
		timestamp := time.Now().UTC().Format("20060102-150405")
		name := strings.TrimSuffix(b.filename, ".json")
		filename = fmt.Sprintf("%s-%s.json", name, timestamp)
	}
	
	return fmt.Sprintf("%s/%s/%s", b.category, b.verificationID, filename)
}