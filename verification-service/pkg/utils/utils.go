package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// GenerateID generates a unique ID with the given prefix
func GenerateID(prefix string) string {
	// Generate random bytes
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		// Fallback to time-based ID if random generation fails
		return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
	}
	
	// Format as UUID-like string
	uuid := fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	
	// Return ID with prefix
	return fmt.Sprintf("%s-%s", prefix, uuid)
}

// IsValidS3URL checks if a string is a valid S3 URL
func IsValidS3URL(url string) bool {
	return strings.HasPrefix(url, "s3://") && strings.Contains(url[5:], "/")
}

// PrettyJSON converts an object to a formatted JSON string
func PrettyJSON(obj interface{}) (string, error) {
	jsonBytes, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// Base64Encode encodes a byte array to base64
func Base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// Base64Decode decodes a base64 string to a byte array
func Base64Decode(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}

// FormatTimestamp formats a time.Time as an ISO 8601 string
func FormatTimestamp(t time.Time) string {
	return t.Format(time.RFC3339)
}

// ParseTimestamp parses an ISO 8601 string to a time.Time
func ParseTimestamp(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

// TruncateString truncates a string to the given max length
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

// Contains checks if a string is in a slice of strings
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// MapOrDefault gets a value from a map or returns the default value
func MapOrDefault(m map[string]interface{}, key string, defaultValue interface{}) interface{} {
	if value, ok := m[key]; ok {
		return value
	}
	return defaultValue
}

// SafeFloat64ToInt converts a float64 to an int, defaulting to 0 if nil
func SafeFloat64ToInt(value interface{}) int {
	if v, ok := value.(float64); ok {
		return int(v)
	}
	return 0
}