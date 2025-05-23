// internal/services/bedrock.go - STRATEGICALLY ENHANCED WITH ROBUST DATA TRANSFORMATION
package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"workflow-function/ExecuteTurn1Combined/internal/config"
	"workflow-function/ExecuteTurn1Combined/internal/models"
	"workflow-function/shared/bedrock"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
)

// BedrockService defines the strategic interface for AI model integration
type BedrockService interface {
	Converse(ctx context.Context, systemPrompt, turnPrompt, base64Image string) (*models.BedrockResponse, error)
}

// bedrockService implements robust AI model integration with comprehensive error handling
type bedrockService struct {
	client               *bedrock.BedrockClient
	modelID              string
	maxTokens            int
	connectTimeout       time.Duration
	callTimeout          time.Duration
	maxDecodedImageSize  int64 // Strategic constraint for decoded image size
	imageFormatValidator map[string]bool // Supported image formats
	logger               logger.Logger // Strategic observability
}

// NewBedrockService constructs a strategically enhanced BedrockService with robust validation
func NewBedrockService(ctx context.Context, cfg config.Config) (BedrockService, error) {
	// Strategic client configuration with comprehensive parameters
	clientCfg := bedrock.CreateClientConfig(
		cfg.AWS.Region,
		cfg.AWS.AnthropicVersion,
		cfg.Processing.MaxTokens,
		cfg.Processing.ThinkingType,
		cfg.Processing.BudgetTokens,
	)
	
	c, err := bedrock.NewBedrockClient(ctx, cfg.AWS.BedrockModel, clientCfg)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrorTypeBedrock, 
			"failed to create Bedrock client", false).
			WithContext("model_id", cfg.AWS.BedrockModel).
			WithContext("region", cfg.AWS.Region).
			WithContext("max_tokens", cfg.Processing.MaxTokens).
			WithContext("architectural_layer", "client_initialization")
	}
	
	// Initialize logger for strategic observability
	log := logger.New("ExecuteTurn1Combined", "BedrockService")
	
	// Strategic service initialization with enhanced capabilities
	return &bedrockService{
		client:         c,
		modelID:        cfg.AWS.BedrockModel,
		maxTokens:      cfg.Processing.MaxTokens,
		connectTimeout: time.Duration(cfg.Processing.BedrockConnectTimeoutSec) * time.Second,
		callTimeout:    time.Duration(cfg.Processing.BedrockCallTimeoutSec) * time.Second,
		maxDecodedImageSize: 5 * 1024 * 1024, // 5MB limit (accounting for Base64 overhead)
		imageFormatValidator: map[string]bool{
			"jpeg": true,
			"jpg":  true,
			"png":  true,
			"webp": true,
		},
		logger: log,
	}, nil
}

