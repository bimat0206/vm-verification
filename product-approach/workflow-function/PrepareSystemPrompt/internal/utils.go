package internal

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"
)

// LogLevel represents the logging level
type LogLevel string

const (
	// Log levels
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
	
	// Default values
	DefaultTokenBudget = 16000
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp       string      `json:"timestamp"`
	Level           LogLevel    `json:"level"`
	Message         string      `json:"message"`
	VerificationID  string      `json:"verificationId,omitempty"`
	VerificationType string     `json:"verificationType,omitempty"`
	Component       string      `json:"component,omitempty"`
	Function        string      `json:"function,omitempty"`
	Details         interface{} `json:"details,omitempty"`
	Error           string      `json:"error,omitempty"`
}

// Logger provides structured logging functionality
func Logger(level LogLevel, message string, verificationID, verificationType string, details interface{}, err error) {
	// Skip debug logs unless DEBUG mode is enabled
	if level == LogLevelDebug && os.Getenv("DEBUG") != "true" {
		return
	}
	
	entry := LogEntry{
		Timestamp:       time.Now().UTC().Format(time.RFC3339),
		Level:           level,
		Message:         message,
		VerificationID:  verificationID,
		VerificationType: verificationType,
		Component:       "PrepareSystemPrompt",
		Function:        getFunctionName(2),
	}
	
	if details != nil {
		entry.Details = details
	}
	
	if err != nil {
		entry.Error = err.Error()
	}
	
	// Marshal to JSON for structured logging
	jsonEntry, jsonErr := json.Marshal(entry)
	if jsonErr != nil {
		log.Printf("Error marshaling log entry: %v", jsonErr)
		log.Printf("[%s] %s: %s", level, message, err)
		return
	}
	
	// Write to stdout for Lambda logging
	fmt.Println(string(jsonEntry))
}

// LogDebug logs a debug message
func LogDebug(message string, verificationID, verificationType string, details interface{}) {
	Logger(LogLevelDebug, message, verificationID, verificationType, details, nil)
}

// LogInfo logs an info message
func LogInfo(message string, verificationID, verificationType string, details interface{}) {
	Logger(LogLevelInfo, message, verificationID, verificationType, details, nil)
}

// LogWarn logs a warning message
func LogWarn(message string, verificationID, verificationType string, details interface{}, err error) {
	Logger(LogLevelWarn, message, verificationID, verificationType, details, err)
}

// LogError logs an error message
func LogError(message string, verificationID, verificationType string, details interface{}, err error) {
	Logger(LogLevelError, message, verificationID, verificationType, details, err)
}

// getFunctionName returns the name of the calling function
func getFunctionName(skip int) string {
	// This is a simplified version that just returns the component name
	// In a production environment, you would use runtime.Caller
	return "PrepareSystemPrompt"
}

// FormatArrayToString formats a string array to a comma-separated string
func FormatArrayToString(arr []string) string {
	if len(arr) == 0 {
		return ""
	}
	return strings.Join(arr, ", ")
}

// FormatProductMappings converts a product position map to a formatted array
func FormatProductMappings(positionMap map[string]ProductInfo) []ProductMapping {
	if positionMap == nil {
		return []ProductMapping{}
	}
	
	mappings := make([]ProductMapping, 0, len(positionMap))
	for position, info := range positionMap {
		mappings = append(mappings, ProductMapping{
			Position:    position,
			ProductID:   info.ProductID,
			ProductName: info.ProductName,
		})
	}
	
	return mappings
}

// ProcessTemplate renders a template with the provided data
func ProcessTemplate(tmpl *template.Template, data TemplateData) (string, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template execution failed: %w", err)
	}
	return buf.String(), nil
}

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

