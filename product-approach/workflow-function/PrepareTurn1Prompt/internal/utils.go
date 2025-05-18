package internal

import (
   "context"
   "encoding/base64"
   "encoding/json"
   "fmt"
   "io"
   "net/url"
   "os"
   "path/filepath"
   "runtime"
   "strconv"
   "strings"
   "time"

   "github.com/aws/aws-sdk-go-v2/config"
   "github.com/aws/aws-sdk-go-v2/service/s3"

   wflogger "workflow-function/shared/logger"
)

// AWS S3 Client Management

// CreateS3Client creates a new S3 client with default configuration
func CreateS3Client() (*s3.Client, error) {
   cfg, err := config.LoadDefaultConfig(context.TODO())
   if err != nil {
   	return nil, fmt.Errorf("failed to load AWS config: %w", err)
   }
   return s3.NewFromConfig(cfg), nil
}

// DownloadS3Object downloads an object from S3 and returns the data
func DownloadS3Object(bucket, key string) ([]byte, error) {
   client, err := CreateS3Client()
   if err != nil {
   	return nil, err
   }

   ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
   defer cancel()

   result, err := client.GetObject(ctx, &s3.GetObjectInput{
   	Bucket: &bucket,
   	Key:    &key,
   })
   if err != nil {
   	return nil, fmt.Errorf("failed to download object s3://%s/%s: %w", bucket, key, err)
   }
   defer result.Body.Close()

   data, err := io.ReadAll(result.Body)
   if err != nil {
   	return nil, fmt.Errorf("failed to read object body: %w", err)
   }

   return data, nil
}

// ValidateS3Access checks if an S3 object exists and is accessible
func ValidateS3Access(bucket, key string) error {
   client, err := CreateS3Client()
   if err != nil {
   	return err
   }

   ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
   defer cancel()

   _, err = client.HeadObject(ctx, &s3.HeadObjectInput{
   	Bucket: &bucket,
   	Key:    &key,
   })
   if err != nil {
   	return fmt.Errorf("object not accessible at s3://%s/%s: %w", bucket, key, err)
   }

   return nil
}

// Base64 Utilities

// EncodeToBase64 encodes binary data to Base64 string
func EncodeToBase64(data []byte) string {
   return base64.StdEncoding.EncodeToString(data)
}

// DecodeFromBase64 decodes Base64 string to binary data
func DecodeFromBase64(base64String string) ([]byte, error) {
   return base64.StdEncoding.DecodeString(base64String)
}

// IsValidBase64 checks if a string is valid Base64
func IsValidBase64(s string) bool {
   _, err := base64.StdEncoding.DecodeString(s)
   return err == nil
}

// GetBase64Size returns the decoded size of a Base64 string without decoding
func GetBase64Size(base64String string) int {
   // Base64 encoding adds ~33% overhead
   // More precise calculation considering padding
   return len(base64String) * 3 / 4
}

// Image Processing Utilities

// ValidateImageSize validates if image size is within limits
func ValidateImageSize(data []byte, maxSizeMB int) error {
   sizeMB := len(data) / (1024 * 1024)
   if sizeMB > maxSizeMB {
   	return fmt.Errorf("image size %dMB exceeds limit of %dMB", sizeMB, maxSizeMB)
   }
   return nil
}

// ExtractImageFormat detects image format from binary data
func ExtractImageFormat(data []byte) (string, error) {
   if len(data) < 4 {
   	return "", fmt.Errorf("insufficient data to determine image format")
   }

   // Check JPEG magic number
   if data[0] == 0xFF && data[1] == 0xD8 {
   	return "jpeg", nil
   }

   // Check PNG magic number
   if len(data) >= 8 && 
   	data[0] == 0x89 && data[1] == 0x50 && 
   	data[2] == 0x4E && data[3] == 0x47 {
   	return "png", nil
   }

   return "", fmt.Errorf("unsupported image format")
}

