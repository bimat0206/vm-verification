package main

import (
	"os"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"workflow-function/shared/dbutils"
	"workflow-function/shared/logger"
	"workflow-function/shared/s3utils"
)

// ConfigVars contains configuration for the lambda function
type ConfigVars struct {
	LayoutTable        string
	VerificationTable  string
	MaxImageSize       int64
}

// Dependencies contains all external dependencies required by the service
type Dependencies struct {
	logger    logger.Logger
	s3Client  *s3.Client
	dbClient  *dynamodb.Client
	s3Utils   *s3utils.S3Utils
	dbUtils   *dbutils.DynamoDBUtils
	awsConfig aws.Config // Store the AWS config for creating services
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
	
	return &Dependencies{
		logger:    log,
		s3Client:  s3Client,
		dbClient:  dbClient,
		s3Utils:   s3Util,
		awsConfig: awsConfig, // Store the config
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

// LoadConfig loads configuration from environment variables
func LoadConfig() ConfigVars {
	return ConfigVars{
		LayoutTable:       getEnvWithDefault("DYNAMODB_LAYOUT_TABLE", "LayoutMetadata"),
		VerificationTable: getEnvWithDefault("DYNAMODB_VERIFICATION_TABLE", "VerificationResults"),
		MaxImageSize:      10 * 1024 * 1024, // 10MB default
	}
}

// getEnvWithDefault gets an environment variable with a default value
func getEnvWithDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}