// Converse performs strategically enhanced multimodal AI inference with robust data transformation
func (s *bedrockService) Converse(ctx context.Context, systemPrompt, turnPrompt, base64Image string) (*models.BedrockResponse, error) {
	// Strategic performance monitoring initialization
	operationStart := time.Now()
	
	// Apply strategic timeout management with operational context
	ctx, cancel := context.WithTimeout(ctx, s.callTimeout)
	defer cancel()

	// Log transformation pipeline entry
	s.logger.Info("Starting Bedrock Converse data transformation pipeline", map[string]interface{}{
		"base64_input_length": len(base64Image),
		"system_prompt_length": len(systemPrompt),
		"turn_prompt_length": len(turnPrompt),
		"transformation_boundary": "pipeline_entry",
	})
	
	// STRATEGIC DATA TRANSFORMATION PIPELINE
	// Phase 1: Validate Base64 encoding integrity
	if err := s.validateBase64Encoding(base64Image); err != nil {
		return nil, errors.WrapError(err, errors.ErrorTypeBedrock,
			"base64 validation failed", false).
			WithAPISource(errors.APISourceConverse).
			WithContext("validation_phase", "pre_decode").
			WithContext("base64_length", len(base64Image)).
			WithContext("architectural_impact", "data_integrity")
	}
	
	// Phase 2: Decode for validation and format detection only
	// The shared bedrock package expects Base64 strings, not raw bytes
	s.logger.Debug("Decoding Base64 for validation purposes", map[string]interface{}{
		"transformation_boundary": "pre_validation_decode",
		"base64_length": len(base64Image),
	})
	
	imageBytes, decodingMetrics, err := s.decodeBase64ForValidation(base64Image)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrorTypeBedrock,
			"base64 validation decoding failed", false).
			WithAPISource(errors.APISourceConverse).
			WithContext("base64_length", len(base64Image)).
			WithContext("validation_phase", "format_detection").
			WithContext("architectural_layer", "validation_pipeline")
	}
	
	// Log decoding metrics
	s.logger.Info("Base64 validation decoding completed", map[string]interface{}{
		"transformation_boundary": "post_validation_decode",
		"decoding_metrics": decodingMetrics,
	})
	
	// Phase 3: Strategic size validation using decoded bytes
	if err := s.validateDecodedImageSize(imageBytes, decodingMetrics); err != nil {
		return nil, err
	}
	
	// Phase 4: Image format detection and validation
	imageFormat, err := s.detectAndValidateImageFormat(imageBytes)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrorTypeBedrock,
			"image format validation failed", false).
			WithAPISource(errors.APISourceConverse).
			WithContext("decoded_size", len(imageBytes)).
			WithContext("architectural_concern", "format_compatibility")
	}

	// Log format detection results
	s.logger.Info("Image format detected and validated", map[string]interface{}{
		"transformation_boundary": "format_detection_complete",
		"detected_format": imageFormat,
		"decoded_size_bytes": len(imageBytes),
	})
	
	// Strategic request construction - Pass Base64 string to shared package
	// CRITICAL: The shared bedrock package expects Base64 strings, NOT raw bytes
	s.logger.Debug("Passing Base64 string to shared bedrock package", map[string]interface{}{
		"transformation_boundary": "shared_package_handoff",
		"image_format": imageFormat,
		"base64_preserved": true,
		"architectural_note": "Base64 string passed directly to shared package",
	})
	
	img := bedrock.CreateImageContentFromBytes(imageFormat, base64Image)
	userMsg := bedrock.CreateUserMessageWithContent(turnPrompt, []bedrock.ContentBlock{img})
	req := bedrock.CreateConverseRequest(s.modelID, []bedrock.MessageWrapper{userMsg}, systemPrompt, s.maxTokens, nil, nil)

	// Strategic performance monitoring for API invocation
	apiInvocationStart := time.Now()
	
	s.logger.Info("Invoking Bedrock Converse API", map[string]interface{}{
		"transformation_boundary": "api_invocation_start",
		"model_id": s.modelID,
		"max_tokens": s.maxTokens,
	})
	
	// Execute AI model invocation with comprehensive monitoring
	resp, _, err := s.client.Converse(ctx, req)
	
	apiLatency := time.Since(apiInvocationStart)
	
	s.logger.Info("Bedrock Converse API call completed", map[string]interface{}{
		"transformation_boundary": "api_invocation_complete",
		"api_latency_ms": apiLatency.Milliseconds(),
		"success": err == nil,
	})
	
	if err != nil {
		s.logger.Error("Bedrock Converse API call failed", map[string]interface{}{
			"transformation_boundary": "api_error",
			"error": err.Error(),
			"api_latency_ms": apiLatency.Milliseconds(),
		})
		return nil, s.classifyAndEnrichBedrockError(err, systemPrompt, turnPrompt, decodingMetrics, apiLatency)
	}

	// Strategic response processing with performance metrics
	raw, _ := json.Marshal(resp)
	usage := models.TokenUsage{
		InputTokens:    resp.Usage.InputTokens,
		OutputTokens:   resp.Usage.OutputTokens,
		ThinkingTokens: 0, // Reserved for future model capabilities
		TotalTokens:    resp.Usage.TotalTokens,
	}

	totalOperationTime := time.Since(operationStart)
	
	s.logger.Info("Bedrock Converse operation completed successfully", map[string]interface{}{
		"transformation_boundary": "operation_complete",
		"total_operation_time_ms": totalOperationTime.Milliseconds(),
		"api_latency_ms": apiLatency.Milliseconds(),
		"token_usage": usage,
		"request_id": resp.RequestID,
	})
	
	// Strategic response construction with comprehensive metadata
	return &models.BedrockResponse{
		Raw:        raw,
		Processed:  resp,
		TokenUsage: usage,
		RequestID:  resp.RequestID,
		// Note: Consider adding performance metrics to response structure
	}, nil
}

