package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"workflow-function/shared/logger"
	"workflow-function/shared/s3utils"
	"workflow-function/shared/dbutils"
)

// Dependencies holds all AWS service clients and utilities
type Dependencies struct {
	S3Client     *s3.Client
	DynamoClient *dynamodb.Client
	S3Utils      *s3utils.S3Utils
	DBUtils      *dbutils.DynamoDBUtils
	Logger       logger.Logger
	Config       *AppConfig
}

// AppConfig holds application configuration
type AppConfig struct {
	// AWS Configuration
	AWSRegion string `json:"awsRegion"`
	
	// DynamoDB Tables
	VerificationTable  string `json:"verificationTable"`
	LayoutTable        string `json:"layoutTable"`
	ConversationTable  string `json:"conversationTable"`
	
	// S3 Buckets
	ReferenceBucket string `json:"referenceBucket"`
	CheckingBucket  string `json:"checkingBucket"`
	ResultsBucket   string `json:"resultsBucket"`
	
	// Processing Configuration
	MaxResponseSize    int64         `json:"maxResponseSize"`
	ProcessingTimeout  time.Duration `json:"processingTimeout"`
	EnableCaching      bool          `json:"enableCaching"`
	ValidateResponses  bool          `json:"validateResponses"`
	
	// Logging Configuration
	LogLevel          string `json:"logLevel"`
	ServiceName       string `json:"serviceName"`
	FunctionName      string `json:"functionName"`
	CorrelationIdKey  string `json:"correlationIdKey"`
	
	// Feature Flags
	EnableFallbackParsing  bool `json:"enableFallbackParsing"`
	EnableStrictValidation bool `json:"enableStrictValidation"`
	EnableMetrics          bool `json:"enableMetrics"`
}

