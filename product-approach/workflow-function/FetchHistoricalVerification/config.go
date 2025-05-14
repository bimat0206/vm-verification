package main

import "os"

// Environment variable names
const (
	EnvRegion                = "AWS_REGION"
	EnvVerificationTableName = "DYNAMODB_VERIFICATION_TABLE"
	EnvCheckingBucketName    = "CHECKING_BUCKET"
	EnvLogLevel              = "LOG_LEVEL"
)

// getEnv is a wrapper around os.Getenv for consistency
func getEnv(key string) string {
	return os.Getenv(key)
}