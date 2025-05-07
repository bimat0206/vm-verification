package main

import "os"

// Environment variable names
const (
	EnvRegion                = "AWS_REGION"
	EnvVerificationTableName = "DYNAMODB_VERIFICATION_TABLE"
	EnvCheckingBucketName    = "CHECKING_BUCKET"
	EnvLogLevel              = "LOG_LEVEL"
)

// getRegion returns the AWS region from environment variable
func getRegion() string {
	return os.Getenv(EnvRegion)
}

// getVerificationTableName returns the DynamoDB table name from environment variable
func getVerificationTableName() string {
	return os.Getenv(EnvVerificationTableName)
}

// getCheckingBucketName returns the S3 bucket name from environment variable
func getCheckingBucketName() string {
	return os.Getenv(EnvCheckingBucketName)
}

// getLogLevel returns the log level from environment variable
func getLogLevel() string {
	return os.Getenv(EnvLogLevel)
}