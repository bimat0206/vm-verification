package internal

import (
   "context"
   "encoding/base64"
   "fmt"
   "io"
   "strings"
   "time"

   "github.com/aws/aws-sdk-go-v2/config"
   "github.com/aws/aws-sdk-go-v2/service/s3"

   "workflow-function/shared/errors"
)

// S3ImageProcessor handles S3 operations for image processing
type S3ImageProcessor struct {
   client         *s3.Client
   timeout        time.Duration
   maxImageSizeMB int
}

// NewS3ImageProcessor creates a new S3 image processor with default configuration
func NewS3ImageProcessor() (*S3ImageProcessor, error) {
   // Load AWS configuration
   cfg, err := config.LoadDefaultConfig(context.TODO())
   if err != nil {
   	return nil, fmt.Errorf("failed to load AWS config: %w", err)
   }

   // Get timeout from environment variable or use default (30 seconds)
   timeout := time.Duration(GetIntEnvWithDefault("BASE64_RETRIEVAL_TIMEOUT", 30000)) * time.Millisecond

   // Get max image size from environment variable or use default (10MB)
   maxImageSizeMB := GetIntEnvWithDefault("MAX_IMAGE_SIZE_MB", 10)

   return &S3ImageProcessor{
   	client:         s3.NewFromConfig(cfg),
   	timeout:        timeout,
   	maxImageSizeMB: maxImageSizeMB,
   }, nil
}

// RetrieveBase64FromTemp retrieves Base64 string from S3 temporary storage
func (p *S3ImageProcessor) RetrieveBase64FromTemp(bucket, key string) (string, error) {
   if bucket == "" || key == "" {
   	return "", errors.NewValidationError("Bucket and key are required", 
   		map[string]interface{}{"bucket": bucket, "key": key})
   }

   // Create context with timeout
   ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
   defer cancel()

   // Download object from S3
   result, err := p.client.GetObject(ctx, &s3.GetObjectInput{
   	Bucket: &bucket,
   	Key:    &key,
   })
   if err != nil {
   	return "", fmt.Errorf("failed to retrieve Base64 from s3://%s/%s: %w", bucket, key, err)
   }
   defer result.Body.Close()

   // Read the content as string (should already be Base64)
   data, err := io.ReadAll(result.Body)
   if err != nil {
   	return "", fmt.Errorf("failed to read Base64 content: %w", err)
   }

   base64String := strings.TrimSpace(string(data))

   // Validate it's proper Base64
   if !p.isValidBase64(base64String) {
   	return "", errors.NewValidationError("Retrieved content is not valid Base64", nil)
   }

   // Validate decoded size
   if err := p.validateBase64Size(base64String); err != nil {
   	return "", err
   }

   return base64String, nil
}

// DownloadAndEncode downloads an image from S3 and encodes it to Base64
func (p *S3ImageProcessor) DownloadAndEncode(bucket, key string) (string, error) {
   if bucket == "" || key == "" {
   	return "", errors.NewValidationError("Bucket and key are required", 
   		map[string]interface{}{"bucket": bucket, "key": key})
   }

   // Create context with timeout
   ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
   defer cancel()

   // Download object from S3
   result, err := p.client.GetObject(ctx, &s3.GetObjectInput{
   	Bucket: &bucket,
   	Key:    &key,
   })
   if err != nil {
   	return "", fmt.Errorf("failed to download image from s3://%s/%s: %w", bucket, key, err)
   }
   defer result.Body.Close()

   // Read the binary data
   imageData, err := io.ReadAll(result.Body)
   if err != nil {
   	return "", fmt.Errorf("failed to read image data: %w", err)
   }

   // Validate image size
   if err := p.validateImageSize(imageData); err != nil {
   	return "", err
   }

   // Validate image format
   if err := p.validateImageFormat(imageData); err != nil {
   	return "", err
   }

   // Encode to Base64
   base64String := base64.StdEncoding.EncodeToString(imageData)

   return base64String, nil
}

// ValidateImageAccess checks if an image exists and is accessible
func (p *S3ImageProcessor) ValidateImageAccess(bucket, key string) error {
   if bucket == "" || key == "" {
   	return errors.NewValidationError("Bucket and key are required", 
   		map[string]interface{}{"bucket": bucket, "key": key})
   }

   // Create context with shorter timeout for head operation
   ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
   defer cancel()

   // Check if object exists using HeadObject (lighter operation)
   _, err := p.client.HeadObject(ctx, &s3.HeadObjectInput{
   	Bucket: &bucket,
   	Key:    &key,
   })
   if err != nil {
   	return fmt.Errorf("image not accessible at s3://%s/%s: %w", bucket, key, err)
   }

   return nil
}

// Helper methods

// isValidBase64 checks if a string is valid Base64
func (p *S3ImageProcessor) isValidBase64(s string) bool {
   _, err := base64.StdEncoding.DecodeString(s)
   return err == nil
}

// validateBase64Size validates the size of Base64 encoded data
func (p *S3ImageProcessor) validateBase64Size(base64String string) error {
   // Decode to check actual size
   decoded, err := base64.StdEncoding.DecodeString(base64String)
   if err != nil {
   	return errors.NewValidationError("Invalid Base64 encoding", 
   		map[string]interface{}{"error": err.Error()})
   }

   // Check size limit
   sizeMB := len(decoded) / (1024 * 1024)
   if sizeMB > p.maxImageSizeMB {
   	return errors.NewValidationError("Image size exceeds limit", 
   		map[string]interface{}{
   			"sizeMB":    sizeMB,
   			"limitMB":   p.maxImageSizeMB,
   		})
   }

   return nil
}

// validateImageSize validates the size of raw image data
func (p *S3ImageProcessor) validateImageSize(imageData []byte) error {
   sizeMB := len(imageData) / (1024 * 1024)
   if sizeMB > p.maxImageSizeMB {
   	return errors.NewValidationError("Image size exceeds limit", 
   		map[string]interface{}{
   			"sizeMB":    sizeMB,
   			"limitMB":   p.maxImageSizeMB,
   		})
   }
   return nil
}

// validateImageFormat validates that the image format is supported by Bedrock
func (p *S3ImageProcessor) validateImageFormat(imageData []byte) error {
   format := p.detectImageFormat(imageData)
   if !p.isValidBedrockFormat(format) {
   	return errors.NewValidationError("Unsupported image format for Bedrock", 
   		map[string]interface{}{
   			"detectedFormat": format,
   			"supportedFormats": []string{"jpeg", "png"},
   		})
   }
   return nil
}

// detectImageFormat detects image format from binary data
func (p *S3ImageProcessor) detectImageFormat(data []byte) string {
   if len(data) < 4 {
   	return "unknown"
   }

   // Check JPEG magic number
   if data[0] == 0xFF && data[1] == 0xD8 {
   	return "jpeg"
   }

   // Check PNG magic number
   if len(data) >= 8 && 
   	data[0] == 0x89 && data[1] == 0x50 && 
   	data[2] == 0x4E && data[3] == 0x47 {
   	return "png"
   }

   return "unknown"
}

// isValidBedrockFormat checks if the format is supported by Bedrock
func (p *S3ImageProcessor) isValidBedrockFormat(format string) bool {
   format = strings.ToLower(format)
   return format == "jpeg" || format == "png"
}

// Static methods that delegate to processor instance

// For direct S3 operations, use the equivalent functions from utils.go:
// - DownloadS3Object
// - EncodeToBase64
// - ValidateS3Access
// - GetIntEnvWithDefault