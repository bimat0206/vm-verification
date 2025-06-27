package config

import (
	"context"
	"fmt"
	"os"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type LambdaConfig struct {
	VerificationResultsTable string
	ConversationHistoryTable string
	StateBucket              string
}

func LoadEnvConfig() (*LambdaConfig, error) {
	cfg := &LambdaConfig{}
	cfg.VerificationResultsTable = os.Getenv("DYNAMODB_VERIFICATION_TABLE")
	cfg.ConversationHistoryTable = os.Getenv("DYNAMODB_CONVERSATION_TABLE")
	cfg.StateBucket = os.Getenv("STATE_BUCKET")

	if cfg.VerificationResultsTable == "" {
		return nil, fmt.Errorf("missing env DYNAMODB_VERIFICATION_TABLE")
	}
	if cfg.ConversationHistoryTable == "" {
		return nil, fmt.Errorf("missing env DYNAMODB_CONVERSATION_TABLE")
	}
	if cfg.StateBucket == "" {
		return nil, fmt.Errorf("missing env STATE_BUCKET")
	}
	return cfg, nil
}

type AWSClients struct {
	S3Client       *s3.Client
	DynamoDBClient *dynamodb.Client
}

func NewAWSClients(ctx context.Context) (*AWSClients, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &AWSClients{
		S3Client:       s3.NewFromConfig(cfg),
		DynamoDBClient: dynamodb.NewFromConfig(cfg),
	}, nil
}
