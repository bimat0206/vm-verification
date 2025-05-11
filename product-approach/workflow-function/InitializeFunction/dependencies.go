package main

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Dependencies contains all external dependencies required by the service
type Dependencies struct {
	Logger       Logger
	S3Client     *s3.Client
	DynamoClient *dynamodb.Client
}

// NewDependencies creates a new Dependencies instance with all required dependencies
func NewDependencies(awsConfig aws.Config) *Dependencies {
	// Initialize clients
	s3Client := s3.NewFromConfig(awsConfig)
	dynamoClient := dynamodb.NewFromConfig(awsConfig)
	logger := NewStructuredLogger()

	// Return dependencies container
	return &Dependencies{
		Logger:       logger,
		S3Client:     s3Client,
		DynamoClient: dynamoClient,
	}
}

// GetS3Util returns an instance of S3Utils
func (d *Dependencies) GetS3Util() *S3Utils {
	s3Util := NewS3Utils(d.S3Client, d.Logger)
	return s3Util
}

// GetDynamoUtil returns an instance of DynamoDBUtils
func (d *Dependencies) GetDynamoUtil() *DynamoDBUtils {
	dbUtil := NewDynamoDBUtils(d.DynamoClient, d.Logger)
	return dbUtil
}

// GetLogger returns the logger instance
func (d *Dependencies) GetLogger() Logger {
	return d.Logger
}