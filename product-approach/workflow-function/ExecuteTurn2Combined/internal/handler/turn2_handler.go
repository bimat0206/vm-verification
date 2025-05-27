package handler

import (
	"context"
	"time"

	"workflow-function/ExecuteTurn2Combined/internal/bedrockparser"
	"workflow-function/ExecuteTurn2Combined/internal/config"
	"workflow-function/ExecuteTurn2Combined/internal/models"
	"workflow-function/ExecuteTurn2Combined/internal/services"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
)

// Turn2Handler handles the complete Turn2 processing flow
type Turn2Handler struct {
	contextLoader *ContextLoader
	promptService services.PromptServiceTurn2
	bedrockService services.BedrockServiceTurn2
	s3Service     services.S3StateManager
	log           logger.Logger
	cfg           config.Config
}

// NewTurn2Handler creates a new Turn2Handler instance
func NewTurn2Handler(
	contextLoader *ContextLoader,
	promptService services.PromptServiceTurn2,
	bedrockService services.BedrockServiceTurn2,
	s3Service services.S3StateManager,
	log logger.Logger,
	cfg config.Config,
) *Turn2Handler {
	return &Turn2Handler{
		contextLoader:  contextLoader,
		promptService:  promptService,
		bedrockService: bedrockService,
		s3Service:      s3Service,
		log:            log,
		cfg:            cfg,
	}
}

