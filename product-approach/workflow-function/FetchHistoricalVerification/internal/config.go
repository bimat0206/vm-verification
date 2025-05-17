package internal

import "os"

// Environment variable names
const (
	EnvRegion                = "AWS_REGION"
	EnvVerificationTableName = "DYNAMODB_VERIFICATION_TABLE"
	EnvCheckingBucketName    = "CHECKING_BUCKET"
	EnvLogLevel              = "LOG_LEVEL"
)

// GetEnv is a wrapper around os.Getenv for consistency
func GetEnv(key string) string {
	return os.Getenv(key)
}

// getEnv is an alias for GetEnv for backwards compatibility
func getEnv(key string) string {
	return GetEnv(key)
}

// GetEnvWithDefault returns environment variable value or default if not set
func GetEnvWithDefault(key, defaultValue string) string {
	value := GetEnv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvWithDefault is an alias for GetEnvWithDefault for backwards compatibility
func getEnvWithDefault(key, defaultValue string) string {
	return GetEnvWithDefault(key, defaultValue)
}