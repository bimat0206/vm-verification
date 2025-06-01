package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"workflow-function/ExecuteTurn1Combined/internal/bedrockparser"
	"workflow-function/ExecuteTurn1Combined/internal/config"
	"workflow-function/ExecuteTurn1Combined/internal/models"
	"workflow-function/ExecuteTurn1Combined/internal/services"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
)

// StorageManager handles S3 storage operations for responses
type StorageManager struct {
	s3  services.S3StateManager
	cfg config.Config
	log logger.Logger
}

// NewStorageManager creates a new instance of StorageManager
func NewStorageManager(s3 services.S3StateManager, cfg config.Config, log logger.Logger) *StorageManager {
	return &StorageManager{
		s3:  s3,
		cfg: cfg,
		log: log,
	}
}

// StorageResult contains the results of storage operations
type StorageResult struct {
	RawRef       models.S3Reference
	ProcessedRef models.S3Reference
	RawSize      int
	Duration     time.Duration
	Error        error
}

// StorePrompt stores the rendered prompt in structured schema format
func (m *StorageManager) StorePrompt(ctx context.Context, req *models.Turn1Request, turn int, result *PromptResult) (models.S3Reference, error) {
	key := fmt.Sprintf("prompts/turn%d-prompt.json", turn)

	contextLogger := m.log.WithCorrelationId(req.VerificationID)

	// Build message structure for Bedrock
	messageStructure := map[string]interface{}{
		"role":    "user",
		"content": []map[string]interface{}{{"type": "text", "text": result.Prompt}},
	}

	// Build contextual instructions - minimal for now
	contextual := map[string]interface{}{
		"analysisObjective": "Analyze reference image in detail",
	}

	// Build image reference with base64 location
	imageRef := map[string]interface{}{
		"imageType": "reference",
		"base64StorageReference": map[string]interface{}{
			"bucket": req.S3Refs.Images.ReferenceBase64.Bucket,
			"key":    req.S3Refs.Images.ReferenceBase64.Key,
		},
	}
	if url, ok := req.VerificationContext.LayoutMetadata["referenceImageUrl"].(string); ok && url != "" {
		imageRef["sourceUrl"] = url
	} else {
		imageRef["sourceUrl"] = req.S3Refs.Images.ReferenceBase64.Key
	}

	// Determine context sources
	contextSources := []string{"INITIALIZATION", "IMAGE_METADATA"}
	if req.VerificationContext.LayoutMetadata != nil {
		contextSources = append(contextSources, "LAYOUT_METADATA")
	}
	if req.VerificationContext.HistoricalContext != nil {
		contextSources = append(contextSources, "HISTORICAL_CONTEXT")
	}

	// Generation metadata
	generationMetadata := map[string]interface{}{
		"processingTimeMs": result.Duration.Milliseconds(),
		"promptSource":     "TEMPLATE_BASED",
		"contextSources":   contextSources,
	}

	promptData := map[string]interface{}{
		"verificationId":         req.VerificationID,
		"promptType":             "TURN1",
		"verificationType":       req.VerificationContext.VerificationType,
		"messageStructure":       messageStructure,
		"contextualInstructions": contextual,
		"imageReference":         imageRef,
		"templateVersion":        m.cfg.Prompts.TemplateVersion,
		"createdAt":              schema.FormatISO8601(),
		"generationMetadata":     generationMetadata,
	}

	ref, err := m.s3.StorePrompt(ctx, req.VerificationID, turn, promptData)
	if err != nil {
		s3Err := errors.WrapError(err, errors.ErrorTypeS3,
			"store prompt failed", true).
			WithContext("verification_id", req.VerificationID).
			WithContext("prompt_size", len(result.Prompt))

		enrichedErr := errors.SetVerificationID(s3Err, req.VerificationID)

		contextLogger.Warn("s3 prompt-store warning", map[string]interface{}{
			"prompt_size_bytes": len(result.Prompt),
			"bucket":            m.cfg.AWS.S3Bucket,
			"key":               key,
		})

		return models.S3Reference{}, enrichedErr
	}

	contextLogger.Debug("stored prompt", map[string]interface{}{
		"key":        ref.Key,
		"size_bytes": len(result.Prompt),
		"turn_id":    turn,
	})

	return ref, nil
}

