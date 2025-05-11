package main

import (
    "os"
)

// Config holds configuration values for the Lambda.
type Config struct {
    LayoutTableName         string
    VerificationTableName   string
    // Add other config as needed
}

// LoadConfig loads configuration from environment variables.
func LoadConfig() Config {
    return Config{
        LayoutTableName:       getEnv("DYNAMODB_LAYOUT_TABLE", "LayoutMetadata"),
        VerificationTableName: getEnv("DYNAMODB_VERIFICATION_TABLE", "VerificationResults"),
    }
}

func getEnv(key, defaultVal string) string {
    if val := os.Getenv(key); val != "" {
        return val
    }
    return defaultVal
}
