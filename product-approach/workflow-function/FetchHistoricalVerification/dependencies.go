package main

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"workflow-function/shared/dbutils"
	"workflow-function/shared/logger"
)

// ConfigVars contains environment configuration values
type ConfigVars struct {
	VerificationTable  string
	CheckingBucket     string
	Region             string
	LogLevel           string
}

// Dependencies contains all external dependencies required by the service
type Dependencies struct {
	logger   logger.Logger
	dbClient *dynamodb.Client
	dbUtils  *dbutils.DynamoDBUtils
}

// NewDependencies creates a new Dependencies instance with all required dependencies
func NewDependencies(awsConfig aws.Config) *Dependencies {
	// Initialize DynamoDB client
	dbClient := dynamodb.NewFromConfig(awsConfig)
	
	// Create logger
	log := logger.New("kootoro-verification", "FetchHistoricalVerification")
	
	return &Dependencies{
		logger:   log,
		dbClient: dbClient,
	}
}

// ConfigureDbUtils sets up the DynamoDB utilities with config
func (d *Dependencies) ConfigureDbUtils(config ConfigVars) {
	// Convert our ConfigVars to the format expected by dbutils
	dbConfig := dbutils.Config{
		VerificationTable: config.VerificationTable,
		DefaultTTLDays:    30, // 30 days default TTL
	}
	
	d.dbUtils = dbutils.New(d.dbClient, d.logger, dbConfig)
}

// GetLogger returns the logger
func (d *Dependencies) GetLogger() logger.Logger {
	return d.logger
}

// GetDynamoUtil returns the DynamoDB utilities
func (d *Dependencies) GetDynamoUtil() *dbutils.DynamoDBUtils {
	return d.dbUtils
}

// GetDBClient returns the raw DynamoDB client (for backward compatibility)
func (d *Dependencies) GetDBClient() *dynamodb.Client {
	return d.dbClient
}

// LoadConfig loads configuration from environment variables
func LoadConfig() ConfigVars {
	return ConfigVars{
		VerificationTable: getEnvWithDefault("DYNAMODB_VERIFICATION_TABLE", "VerificationResults"),
		CheckingBucket:    getEnvWithDefault("CHECKING_BUCKET", "kootoro-checking-bucket"),
		Region:            getEnvWithDefault("AWS_REGION", "us-east-1"),
		LogLevel:          getEnvWithDefault("LOG_LEVEL", "INFO"),
	}
}

// getEnvWithDefault gets an environment variable with a default value
func getEnvWithDefault(key, defaultValue string) string {
	value := getEnv(key)
	if value == "" {
		return defaultValue
	}
	return value
}