// StoreConversation stores turn1 conversation messages in S3 with complete schema compliance
func (m *StorageManager) StoreConversation(ctx context.Context, verificationID string, systemPrompt string, userPrompt string, base64Image string, assistantResponse string, thinkingContent string, tokenUsage *schema.TokenUsage, latencyMs int64, bedrockRequestId string, modelId string, bedrockResponseMetadata map[string]interface{}) (models.S3Reference, error) {
	if verificationID == "" {
		return models.S3Reference{}, errors.NewValidationError(
			"verification ID required for conversation storage",
			map[string]interface{}{"operation": "store_conversation"})
	}

	start := time.Now()
	ref, err := m.s3.StoreTurn1Conversation(ctx, verificationID, systemPrompt, userPrompt, base64Image, assistantResponse, thinkingContent, tokenUsage, latencyMs, bedrockRequestId, modelId, bedrockResponseMetadata)
	if err != nil {
		m.log.Warn("s3 conversation-store warning", map[string]interface{}{
			"error":  err.Error(),
			"bucket": m.cfg.AWS.S3Bucket,
		})
		return models.S3Reference{}, err
	}

	m.log.Debug("stored conversation", map[string]interface{}{
		"key":        ref.Key,
		"durationMs": time.Since(start).Milliseconds(),
	})
	return ref, nil
}

