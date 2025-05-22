package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"

	internalConfig "ExecuteTurn1Combined/internal/config"
	"ExecuteTurn1Combined/internal/handler"
	"ExecuteTurn1Combined/internal/logger"
	"ExecuteTurn1Combined/internal/services"
	"ExecuteTurn1Combined/internal/utils"
)

// ApplicationContainer represents the strategic dependency orchestration framework
// implementing the Composition Root pattern for centralized dependency management
// while maintaining architectural flexibility for future evolution patterns
type ApplicationContainer struct {
	config    *internalConfig.Config
	logger    *logger.Logger
	awsConfig aws.Config
	handler   *handler.Handler

	// Strategic service abstractions with lifecycle management
	s3Service      services.S3StateManager
	bedrockService services.BedrockService
	dynamoService  services.DynamoDBService
	promptService  services.PromptService
}

// SystemInitializationMetrics captures critical bootstrap performance indicators
// enabling operational visibility into cold start optimization patterns and
// infrastructure readiness assessment for strategic scaling decisions
type SystemInitializationMetrics struct {
	InitializationStartTime   time.Time
	ConfigurationLoadTime     time.Duration
	AWSClientSetupTime        time.Duration
	ServiceInitializationTime time.Duration
	TotalBootstrapTime        time.Duration
	MemoryUtilization         int64
	ColdStartIndicator        bool
}

// Strategic global container instance for Lambda runtime lifecycle management
// implementing singleton pattern with thread-safe initialization to optimize
// container reuse patterns and minimize cold start performance impacts
var applicationContainer *ApplicationContainer
var initializationMetrics SystemInitializationMetrics

// init performs strategic system initialization with comprehensive error boundary
// management and performance optimization for Lambda container reuse patterns
func init() {
	initializationMetrics.InitializationStartTime = time.Now()
	initializationMetrics.ColdStartIndicator = true

	// Initialize application container with strategic error handling
	container, err := initializeApplicationContainer()
	if err != nil {
		// Strategic failure isolation - log critical error and allow Lambda
		// runtime to handle initialization failure gracefully
		log.Fatalf("CRITICAL: Application container initialization failed: %v", err)
	}

	applicationContainer = container
	initializationMetrics.TotalBootstrapTime = time.Since(initializationMetrics.InitializationStartTime)

	// Strategic performance monitoring integration
	applicationContainer.logger.Info("system_initialization_completed", map[string]interface{}{
		"bootstrap_time_ms":        initializationMetrics.TotalBootstrapTime.Milliseconds(),
		"config_load_time_ms":      initializationMetrics.ConfigurationLoadTime.Milliseconds(),
		"aws_client_setup_time_ms": initializationMetrics.AWSClientSetupTime.Milliseconds(),
		"service_init_time_ms":     initializationMetrics.ServiceInitializationTime.Milliseconds(),
		"cold_start":               initializationMetrics.ColdStartIndicator,
		"aws_region":               applicationContainer.config.AWS.Region,
		"bedrock_model":            applicationContainer.config.AWS.BedrockModel,
	})
}

