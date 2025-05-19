package integration

import (
	"fmt"
	"os"
	"strconv"
	"time"
	"math/rand"
	"strings"
	"workflow-function/shared/schema"
)

// GeneratePromptID generates a unique ID for a prompt
func GeneratePromptID(verificationID string, turnNumber int) string {
	// Format: verification-id_turn-1_timestamp_random
	timestamp := time.Now().UTC().Format("20060102150405")
	random := rand.Intn(10000)
	return fmt.Sprintf("%s_turn-%d_%s_%04d", verificationID, turnNumber, timestamp, random)
}

// FormatTimestamp formats the current time as ISO8601
func FormatTimestamp(t time.Time) string {
	return t.Format(time.RFC3339)
}

// GetEnvWithDefault gets an environment variable with a default value
func GetEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetIntEnvWithDefault gets an integer environment variable with a default value
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

// GetBoolEnvWithDefault gets a boolean environment variable with a default value
func GetBoolEnvWithDefault(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	
	valueStr = strings.ToLower(valueStr)
	return valueStr == "true" || valueStr == "yes" || valueStr == "1"
}

// UpdateVerificationStatus updates the verification status in the workflow state
func UpdateVerificationStatus(state *schema.WorkflowState, status string) {
	if state != nil && state.VerificationContext != nil {
		state.VerificationContext.Status = status
	}
}

// FormatArrayToString formats a string array to a comma-separated string
func FormatArrayToString(arr []string) string {
	if len(arr) == 0 {
		return ""
	}
	return strings.Join(arr, ", ")
}

// RecoverFromPanic recovers from a panic and returns an error
func RecoverFromPanic() error {
	if r := recover(); r != nil {
		var err error
		switch x := r.(type) {
		case string:
			err = fmt.Errorf("panic: %s", x)
		case error:
			err = fmt.Errorf("panic: %w", x)
		default:
			err = fmt.Errorf("panic: unknown error")
		}
		return err
	}
	return nil
}