// ProcessTurn2Request handles the complete Turn2 processing flow
func (h *Turn2Handler) ProcessTurn2Request(ctx context.Context, req *models.Turn2Request) (*models.Turn2Response, error) {
	startTime := time.Now()
	h.log.Info("turn2_processing_started", map[string]interface{}{
		"verification_id":    req.VerificationID,
		"verification_type":  req.VerificationType,
		"checking_image_key": req.S3Refs.Images.CheckingBase64.Key,
		"turn1_response_key": req.S3Refs.Turn1References.ProcessedResponse.Key,
	})

	// Load context (system prompt, checking image, Turn1 results)
	loadResult := h.contextLoader.LoadContextTurn2(ctx, req)
	if loadResult.Error != nil {
		return nil, errors.WrapError(loadResult.Error, errors.ErrorTypeS3,
			"failed to load Turn2 context", true).
			WithContext("verification_id", req.VerificationID).
			WithContext("stage", "context_loading")
	}

	// Create verification context
	vCtx := &schema.VerificationContext{
		VerificationId:   req.VerificationID,
		VerificationType: req.VerificationType,
		LayoutId:         req.LayoutID,
		VendingMachineId: req.VendingMachineID,
		Location:         req.Location,
		Timestamp:        time.Now().Format(time.RFC3339),
	}

	// Generate Turn2 prompt
	prompt, promptProcessor, err := h.promptService.GenerateTurn2PromptWithMetrics(
		ctx,
		vCtx,
		loadResult.SystemPrompt,
		loadResult.Turn1Response,
		loadResult.Turn1RawResponse,
	)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrorTypeTemplate,
			"failed to generate Turn2 prompt", false).
			WithContext("verification_id", req.VerificationID).
			WithContext("stage", "prompt_generation")
	}

	// Store prompt processor metrics
	promptProcessorRef, err := h.s3Service.StoreTemplateProcessor(ctx, req.VerificationID, promptProcessor)
	if err != nil {
		h.log.Warn("failed_to_store_prompt_processor", map[string]interface{}{
			"error":           err.Error(),
			"verification_id": req.VerificationID,
		})
		// Non-critical error, continue processing
	}

	// Invoke Bedrock with conversation history
	bedrockResponse, err := h.bedrockService.ConverseWithHistory(
		ctx,
		loadResult.SystemPrompt,
		prompt,
		loadResult.Base64Image,
		loadResult.Turn1Response,
	)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrorTypeBedrock,
			"failed to invoke Bedrock for Turn2", true).
			WithContext("verification_id", req.VerificationID).
			WithContext("stage", "bedrock_invocation")
	}

	// Store raw Bedrock response
	rawResponseRef, err := h.s3Service.StoreRawResponse(ctx, req.VerificationID, bedrockResponse)
	if err != nil {
		h.log.Warn("failed_to_store_raw_response", map[string]interface{}{
			"error":           err.Error(),
			"verification_id": req.VerificationID,
		})
		// Non-critical error, continue processing
	}

	// Parse Bedrock response
	markdownResponse, err := bedrockparser.ParseBedrockResponseAsMarkdown(bedrockResponse.Content)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrorTypeParser,
			"failed to parse Bedrock response as markdown", false).
			WithContext("verification_id", req.VerificationID).
			WithContext("stage", "response_parsing")
	}

	// Store markdown response
	markdownRef, err := h.s3Service.StoreTurn2Markdown(ctx, req.VerificationID, markdownResponse)
	if err != nil {
		h.log.Warn("failed_to_store_markdown_response", map[string]interface{}{
			"error":           err.Error(),
			"verification_id": req.VerificationID,
		})
		// Non-critical error, continue processing
	}

	// Parse structured data from response
	parsedData, err := bedrockparser.ParseTurn2Response(bedrockResponse.Content)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrorTypeParser,
			"failed to parse Turn2 response", false).
			WithContext("verification_id", req.VerificationID).
			WithContext("stage", "response_parsing")
	}

	// Store processed Turn2 response
	processedRef, err := h.s3Service.StoreTurn2Response(ctx, req.VerificationID, parsedData)
	if err != nil {
		h.log.Warn("failed_to_store_processed_response", map[string]interface{}{
			"error":           err.Error(),
			"verification_id": req.VerificationID,
		})
		// Non-critical error, continue processing
	}

	// Create processing metrics
	processingMetrics := &schema.ProcessingMetrics{
		VerificationId:   req.VerificationID,
		VerificationType: req.VerificationType,
		Turn:             2,
		TotalTimeMs:      time.Since(startTime).Milliseconds(),
		ContextLoadingMs: loadResult.Duration.Milliseconds(),
		PromptGenerationMs: promptProcessor.ProcessingTimeMs,
		BedrockInvocationMs: bedrockResponse.LatencyMs,
		ResponseParsingMs:   0, // Not tracked separately
		TokenUsage: schema.TokenUsage{
			InputTokens:  bedrockResponse.InputTokens,
			OutputTokens: bedrockResponse.OutputTokens,
			TotalTokens:  bedrockResponse.InputTokens + bedrockResponse.OutputTokens,
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Store processing metrics
	metricsRef, err := h.s3Service.StoreProcessingMetrics(ctx, req.VerificationID, processingMetrics)
	if err != nil {
		h.log.Warn("failed_to_store_processing_metrics", map[string]interface{}{
			"error":           err.Error(),
			"verification_id": req.VerificationID,
		})
		// Non-critical error, continue processing
	}

	// Create Turn2 response
	response := &models.Turn2Response{
		VerificationID:    req.VerificationID,
		VerificationType:  req.VerificationType,
		VerificationOutcome: parsedData.VerificationOutcome,
		Discrepancies:     parsedData.Discrepancies,
		ComparisonSummary: parsedData.ComparisonSummary,
		S3Refs: models.Turn2ResponseS3Refs{
			RawResponse:       rawResponseRef,
			MarkdownResponse:  markdownRef,
			ProcessedResponse: processedRef,
			PromptProcessor:   promptProcessorRef,
			ProcessingMetrics: metricsRef,
		},
		ProcessingMetrics: processingMetrics,
	}

	h.log.Info("turn2_processing_completed", map[string]interface{}{
		"verification_id":       req.VerificationID,
		"verification_type":     req.VerificationType,
		"verification_outcome":  parsedData.VerificationOutcome,
		"discrepancy_count":     len(parsedData.Discrepancies),
		"total_processing_time": processingMetrics.TotalTimeMs,
		"input_tokens":          processingMetrics.TokenUsage.InputTokens,
		"output_tokens":         processingMetrics.TokenUsage.OutputTokens,
		"total_tokens":          processingMetrics.TokenUsage.TotalTokens,
	})

	return response, nil
}