// initializeApplicationContainer orchestrates comprehensive system bootstrap
// implementing strategic dependency injection patterns with proper error
// boundary management and performance optimization for Lambda runtime efficiency
func initializeApplicationContainer() (*ApplicationContainer, error) {
	configStartTime := time.Now()

	// Strategic configuration initialization with environment validation
	cfg, err := internalConfig.LoadConfiguration()
	if err != nil {
		return nil, fmt.Errorf("configuration initialization failed: %w", err)
	}

	initializationMetrics.ConfigurationLoadTime = time.Since(configStartTime)

	// Strategic logging infrastructure initialization with correlation context
	logger := logger.New("ExecuteTurn1Combined")

	logger.Info("configuration_loaded_successfully", map[string]interface{}{
		"aws_region":         cfg.AWS.Region,
		"state_bucket":       cfg.AWS.S3Bucket,
		"verification_table": cfg.AWS.DynamoDBVerificationTable,
		"bedrock_model":      cfg.AWS.BedrockModel,
		"log_level":          cfg.Logging.Level,
	})

	awsSetupStartTime := time.Now()

	// Strategic AWS configuration initialization with region and credential optimization
	awsConfig, err := initializeAWSConfiguration(cfg)
	if err != nil {
		logger.Error("aws_configuration_initialization_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("AWS configuration initialization failed: %w", err)
	}

	initializationMetrics.AWSClientSetupTime = time.Since(awsSetupStartTime)

	serviceInitStartTime := time.Now()

	// Strategic service layer initialization with dependency orchestration
	services, err := initializeServiceLayer(awsConfig, cfg, logger)
	if err != nil {
		logger.Error("service_layer_initialization_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("service layer initialization failed: %w", err)
	}

	// Strategic handler initialization with comprehensive dependency injection
	handlerInstance, err := handler.NewHandler(
		services.s3Service,
		services.bedrockService,
		services.dynamoService,
		services.promptService,
		logger,
		cfg,
	)
	if err != nil {
		logger.Error("handler_initialization_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("handler initialization failed: %w", err)
	}

	initializationMetrics.ServiceInitializationTime = time.Since(serviceInitStartTime)

	return &ApplicationContainer{
		config:         cfg,
		logger:         logger,
		awsConfig:      awsConfig,
		handler:        handlerInstance,
		s3Service:      services.s3Service,
		bedrockService: services.bedrockService,
		dynamoService:  services.dynamoService,
		promptService:  services.promptService,
	}, nil
}

// ServiceLayerComponents encapsulates initialized service dependencies
// implementing strategic abstraction patterns for clean architecture boundaries
type ServiceLayerComponents struct {
	s3Service      services.S3StateManager
	bedrockService services.BedrockService
	dynamoService  services.DynamoDBService
	promptService  services.PromptService
}

// initializeServiceLayer orchestrates comprehensive service dependency initialization
// with strategic error handling and performance optimization patterns
func initializeServiceLayer(awsConfig aws.Config, cfg *internalConfig.Config, logger *logger.Logger) (*ServiceLayerComponents, error) {
	// Strategic S3 service initialization with optimized client configuration
	s3Service, err := services.NewS3StateManager(cfg.AWS.S3Bucket)
	if err != nil {
		return nil, fmt.Errorf("S3 service initialization failed: %w", err)
	}

	// Strategic Bedrock service initialization with runtime optimization
	bedrockService, err := services.NewBedrockService(context.Background(), *cfg)
	if err != nil {
		return nil, fmt.Errorf("Bedrock service initialization failed: %w", err)
	}

	// Strategic DynamoDB service initialization with performance optimization
	dynamoService := services.NewDynamoDBService(*cfg)

	// Strategic prompt service initialization with template management
	promptService, err := services.NewPromptService(cfg.Prompts.TemplateVersion, logger)
	if err != nil {
		return nil, fmt.Errorf("Prompt service initialization failed: %w", err)
	}

	return &ServiceLayerComponents{
		s3Service:      s3Service,
		bedrockService: bedrockService,
		dynamoService:  dynamoService,
		promptService:  promptService,
	}, nil
}

// initializeAWSConfiguration establishes strategic AWS SDK configuration
// with comprehensive credential management and regional optimization patterns
func initializeAWSConfiguration(cfg *internalConfig.Config) (aws.Config, error) {
	// Strategic AWS configuration with explicit region specification
	awsConfig, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.AWS.Region),
		config.WithRetryMaxAttempts(cfg.Processing.MaxRetries),
		config.WithRetryMode(aws.RetryModeAdaptive),
	)
	if err != nil {
		return aws.Config{}, fmt.Errorf("AWS configuration loading failed: %w", err)
	}

	// Strategic credential validation with security boundary enforcement
	credentials, err := awsConfig.Credentials.Retrieve(context.Background())
	if err != nil {
		return aws.Config{}, fmt.Errorf("AWS credentials validation failed: %w", err)
	}

	// Strategic security validation without credential exposure
	if credentials.AccessKeyID == "" {
		return aws.Config{}, fmt.Errorf("AWS credentials validation failed: empty access key")
	}

	return awsConfig, nil
}

// LambdaExecutionContext encapsulates comprehensive execution environment
// context providing strategic visibility into runtime characteristics and
// performance optimization opportunities for operational excellence
type LambdaExecutionContext struct {
	RequestID        string           `json:"requestId"`
	FunctionName     string           `json:"functionName"`
	FunctionVersion  string           `json:"functionVersion"`
	MemoryLimitMB    int              `json:"memoryLimitMB"`
	RemainingTimeMS  int64            `json:"remainingTimeMS"`
	ColdStart        bool             `json:"coldStart"`
	ExecutionMetrics ExecutionMetrics `json:"executionMetrics"`
	CorrelationID    string           `json:"correlationId"`
}

// ExecutionMetrics captures strategic performance indicators for
// operational visibility and continuous optimization patterns
type ExecutionMetrics struct {
	HandlerStartTime   time.Time     `json:"handlerStartTime"`
	ProcessingDuration time.Duration `json:"processingDuration"`
	MemoryUtilization  int64         `json:"memoryUtilization"`
	ColdStartOverhead  time.Duration `json:"coldStartOverhead"`
}

// HandleRequest represents the strategic Lambda execution entry point
// implementing comprehensive request lifecycle management with performance
// monitoring, error boundary management, and operational observability
func HandleRequest(ctx context.Context, event json.RawMessage) (interface{}, error) {
	executionStartTime := time.Now()

	// Strategic Lambda context extraction for operational visibility
	lambdaCtx, exists := lambdacontext.FromContext(ctx)
	if !exists {
		applicationContainer.logger.Warn("lambda_context_extraction_failed", map[string]interface{}{
			"execution_context": "non_lambda_environment",
		})
	}

	// Strategic execution context construction with performance tracking
	executionContext := &LambdaExecutionContext{
		ColdStart: initializationMetrics.ColdStartIndicator,
		ExecutionMetrics: ExecutionMetrics{
			HandlerStartTime:  executionStartTime,
			ColdStartOverhead: initializationMetrics.TotalBootstrapTime,
		},
		CorrelationID: generateCorrelationID(),
	}

	if exists {
		executionContext.RequestID = lambdaCtx.AwsRequestID
		executionContext.FunctionName = lambdaCtx.InvokedFunctionArn
		executionContext.MemoryLimitMB = lambdaCtx.MemoryLimitInMB
		deadline, hasDeadline := ctx.Deadline()
		if hasDeadline {
			executionContext.RemainingTimeMS = time.Until(deadline).Milliseconds()
		}
	}

	// Strategic cold start optimization - mark subsequent executions as warm
	if initializationMetrics.ColdStartIndicator {
		initializationMetrics.ColdStartIndicator = false
	}

	// Strategic correlation context injection for distributed tracing
	enrichedCtx := utils.WithCorrelationID(ctx, executionContext.CorrelationID)

	applicationContainer.logger.Info("execution_context_established", map[string]interface{}{
		"correlation_id":         executionContext.CorrelationID,
		"request_id":             executionContext.RequestID,
		"function_name":          executionContext.FunctionName,
		"memory_limit_mb":        executionContext.MemoryLimitMB,
		"remaining_time_ms":      executionContext.RemainingTimeMS,
		"cold_start":             executionContext.ColdStart,
		"cold_start_overhead_ms": executionContext.ExecutionMetrics.ColdStartOverhead.Milliseconds(),
	})

	// Strategic handler execution with comprehensive error boundary management
	response, err := applicationContainer.handler.HandleTurn1Combined(enrichedCtx, event)

	executionContext.ExecutionMetrics.ProcessingDuration = time.Since(executionStartTime)

	// Strategic execution metrics collection for operational optimization
	applicationContainer.logger.Info("execution_completed", map[string]interface{}{
		"correlation_id":          executionContext.CorrelationID,
		"processing_duration_ms":  executionContext.ExecutionMetrics.ProcessingDuration.Milliseconds(),
		"success":                 err == nil,
		"total_execution_time_ms": time.Since(executionStartTime).Milliseconds(),
	})

	if err != nil {
		// Strategic error context preservation for operational debugging
		applicationContainer.logger.Error("execution_failed", map[string]interface{}{
			"correlation_id":    executionContext.CorrelationID,
			"error":             err.Error(),
			"execution_time_ms": executionContext.ExecutionMetrics.ProcessingDuration.Milliseconds(),
		})
		return nil, err
	}

	return response, nil
}

// generateCorrelationID creates strategic correlation identifiers for
// distributed request tracing and operational debugging capabilities
func generateCorrelationID() string {
	return fmt.Sprintf("turn1-%d-%d",
		time.Now().UnixNano()/int64(time.Millisecond),
		os.Getpid(),
	)
}

// main represents the strategic Lambda runtime integration point
// implementing clean shutdown patterns and operational lifecycle management
func main() {
	// Strategic runtime validation with comprehensive error handling
	if applicationContainer == nil {
		log.Fatal("CRITICAL: Application container not initialized - system cannot proceed")
	}

	applicationContainer.logger.Info("lambda_runtime_initialization", map[string]interface{}{
		"function_ready":         true,
		"initialization_time_ms": initializationMetrics.TotalBootstrapTime.Milliseconds(),
		"aws_region":             applicationContainer.config.AWS.Region,
		"bedrock_model":          applicationContainer.config.AWS.BedrockModel,
		"service_health":         "operational",
	})

	// Strategic Lambda runtime registration with handler delegation
	lambda.Start(HandleRequest)
}
