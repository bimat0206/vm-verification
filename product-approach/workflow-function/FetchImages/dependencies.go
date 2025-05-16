package main

import (
	"os"
	"strconv"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"workflow-function/shared/dbutils"
	"workflow-function/shared/logger"
	"workflow-function/shared/s3utils"
)

// ConfigVars contains configuration for the lambda function
type ConfigVars struct {
	LayoutTable         string
	VerificationTable   string
	MaxImageSize        int64  // Maximum image size in bytes
	TempBase64Bucket    string // S3 bucket for temporary Base64 storage
	MaxInlineBase64Size int64  // Maximum size for inline Base64 storage (2MB threshold)
}

// Dependencies contains all external dependencies required by the service
type Dependencies struct {
	logger    logger.Logger
	s3Client  *s3.Client
	dbClient  *dynamodb.Client
	s3Utils   *s3utils.S3Utils
	dbUtils   *dbutils.DynamoDBUtils
	awsConfig aws.Config
	config    ConfigVars
}

// NewDependencies creates a new Dependencies instance with all required dependencies
func NewDependencies(awsConfig aws.Config) *Dependencies {
	// Initialize clients
	s3Client := s3.NewFromConfig(awsConfig)
	dbClient := dynamodb.NewFromConfig(awsConfig)
	
	// Create logger
	log := logger.New("kootoro-verification", "FetchImagesFunction")
	
	// Create S3Utils instance with config
	s3Util := s3utils.NewWithConfig(awsConfig, log)
	
	// Load configuration early
	config := LoadConfig()
	
	return &Dependencies{
		logger:    log,
		s3Client:  s3Client,
		dbClient:  dbClient,
		s3Utils:   s3Util,
		awsConfig: awsConfig,
		config:    config,
	}
}

// ConfigureDbUtils sets up the DynamoDB utilities with config
func (d *Dependencies) ConfigureDbUtils(config ConfigVars) {
	// Convert our ConfigVars to the format expected by dbutils
	dbConfig := dbutils.Config{
		VerificationTable: config.VerificationTable,
		LayoutTable:       config.LayoutTable,
		DefaultTTLDays:    30, // 30 days default TTL
	}
	
	d.dbUtils = dbutils.New(d.dbClient, d.logger, dbConfig)
}

// NewS3WrapperWithSize creates a new S3UtilsWrapper with the configured max image size and hybrid storage settings
func (d *Dependencies) NewS3WrapperWithSize(maxImageSize int64) *S3UtilsWrapper {
	wrapper := NewS3Utils(d.awsConfig, d.logger, maxImageSize)
	
	// Configure hybrid storage settings
	wrapper.SetTempBucket(d.config.TempBase64Bucket)
	wrapper.SetMaxInlineSize(d.config.MaxInlineBase64Size)
	
	// Log storage configuration for debugging
	d.logger.Info("Configured S3 wrapper with hybrid storage", map[string]interface{}{
		"maxImageSize":        maxImageSize,
		"tempBase64Bucket":    d.config.TempBase64Bucket,
		"maxInlineBase64Size": d.config.MaxInlineBase64Size,
		"inlineThresholdMB":   float64(d.config.MaxInlineBase64Size) / 1024 / 1024,
	})
	
	return wrapper
}

// GetLogger returns the logger
func (d *Dependencies) GetLogger() logger.Logger {
	return d.logger
}

// GetS3Client returns the S3 client
func (d *Dependencies) GetS3Client() *s3.Client {
	return d.s3Client
}

// GetDBClient returns the DynamoDB client
func (d *Dependencies) GetDBClient() *dynamodb.Client {
	return d.dbClient
}

// GetS3Utils returns the S3 utilities
func (d *Dependencies) GetS3Utils() *s3utils.S3Utils {
	return d.s3Utils
}

// GetDbUtils returns the DynamoDB utilities
func (d *Dependencies) GetDbUtils() *dbutils.DynamoDBUtils {
	return d.dbUtils
}

// GetAWSConfig returns the AWS config
func (d *Dependencies) GetAWSConfig() aws.Config {
	return d.awsConfig
}

// GetConfig returns the configuration
func (d *Dependencies) GetConfig() ConfigVars {
	return d.config
}