// NormalizeImageFormat normalizes image format string
func NormalizeImageFormat(format string) string {
   format = strings.ToLower(strings.TrimSpace(format))
   if format == "jpg" {
   	return "jpeg"
   }
   return format
}

// GetImageFormatFromPath extracts image format from file path or URL
func GetImageFormatFromPath(path string) string {
   ext := strings.ToLower(filepath.Ext(path))
   if ext == "" {
   	return ""
   }
   
   // Remove leading dot and normalize
   ext = ext[1:]
   
   switch ext {
   case "jpg", "jpeg":
   	return "jpeg"
   case "png":
   	return "png"
   default:
   	return ""
   }
}

// S3 URL Processing

// CleanS3URL cleans and normalizes an S3 URL
func CleanS3URL(s3URL string) string {
   // Trim whitespace
   s3URL = strings.TrimSpace(s3URL)
   
   // Ensure s3:// prefix
   if !strings.HasPrefix(s3URL, "s3://") {
   	if strings.HasPrefix(s3URL, "s3:/") {
   		s3URL = "s3://" + strings.TrimPrefix(s3URL, "s3:/")
   	} else {
   		s3URL = "s3://" + strings.TrimPrefix(s3URL, "s3:")
   	}
   }
   
   // Remove double slashes (except in protocol)
   parts := strings.SplitN(s3URL, "//", 2)
   if len(parts) == 2 {
   	protocol := parts[0] + "//"
   	path := strings.Replace(parts[1], "//", "/", -1)
   	s3URL = protocol + path
   }
   
   return s3URL
}

// ExtractS3BucketAndKey extracts the bucket name and key from an S3 URL
func ExtractS3BucketAndKey(s3URL string) (string, string, error) {
   // Clean URL first
   s3URL = CleanS3URL(s3URL)
   
   // Parse URL
   parsedURL, err := url.Parse(s3URL)
   if err != nil {
   	return "", "", fmt.Errorf("invalid S3 URL: %w", err)
   }
   
   if parsedURL.Scheme != "s3" {
   	return "", "", fmt.Errorf("not an S3 URL: %s", s3URL)
   }
   
   bucket := parsedURL.Host
   key := strings.TrimPrefix(parsedURL.Path, "/")
   
   if bucket == "" {
   	return "", "", fmt.Errorf("S3 URL missing bucket: %s", s3URL)
   }
   
   return bucket, key, nil
}

// IsValidS3URL checks if a string is a valid S3 URL
func IsValidS3URL(s3URL string) bool {
   _, _, err := ExtractS3BucketAndKey(s3URL)
   return err == nil
}

// BuildS3TempKey builds a temporary S3 key for storing Base64 data
func BuildS3TempKey(verificationId, imageType string) string {
   timestamp := time.Now().UTC().Format("20060102-150405")
   return fmt.Sprintf("temp-base64/%s/%s-%s.base64", verificationId, imageType, timestamp)
}

// Storage Method Detection

// DetectStorageMethod determines storage method from image info
func DetectStorageMethod(base64Data, s3TempBucket, s3TempKey string) string {
   if base64Data != "" {
   	return "inline"
   }
   if s3TempBucket != "" && s3TempKey != "" {
   	return "s3-temporary"
   }
   return "s3-direct"
}

// Environment Variable Utilities

// GetEnvWithDefault gets an environment variable or returns a default value
func GetEnvWithDefault(key, defaultValue string) string {
   value := os.Getenv(key)
   if value == "" {
   	return defaultValue
   }
   return value
}

// GetIntEnvWithDefault gets an integer environment variable or returns a default value
func GetIntEnvWithDefault(key string, defaultValue int) int {
   valueStr := os.Getenv(key)
   if valueStr == "" {
   	return defaultValue
   }
   
   value, err := strconv.Atoi(valueStr)
   if err != nil {
   	return defaultValue
   }
   return value
}