// ExtractS3BucketAndKey extracts bucket and key from an S3 URL
func ExtractS3BucketAndKey(s3URL string) (string, string, error) {
	// Clean URL first
	s3URL = CleanS3URL(s3URL)
	
	// Parse URL
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

// GetImageTypeFromS3Key extracts the image type from an S3 key
func GetImageTypeFromS3Key(key string) string {
	ext := strings.ToLower(filepath.Ext(key))
	if ext == "" {
		return "" // No extension found
	}
	
	ext = strings.TrimPrefix(ext, ".")
	
	switch ext {
	case "jpg", "jpeg":
		return "jpeg"
	case "png":
		return "png"
	default:
		return "" // Unsupported extension
	}
}

// IsValidImageType checks if an image type is supported by Bedrock
func IsValidImageType(imageType string) bool {
	imageType = strings.ToLower(imageType)
	return imageType == "jpeg" || imageType == "jpg" || imageType == "png"
}

// DecodeBase64Image decodes a base64-encoded image and returns image metadata
func DecodeBase64Image(data string) (image.Image, string, error) {
	// Remove data URL prefix if present
	if strings.Contains(data, ";base64,") {
		data = strings.SplitN(data, ";base64,", 2)[1]
	}
	
	// Decode base64
	imgData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode base64: %w", err)
	}
	
	// Detect image format
	imgFormat := ""
	if len(imgData) > 2 {
		if imgData[0] == 0xFF && imgData[1] == 0xD8 {
			imgFormat = "jpeg"
		} else if imgData[0] == 0x89 && imgData[1] == 0x50 {
			imgFormat = "png"
		} else {
			return nil, "", fmt.Errorf("unsupported image format (not JPEG or PNG)")
		}
	}
	
	// Decode image
	reader := bytes.NewReader(imgData)
	var img image.Image
	
	if imgFormat == "jpeg" {
		img, err = jpeg.Decode(reader)
		if err != nil {
			return nil, "", fmt.Errorf("failed to decode JPEG image: %w", err)
		}
	} else if imgFormat == "png" {
		img, err = png.Decode(reader)
		if err != nil {
			return nil, "", fmt.Errorf("failed to decode PNG image: %w", err)
		}
	} else {
		return nil, "", fmt.Errorf("unsupported image format (not JPEG or PNG)")
	}
	
	return img, imgFormat, nil
}

// EncodeImageToBase64 encodes an image to base64
func EncodeImageToBase64(img image.Image, format string) (string, error) {
	var buf bytes.Buffer
	
	// Encode based on format
	var err error
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	case "png":
		err = png.Encode(&buf, img)
	default:
		return "", fmt.Errorf("unsupported image format: %s (only JPEG or PNG are supported)", format)
	}
	
	if err != nil {
		return "", fmt.Errorf("failed to encode image: %w", err)
	}
	
	// Encode to base64
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// ValidateImageBytes checks if image bytes represent a valid JPEG or PNG image
func ValidateImageBytes(data []byte) (string, error) {
	if len(data) < 8 {
		return "", fmt.Errorf("image data too short")
	}
	
	// Check JPEG magic number
	if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return "jpeg", nil
	}
	
	// Check PNG magic number
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return "png", nil
	}
	
	return "", fmt.Errorf("unsupported image format (not JPEG or PNG)")
}

// FormatTimestamp formats a Unix timestamp as an ISO8601 string
func FormatTimestamp(timestamp int64) string {
	return time.Unix(timestamp, 0).UTC().Format(time.RFC3339)
}

// ParseTimestamp parses an ISO8601 string to a Unix timestamp
func ParseTimestamp(timestamp string) (int64, error) {
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return 0, fmt.Errorf("failed to parse timestamp: %w", err)
	}
	
	return t.Unix(), nil
}

// CalculateHoursBetween calculates hours between two ISO8601 timestamps
func CalculateHoursBetween(start, end string) (float64, error) {
	startTime, err := time.Parse(time.RFC3339, start)
	if err != nil {
		return 0, fmt.Errorf("failed to parse start time: %w", err)
	}
	
	endTime, err := time.Parse(time.RFC3339, end)
	if err != nil {
		return 0, fmt.Errorf("failed to parse end time: %w", err)
	}
	
	duration := endTime.Sub(startTime)
	return duration.Hours(), nil
}

