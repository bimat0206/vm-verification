package internal

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
	
	wflogger "workflow-function/shared/logger"
)

// LogLevel represents the logging level
type LogLevel string

const (
	// Log levels
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
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
		Component:       "PrepareTurn1Prompt",
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
	return "PrepareTurn1Prompt"
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


// GeneratePromptID generates a unique ID for a prompt
func GeneratePromptID(verificationID string, turnNumber int) string {
	timestamp := time.Now().UTC().Format("20060102-150405")
	return fmt.Sprintf("prompt-%s-turn%d-%s", verificationID, turnNumber, timestamp)
}

// FormatTimestamp formats a time.Time as an ISO 8601 string
func FormatTimestamp(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

// ExtractS3BucketAndKey extracts the bucket name and key from an S3 URL
func ExtractS3BucketAndKey(s3URL string) (string, string, error) {
	// Clean URL first
	s3URL = CleanS3URL(s3URL)
	
	// Check prefix
	if !strings.HasPrefix(s3URL, "s3://") {
		return "", "", fmt.Errorf("invalid S3 URL format: %s", s3URL)
	}
	
	// Remove prefix
	s3Path := strings.TrimPrefix(s3URL, "s3://")
	
	// Split into bucket and key
	parts := strings.SplitN(s3Path, "/", 2)
	if len(parts) < 2 {
		return parts[0], "", fmt.Errorf("S3 URL has bucket but no key: %s", s3URL)
	}
	
	return parts[0], parts[1], nil
}

// GetFileExtension returns the file extension from a path
func GetFileExtension(path string) string {
	return strings.ToLower(filepath.Ext(path))
}

// IsValidImageFormat checks if a file extension is a valid image format for Bedrock
func IsValidImageFormat(ext string) bool {
	// Remove leading dot if present
	ext = strings.TrimPrefix(ext, ".")
	ext = strings.ToLower(ext)
	
	// Check against supported formats
	return ext == "jpg" || ext == "jpeg" || ext == "png"
}

// truncateString truncates a string to the given length with ellipsis
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

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

// GetRequestUUID generates a unique identifier for a request
func GetRequestUUID() string {
	return fmt.Sprintf("req-%d", time.Now().UnixNano())
}

// ReadLocalFile reads a file from the local file system
func ReadLocalFile(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

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

// IsJSONValid checks if a string is valid JSON
func IsJSONValid(s string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(s), &js) == nil
}

// GetDefaultTemplateBasePath returns the default template base path
func GetDefaultTemplateBasePath() string {
	return "/opt/templates"
}

// SafeJSON safely marshals a struct to a JSON string
func SafeJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("Error marshaling JSON: %v", err)
	}
	return string(data)
}

// PrettyJSON marshals a struct to an indented JSON string
func PrettyJSON(v interface{}) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error marshaling JSON: %v", err)
	}
	return string(data)
}

// HasValidPrefix checks if a string has a given prefix
func HasValidPrefix(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

// RecoverFromPanic provides a standardized way to recover from panics
// It logs the error with stack trace and returns an appropriate error
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