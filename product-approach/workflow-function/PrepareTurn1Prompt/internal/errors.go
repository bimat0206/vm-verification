package internal

import (
   "fmt"
)

// ImageProcessingError represents errors that occur during image processing
type ImageProcessingError struct {
   Type    string
   Message string
   Details map[string]interface{}
}

// Error implements the error interface
func (e ImageProcessingError) Error() string {
   if len(e.Details) > 0 {
   	return fmt.Sprintf("image processing error [%s]: %s (details: %v)", e.Type, e.Message, e.Details)
   }
   return fmt.Sprintf("image processing error [%s]: %s", e.Type, e.Message)
}

// NewImageProcessingError creates a new image processing error
func NewImageProcessingError(errorType, message string, details map[string]interface{}) *ImageProcessingError {
   return &ImageProcessingError{
   	Type:    errorType,
   	Message: message,
   	Details: details,
   }
}

// StorageMethodError represents errors related to storage method handling
type StorageMethodError struct {
   Method   string
   Expected string
   Actual   string
   Details  map[string]interface{}
}

// Error implements the error interface
func (e StorageMethodError) Error() string {
   if e.Expected != "" {
   	return fmt.Sprintf("storage method error: expected '%s', got '%s' for method '%s'", e.Expected, e.Actual, e.Method)
   }
   return fmt.Sprintf("storage method error: unsupported method '%s'", e.Method)
}

// NewStorageMethodError creates a new storage method error
func NewStorageMethodError(method, expected, actual string, details map[string]interface{}) *StorageMethodError {
   return &StorageMethodError{
   	Method:   method,
   	Expected: expected,
   	Actual:   actual,
   	Details:  details,
   }
}

// ImageFormatError represents errors related to unsupported image formats
type ImageFormatError struct {
   Format    string
   Supported []string
   Details   map[string]interface{}
}

// Error implements the error interface
func (e ImageFormatError) Error() string {
   return fmt.Sprintf("unsupported image format '%s', supported formats: %v", e.Format, e.Supported)
}

// NewImageFormatError creates a new image format error
func NewImageFormatError(format string, supported []string, details map[string]interface{}) *ImageFormatError {
   return &ImageFormatError{
   	Format:    format,
   	Supported: supported,
   	Details:   details,
   }
}

// S3AccessError represents errors when accessing S3 objects
type S3AccessError struct {
   Bucket    string
   Key       string
   Operation string
   Reason    string
   Details   map[string]interface{}
}

// Error implements the error interface
func (e S3AccessError) Error() string {
   return fmt.Sprintf("S3 %s failed for s3://%s/%s: %s", e.Operation, e.Bucket, e.Key, e.Reason)
}

// NewS3AccessError creates a new S3 access error
func NewS3AccessError(bucket, key, operation, reason string, details map[string]interface{}) *S3AccessError {
   return &S3AccessError{
   	Bucket:    bucket,
   	Key:       key,
   	Operation: operation,
   	Reason:    reason,
   	Details:   details,
   }
}

// Base64Error represents errors related to Base64 encoding/decoding
type Base64Error struct {
   Operation string
   Reason    string
   Details   map[string]interface{}
}

// Error implements the error interface
func (e Base64Error) Error() string {
   return fmt.Sprintf("Base64 %s error: %s", e.Operation, e.Reason)
}

// NewBase64Error creates a new Base64 error
func NewBase64Error(operation, reason string, details map[string]interface{}) *Base64Error {
   return &Base64Error{
   	Operation: operation,
   	Reason:    reason,
   	Details:   details,
   }
}

// TemplateError represents errors during template processing
type TemplateError struct {
   TemplateName string
   Operation    string
   Reason       string
   Details      map[string]interface{}
}

// Error implements the error interface
func (e TemplateError) Error() string {
   return fmt.Sprintf("template %s error for '%s': %s", e.Operation, e.TemplateName, e.Reason)
}

// NewTemplateError creates a new template error
func NewTemplateError(templateName, operation, reason string, details map[string]interface{}) *TemplateError {
   return &TemplateError{
   	TemplateName: templateName,
   	Operation:    operation,
   	Reason:       reason,
   	Details:      details,
   }
}

// BedrockMessageError represents errors when creating Bedrock messages
type BedrockMessageError struct {
   Component string
   Reason    string
   Details   map[string]interface{}
}

