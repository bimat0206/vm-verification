package config

import (
	"os"
)

// Config holds basic configuration for the Lambda
type Config struct {
	AWS struct {
		Region   string
		S3Bucket string
	}
	Logging struct {
		Level string
	}
}

// LoadConfiguration loads configuration from environment variables
func LoadConfiguration() Config {
	var cfg Config
	cfg.AWS.Region = getEnv("AWS_REGION", "us-east-1")
	cfg.AWS.S3Bucket = getEnv("STATE_BUCKET", "")
	cfg.Logging.Level = getEnv("LOG_LEVEL", "INFO")
	return cfg
}

func getEnv(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}
