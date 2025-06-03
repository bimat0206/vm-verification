package config

import (
	"context"
	"fmt"
	"os"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// LambdaConfig holds environment configuration for the FinalizeWithError function.
type LambdaConfig struct {
	VerificationResultsTable string
	ConversationHistoryTable string
	StateManagementBucket    string
}

// LoadEnvConfig loads configuration from environment variables.
func LoadEnvConfig() (*LambdaConfig, error) {
	cfg := &LambdaConfig{
		VerificationResultsTable: os.Getenv("DYNAMODB_VERIFICATION_TABLE"),
		ConversationHistoryTable: os.Getenv("DYNAMODB_CONVERSATION_TABLE"),
		StateManagementBucket:    os.Getenv("STATE_BUCKET"),
	}

	if cfg.VerificationResultsTable == "" {
		return nil, fmt.Errorf("DYNAMODB_VERIFICATION_TABLE is required")
	}
	if cfg.StateManagementBucket == "" {
		return nil, fmt.Errorf("STATE_BUCKET is required")
	}
	return cfg, nil
}

// AWSClients bundles AWS service clients used by the function.
type AWSClients struct {
	DynamoDBClient *dynamodb.Client
	S3Client       *s3.Client
}

// NewAWSClients creates AWS service clients using default configuration.
func NewAWSClients(ctx context.Context) (*AWSClients, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &AWSClients{
		DynamoDBClient: dynamodb.NewFromConfig(cfg),
		S3Client:       s3.NewFromConfig(cfg),
	}, nil
}