// NewDependencies creates a new Dependencies instance
func NewDependencies(log logger.Logger) *Dependencies {
	ctx := context.Background()
	deps, err := InitializeDependencies(ctx)
	if err != nil {
		log.Error("Failed to initialize dependencies", map[string]interface{}{
			"error": err.Error(),
		})
		// Return a basic dependencies struct for graceful degradation
		return &Dependencies{
			Logger: log,
			Config: LoadConfig(),
		}
	}
	return deps
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *AppConfig {
	return &AppConfig{
		// AWS Configuration
		AWSRegion: getEnvOrDefault("AWS_REGION", "us-east-1"),
		
		// DynamoDB Tables
		VerificationTable: getEnvOrDefault("DYNAMODB_VERIFICATION_TABLE", "VerificationResults"),
		LayoutTable:       getEnvOrDefault("DYNAMODB_LAYOUT_TABLE", "LayoutMetadata"),
		ConversationTable: getEnvOrDefault("DYNAMODB_CONVERSATION_TABLE", "ConversationHistory"),
		
		// S3 Buckets  
		ReferenceBucket: getEnvOrDefault("REFERENCE_BUCKET", "kootoro-reference-bucket"),
		CheckingBucket:  getEnvOrDefault("CHECKING_BUCKET", "kootoro-checking-bucket"),
		ResultsBucket:   getEnvOrDefault("RESULTS_BUCKET", "kootoro-results-bucket"),
		
		// Processing Configuration
		MaxResponseSize:   getEnvAsInt64("MAX_RESPONSE_SIZE", 1024*1024), // 1MB default
		ProcessingTimeout: getEnvAsDuration("PROCESSING_TIMEOUT", "60s"),
		EnableCaching:     getEnvAsBool("ENABLE_CACHING", true),
		ValidateResponses: getEnvAsBool("VALIDATE_RESPONSES", true),
		
		// Logging Configuration
		LogLevel:         getEnvOrDefault("LOG_LEVEL", "INFO"),
		ServiceName:      getEnvOrDefault("SERVICE_NAME", "verification-service"),
		FunctionName:     getEnvOrDefault("FUNCTION_NAME", "ProcessTurn1Response"),
		CorrelationIdKey: getEnvOrDefault("CORRELATION_ID_KEY", "verificationId"),
		
		// Feature Flags
		EnableFallbackParsing:  getEnvAsBool("ENABLE_FALLBACK_PARSING", true),
		EnableStrictValidation: getEnvAsBool("ENABLE_STRICT_VALIDATION", false),
		EnableMetrics:          getEnvAsBool("ENABLE_METRICS", true),
	}
}

// InitializeDependencies initializes all dependencies
func InitializeDependencies(ctx context.Context) (*Dependencies, error) {
	// Load configuration
	appConfig := LoadConfig()
	
	// Initialize logger
	log := logger.New(appConfig.ServiceName, appConfig.FunctionName)
	
	log.Info("Initializing dependencies", map[string]interface{}{
		"region":           appConfig.AWSRegion,
		"functionName":     appConfig.FunctionName,
		"logLevel":         appConfig.LogLevel,
		"verificationTable": appConfig.VerificationTable,
	})
	
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(appConfig.AWSRegion),
		config.WithClientLogMode(aws.LogRetries | aws.LogRequest),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	
	// Initialize AWS service clients
	s3Client := s3.NewFromConfig(cfg)
	dynamoClient := dynamodb.NewFromConfig(cfg)
	
	// Initialize utility clients
	s3Utils := s3utils.New(s3Client, log)
	
	// Configure DynamoDB utils
	dbConfig := dbutils.Config{
		VerificationTable: appConfig.VerificationTable,
		LayoutTable:       appConfig.LayoutTable,
		ConversationTable: appConfig.ConversationTable,
		DefaultTTLDays:    30,
	}
	dbUtils := dbutils.New(dynamoClient, log, dbConfig)
	
	// Create dependencies struct
	deps := &Dependencies{
		S3Client:     s3Client,
		DynamoClient: dynamoClient,
		S3Utils:      s3Utils,
		DBUtils:      dbUtils,
		Logger:       log,
		Config:       appConfig,
	}
	
	log.Info("Dependencies initialized successfully", map[string]interface{}{
		"s3ClientReady":     deps.S3Client != nil,
		"dynamoClientReady": deps.DynamoClient != nil,
		"s3UtilsReady":      deps.S3Utils != nil,
		"dbUtilsReady":      deps.DBUtils != nil,
	})
	
	return deps, nil
}

// ValidateConfiguration validates the application configuration
func (deps *Dependencies) ValidateConfiguration() error {
	config := deps.Config
	
	// Required environment variables
	required := map[string]string{
		"DYNAMODB_VERIFICATION_TABLE": config.VerificationTable,
		"REFERENCE_BUCKET":            config.ReferenceBucket,
		"CHECKING_BUCKET":             config.CheckingBucket,
	}
	
	for name, value := range required {
		if value == "" {
			return fmt.Errorf("required environment variable %s is not set", name)
		}
	}
	
	// Validate AWS region
	if config.AWSRegion == "" {
		return fmt.Errorf("AWS region is not configured")
	}
	
	// Validate timeout
	if config.ProcessingTimeout <= 0 {
		return fmt.Errorf("processing timeout must be positive, got %v", config.ProcessingTimeout)
	}
	
	// Validate max response size
	if config.MaxResponseSize <= 0 {
		return fmt.Errorf("max response size must be positive, got %d", config.MaxResponseSize)
	}
	
	deps.Logger.Info("Configuration validation successful", map[string]interface{}{
		"region":              config.AWSRegion,
		"processingTimeout":   config.ProcessingTimeout,
		"maxResponseSize":     config.MaxResponseSize,
		"enableCaching":       config.EnableCaching,
		"validateResponses":   config.ValidateResponses,
	})
	
	return nil
}

// TestConnections tests connectivity to AWS services
func (deps *Dependencies) TestConnections(ctx context.Context) error {
	log := deps.Logger
	
	// Skip tests if clients are nil (graceful degradation)
	if deps.DynamoClient == nil || deps.S3Client == nil {
		log.Warn("Skipping connection tests due to missing clients", nil)
		return nil
	}
	
	// Test DynamoDB connection
	log.Debug("Testing DynamoDB connection", nil)
	_, err := deps.DynamoClient.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(deps.Config.VerificationTable),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to DynamoDB table %s: %w", 
			deps.Config.VerificationTable, err)
	}
	
	// Test S3 connection
	log.Debug("Testing S3 connection", nil)
	_, err = deps.S3Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(deps.Config.ReferenceBucket),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to S3 bucket %s: %w", 
			deps.Config.ReferenceBucket, err)
	}
	
	log.Info("All service connections tested successfully", nil)
	return nil
}