// GetBoolEnvWithDefault gets a boolean environment variable or returns a default value
func GetBoolEnvWithDefault(key string, defaultValue bool) bool {
   valueStr := os.Getenv(key)
   if valueStr == "" {
   	return defaultValue
   }
   
   value, err := strconv.ParseBool(valueStr)
   if err != nil {
   	return defaultValue
   }
   return value
}

// GetDurationEnvWithDefault gets a duration environment variable or returns a default value
func GetDurationEnvWithDefault(key string, defaultValue time.Duration) time.Duration {
   valueStr := os.Getenv(key)
   if valueStr == "" {
   	return defaultValue
   }
   
   // Try parsing as duration string first (e.g., "30s", "5m")
   if duration, err := time.ParseDuration(valueStr); err == nil {
   	return duration
   }
   
   // Try parsing as milliseconds
   if ms, err := strconv.Atoi(valueStr); err == nil {
   	return time.Duration(ms) * time.Millisecond
   }
   
   return defaultValue
}

// JSON Utilities

// IsJSONValid checks if a string is valid JSON
func IsJSONValid(s string) bool {
   var js json.RawMessage
   return json.Unmarshal([]byte(s), &js) == nil
}

// SafeJSONMarshal safely marshals a struct to a JSON string
func SafeJSONMarshal(v interface{}) string {
   data, err := json.Marshal(v)
   if err != nil {
   	return fmt.Sprintf("Error marshaling JSON: %v", err)
   }
   return string(data)
}

// PrettyJSONMarshal marshals a struct to an indented JSON string
func PrettyJSONMarshal(v interface{}) string {
   data, err := json.MarshalIndent(v, "", "  ")
   if err != nil {
   	return fmt.Sprintf("Error marshaling JSON: %v", err)
   }
   return string(data)
}

// ExtractFromJSON extracts a specific field from a JSON string
func ExtractFromJSON(jsonStr, path string) (string, error) {
   var data map[string]interface{}
   if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
   	return "", err
   }
   
   parts := strings.Split(path, ".")
   current := data
   
   for i, part := range parts {
   	if i == len(parts)-1 {
   		// Final part, extract value
   		if val, ok := current[part]; ok {
   			switch v := val.(type) {
   			case string:
   				return v, nil
   			case float64:
   				return fmt.Sprintf("%g", v), nil
   			case bool:
   				return fmt.Sprintf("%t", v), nil
   			default:
   				valJSON, err := json.Marshal(val)
   				if err != nil {
   					return "", err
   				}
   				return string(valJSON), nil
   			}
   		}
   		return "", fmt.Errorf("field %s not found", part)
   	}
   	
   	// Not final part, navigate deeper
   	if nextLevel, ok := current[part].(map[string]interface{}); ok {
   		current = nextLevel
   	} else {
   		return "", fmt.Errorf("field %s not found or not an object", part)
   	}
   }
   
   return "", fmt.Errorf("invalid path")
}

// ID and Timestamp Utilities

// GeneratePromptID generates a unique ID for a prompt
func GeneratePromptID(verificationID string, turnNumber int) string {
   timestamp := time.Now().UTC().Format("20060102-150405")
   return fmt.Sprintf("prompt-%s-turn%d-%s", verificationID, turnNumber, timestamp)
}

// FormatTimestamp formats a time.Time as an ISO 8601 string
func FormatTimestamp(t time.Time) string {
   return t.UTC().Format(time.RFC3339)
}

// IsValidISOTimestamp checks if a string is a valid ISO8601 timestamp
func IsValidISOTimestamp(timestamp string) bool {
   _, err := time.Parse(time.RFC3339, timestamp)
   return err == nil
}

// CalculateHoursBetween calculates the hours between two ISO8601 timestamps
func CalculateHoursBetween(start, end string) (float64, error) {
   startTime, err := time.Parse(time.RFC3339, start)
   if err != nil {
   	return 0, fmt.Errorf("invalid start timestamp: %w", err)
   }
   
   endTime, err := time.Parse(time.RFC3339, end)
   if err != nil {
   	return 0, fmt.Errorf("invalid end timestamp: %w", err)
   }
   
   duration := endTime.Sub(startTime)
   return duration.Hours(), nil
}