// Error implements the error interface
func (e BedrockMessageError) Error() string {
   return fmt.Sprintf("Bedrock message error (%s): %s", e.Component, e.Reason)
}

// NewBedrockMessageError creates a new Bedrock message error
func NewBedrockMessageError(component, reason string, details map[string]interface{}) *BedrockMessageError {
   return &BedrockMessageError{
   	Component: component,
   	Reason:    reason,
   	Details:   details,
   }
}

// ConfigurationError represents errors related to missing or invalid configuration
type ConfigurationError struct {
   Component string
   Setting   string
   Reason    string
   Details   map[string]interface{}
}

// Error implements the error interface
func (e ConfigurationError) Error() string {
   return fmt.Sprintf("configuration error [%s.%s]: %s", e.Component, e.Setting, e.Reason)
}

// NewConfigurationError creates a new configuration error
func NewConfigurationError(component, setting, reason string, details map[string]interface{}) *ConfigurationError {
   return &ConfigurationError{
   	Component: component,
   	Setting:   setting,
   	Reason:    reason,
   	Details:   details,
   }
}

// Helper functions for common error scenarios

// WrapError wraps an existing error with additional context
func WrapError(err error, context string) error {
   if err == nil {
   	return nil
   }
   return fmt.Errorf("%s: %w", context, err)
}

// IsImageProcessingError checks if an error is an ImageProcessingError
func IsImageProcessingError(err error) bool {
   _, ok := err.(*ImageProcessingError)
   return ok
}

// IsStorageMethodError checks if an error is a StorageMethodError
func IsStorageMethodError(err error) bool {
   _, ok := err.(*StorageMethodError)
   return ok
}

// IsImageFormatError checks if an error is an ImageFormatError
func IsImageFormatError(err error) bool {
   _, ok := err.(*ImageFormatError)
   return ok
}

// IsS3AccessError checks if an error is an S3AccessError
func IsS3AccessError(err error) bool {
   _, ok := err.(*S3AccessError)
   return ok
}

// IsBase64Error checks if an error is a Base64Error
func IsBase64Error(err error) bool {
   _, ok := err.(*Base64Error)
   return ok
}

// IsTemplateError checks if an error is a TemplateError
func IsTemplateError(err error) bool {
   _, ok := err.(*TemplateError)
   return ok
}

// IsBedrockMessageError checks if an error is a BedrockMessageError
func IsBedrockMessageError(err error) bool {
   _, ok := err.(*BedrockMessageError)
   return ok
}

// IsConfigurationError checks if an error is a ConfigurationError
func IsConfigurationError(err error) bool {
   _, ok := err.(*ConfigurationError)
   return ok
}

// Common error constructors for frequently used scenarios

// NewImageTooLargeError creates an error for images that exceed size limits
func NewImageTooLargeError(sizeMB, limitMB int) *ImageProcessingError {
   return NewImageProcessingError("size_limit", 
   	fmt.Sprintf("Image size %dMB exceeds limit of %dMB", sizeMB, limitMB),
   	map[string]interface{}{
   		"sizeMB":  sizeMB,
   		"limitMB": limitMB,
   	})
}

// NewUnsupportedImageFormatError creates an error for unsupported image formats
func NewUnsupportedImageFormatError(format string) *ImageFormatError {
   return NewImageFormatError(format, []string{"jpeg", "png"}, nil)
}

// NewInvalidBase64Error creates an error for invalid Base64 data
func NewInvalidBase64Error(reason string) *Base64Error {
   return NewBase64Error("decode", reason, nil)
}

// NewS3NotFoundError creates an error for S3 objects that don't exist
func NewS3NotFoundError(bucket, key string) *S3AccessError {
   return NewS3AccessError(bucket, key, "GetObject", "object not found", nil)
}

// NewMissingConfigError creates an error for missing configuration
func NewMissingConfigError(component, setting string) *ConfigurationError {
   return NewConfigurationError(component, setting, "required configuration not provided", nil)
}

// NewInvalidConfigError creates an error for invalid configuration values
func NewInvalidConfigError(component, setting, value, expectedFormat string) *ConfigurationError {
   return NewConfigurationError(component, setting, 
   	fmt.Sprintf("invalid value '%s', expected format: %s", value, expectedFormat),
   	map[string]interface{}{
   		"value":          value,
   		"expectedFormat": expectedFormat,
   	})
}