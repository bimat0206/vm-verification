package internal

import (
	"os"
	"workflow-function/shared/logger"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"workflow-function/shared/s3state"
)

// ConfigVars contains environment configuration values
type ConfigVars struct {
	VerificationTable string
	CheckingBucket    string
	StateBucket       string
	Region            string
	LogLevel          string
}

// Dependencies contains all external dependencies required by the service
type Dependencies struct {
	logger     logger.Logger
	dbClient   *dynamodb.Client
	dynamoRepo *DynamoDBRepository // Replace dbUtils with dynamoRepo
	stateMgr   s3state.Manager
}

// NewDependencies creates a new Dependencies instance with all required dependencies
func NewDependencies(awsConfig aws.Config, config ConfigVars) (*Dependencies, error) {
	// Initialize DynamoDB client
	dbClient := dynamodb.NewFromConfig(awsConfig)

	// Create logger
	log := logger.New("kootoro-verification", "FetchHistoricalVerification")

	// Create DynamoDB repository
	dynamoRepo := NewDynamoDBRepository(dbClient, config.VerificationTable, log)

	// Initialize S3 state manager
	mgr, err := s3state.New(config.StateBucket)
	if err != nil {
		return nil, err
	}

	return &Dependencies{
		logger:     log,
		dbClient:   dbClient,
		dynamoRepo: dynamoRepo,
		stateMgr:   mgr,
	}, nil
}

// GetLogger returns the logger
func (d *Dependencies) GetLogger() logger.Logger {
	return d.logger
}

// GetDynamoRepo returns the DynamoDB repository
func (d *Dependencies) GetDynamoRepo() *DynamoDBRepository {
	return d.dynamoRepo
}

// GetStateManager returns the S3 state manager
func (d *Dependencies) GetStateManager() s3state.Manager {
	return d.stateMgr
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
		StateBucket:       getEnvWithDefault("STATE_BUCKET", ""),
		Region:            getEnvWithDefault("AWS_REGION", "us-east-1"),
		LogLevel:          getEnvWithDefault("LOG_LEVEL", "INFO"),
	}
}

// getEnvWithDefault returns environment variable value or default if not set
func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