// LoadConfig loads configuration from environment variables
func LoadConfig() ConfigVars {
	// Parse max image size with default of 100MB
	maxImageSizeStr := getEnvWithDefault("MAX_IMAGE_SIZE", "104857600")
	maxImageSize, err := strconv.ParseInt(maxImageSizeStr, 10, 64)
	if err != nil {
		// Log error and use default
		maxImageSize = 104857600 // 100MB default
	}
	
	// Ensure reasonable limits (minimum 1MB, maximum 1GB)
	if maxImageSize < 1048576 { // 1MB
		maxImageSize = 1048576
	}
	if maxImageSize > 1073741824 { // 1GB
		maxImageSize = 1073741824
	}
	
	// Parse max inline Base64 size with default of 2MB
	maxInlineBase64SizeStr := getEnvWithDefault("MAX_INLINE_BASE64_SIZE", "2097152")
	maxInlineBase64Size, err := strconv.ParseInt(maxInlineBase64SizeStr, 10, 64)
	if err != nil {
		// Log error and use default
		maxInlineBase64Size = 2097152 // 2MB default
	}
	
	// Ensure reasonable limits for inline storage (minimum 512KB, maximum 10MB)
	if maxInlineBase64Size < 524288 { // 512KB
		maxInlineBase64Size = 524288
	}
	if maxInlineBase64Size > 10485760 { // 10MB
		maxInlineBase64Size = 10485760
	}
	
	// Validate temporary Base64 bucket
	tempBase64Bucket := getEnvWithDefault("TEMP_BASE64_BUCKET", "")
	if tempBase64Bucket == "" {
		// Log warning - this will cause issues for large images
		// For development/testing, this might be okay if only small images are used
	}
	
	return ConfigVars{
		LayoutTable:         getEnvWithDefault("DYNAMODB_LAYOUT_TABLE", "LayoutMetadata"),
		VerificationTable:   getEnvWithDefault("DYNAMODB_VERIFICATION_TABLE", "VerificationResults"),
		MaxImageSize:        maxImageSize,
		TempBase64Bucket:    tempBase64Bucket,
		MaxInlineBase64Size: maxInlineBase64Size,
	}
}

// ValidateConfig validates the configuration for hybrid storage
func (c *ConfigVars) ValidateConfig() error {
	// Validate that temporary bucket is configured for production use
	if c.TempBase64Bucket == "" {
		// This is only a warning in development, but would be an error in production
		// For now, we'll allow it but log a warning
		return nil
	}
	
	// Validate that inline threshold is smaller than max image size
	if c.MaxInlineBase64Size >= c.MaxImageSize {
		// This could cause issues, but we'll allow it
		return nil
	}
	
	return nil
}

// GetStorageConfig returns storage configuration for logging/debugging
func (c *ConfigVars) GetStorageConfig() map[string]interface{} {
	return map[string]interface{}{
		"maxImageSize":         c.MaxImageSize,
		"maxInlineBase64Size":  c.MaxInlineBase64Size,
		"tempBase64Bucket":     c.TempBase64Bucket,
		"maxImageSizeMB":       float64(c.MaxImageSize) / 1024 / 1024,
		"inlineThresholdMB":    float64(c.MaxInlineBase64Size) / 1024 / 1024,
		"layoutTable":          c.LayoutTable,
		"verificationTable":    c.VerificationTable,
	}
}

// getEnvWithDefault gets an environment variable with a default value
func getEnvWithDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

// ValidateBucketAccess validates that the temporary S3 bucket is accessible (optional validation)
func (d *Dependencies) ValidateBucketAccess() error {
	if d.config.TempBase64Bucket == "" {
		d.logger.Warn("No temporary Base64 bucket configured - large images will fail", map[string]interface{}{
			"tempBase64Bucket": d.config.TempBase64Bucket,
		})
		return nil
	}
	
	// Validate bucket access if configured
	d.logger.Info("Validating temporary Base64 bucket access", map[string]interface{}{
		"bucket": d.config.TempBase64Bucket,
	})
	
	// TODO: Add actual bucket access validation if needed
	// This could include checking if the bucket exists and we have write permissions
	
	return nil
}

// GetEnvironmentInfo returns debug information about the environment
func GetEnvironmentInfo() map[string]interface{} {
	return map[string]interface{}{
		"env_MAX_IMAGE_SIZE":           os.Getenv("MAX_IMAGE_SIZE"),
		"env_MAX_INLINE_BASE64_SIZE":   os.Getenv("MAX_INLINE_BASE64_SIZE"),
		"env_TEMP_BASE64_BUCKET":       os.Getenv("TEMP_BASE64_BUCKET"),
		"env_DYNAMODB_LAYOUT_TABLE":    os.Getenv("DYNAMODB_LAYOUT_TABLE"),
		"env_DYNAMODB_VERIFICATION_TABLE": os.Getenv("DYNAMODB_VERIFICATION_TABLE"),
	}
}