// Package main implements the strategic Lambda entry point with enhanced
// deterministic control architecture for Bedrock integration
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync/atomic"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"

	internalConfig "workflow-function/ExecuteTurn1Combined/internal/config"
	"workflow-function/ExecuteTurn1Combined/internal/handler"
	"workflow-function/ExecuteTurn1Combined/internal/services"
	"workflow-function/ExecuteTurn1Combined/internal/utils"

	// Strategic addition: Local bedrock package for deterministic control
	localBedrock "workflow-function/ExecuteTurn1Combined/internal/bedrock"

	// Shared packages with strategic integration
	sharedBedrock "workflow-function/shared/bedrock"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
)

// ApplicationContainer represents the enhanced dependency orchestration framework
// with strategic separation between shared infrastructure and local control
type ApplicationContainer struct {
	config    *internalConfig.Config
	logger    logger.Logger
	awsConfig aws.Config
	handler   *handler.Handler

	// Strategic service abstractions with deterministic control patterns
	s3Service      services.S3StateManager
	bedrockService services.BedrockService // Now using local implementation
	dynamoService  services.DynamoDBService
	promptService  services.PromptService
}

// SystemInitializationMetrics captures comprehensive bootstrap telemetry
type SystemInitializationMetrics struct {
	InitializationStartTime   time.Time
	ConfigurationLoadTime     time.Duration
	AWSClientSetupTime        time.Duration
	ServiceInitializationTime time.Duration
	TotalBootstrapTime        time.Duration
	MemoryUtilization         int64
	ColdStartIndicator        bool
	// Strategic addition: Bedrock initialization metrics
	BedrockSetupTime    time.Duration
	ArchitecturePattern string
}

// Global container instance with enhanced lifecycle management
var applicationContainer *ApplicationContainer
var initializationMetrics SystemInitializationMetrics

// init performs strategic system initialization with deterministic control architecture
func init() {
	initializationMetrics.InitializationStartTime = time.Now()
	initializationMetrics.ColdStartIndicator = true
	initializationMetrics.ArchitecturePattern = "deterministic_control"

	// Initialize application container with enhanced error boundaries
	container, err := initializeApplicationContainer()
	if err != nil {
		// Strategic error categorization for operational visibility
		if errors.IsConfigError(err) {
			if workflowErr, ok := err.(*errors.WorkflowError); ok {
				errJSON, _ := json.Marshal(map[string]interface{}{
					"level":        "ERROR",
					"msg":          "config_load_failed",
					"errorType":    string(workflowErr.Type),
					"errorCode":    workflowErr.Code,
					"error":        workflowErr.Message,
					"var":          workflowErr.Details["variable"],
					"severity":     "ERROR",
					"architecture": "deterministic_control",
				})
				fmt.Fprintf(os.Stderr, "%s\n", errJSON)
			}
			log.Fatalf("CRITICAL: Configuration error: %v", err.Error())
		}

		criticalErr := errors.NewInternalError("application_bootstrap", err).
			WithContext("stage", "initialization").
			WithContext("architecture", "deterministic_control")

		log.Fatalf("CRITICAL: Application container initialization failed: %v", criticalErr.Error())
	}

	applicationContainer = container
	initializationMetrics.TotalBootstrapTime = time.Since(initializationMetrics.InitializationStartTime)

	// Enhanced initialization telemetry with architectural insights
	applicationContainer.logger.Info("system_initialization_completed", map[string]interface{}{
		"bootstrap_time_ms":        initializationMetrics.TotalBootstrapTime.Milliseconds(),
		"config_load_time_ms":      initializationMetrics.ConfigurationLoadTime.Milliseconds(),
		"aws_client_setup_time_ms": initializationMetrics.AWSClientSetupTime.Milliseconds(),
		"service_init_time_ms":     initializationMetrics.ServiceInitializationTime.Milliseconds(),
		"bedrock_setup_time_ms":    initializationMetrics.BedrockSetupTime.Milliseconds(),
		"cold_start":               initializationMetrics.ColdStartIndicator,
		"architecture_pattern":     initializationMetrics.ArchitecturePattern,
		"aws_region":               applicationContainer.config.AWS.Region,
		"bedrock_model":            applicationContainer.config.AWS.BedrockModel,
		"template_base_path":       applicationContainer.config.Prompts.TemplateBasePath,
		"template_version":         applicationContainer.config.Prompts.TemplateVersion,
		"services_initialized": map[string]interface{}{
			"s3_service":      applicationContainer.s3Service != nil,
			"bedrock_service": applicationContainer.bedrockService != nil,
			"dynamo_service":  applicationContainer.dynamoService != nil,
			"prompt_service":  applicationContainer.promptService != nil,
			"architecture":    "deterministic_local_control",
		},
	})
}