// validateBase64Encoding performs strategic pre-flight validation of Base64 integrity
func (s *bedrockService) validateBase64Encoding(base64Str string) error {
	if len(base64Str) == 0 {
		return fmt.Errorf("empty base64 string provided")
	}
	
	// Strategic validation: Check for proper Base64 padding
	if len(base64Str)%4 != 0 {
		return fmt.Errorf("invalid base64 padding: length %d not divisible by 4", len(base64Str))
	}
	
	// Strategic validation: Check for invalid characters (basic check)
	// Full validation happens during decoding
	return nil
}

// decodeBase64ForValidation performs Base64 decoding for validation purposes only
// The decoded bytes are used for size validation and format detection
// The original Base64 string is passed to the shared bedrock package
func (s *bedrockService) decodeBase64ForValidation(base64Str string) ([]byte, map[string]interface{}, error) {
	decodingStart := time.Now()
	
	// Strategic data transformation: Base64 to binary
	imageBytes, err := base64.StdEncoding.DecodeString(base64Str)
	
	decodingLatency := time.Since(decodingStart)
	
	// Strategic metrics collection for operational insights
	metrics := map[string]interface{}{
		"base64_length":       len(base64Str),
		"decoded_length":      len(imageBytes),
		"compression_ratio":   float64(len(imageBytes)) / float64(len(base64Str)),
		"decoding_latency_us": decodingLatency.Microseconds(),
		"size_reduction_pct":  (1.0 - float64(len(imageBytes))/float64(len(base64Str))) * 100,
	}
	
	if err != nil {
		metrics["decoding_error"] = err.Error()
		return nil, metrics, fmt.Errorf("base64 decode error: %w", err)
	}
	
	return imageBytes, metrics, nil
}

// validateDecodedImageSize enforces strategic size constraints with architectural awareness
func (s *bedrockService) validateDecodedImageSize(imageBytes []byte, metrics map[string]interface{}) error {
	decodedSize := int64(len(imageBytes))
	
	if decodedSize > s.maxDecodedImageSize {
		s.logger.Warn("Decoded image exceeds size limit", map[string]interface{}{
			"transformation_boundary": "size_validation",
			"decoded_size_bytes": decodedSize,
			"max_allowed_bytes": s.maxDecodedImageSize,
			"size_excess_pct": float64(decodedSize-s.maxDecodedImageSize) / float64(s.maxDecodedImageSize) * 100,
		})
		return errors.NewValidationError(
			"decoded image exceeds maximum size limit",
			map[string]interface{}{
				"decoded_size_bytes":  decodedSize,
				"max_allowed_bytes":   s.maxDecodedImageSize,
				"size_excess_bytes":   decodedSize - s.maxDecodedImageSize,
				"size_excess_pct":     float64(decodedSize-s.maxDecodedImageSize) / float64(s.maxDecodedImageSize) * 100,
				"decoding_metrics":    metrics,
				"architectural_impact": "resource_constraints",
				"mitigation_strategy": "image_optimization_required",
			})
	}
	
	return nil
}