// String Utilities

// TruncateString truncates a string to the given length with ellipsis
func TruncateString(s string, maxLen int) string {
   if len(s) <= maxLen {
   	return s
   }
   if maxLen <= 3 {
   	return s[:maxLen]
   }
   return s[:maxLen-3] + "..."
}

// HasValidPrefix checks if a string has a given prefix
func HasValidPrefix(s, prefix string) bool {
   return strings.HasPrefix(s, prefix)
}

// SanitizeString removes potentially harmful characters from a string
func SanitizeString(s string) string {
   // Remove null bytes and other control characters
   s = strings.ReplaceAll(s, "\x00", "")
   
   // Trim whitespace
   return strings.TrimSpace(s)
}

// Error Recovery

// RecoverFromPanic provides a standardized way to recover from panics
func RecoverFromPanic(log wflogger.Logger, verificationId string) {
   if r := recover(); r != nil {
   	// Get stack trace
   	buf := make([]byte, 4096)
   	n := runtime.Stack(buf, false)
   	stackTrace := string(buf[:n])
   	
   	// Log the panic with stack trace
   	log.Error("Recovered from panic", map[string]interface{}{
   		"error":          fmt.Sprint(r),
   		"verificationId": verificationId,
   		"stackTrace":     stackTrace,
   	})
   }
}

// File Path Utilities

// GetFileExtension returns the file extension from a path
func GetFileExtension(path string) string {
   return strings.ToLower(filepath.Ext(path))
}

// IsValidImageExtension checks if a file extension is a valid image format
func IsValidImageExtension(ext string) bool {
   // Remove leading dot if present
   ext = strings.TrimPrefix(ext, ".")
   ext = strings.ToLower(ext)
   
   // Check against supported formats
   return ext == "jpg" || ext == "jpeg" || ext == "png"
}

// GetDefaultTemplateBasePath returns the default template base path
func GetDefaultTemplateBasePath() string {
   return "/opt/templates"
}

// Validation Utilities

// ValidateRequired checks if a value is not empty/nil
func ValidateRequired(value interface{}, fieldName string) error {
   if value == nil {
   	return fmt.Errorf("%s is required", fieldName)
   }
   
   switch v := value.(type) {
   case string:
   	if v == "" {
   		return fmt.Errorf("%s cannot be empty", fieldName)
   	}
   case []string:
   	if len(v) == 0 {
   		return fmt.Errorf("%s cannot be empty", fieldName)
   	}
   }
   
   return nil
}

// For ValidatePositiveInt, use the function from validator.go

// ValidateInRange checks if an integer is within a specified range
func ValidateInRange(value, min, max int, fieldName string) error {
   if value < min || value > max {
   	return fmt.Errorf("%s must be between %d and %d, got %d", fieldName, min, max, value)
   }
   return nil
}

// Array Utilities

// GetRowLabel gets the label for a specific row index
func GetRowLabel(rowOrder []string, index int) string {
   if index < 0 || index >= len(rowOrder) {
   	return ""
   }
   return rowOrder[index]
}

// GetTopAndBottomRows gets the top and bottom row labels from row order
func GetTopAndBottomRows(rowOrder []string) (string, string) {
   if len(rowOrder) == 0 {
   	return "", ""
   }
   return rowOrder[0], rowOrder[len(rowOrder)-1]
}

// ContainsString checks if a slice contains a specific string
func ContainsString(slice []string, item string) bool {
   for _, s := range slice {
   	if s == item {
   		return true
   	}
   }
   return false
}

// UniqueStrings returns a slice with duplicate strings removed
func UniqueStrings(slice []string) []string {
   seen := make(map[string]bool)
   result := make([]string, 0, len(slice))
   
   for _, s := range slice {
   	if !seen[s] {
   		seen[s] = true
   		result = append(result, s)
   	}
   }
   
   return result
}