// initializeApplicationContainer orchestrates strategic system bootstrap with enhanced architecture
func initializeApplicationContainer() (*ApplicationContainer, error) {
	configStartTime := time.Now()

	// Strategic configuration initialization
	cfg, err := internalConfig.LoadConfiguration()
	if err != nil {
		if _, ok := err.(*errors.WorkflowError); ok {
			return nil, err
		}
		return nil, errors.WrapError(err, errors.ErrorTypeConfig,
			"configuration initialization failed", false)
	}

	initializationMetrics.ConfigurationLoadTime = time.Since(configStartTime)

	// Initialize structured logger with architectural context
	logger := logger.New("ExecuteTurn1Combined", "main").
		WithFields(map[string]interface{}{
			"architecture": "deterministic_control",
			"version":      "2.1.0", // Updated to match schema version
		})

	logger.Info("configuration_loaded_successfully", map[string]interface{}{
		"aws_region":         cfg.AWS.Region,
		"state_bucket":       cfg.AWS.S3Bucket,
		"verification_table": cfg.AWS.DynamoDBVerificationTable,
		"bedrock_model":      cfg.AWS.BedrockModel,
		"log_level":          cfg.Logging.Level,
		"architecture":       "deterministic_control",
	})

	awsSetupStartTime := time.Now()

	// Strategic AWS configuration with enhanced resilience
	awsConfig, err := initializeAWSConfiguration(cfg, logger)
	if err != nil {
		logger.Error("aws_configuration_initialization_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, errors.WrapError(err, errors.ErrorTypeAPI,
			"AWS configuration initialization failed", false)
	}

	initializationMetrics.AWSClientSetupTime = time.Since(awsSetupStartTime)

	serviceInitStartTime := time.Now()

	// Strategic service layer initialization with deterministic architecture
	services, err := initializeServiceLayerWithLocalBedrock(awsConfig, cfg, logger)
	if err != nil {
		logger.Error("service_layer_initialization_failed", map[string]interface{}{
			"error":        err.Error(),
			"architecture": "deterministic_control",
		})
		return nil, errors.WrapError(err, errors.ErrorTypeInternal,
			"service layer initialization failed", false)
	}

	initializationMetrics.ServiceInitializationTime = time.Since(serviceInitStartTime)

	// Strategic handler initialization with enhanced services
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
		return nil, errors.WrapError(err, errors.ErrorTypeInternal,
			"handler initialization failed", false)
	}

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

// ServiceLayerComponents encapsulates service dependencies with architectural metadata
type ServiceLayerComponents struct {
	s3Service      services.S3StateManager
	bedrockService services.BedrockService
	dynamoService  services.DynamoDBService
	promptService  services.PromptService
}