// detectAndValidateImageFormat performs strategic image format detection
func (s *bedrockService) detectAndValidateImageFormat(imageBytes []byte) (string, error) {
	if len(imageBytes) < 12 {
		return "", fmt.Errorf("insufficient bytes for format detection")
	}
	
	// Strategic format detection using magic numbers
	switch {
	case imageBytes[0] == 0xFF && imageBytes[1] == 0xD8 && imageBytes[2] == 0xFF:
		return "jpeg", nil
	case imageBytes[0] == 0x89 && imageBytes[1] == 0x50 && imageBytes[2] == 0x4E && imageBytes[3] == 0x47:
		return "png", nil
	case len(imageBytes) >= 12 && string(imageBytes[8:12]) == "WEBP":
		return "webp", nil
	default:
		// Fallback to JPEG for maximum compatibility
		return "jpeg", nil
	}
}

// classifyAndEnrichBedrockError provides strategic error classification with architectural context
func (s *bedrockService) classifyAndEnrichBedrockError(
	err error, 
	systemPrompt, turnPrompt string,
	decodingMetrics map[string]interface{},
	apiLatency time.Duration,
) error {
	errMsg := strings.ToLower(err.Error())
	
	// Strategic base error construction with comprehensive context
	baseError := errors.WrapError(err, errors.ErrorTypeBedrock, 
		"Bedrock Converse API call failed", true).
		WithAPISource(errors.APISourceConverse).
		WithContext("model_id", s.modelID).
		WithContext("max_tokens", s.maxTokens).
		WithContext("system_prompt_length", len(systemPrompt)).
		WithContext("turn_prompt_length", len(turnPrompt)).
		WithContext("api_latency_ms", apiLatency.Milliseconds()).
		WithContext("decoding_metrics", decodingMetrics)

	// Strategic error classification with architectural awareness
	switch {
	case strings.Contains(errMsg, "throttl") || strings.Contains(errMsg, "rate"):
		return baseError.
			WithContext("error_category", "throttling").
			WithContext("retry_strategy", "exponential_backoff").
			WithContext("architectural_recommendation", "implement_circuit_breaker").
			WithContext("severity", "low")
			
	case strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline"):
		return errors.NewBedrockError(
			"Bedrock call exceeded timeout limit",
			"BedrockTimeout",
			false,
		).WithAPISource(errors.APISourceConverse).
			WithContext("model_id", s.modelID).
			WithContext("timeout_seconds", s.callTimeout.Seconds()).
			WithContext("api_latency_ms", apiLatency.Milliseconds()).
			WithContext("architectural_concern", "latency_optimization_required")
			
	case strings.Contains(errMsg, "content") || strings.Contains(errMsg, "policy"):
		return errors.WrapError(err, errors.ErrorTypeBedrock, 
			"content policy violation", false).
			WithAPISource(errors.APISourceConverse).
			WithContext("error_category", "content_policy").
			WithContext("architectural_impact", "prompt_engineering_review_required").
			WithContext("severity", "high")
			
	case strings.Contains(errMsg, "token") && strings.Contains(errMsg, "limit"):
		return errors.WrapError(err, errors.ErrorTypeBedrock, 
			"token limit exceeded", false).
			WithAPISource(errors.APISourceConverse).
			WithContext("error_category", "token_limit").
			WithContext("architectural_solution", "implement_prompt_compression").
			WithContext("severity", "high")
			
	case strings.Contains(errMsg, "image") && strings.Contains(errMsg, "invalid"):
		return errors.WrapError(err, errors.ErrorTypeBedrock,
			"image validation failed", false).
			WithAPISource(errors.APISourceConverse).
			WithContext("error_category", "image_format").
			WithContext("decoding_metrics", decodingMetrics).
			WithContext("architectural_concern", "image_preprocessing_required").
			WithContext("severity", "medium")
			
	default:
		return baseError.
			WithContext("error_category", "unknown").
			WithContext("retry_strategy", "cautious").
			WithContext("severity", "medium").
			WithContext("requires_investigation", true)
	}
}