// GetS3Client returns the S3 client
func (deps *Dependencies) GetS3Client() *s3.Client {
	return deps.S3Client
}

// GetDynamoDBClient returns the DynamoDB client
func (deps *Dependencies) GetDynamoDBClient() *dynamodb.Client {
	return deps.DynamoClient
}

// GetS3Utils returns the S3 utilities
func (deps *Dependencies) GetS3Utils() *s3utils.S3Utils {
	return deps.S3Utils
}

// GetDBUtils returns the DynamoDB utilities
func (deps *Dependencies) GetDBUtils() *dbutils.DynamoDBUtils {
	return deps.DBUtils
}

// GetLogger returns the logger
func (deps *Dependencies) GetLogger() logger.Logger {
	return deps.Logger
}

// GetConfig returns the application configuration
func (deps *Dependencies) GetConfig() *AppConfig {
	return deps.Config
}

// CreateProcessingConfig creates processing configuration based on verification context
func (deps *Dependencies) CreateProcessingConfig(verificationType string, hasHistorical bool) *ProcessingConfig {
	// Get base configuration from use case
	config := GetProcessingConfigForUseCase(verificationType, hasHistorical)
	
	// Apply feature flags from app config
	config.FallbackToTextParsing = deps.Config.EnableFallbackParsing
	config.StrictValidation = deps.Config.EnableStrictValidation
	config.MaxResponseSize = deps.Config.MaxResponseSize
	
	deps.Logger.Debug("Created processing configuration", map[string]interface{}{
		"verificationType":       verificationType,
		"hasHistorical":          hasHistorical,
		"extractMachineStructure": config.ExtractMachineStructure,
		"fallbackToTextParsing":  config.FallbackToTextParsing,
		"strictValidation":       config.StrictValidation,
	})
	
	return config
}

// Cleanup performs cleanup operations
func (deps *Dependencies) Cleanup() error {
	deps.Logger.Info("Performing cleanup operations", nil)
	
	// Close any resources if needed
	// Currently, AWS SDK clients don't require explicit cleanup
	
	deps.Logger.Info("Cleanup completed", nil)
	return nil
}

// Helper functions for environment variable parsing

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := parseSize(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvAsDuration(key, defaultValue string) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	duration, _ := time.ParseDuration(defaultValue)
	return duration
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		switch value {
		case "true", "TRUE", "1", "yes", "YES":
			return true
		case "false", "FALSE", "0", "no", "NO":
			return false
		}
	}
	return defaultValue
}

// parseSize parses size strings like "1MB", "512KB", etc.
func parseSize(s string) (int64, error) {
	// Simple implementation - can be extended for more formats
	if s == "" {
		return 0, fmt.Errorf("empty size string")
	}
	
	// Check for common suffixes
	switch {
	case len(s) > 2 && s[len(s)-2:] == "MB":
		val := s[:len(s)-2]
		if parsed, err := parseInt64(val); err == nil {
			return parsed * 1024 * 1024, nil
		}
	case len(s) > 2 && s[len(s)-2:] == "KB":
		val := s[:len(s)-2]
		if parsed, err := parseInt64(val); err == nil {
			return parsed * 1024, nil
		}
	case len(s) > 1 && s[len(s)-1:] == "B":
		val := s[:len(s)-1]
		return parseInt64(val)
	default:
		// Try parsing as plain number (bytes)
		return parseInt64(s)
	}
	
	return 0, fmt.Errorf("invalid size format: %s", s)
}

func parseInt64(s string) (int64, error) {
	var result int64
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0, fmt.Errorf("invalid number: %s", s)
		}
		result = result*10 + int64(ch-'0')
	}
	return result, nil
}

// Global dependencies instance (initialized in main.go)
var GlobalDependencies *Dependencies