// initializeServiceLayerWithLocalBedrock implements strategic service initialization
// with deterministic local Bedrock control architecture
func initializeServiceLayerWithLocalBedrock(awsConfig aws.Config, cfg *internalConfig.Config, logger logger.Logger) (*ServiceLayerComponents, error) {
	// S3 service initialization
	s3Service, err := services.NewS3StateManager(*cfg, logger)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrorTypeS3,
			"S3 service initialization failed", false)
	}

	// Strategic Bedrock service initialization with local control pattern
	bedrockInitStart := time.Now()

	// Phase 1: Initialize shared Bedrock client for infrastructure
	sharedClientConfig := sharedBedrock.CreateClientConfig(
		cfg.AWS.Region,
		cfg.AWS.AnthropicVersion,
		cfg.Processing.MaxTokens,
		cfg.Processing.ThinkingType,
		cfg.Processing.BudgetTokens,
	)

	sharedClient, err := sharedBedrock.NewBedrockClient(
		context.Background(),
		cfg.AWS.BedrockModel,
		sharedClientConfig,
	)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrorTypeBedrock,
			"shared Bedrock client initialization failed", false).
			WithContext("architecture", "shared_infrastructure")
	}

	// Phase 2: Configure local Bedrock client
	localConfig := &localBedrock.Config{
		ModelID:          cfg.AWS.BedrockModel,
		MaxTokens:        cfg.Processing.MaxTokens,
		Temperature:      0.7, // Strategic default for balanced output
		Timeout:          time.Duration(cfg.Processing.BedrockCallTimeoutSec) * time.Second,
		Region:           cfg.AWS.Region,
		AnthropicVersion: cfg.AWS.AnthropicVersion,
	}

	localClient := localBedrock.NewClient(sharedClient, localConfig, logger)

	// Phase 3: Validate configuration
	if err := localClient.ValidateConfiguration(); err != nil {
		return nil, errors.WrapError(err, errors.ErrorTypeBedrock,
			"Bedrock configuration validation failed", false).
			WithContext("architecture", "local_validation")
	}

	// Phase 4: Create service wrapper with local client
	bedrockService := services.NewBedrockServiceWithLocalClient(localClient, *cfg, logger)

	initializationMetrics.BedrockSetupTime = time.Since(bedrockInitStart)

	logger.Info("bedrock_service_initialized", map[string]interface{}{
		"architecture":        "deterministic_local_control",
		"setup_time_ms":       initializationMetrics.BedrockSetupTime.Milliseconds(),
		"connect_timeout_sec": cfg.Processing.BedrockConnectTimeoutSec,
		"call_timeout_sec":    cfg.Processing.BedrockCallTimeoutSec,
		"model":               cfg.AWS.BedrockModel,
		"max_tokens":          cfg.Processing.MaxTokens,
		"operational_metrics": localClient.GetOperationalMetrics(),
	})

	// DynamoDB service initialization
	dynamoService, err := services.NewDynamoDBService(cfg)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrorTypeDynamoDB,
			"DynamoDB service initialization failed", false)
	}

	// Prompt service initialization
	promptService, err := services.NewPromptService(cfg, logger)
	if err != nil {
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

// initializeAWSConfiguration remains unchanged - provides base AWS SDK configuration
func initializeAWSConfiguration(cfg *internalConfig.Config, logger logger.Logger) (aws.Config, error) {
	awsConfig, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(cfg.AWS.Region),
		awsconfig.WithRetryMaxAttempts(cfg.Processing.MaxRetries),
		awsconfig.WithRetryMode(aws.RetryModeAdaptive),
	)
	if err != nil {
		return aws.Config{}, errors.WrapError(err, errors.ErrorTypeAPI,
			"AWS configuration loading failed", false).
			WithContext("region", cfg.AWS.Region).
			WithContext("max_retries", cfg.Processing.MaxRetries)
	}

	credentials, err := awsConfig.Credentials.Retrieve(context.Background())
	if err != nil {
		return aws.Config{}, errors.WrapError(err, errors.ErrorTypeAPI,
			"AWS credentials validation failed", false).
			WithContext("region", cfg.AWS.Region)
	}

	if credentials.AccessKeyID == "" {
		return aws.Config{}, errors.NewValidationError(
			"AWS credentials validation failed: empty access key",
			map[string]interface{}{
				"component": "aws_config",
				"region":    cfg.AWS.Region,
			})
	}

	return awsConfig, nil
}

// LambdaExecutionContext with enhanced architectural metadata
type LambdaExecutionContext struct {
	RequestID        string           `json:"requestId"`
	FunctionName     string           `json:"functionName"`
	FunctionVersion  string           `json:"functionVersion"`
	MemoryLimitMB    int              `json:"memoryLimitMB"`
	RemainingTimeMS  int64            `json:"remainingTimeMS"`
	ColdStart        bool             `json:"coldStart"`
	ExecutionMetrics ExecutionMetrics `json:"executionMetrics"`
	CorrelationID    string           `json:"correlationId"`
	Architecture     string           `json:"architecture"`
}

// ExecutionMetrics with architectural insights
type ExecutionMetrics struct {
	HandlerStartTime   time.Time     `json:"handlerStartTime"`
	ProcessingDuration time.Duration `json:"processingDuration"`
	MemoryUtilization  int64         `json:"memoryUtilization"`
	ColdStartOverhead  time.Duration `json:"coldStartOverhead"`
	ArchitectureMode   string        `json:"architectureMode"`
}

