// cmd/main.go - CORRECTED WITH PROPER FLUENT INTERFACE USAGE
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"

	internalConfig "workflow-function/ExecuteTurn1Combined/internal/config"
	"workflow-function/ExecuteTurn1Combined/internal/handler"
	"workflow-function/ExecuteTurn1Combined/internal/services"
	"workflow-function/ExecuteTurn1Combined/internal/utils"
	
	// Using shared packages correctly
	"workflow-function/shared/logger"
	"workflow-function/shared/errors"
)

// ApplicationContainer represents the strategic dependency orchestration framework
type ApplicationContainer struct {
	config    *internalConfig.Config
	// CORRECT: Logger is an interface, not a pointer
	logger    logger.Logger
	awsConfig aws.Config
	handler   *handler.Handler

	// Strategic service abstractions with lifecycle management
	s3Service      services.S3StateManager
	bedrockService services.BedrockService
	dynamoService  services.DynamoDBService
	promptService  services.PromptService
}

// SystemInitializationMetrics captures critical bootstrap performance indicators
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
var applicationContainer *ApplicationContainer
var initializationMetrics SystemInitializationMetrics

// init performs strategic system initialization with comprehensive error boundary
func init() {
	initializationMetrics.InitializationStartTime = time.Now()
	initializationMetrics.ColdStartIndicator = true

	// Initialize application container with strategic error handling
	container, err := initializeApplicationContainer()
	if err != nil {
		// FIXED: Remove the unnecessary type assertion
		// The fluent interface already returns the correct type
		criticalErr := errors.NewInternalError("application_bootstrap", err).
			WithContext("stage", "initialization")
		// No need for (*errors.WorkflowError) - it's already that type!
		
		log.Fatalf("CRITICAL: Application container initialization failed: %v", criticalErr.Error())
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
func initializeApplicationContainer() (*ApplicationContainer, error) {
	configStartTime := time.Now()

	// Strategic configuration initialization with environment validation
	cfg, err := internalConfig.LoadConfiguration()
	if err != nil {
		// FIXED: Clean fluent interface usage - no type assertions needed
		return nil, errors.WrapError(err, errors.ErrorTypeValidation, 
			"configuration initialization failed", false)
	}

	initializationMetrics.ConfigurationLoadTime = time.Since(configStartTime)

	// CORRECT: Using shared logger constructor with service and function names
	logger := logger.New("ExecuteTurn1Combined", "main")

	logger.Info("configuration_loaded_successfully", map[string]interface{}{
		"aws_region":         cfg.AWS.Region,
		"state_bucket":       cfg.AWS.S3Bucket,
		"verification_table": cfg.AWS.DynamoDBVerificationTable,
		"bedrock_model":      cfg.AWS.BedrockModel,
		"log_level":          cfg.Logging.Level,
	})

	awsSetupStartTime := time.Now()

	// Strategic AWS configuration initialization
	awsConfig, err := initializeAWSConfiguration(cfg, logger)
	if err != nil {
		logger.Error("aws_configuration_initialization_failed", map[string]interface{}{
			"error": err.Error(),
		})
		// FIXED: Clean fluent interface - the chain returns the right type
		return nil, errors.WrapError(err, errors.ErrorTypeAPI, 
			"AWS configuration initialization failed", false)
	}

	initializationMetrics.AWSClientSetupTime = time.Since(awsSetupStartTime)

	serviceInitStartTime := time.Now()

	// Strategic service layer initialization
	services, err := initializeServiceLayer(awsConfig, cfg, logger)
	if err != nil {
		logger.Error("service_layer_initialization_failed", map[string]interface{}{
			"error": err.Error(),
		})
		// FIXED: Fluent interface works perfectly without type assertions
		return nil, errors.WrapError(err, errors.ErrorTypeInternal, 
			"service layer initialization failed", false)
	}

	// Strategic handler initialization
	handlerInstance, err := handler.NewHandler(
		services.s3Service,
		services.bedrockService,
		services.dynamoService,
		services.promptService,
		logger, // Passing shared logger interface correctly
		cfg,
	)
	if err != nil {
		logger.Error("handler_initialization_failed", map[string]interface{}{
			"error": err.Error(),
		})
		// FIXED: Clean method chaining - Go's type system handles this elegantly
		return nil, errors.WrapError(err, errors.ErrorTypeInternal, 
			"handler initialization failed", false)
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
type ServiceLayerComponents struct {
	s3Service      services.S3StateManager
	bedrockService services.BedrockService
	dynamoService  services.DynamoDBService
	promptService  services.PromptService
}

// initializeServiceLayer orchestrates comprehensive service dependency initialization
func initializeServiceLayer(awsConfig aws.Config, cfg *internalConfig.Config, logger logger.Logger) (*ServiceLayerComponents, error) {
	// Strategic S3 service initialization
	s3Service, err := services.NewS3StateManager(cfg.AWS.S3Bucket)
	if err != nil {
		// FIXED: Simple, clean error creation - no type gymnastics needed
		return nil, errors.WrapError(err, errors.ErrorTypeS3, 
			"S3 service initialization failed", false)
	}

	// Strategic Bedrock service initialization
	bedrockService, err := services.NewBedrockService(context.Background(), *cfg)
	if err != nil {
		// FIXED: The fluent interface is designed to work naturally
		return nil, errors.WrapError(err, errors.ErrorTypeBedrock, 
			"Bedrock service initialization failed", false)
	}

	// Strategic DynamoDB service initialization
	dynamoService, err := services.NewDynamoDBService(cfg)
	if err != nil {
		// FIXED: The fluent interface is designed to work naturally
		return nil, errors.WrapError(err, errors.ErrorTypeDynamoDB, 
			"DynamoDB service initialization failed", false)
	}

	// Strategic prompt service initialization - pass the logger correctly
	promptService, err := services.NewPromptService(cfg.Prompts.TemplateVersion, logger)
	if err != nil {
		// FIXED: Method chaining works seamlessly
		return nil, errors.WrapError(err, errors.ErrorTypeInternal, 
			"Prompt service initialization failed", false)
	}

	return &ServiceLayerComponents{
		s3Service:      s3Service,
		bedrockService: bedrockService,
		dynamoService:  dynamoService,
		promptService:  promptService,
	}, nil
}

// initializeAWSConfiguration establishes strategic AWS SDK configuration
func initializeAWSConfiguration(cfg *internalConfig.Config, logger logger.Logger) (aws.Config, error) {
	// Strategic AWS configuration with explicit region specification
	awsConfig, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.AWS.Region),
		config.WithRetryMaxAttempts(cfg.Processing.MaxRetries),
		config.WithRetryMode(aws.RetryModeAdaptive),
	)
	if err != nil {
		// FIXED: Clean error creation with proper context
		return aws.Config{}, errors.WrapError(err, errors.ErrorTypeAPI, 
			"AWS configuration loading failed", false).
			WithContext("region", cfg.AWS.Region).
			WithContext("max_retries", cfg.Processing.MaxRetries)
	}

	// Strategic credential validation
	credentials, err := awsConfig.Credentials.Retrieve(context.Background())
	if err != nil {
		// FIXED: Fluent interface chain - each method returns the right type
		return aws.Config{}, errors.WrapError(err, errors.ErrorTypeAPI, 
			"AWS credentials validation failed", false).
			WithContext("region", cfg.AWS.Region)
	}

	// Strategic security validation without credential exposure
	if credentials.AccessKeyID == "" {
		// FIXED: Factory function creates the right type directly
		return aws.Config{}, errors.NewValidationError(
			"AWS credentials validation failed: empty access key",
			map[string]interface{}{
				"component": "aws_config",
				"region":    cfg.AWS.Region,
			})
	}

	return awsConfig, nil
}

// LambdaExecutionContext encapsulates comprehensive execution environment context
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

// ExecutionMetrics captures strategic performance indicators
type ExecutionMetrics struct {
	HandlerStartTime   time.Time     `json:"handlerStartTime"`
	ProcessingDuration time.Duration `json:"processingDuration"`
	MemoryUtilization  int64         `json:"memoryUtilization"`
	ColdStartOverhead  time.Duration `json:"coldStartOverhead"`
}

// HandleRequest represents the strategic Lambda execution entry point
func HandleRequest(ctx context.Context, event json.RawMessage) (interface{}, error) {
	executionStartTime := time.Now()

	// Strategic Lambda context extraction
	lambdaCtx, exists := lambdacontext.FromContext(ctx)
	if !exists {
		applicationContainer.logger.Warn("lambda_context_extraction_failed", map[string]interface{}{
			"execution_context": "non_lambda_environment",
		})
	}

	// Strategic execution context construction
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
		if memoryStr := os.Getenv("AWS_LAMBDA_FUNCTION_MEMORY_SIZE"); memoryStr != "" {
			var memoryLimit int
			if _, err := fmt.Sscanf(memoryStr, "%d", &memoryLimit); err == nil {
				executionContext.MemoryLimitMB = memoryLimit
			}
		}
		deadline, hasDeadline := ctx.Deadline()
		if hasDeadline {
			executionContext.RemainingTimeMS = time.Until(deadline).Milliseconds()
		}
	}

	// Strategic cold start optimization
	if initializationMetrics.ColdStartIndicator {
		initializationMetrics.ColdStartIndicator = false
	}

	// Create logger with correlation ID using shared logger's fluent interface
	contextLogger := applicationContainer.logger.WithCorrelationId(executionContext.CorrelationID)
	
	// Strategic correlation context injection
	enrichedCtx := utils.WithCorrelationID(ctx, executionContext.CorrelationID)

	contextLogger.Info("execution_context_established", map[string]interface{}{
		"request_id":             executionContext.RequestID,
		"function_name":          executionContext.FunctionName,
		"memory_limit_mb":        executionContext.MemoryLimitMB,
		"remaining_time_ms":      executionContext.RemainingTimeMS,
		"cold_start":             executionContext.ColdStart,
		"cold_start_overhead_ms": executionContext.ExecutionMetrics.ColdStartOverhead.Milliseconds(),
	})

	// Log the received event using shared logger's event logging capability
	contextLogger.LogReceivedEvent(event)

	// Strategic handler execution with comprehensive error boundary management
	response, err := applicationContainer.handler.HandleTurn1Combined(enrichedCtx, event)

	executionContext.ExecutionMetrics.ProcessingDuration = time.Since(executionStartTime)

	// Strategic execution metrics collection
	contextLogger.Info("execution_completed", map[string]interface{}{
		"processing_duration_ms":  executionContext.ExecutionMetrics.ProcessingDuration.Milliseconds(),
		"success":                 err == nil,
		"total_execution_time_ms": time.Since(executionStartTime).Milliseconds(),
	})

	if err != nil {
		// FIXED: Proper error type checking and enrichment without unnecessary assertions
		if workflowErr, ok := err.(*errors.WorkflowError); ok {
			// Enrich the error with execution context using fluent interface
			enrichedErr := workflowErr.WithContext("execution_time_ms", 
				executionContext.ExecutionMetrics.ProcessingDuration.Milliseconds()).
				WithContext("correlation_id", executionContext.CorrelationID)
			
			contextLogger.Error("execution_failed", map[string]interface{}{
				"error_type":    string(enrichedErr.Type),
				"error_code":    enrichedErr.Code,
				"retryable":     enrichedErr.Retryable,
				"severity":      string(enrichedErr.Severity),
				"api_source":    string(enrichedErr.APISource),
			})
			return nil, enrichedErr
		} else {
			// FIXED: Clean error wrapping without type assertions
			wrappedErr := errors.WrapError(err, errors.ErrorTypeInternal, 
				"execution failed", false).
				WithContext("correlation_id", executionContext.CorrelationID)
			
			contextLogger.Error("execution_failed", map[string]interface{}{
				"error": err.Error(),
			})
			return nil, wrappedErr
		}
	}

	// Log the output event using shared logger's event logging capability
	contextLogger.LogOutputEvent(response)

	return response, nil
}

// generateCorrelationID creates strategic correlation identifiers
func generateCorrelationID() string {
	return fmt.Sprintf("turn1-%d", time.Now().UnixNano()/int64(time.Millisecond))
}

// main represents the strategic Lambda bootstrap entry point
func main() {
	// Strategic Lambda handler registration
	lambda.Start(HandleRequest)
}