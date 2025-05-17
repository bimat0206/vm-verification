package dependencies

import (
	"context"
	//"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"workflow-function/shared/schema"
	"workflow-function/shared/logger"
	"workflow-function/shared/errors"
	"workflow-function/ExecuteTurn1/internal/config" // This is your local config wrapper
)

// Clients centralizes all required AWS and hybrid storage clients/services.
type Clients struct {
	BedrockClient *bedrockruntime.Client
	S3Client      *s3.Client
	HybridConfig  *schema.HybridStorageConfig
	Logger        logger.Logger
}

// New initializes AWS clients and hybrid config, logs issues, returns errors.WorkflowError on failure.
func New(ctx context.Context, cfg *config.Config, log logger.Logger) (*Clients, error) {
	// AWS SDK v2 config loading
	awsCfg, err := loadAWSConfig(ctx, cfg.AWSRegion)
	if err != nil {
		log.Error("Failed to load AWS config", map[string]interface{}{"error": err.Error()})
		return nil, errors.NewInternalError("AWSConfig", err)
	}

	// Bedrock client
	bedrockClient := bedrockruntime.NewFromConfig(awsCfg)
	// S3 client
	s3Client := s3.NewFromConfig(awsCfg)

	// Hybrid storage config as per shared schema (populate from your config)
	hybridCfg := &schema.HybridStorageConfig{
		TempBase64Bucket:       cfg.TempBase64Bucket,
		Base64SizeThreshold:    cfg.Base64SizeThreshold,
		Base64RetrievalTimeout: int(cfg.Base64RetrievalTimeout.Milliseconds()),
		EnableHybridStorage:    cfg.EnableHybridStorage,
	}

	log.Info("AWS clients and hybrid config initialized", map[string]interface{}{
		"region":              cfg.AWSRegion,
		"hybridBase64Enabled": cfg.EnableHybridStorage,
		"tempBase64Bucket":    cfg.TempBase64Bucket,
	})

	return &Clients{
		BedrockClient: bedrockClient,
		S3Client:      s3Client,
		HybridConfig:  hybridCfg,
		Logger:        log,
	}, nil
}

// loadAWSConfig is a helper for AWS config loading with retries and defaults.
func loadAWSConfig(ctx context.Context, region string) (aws.Config, error) {
	return awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
}
