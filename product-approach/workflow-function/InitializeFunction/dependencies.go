package main

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"workflow-function/shared/dbutils"
	"workflow-function/shared/logger"
	"workflow-function/shared/s3utils"
)

// Dependencies contains all external dependencies required by the service
type Dependencies struct {
	logger    logger.Logger
	s3Client  *s3.Client
	dbClient  *dynamodb.Client
	s3Utils   *s3utils.S3Utils
	dbUtils   *dbutils.DynamoDBUtils
}

// NewDependencies creates a new Dependencies instance with all required dependencies
func NewDependencies(awsConfig aws.Config) *Dependencies {
	// Initialize clients
	s3Client := s3.NewFromConfig(awsConfig)
	dbClient := dynamodb.NewFromConfig(awsConfig)
	
	// Create logger
	log := logger.New("kootoro-verification", "InitializeFunction")
	
	// Create utility instances
	s3Util := s3utils.New(s3Client, log)
	
	// Return dependencies without configuring dbUtils yet
	// It will be configured once config is available
	return &Dependencies{
		logger:    log,
		s3Client:  s3Client,
		dbClient:  dbClient,
		s3Utils:   s3Util,
	}
}

// ConfigureDbUtils sets up the DynamoDB utilities with config
func (d *Dependencies) ConfigureDbUtils(config ConfigVars) {
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

// GetS3Util returns the S3 utilities
func (d *Dependencies) GetS3Util() *s3utils.S3Utils {
	return d.s3Utils
}

// GetDynamoUtil returns the DynamoDB utilities
func (d *Dependencies) GetDynamoUtil() *dbutils.DynamoDBUtils {
	return d.dbUtils
}

// For backward compatibility
func (d *Dependencies) GetS3Client() *s3.Client {
	return d.s3Client
}

// For backward compatibility
func (d *Dependencies) GetDBClient() *dynamodb.Client {
	return d.dbClient
}