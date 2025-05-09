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
        LayoutTableName:       getEnv("LAYOUT_TABLE_NAME", "LayoutMetadata"),
        VerificationTableName: getEnv("VERIFICATION_TABLE_NAME", "VerificationResults"),
    }
}

func getEnv(key, defaultVal string) string {
    if val := os.Getenv(key); val != "" {
        return val
    }
    return defaultVal
}
