package dependencies

import (
   "context"
   "time"

   "github.com/aws/aws-sdk-go-v2/aws"
   awsconfig "github.com/aws/aws-sdk-go-v2/config"
   "github.com/aws/aws-sdk-go-v2/service/s3"

   "workflow-function/shared/logger"
   "workflow-function/shared/s3state"
   "workflow-function/shared/bedrock"
   wferrors "workflow-function/shared/errors"

   "workflow-function/ExecuteTurn1/internal"
   localBedrock "workflow-function/ExecuteTurn1/internal/bedrock"
   "workflow-function/ExecuteTurn1/internal/config"
   "workflow-function/ExecuteTurn1/internal/state"
   "workflow-function/ExecuteTurn1/internal/validation"
   "workflow-function/ExecuteTurn1/internal/handler"
)

// Clients centralizes shared dependencies
type Clients struct {
   // Services
   S3Client       *s3.Client
   S3StateManager s3state.Manager
   BedrockClient  *bedrock.BedrockClient
   
   // Components
   StateLoader    *state.Loader
   StateSaver     *state.Saver
   BedrockHandler *localBedrock.Client
   Validator      *validation.Validator
   
   // Main handler
   Handler        *handler.Handler
   
   // Configuration
   Config         *config.Config
   HybridConfig   *internal.HybridStorageConfig
   
   // Logger
   Logger         logger.Logger
}

// New initializes clients and components for dependency injection
func New(ctx context.Context, cfg *config.Config, log logger.Logger) (*Clients, error) {
   // 1. Load AWS config
   awsCfg, err := loadAWSConfig(ctx, cfg.BedrockRegion)
   if err != nil {
   	log.Error("Failed to load AWS config", map[string]interface{}{"error": err.Error()})
   	return nil, wferrors.NewInternalError("AWSConfig", err)
   }

   // 2. Initialize clients
   s3Client := s3.NewFromConfig(awsCfg)
   
   // 3. Set up S3 state manager
   s3StateManager, err := s3state.New(cfg.StateBucket)
   if err != nil {
   	log.Error("Failed to create S3 state manager", map[string]interface{}{"error": err.Error()})
   	return nil, wferrors.NewInternalError("S3StateManager", err)
   }
   
   // 4. Set up Bedrock client with Converse API
   bedrockClientConfig := bedrock.CreateClientConfig(
   	cfg.BedrockRegion, 
   	cfg.AnthropicVersion, 
   	cfg.MaxTokens, 
   	cfg.ThinkingType, 
   	cfg.ThinkingBudgetTokens,
   )
   
   bedrockClient, err := bedrock.NewBedrockClient(ctx, cfg.BedrockModel, bedrockClientConfig)
   if err != nil {
   	log.Error("Failed to create Bedrock client", map[string]interface{}{"error": err.Error()})
   	return nil, wferrors.NewInternalError("BedrockClient", err)
   }
   
   // 5. Set up hybrid storage config for images - without TempBase64Bucket and EnableHybridStorage
   hybridConfig := &internal.HybridStorageConfig{
   	// Using empty string for TempBase64Bucket since it was removed from config
   	TempBase64Bucket:       "",
   	Base64SizeThreshold:    cfg.Base64SizeThreshold,
   	Base64RetrievalTimeout: int(cfg.StateTimeout.Milliseconds()),
   	// We set EnableHybridStorage to false since it was removed from config
   	EnableHybridStorage:    false,
   }
   
   // 6. Create components
   stateLoader := state.NewLoader(s3StateManager, s3Client, log, cfg.StateTimeout)
   stateSaver := state.NewSaver(s3StateManager, s3Client, log, cfg.StateTimeout)
   validator := validation.NewValidator(log)
   
   // 7. Create local Bedrock handler with adapter
   bedrockAdapter := localBedrock.NewBedrockAdapter(bedrockClient)
   bedrockHandler := localBedrock.NewClient(bedrockAdapter, log, &localBedrock.Config{
   	ModelID:          cfg.BedrockModel, // Updated from BedrockModelID
   	AnthropicVersion: cfg.AnthropicVersion,
   	MaxTokens:        cfg.MaxTokens,
   	Temperature:      cfg.Temperature,
   	ThinkingType:     cfg.ThinkingType,
   	ThinkingBudget:   cfg.ThinkingBudgetTokens,
   	Timeout:          cfg.BedrockTimeout,
   })
   
   // 8. Create main handler
   mainHandler := handler.NewHandler(
   	stateLoader,
   	stateSaver,
   	bedrockHandler,
   	validator,
   	s3Client,
   	hybridConfig,
   	log,
   )
   
   // Log initialization success
   log.Info("Dependencies initialized successfully", map[string]interface{}{
   	"region":           cfg.BedrockRegion,
   	"bedrockModel":     cfg.BedrockModel, // Updated from BedrockModelID
   	"stateBucket":      cfg.StateBucket,
   	"hybridEnabled":    false, // Hardcoded since EnableHybridStorage was removed
   })
   
   // Return all clients
   return &Clients{
   	// Services
   	S3Client:       s3Client,
   	S3StateManager: s3StateManager,
   	BedrockClient:  bedrockClient,
   	
   	// Components
   	StateLoader:    stateLoader,
   	StateSaver:     stateSaver,
   	BedrockHandler: bedrockHandler,
   	Validator:      validator,
   	
   	// Main handler
   	Handler:        mainHandler,
   	
   	// Configuration
   	Config:         cfg,
   	HybridConfig:   hybridConfig,
   	
   	// Logger
   	Logger:         log,
   }, nil
}

// loadAWSConfig is a thin wrapper for AWS SDK v2 config loading
func loadAWSConfig(ctx context.Context, region string) (aws.Config, error) {
   // Create a context with timeout
   ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
   defer cancel()
   
   return awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
}