// IsValidISO8601 checks if a string is a valid ISO8601 timestamp
func IsValidISO8601(timestamp string) bool {
	// This is a simplified pattern that matches common ISO8601 formats
	pattern := `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(.\d+)?(Z|[+-]\d{2}:\d{2})?$`
	match, _ := regexp.MatchString(pattern, timestamp)
	return match
}

// GeneratePromptID generates a unique prompt ID
func GeneratePromptID(verificationID string, promptType string) string {
	timestamp := time.Now().UTC().Format("20060102-150405")
	return fmt.Sprintf("prompt-%s-%s-%s", verificationID, promptType, timestamp)
}

// TruncateString truncates a string to a maximum length
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	
	return s[:maxLen-3] + "..."
}

// GetEnvWithFallback retrieves an environment variable with fallbacks
func GetEnvWithFallback(keys []string, defaultValue string) string {
	for _, key := range keys {
		value := os.Getenv(key)
		if value != "" {
			return value
		}
	}
	
	return defaultValue
}

// GetEnvBool retrieves a boolean environment variable
func GetEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	
	return boolValue
}

// GetEnv retrieves an environment variable with a default value
func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetIntEnv retrieves an integer environment variable with a default value
func GetIntEnv(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	
	return intValue
}

// GetFloatEnv retrieves a float environment variable with a default value
func GetFloatEnv(key string, defaultValue float64) float64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return defaultValue
	}
	
	return floatValue
}

// ReadLocalTemplate reads a template file from the local filesystem
func ReadLocalTemplate(templateBasePath, templateType string, version string) (string, error) {
	// Normalize templateType for file system access
	templateType = strings.ReplaceAll(strings.ToLower(templateType), "_", "-")
	
	// Try version-specific file first
	templatePath := filepath.Join(templateBasePath, templateType, fmt.Sprintf("v%s.tmpl", version))
	content, err := ioutil.ReadFile(templatePath)
	if err == nil {
		return string(content), nil
	}
	
	// Try flat file structure if version directory doesn't exist
	altTemplatePath := filepath.Join(templateBasePath, fmt.Sprintf("%s.tmpl", templateType))
	content, err = ioutil.ReadFile(altTemplatePath)
	if err == nil {
		return string(content), nil
	}
	
	// Try default template without version
	defaultTemplatePath := filepath.Join(templateBasePath, fmt.Sprintf("%s-default.tmpl", templateType))
	content, err = ioutil.ReadFile(defaultTemplatePath)
	if err == nil {
		return string(content), nil
	}
	
	return "", fmt.Errorf("template not found for type %s version %s", templateType, version)
}

// ListLocalTemplates lists all available local templates
func ListLocalTemplates(templateBasePath string) (map[string][]string, error) {
	// Check if the base directory exists
	if _, err := os.Stat(templateBasePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("template base directory does not exist: %s", templateBasePath)
	}
	
	result := make(map[string][]string)
	
	// Read the base directory entries
	entries, err := ioutil.ReadDir(templateBasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template directory: %v", err)
	}
	
	// Process each entry
	for _, entry := range entries {
		if entry.IsDir() {
			// If it's a directory, look for version files inside
			templateType := entry.Name()
			versionDir := filepath.Join(templateBasePath, templateType)
			versionFiles, err := ioutil.ReadDir(versionDir)
			if err != nil {
				continue
			}
			
			versions := []string{}
			for _, vFile := range versionFiles {
				if !vFile.IsDir() && strings.HasSuffix(vFile.Name(), ".tmpl") {
					// Extract version from filename (e.g., v1.0.0.tmpl -> 1.0.0)
					if strings.HasPrefix(vFile.Name(), "v") {
						version := strings.TrimSuffix(vFile.Name(), ".tmpl")
						version = strings.TrimPrefix(version, "v")
						versions = append(versions, version)
					}
				}
			}
			
			if len(versions) > 0 {
				result[templateType] = versions
			}
		} else if strings.HasSuffix(entry.Name(), ".tmpl") {
			// If it's a file, add it as a template without version info
			templateType := strings.TrimSuffix(entry.Name(), ".tmpl")
			result[templateType] = []string{"default"}
		}
	}
	
	return result, nil
}

