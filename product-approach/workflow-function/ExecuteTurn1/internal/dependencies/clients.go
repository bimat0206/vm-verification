// internal/dependencies/clients.go

package dependencies

import (
    "context"
    "net/http"
    //"time"

    "github.com/aws/aws-sdk-go-v2/aws"
    awsconfig "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
    "github.com/aws/aws-sdk-go-v2/service/s3"

    "workflow-function/shared/schema"
    "workflow-function/shared/logger"
    wferrors "workflow-function/shared/errors"
    "workflow-function/ExecuteTurn1/internal/config"
)

// Clients centralizes AWS clients and hybrid-storage config.
type Clients struct {
    BedrockClient *bedrockruntime.Client
    S3Client      *s3.Client
    HybridConfig  *schema.HybridStorageConfig
    Logger        logger.Logger
}

// New initializes AWS SDK clients, applying the Bedrock timeout from Lambda env.
// Returns a WorkflowError on any failure.
func New(ctx context.Context, cfg *config.Config, log logger.Logger) (*Clients, error) {
    // 1. Load AWS config for region
    awsCfg, err := loadAWSConfig(ctx, cfg.AWSRegion)
    if err != nil {
        log.Error("Failed to load AWS config", map[string]interface{}{"error": err.Error()})
        return nil, wferrors.NewInternalError("AWSConfig", err)
    }

    // 2. Create a custom HTTP client for Bedrock API calls
    httpClient := &http.Client{Timeout: cfg.BedrockTimeout}

    // 3. Bedrock client with timeout
    bedrockClient := bedrockruntime.NewFromConfig(awsCfg, func(o *bedrockruntime.Options) {
        o.HTTPClient = httpClient
    })

    // 4. Standard S3 client
    s3Client := s3.NewFromConfig(awsCfg)

    // 5. Hybrid-storage config
    hybridCfg := &schema.HybridStorageConfig{
        TempBase64Bucket:       cfg.TempBase64Bucket,
        Base64SizeThreshold:    cfg.Base64SizeThreshold,
        Base64RetrievalTimeout: int(cfg.Base64RetrievalTimeout.Milliseconds()),
        EnableHybridStorage:    cfg.EnableHybridStorage,
    }

    log.Info("AWS clients initialized", map[string]interface{}{
        "region":           cfg.AWSRegion,
        "hybridEnabled":    cfg.EnableHybridStorage,
        "tempBase64Bucket": cfg.TempBase64Bucket,
        "bedrockTimeout":   cfg.BedrockTimeout.String(),
    })

    return &Clients{
        BedrockClient: bedrockClient,
        S3Client:      s3Client,
        HybridConfig:  hybridCfg,
        Logger:        log,
    }, nil
}

// loadAWSConfig is a thin wrapper for AWS SDK v2 config loading.
func loadAWSConfig(ctx context.Context, region string) (aws.Config, error) {
    return awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
}