// HandleRequest implements the strategic Lambda execution entry point
func HandleRequest(ctx context.Context, event json.RawMessage) (interface{}, error) {
	executionStartTime := time.Now()

	lambdaCtx, exists := lambdacontext.FromContext(ctx)
	if !exists {
		applicationContainer.logger.Warn("lambda_context_extraction_failed", map[string]interface{}{
			"execution_context": "non_lambda_environment",
		})
	}

	// Enhanced execution context with architectural metadata
	executionContext := &LambdaExecutionContext{
		ColdStart: initializationMetrics.ColdStartIndicator,
		ExecutionMetrics: ExecutionMetrics{
			HandlerStartTime:  executionStartTime,
			ColdStartOverhead: initializationMetrics.TotalBootstrapTime,
			ArchitectureMode:  "deterministic_control",
		},
		CorrelationID: generateCorrelationID(),
		Architecture:  "deterministic_local_bedrock",
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

	if initializationMetrics.ColdStartIndicator {
		initializationMetrics.ColdStartIndicator = false
	}

	contextLogger := applicationContainer.logger.WithCorrelationId(executionContext.CorrelationID)
	enrichedCtx := utils.WithCorrelationID(ctx, executionContext.CorrelationID)

	contextLogger.Info("execution_context_established", map[string]interface{}{
		"request_id":             executionContext.RequestID,
		"function_name":          executionContext.FunctionName,
		"memory_limit_mb":        executionContext.MemoryLimitMB,
		"remaining_time_ms":      executionContext.RemainingTimeMS,
		"cold_start":             executionContext.ColdStart,
		"cold_start_overhead_ms": executionContext.ExecutionMetrics.ColdStartOverhead.Milliseconds(),
		"architecture":           executionContext.Architecture,
	})

	contextLogger.LogReceivedEvent(event)

	// Execute handler with deterministic architecture
	response, err := applicationContainer.handler.HandleTurn1Combined(enrichedCtx, event)

	executionContext.ExecutionMetrics.ProcessingDuration = time.Since(executionStartTime)

	contextLogger.Info("execution_completed", map[string]interface{}{
		"processing_duration_ms":  executionContext.ExecutionMetrics.ProcessingDuration.Milliseconds(),
		"success":                 err == nil,
		"total_execution_time_ms": time.Since(executionStartTime).Milliseconds(),
		"architecture":            "deterministic_control",
	})

	if err != nil {
		if workflowErr, ok := err.(*errors.WorkflowError); ok {
			enrichedErr := workflowErr.
				WithContext("execution_time_ms", executionContext.ExecutionMetrics.ProcessingDuration.Milliseconds()).
				WithContext("correlation_id", executionContext.CorrelationID).
				WithContext("architecture", "deterministic_control")

			contextLogger.Error("execution_failed", map[string]interface{}{
				"error_type":   string(enrichedErr.Type),
				"error_code":   enrichedErr.Code,
				"retryable":    enrichedErr.Retryable,
				"severity":     string(enrichedErr.Severity),
				"api_source":   string(enrichedErr.APISource),
				"architecture": "deterministic_control",
			})
			return nil, enrichedErr
		} else {
			wrappedErr := errors.WrapError(err, errors.ErrorTypeInternal,
				"execution failed", false).
				WithContext("correlation_id", executionContext.CorrelationID).
				WithContext("architecture", "deterministic_control")

			contextLogger.Error("execution_failed", map[string]interface{}{
				"error":        err.Error(),
				"architecture": "deterministic_control",
			})
			return nil, wrappedErr
		}
	}

	contextLogger.LogOutputEvent(response)

	return response, nil
}

// Global counter for correlation ID uniqueness
var correlationCounter uint64

// generateCorrelationID creates strategic correlation identifiers
func generateCorrelationID() string {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	randomBytes := make([]byte, 4)
	if _, err := rand.Read(randomBytes); err != nil {
		randomBytes = []byte{byte(timestamp), byte(timestamp >> 8), byte(timestamp >> 16), byte(timestamp >> 24)}
	}
	randomHex := hex.EncodeToString(randomBytes)
	counter := atomic.AddUint64(&correlationCounter, 1)

	return fmt.Sprintf("turn1-%d-%s-%d", timestamp, randomHex, counter)
}

// main represents the strategic Lambda bootstrap entry point
func main() {
	lambda.Start(HandleRequest)
}