// MapVerificationTypeToTemplateType converts a verification type to a template type
func MapVerificationTypeToTemplateType(verificationType string) string {
	// Convert to lowercase and replace underscores with hyphens
	return strings.ReplaceAll(strings.ToLower(verificationType), "_", "-")
}

// StripHTMLTags removes HTML tags from a string
func StripHTMLTags(input string) string {
	// Simple regex to remove HTML tags
	re := regexp.MustCompile("<[^>]*>")
	return re.ReplaceAllString(input, "")
}

// FormatMachineStructure formats a machine structure into a human-readable string
func FormatMachineStructure(ms *MachineStructure) string {
	if ms == nil {
		return "Unknown machine structure"
	}
	
	return fmt.Sprintf("%d rows (%s) × %d columns (%s)",
		ms.RowCount,
		FormatArrayToString(ms.RowOrder),
		ms.ColumnsPerRow,
		FormatArrayToString(ms.ColumnOrder))
}

// ConvertBase64ToImageData converts a base64 image to an image object
func ConvertBase64ToImageData(base64Data string) (image.Image, string, error) {
	// Remove data URL prefix if present
	if strings.Contains(base64Data, ";base64,") {
		base64Data = strings.SplitN(base64Data, ";base64,", 2)[1]
	}
	
	// Decode base64
	imgBytes, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode base64 image: %w", err)
	}
	
	// Validate image format (only JPEG and PNG are supported by Bedrock)
	format, err := ValidateImageBytes(imgBytes)
	if err != nil {
		return nil, "", err
	}
	
	// Decode image based on format
	reader := bytes.NewReader(imgBytes)
	var img image.Image
	
	if format == "jpeg" {
		img, err = jpeg.Decode(reader)
		if err != nil {
			return nil, "", fmt.Errorf("failed to decode JPEG image: %w", err)
		}
	} else if format == "png" {
		img, err = png.Decode(reader)
		if err != nil {
			return nil, "", fmt.Errorf("failed to decode PNG image: %w", err)
		}
	}
	
	return img, format, nil
}

// GetImageDimensions returns the dimensions of an image
func GetImageDimensions(img image.Image) (width, height int) {
	if img == nil {
		return 0, 0
	}
	
	bounds := img.Bounds()
	return bounds.Max.X - bounds.Min.X, bounds.Max.Y - bounds.Min.Y
}

// ValidateImageDimensions checks if image dimensions are within allowed limits
func ValidateImageDimensions(width, height int) error {
	// Bedrock has image dimension limits, but these are example values
	// Adjust these based on actual Bedrock limitations
	maxDimension := 4096
	minDimension := 64
	
	if width > maxDimension || height > maxDimension {
		return fmt.Errorf("image dimensions exceed maximum allowed (%dx%d > %dx%d)",
			width, height, maxDimension, maxDimension)
	}
	
	if width < minDimension || height < minDimension {
		return fmt.Errorf("image dimensions below minimum allowed (%dx%d < %dx%d)",
			width, height, minDimension, minDimension)
	}
	
	return nil
}

// IsJSON checks if a string is valid JSON
func IsJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

// WrapDynamoDBError wraps a DynamoDB error with more context
func WrapDynamoDBError(err error, operation, tableName string) error {
	if err == nil {
		return nil
	}
	
	return fmt.Errorf("DynamoDB %s operation on table %s failed: %w", 
		operation, tableName, err)
}

// WrapS3Error wraps an S3 error with more context
func WrapS3Error(err error, operation, bucket, key string) error {
	if err == nil {
		return nil
	}
	
	return fmt.Errorf("S3 %s operation on %s/%s failed: %w", 
		operation, bucket, key, err)
}

// FormatError creates a user-friendly error message
func FormatError(err error) string {
	if err == nil {
		return ""
	}
	
	// Clean up common AWS error messages
	msg := err.Error()
	msg = regexp.MustCompile(`\(Service: [^)]+\)`).ReplaceAllString(msg, "")
	msg = regexp.MustCompile(`\(Request ID: [^)]+\)`).ReplaceAllString(msg, "")
	msg = strings.TrimSpace(msg)
	
	return msg
}