// StoreResponses stores raw and processed responses to S3
func (s *StorageManager) StoreResponses(ctx context.Context, req *models.Turn1Request, invoke *InvokeResult, prompt *PromptResult, imageSize int, parsedMarkdown *bedrockparser.ParsedTurn1Markdown) *StorageResult {
	startTime := time.Now()
	result := &StorageResult{}
	verificationID := req.VerificationID
	resp := invoke.Response
	contextLogger := s.log.WithCorrelationId(verificationID)

	// Build raw response structure according to schema
	var stopReason string
	var rawMap map[string]interface{}
	if err := json.Unmarshal(resp.Raw, &rawMap); err == nil {
		if sr, ok := rawMap["stop_reason"].(string); ok {
			stopReason = sr
		} else if sr, ok := rawMap["stopReason"].(string); ok {
			stopReason = sr
		}
	}

	// Build response content with thinking if available
	// Enhanced to properly structure thinking content from adapter metadata
	responseContent := []map[string]interface{}{
		{"type": "text", "text": resp.Processed.(map[string]interface{})["content"]},
	}

	// Add thinking content if available in metadata
	var thinkingContent string
	var hasThinking bool
	if resp.Metadata != nil {
		if thinking, ok := resp.Metadata["thinking"].(string); ok && thinking != "" {
			thinkingContent = thinking
			hasThinking = true
			responseContent = append(responseContent, map[string]interface{}{
				"type": "thinking",
				"text": thinking,
			})
		}
	}

	// Build enhanced bedrock metadata with thinking support
	bedrockMetadata := map[string]interface{}{
		"modelId":         s.cfg.AWS.BedrockModel,
		"requestId":       resp.RequestID,
		"stopReason":      stopReason,
		"hasThinking":     hasThinking,
		"thinkingEnabled": s.cfg.Processing.ThinkingType == "enable",
	}

	// Add thinking-specific metadata if available
	if hasThinking {
		bedrockMetadata["thinkingLength"] = len(thinkingContent)
		bedrockMetadata["thinkingTokens"] = resp.TokenUsage.ThinkingTokens
	}

	// Extract thinking blocks from metadata if available
	var thinkingBlocks []interface{}
	if resp.Metadata != nil {
		if blocks, ok := resp.Metadata["thinking_blocks"]; ok {
			thinkingBlocks = blocks.([]interface{})
		}
	}

	rawData := map[string]interface{}{
		"verificationId":   verificationID,
		"turnId":           1,
		"analysisStage":    "REFERENCE_ANALYSIS",
		"verificationType": req.VerificationContext.VerificationType,
		"response": map[string]interface{}{
			"content": responseContent,
		},
		"tokenUsage": map[string]interface{}{
			"input":    resp.TokenUsage.InputTokens,
			"output":   resp.TokenUsage.OutputTokens,
			"thinking": resp.TokenUsage.ThinkingTokens,
			"total":    resp.TokenUsage.TotalTokens,
		},
		"latencyMs":       invoke.Duration.Milliseconds(),
		"bedrockMetadata": bedrockMetadata,
		"promptMetadata": map[string]interface{}{
			"imageType":           "reference",
			"promptTokenEstimate": prompt.TemplateProcessor.InputTokens,
			"imageSize":           imageSize,
		},
		"timestamp": schema.FormatISO8601(),
		"status":    "SUCCESS",
		"processingMetadata": map[string]interface{}{
			"executionTimeMs": invoke.Duration.Milliseconds(),
			"retryAttempts":   0,
		},
	}

	// Add thinking blocks if available
	if len(thinkingBlocks) > 0 {
		rawData["thinkingBlocks"] = thinkingBlocks
	}

	rawJSON, _ := json.Marshal(rawData)
	// Store raw response in new structured format
	rawRef, err := s.s3.StoreRawResponse(ctx, verificationID, rawData)
	if err != nil {
		s3Err := errors.WrapError(err, errors.ErrorTypeS3,
			"store raw response failed", true).
			WithContext("verification_id", verificationID).
			WithContext("response_size", len(rawJSON))

		enrichedErr := errors.SetVerificationID(s3Err, verificationID)

		contextLogger.Warn("s3 raw-store warning", map[string]interface{}{
			"response_size_bytes": len(rawJSON),
			"bucket":              s.cfg.AWS.S3Bucket,
		})

		result.Error = enrichedErr
		result.Duration = time.Since(startTime)
		return result
	}

	// Store processed analysis or parsed response
	var procRef models.S3Reference
	var procErr error
	if parsedMarkdown != nil && parsedMarkdown.AnalysisMarkdown != "" {
		procRef, procErr = s.s3.StoreProcessedTurn1Markdown(ctx, verificationID, parsedMarkdown.AnalysisMarkdown)
		if procErr != nil {
			s3Err := errors.WrapError(procErr, errors.ErrorTypeS3,
				"store processed analysis failed", true).
				WithContext("verification_id", verificationID)

			enrichedErr := errors.SetVerificationID(s3Err, verificationID)

			contextLogger.Warn("s3 processed-store warning", map[string]interface{}{
				"bucket": s.cfg.AWS.S3Bucket,
			})

			result.Error = enrichedErr
			result.Duration = time.Since(startTime)
			return result
		}
		result.ProcessedRef = procRef
	} else {
		contextLogger.Warn("Parsed Turn 1 Markdown is nil or empty, skipping S3 storage of processed Markdown response.", map[string]interface{}{"verificationId": verificationID})
	}

	result.RawRef = rawRef
	result.RawSize = len(rawJSON)
	result.Duration = time.Since(startTime)

	return result
}

// GetStorageMetadata returns metadata for tracking storage operations
func (s *StorageManager) GetStorageMetadata(result *StorageResult) map[string]interface{} {
	return map[string]interface{}{
		"s3_objects_created": 2,
		"raw_response_size":  result.RawSize,
		"processed_ref_key":  result.ProcessedRef.Key,
		"raw_ref_key":        result.RawRef.Key,
	}
}
