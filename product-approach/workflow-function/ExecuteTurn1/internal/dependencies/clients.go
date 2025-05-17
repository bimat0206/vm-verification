package dependencies

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	
	appConfig "workflow-function/ExecuteTurn1/internal/config"
)

// Clients holds all AWS clients needed for the ExecuteTurn1 function
type Clients struct {
	BedrockClient *bedrockruntime.Client
	S3Client      *s3.Client
	Config        *appConfig.Config
}

// New creates and initializes all needed AWS clients
func New(ctx context.Context, cfg *appConfig.Config) (*Clients, error) {
	// Set up AWS config with region
	awsCfg, err := config.LoadDefaultConfig(ctx, 
		config.WithRegion(cfg.AWSRegion),
		config.WithRetryMode(aws.RetryModeAdaptive),
		config.WithRetryMaxAttempts(3),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS configuration: %w", err)
	}

	// Create Bedrock client with timeout
	bedrockClientTimeout := cfg.BedrockTimeout
	if bedrockClientTimeout <= 0 {
		bedrockClientTimeout = 5 * time.Minute
	}
	
	bedrockClient := bedrockruntime.NewFromConfig(awsCfg, func(o *bedrockruntime.Options) {
		// Use the default retryer
		o.ClientLogMode = aws.LogRetries | aws.LogRequest | aws.LogResponse
	})
	
	// Create S3 client
	s3Client := s3.NewFromConfig(awsCfg)

	return &Clients{
		BedrockClient: bedrockClient,
		S3Client:      s3Client,
		Config:        cfg,
	}, nil
}

// BedrockService provides methods for interacting with the Bedrock service
type BedrockService struct {
	client *bedrockruntime.Client
	modelID string
	anthropicVersion string
}

// NewBedrockService creates a new BedrockService with the given client and configuration
func NewBedrockService(client *bedrockruntime.Client, cfg *appConfig.Config) *BedrockService {
	return &BedrockService{
		client: client,
		modelID: cfg.BedrockModelID,
		anthropicVersion: cfg.AnthropicVersion,
	}
}

// HybridBase64Service provides methods for handling hybrid Base64 storage
type HybridBase64Service struct {
	s3Client *s3.Client
	bucket string
	sizeThreshold int64
	retrievalTimeout time.Duration
}

// NewHybridBase64Service creates a new HybridBase64Service with the given client and configuration
func NewHybridBase64Service(s3Client *s3.Client, cfg *appConfig.Config) *HybridBase64Service {
	return &HybridBase64Service{
		s3Client: s3Client,
		bucket: cfg.TempBase64Bucket,
		sizeThreshold: cfg.Base64SizeThreshold,
		retrievalTimeout: cfg.Base64RetrievalTimeout,
	}
}
