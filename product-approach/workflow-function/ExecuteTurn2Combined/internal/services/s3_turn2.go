package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"workflow-function/ExecuteTurn2Combined/internal/bedrockparser"
	"workflow-function/ExecuteTurn2Combined/internal/models"
	"workflow-function/shared/bedrock"
	"workflow-function/shared/errors"
	"workflow-function/shared/schema"
)

// TurnConversationDataStore represents the stored conversation format for a turn
type TurnConversationDataStore struct {
	VerificationId     string                  `json:"verificationId"`
	Timestamp          string                  `json:"timestamp"`
	TurnId             int                     `json:"turnId"`
	AnalysisStage      string                  `json:"analysisStage"`
	Messages           []schema.BedrockMessage `json:"messages"`
	TokenUsage         *schema.TokenUsage      `json:"tokenUsage,omitempty"`
	LatencyMs          int64                   `json:"latencyMs,omitempty"`
	ProcessingMetadata map[string]interface{}  `json:"processingMetadata,omitempty"`
	BedrockMetadata    map[string]interface{}  `json:"bedrockMetadata,omitempty"`
}

// buildAssistantContent creates the assistant message content
func buildAssistantContent(assistantResponse string) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"type": "text",
			"text": assistantResponse,
		},
	}
}

func bedrockMessageToMap(msg schema.BedrockMessage) map[string]interface{} {
	b, _ := json.Marshal(msg)
	var result map[string]interface{}
	json.Unmarshal(b, &result)
	return result
}

// LoadTurn1ProcessedResponse loads the processed Turn1 response from S3
func (m *s3Manager) LoadTurn1ProcessedResponse(ctx context.Context, ref models.S3Reference) (*schema.Turn1ProcessedResponse, error) {
	startTime := time.Now()
	m.logger.Info("s3_loading_turn1_processed_response_started", map[string]interface{}{
		"bucket":    ref.Bucket,
		"key":       ref.Key,
		"size":      ref.Size,
		"operation": "turn1_processed_response_load",
	})

	if err := m.validateReference(ref, "turn1_processed_response"); err != nil {
		return nil, err
	}

	// Get raw bytes
	stateRef := m.toStateReference(ref)
	raw, err := m.stateManager.Retrieve(stateRef)
	if err != nil {
		duration := time.Since(startTime)
		m.logger.Error("s3_turn1_processed_response_retrieval_failed", map[string]interface{}{
			"error":       err.Error(),
			"bucket":      ref.Bucket,
			"key":         ref.Key,
			"duration_ms": duration.Milliseconds(),
			"operation":   "get_bytes",
		})
		wfErr := &errors.WorkflowError{
			Type:      errors.ErrorTypeS3,
			Code:      "ReadFailed",
			Message:   fmt.Sprintf("failed to read Turn1 processed response: %v", err),
			Retryable: true,
			Severity:  errors.ErrorSeverityHigh,
			APISource: errors.APISourceUnknown,
			Timestamp: time.Now(),
		}
		return nil, wfErr.WithContext("s3_key", ref.Key).WithContext("bucket", ref.Bucket)
	}

	// Parse the Turn1 processed response
	var parsedData bedrockparser.ParsedTurn1Data
	if err := json.Unmarshal(raw, &parsedData); err != nil {
		m.logger.Error("s3_turn1_processed_response_format_error", map[string]interface{}{
			"error":  err.Error(),
			"bucket": ref.Bucket,
			"key":    ref.Key,
			"bytes":  len(raw),
		})
		return nil, &errors.WorkflowError{
			Type:      errors.ErrorTypeValidation,
			Code:      "BadTurn1ProcessedResponse",
			Message:   fmt.Sprintf("expected valid Turn1 processed response, got err %v", err),
			Retryable: false,
			Severity:  errors.ErrorSeverityCritical,
			APISource: errors.APISourceUnknown,
			Timestamp: time.Now(),
		}
	}

	// Convert to schema.Turn1ProcessedResponse
	response := &schema.Turn1ProcessedResponse{
		InitialConfirmation: parsedData.InitialConfirmation,
		MachineStructure:    fmt.Sprintf("%v", parsedData.MachineStructure),
		ReferenceRowStatus:  fmt.Sprintf("%v", parsedData.ReferenceRowStatus),
	}

	duration := time.Since(startTime)
	m.logger.Info("turn1_processed_response_loaded_successfully", map[string]interface{}{
		"bucket":      ref.Bucket,
		"key":         ref.Key,
		"duration_ms": duration.Milliseconds(),
	})

	return response, nil
}

// LoadTurn1RawResponse loads the raw Turn1 response from S3
func (m *s3Manager) LoadTurn1RawResponse(ctx context.Context, ref models.S3Reference) (json.RawMessage, error) {
	startTime := time.Now()
	m.logger.Info("s3_loading_turn1_raw_response_started", map[string]interface{}{
		"bucket":    ref.Bucket,
		"key":       ref.Key,
		"size":      ref.Size,
		"operation": "turn1_raw_response_load",
	})

	if err := m.validateReference(ref, "turn1_raw_response"); err != nil {
		return nil, err
	}

	// Get raw bytes
	stateRef := m.toStateReference(ref)
	raw, err := m.stateManager.Retrieve(stateRef)
	if err != nil {
		duration := time.Since(startTime)
		m.logger.Error("s3_turn1_raw_response_retrieval_failed", map[string]interface{}{
			"error":       err.Error(),
			"bucket":      ref.Bucket,
			"key":         ref.Key,
			"duration_ms": duration.Milliseconds(),
			"operation":   "get_bytes",
		})
		wfErr := &errors.WorkflowError{
			Type:      errors.ErrorTypeS3,
			Code:      "ReadFailed",
			Message:   fmt.Sprintf("failed to read Turn1 raw response: %v", err),
			Retryable: true,
			Severity:  errors.ErrorSeverityHigh,
			APISource: errors.APISourceUnknown,
			Timestamp: time.Now(),
		}
		return nil, wfErr.WithContext("s3_key", ref.Key).WithContext("bucket", ref.Bucket)
	}

	// Validate it's valid JSON
	var jsonObj interface{}
	if err := json.Unmarshal(raw, &jsonObj); err != nil {
		m.logger.Error("s3_turn1_raw_response_format_error", map[string]interface{}{
			"error":  err.Error(),
			"bucket": ref.Bucket,
			"key":    ref.Key,
			"bytes":  len(raw),
		})
		return nil, &errors.WorkflowError{
			Type:      errors.ErrorTypeValidation,
			Code:      "BadTurn1RawResponse",
			Message:   fmt.Sprintf("expected valid JSON, got err %v", err),
			Retryable: false,
			Severity:  errors.ErrorSeverityCritical,
			APISource: errors.APISourceUnknown,
			Timestamp: time.Now(),
		}
	}

	duration := time.Since(startTime)
	m.logger.Info("turn1_raw_response_loaded_successfully", map[string]interface{}{
		"bucket":      ref.Bucket,
		"key":         ref.Key,
		"bytes":       len(raw),
		"duration_ms": duration.Milliseconds(),
	})

	return raw, nil
}

// LoadTurn1SchemaResponse loads the raw Turn1 response and unmarshals it into schema.TurnResponse
func (m *s3Manager) LoadTurn1SchemaResponse(ctx context.Context, ref models.S3Reference, conversationRef *models.S3Reference) (*schema.TurnResponse, error) {
	raw, err := m.LoadTurn1RawResponse(ctx, ref)
	if err != nil {
		return nil, err
	}

	var resp schema.TurnResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		// Attempt to parse legacy structured format where response.content is an array
		var alt struct {
			TurnId    int               `json:"turnId"`
			Timestamp string            `json:"timestamp"`
			Prompt    string            `json:"prompt"`
			ImageUrls map[string]string `json:"imageUrls,omitempty"`
			Response  struct {
				Content []map[string]interface{} `json:"content"`
			} `json:"response"`
			LatencyMs       int64                  `json:"latencyMs"`
			TokenUsage      *schema.TokenUsage     `json:"tokenUsage,omitempty"`
			Stage           string                 `json:"analysisStage"`
			Metadata        map[string]interface{} `json:"metadata,omitempty"`
			BedrockMetadata struct {
				ModelId    string `json:"modelId"`
				RequestId  string `json:"requestId"`
				StopReason string `json:"stopReason,omitempty"`
			} `json:"bedrockMetadata,omitempty"`
		}

		if altErr := json.Unmarshal(raw, &alt); altErr == nil {
			resp.TurnId = alt.TurnId
			resp.Timestamp = alt.Timestamp
			resp.Prompt = alt.Prompt
			resp.ImageUrls = alt.ImageUrls
			resp.LatencyMs = alt.LatencyMs
			resp.TokenUsage = alt.TokenUsage
			resp.Stage = alt.Stage
			resp.Metadata = alt.Metadata
			resp.Response.ModelId = alt.BedrockMetadata.ModelId
			resp.Response.RequestId = alt.BedrockMetadata.RequestId
			resp.Response.StopReason = alt.BedrockMetadata.StopReason
			for _, c := range alt.Response.Content {
				if typ, ok := c["type"].(string); ok {
					switch typ {
					case "text":
						if s, ok := c["text"].(string); ok {
							resp.Response.Content = s
						}
					case "thinking":
						if s, ok := c["text"].(string); ok {
							resp.Response.Thinking = s
						}
					}
				}
			}
		} else {
			m.logger.Error("turn1_schema_unmarshal_failed", map[string]interface{}{
				"error":  err.Error(),
				"bucket": ref.Bucket,
				"key":    ref.Key,
			})
			return nil, &errors.WorkflowError{
				Type:      errors.ErrorTypeValidation,
				Code:      "BadTurn1SchemaResponse",
				Message:   fmt.Sprintf("failed to parse Turn1 raw response: %v", err),
				Retryable: false,
				Severity:  errors.ErrorSeverityCritical,
				APISource: errors.APISourceUnknown,
				Timestamp: time.Now(),
			}
		}
	}

	// If prompt or response content is empty, attempt to recover from
	// conversation history when provided
	if (resp.Prompt == "" || resp.Response.Content == "") && conversationRef != nil && conversationRef.Key != "" {
		var conv struct {
			Messages []schema.BedrockMessage `json:"messages"`
		}
		if err := m.LoadJSON(ctx, *conversationRef, &conv); err == nil {
			var userText, assistantText string
			for _, msg := range conv.Messages {
				if msg.Role == "user" && userText == "" {
					for _, c := range msg.Content {
						if c.Type == "text" && c.Text != "" {
							userText = c.Text
							break
						}
					}
				}
				if msg.Role == "assistant" && assistantText == "" {
					for _, c := range msg.Content {
						if c.Type == "text" && c.Text != "" {
							assistantText = c.Text
							break
						}
					}
				}
				if userText != "" && assistantText != "" {
					break
				}
			}
			if resp.Prompt == "" && userText != "" {
				resp.Prompt = userText
				m.logger.Info("turn1_prompt_loaded_for_history", map[string]interface{}{
					"bucket": conversationRef.Bucket,
					"key":    conversationRef.Key,
					"length": len(userText),
				})
			}
			if resp.Response.Content == "" && assistantText != "" {
				resp.Response.Content = assistantText
			}
		} else {
			m.logger.Warn("turn1_conversation_load_failed", map[string]interface{}{
				"error":  err.Error(),
				"bucket": conversationRef.Bucket,
				"key":    conversationRef.Key,
			})
		}
	}

	// Validate the loaded Turn1 response to ensure required fields are present
	// Convert schema.TurnResponse to bedrock.Turn1Response for validation
	bResp := bedrock.Turn1Response{
		TurnID:    resp.TurnId,
		Timestamp: resp.Timestamp,
		Prompt:    resp.Prompt,
		Response: bedrock.TextResponse{
			Content:    resp.Response.Content,
			StopReason: resp.Response.StopReason,
			Thinking:   resp.Response.Thinking,
		},
		Thinking:      resp.Response.Thinking,
		LatencyMs:     resp.LatencyMs,
		AnalysisStage: resp.Stage,
		BedrockMetadata: bedrock.BedrockMetadata{
			ModelID:         resp.Response.ModelId,
			RequestID:       resp.Response.RequestId,
			InvokeLatencyMs: resp.LatencyMs,
			APIType:         bedrock.APITypeConverse,
		},
		APIType: bedrock.APITypeConverse,
	}
	if resp.TokenUsage != nil {
		bResp.TokenUsage = bedrock.TokenUsage{
			InputTokens:    resp.TokenUsage.InputTokens,
			OutputTokens:   resp.TokenUsage.OutputTokens,
			ThinkingTokens: resp.TokenUsage.ThinkingTokens,
			TotalTokens:    resp.TokenUsage.TotalTokens,
		}
	}

	if err := bedrock.ValidateTurn1Response(&bResp); err != nil {
		m.logger.Error("turn1_schema_validation_failed", map[string]interface{}{
			"error":  err.Error(),
			"bucket": ref.Bucket,
			"key":    ref.Key,
		})
		return nil, &errors.WorkflowError{
			Type:      errors.ErrorTypeValidation,
			Code:      "InvalidTurn1Response",
			Message:   err.Error(),
			Retryable: false,
			Severity:  errors.ErrorSeverityCritical,
			APISource: errors.APISourceUnknown,
			Timestamp: time.Now(),
		}
	}

	return &resp, nil
}

// StoreTurn2Response stores the Turn2 response
func (m *s3Manager) StoreTurn2Response(ctx context.Context, verificationID string, response *bedrockparser.ParsedTurn2Data) (models.S3Reference, error) {
	if verificationID == "" {
		return models.S3Reference{}, errors.NewValidationError(
			"verification ID required for storing Turn2 response",
			map[string]interface{}{"operation": "store_turn2_response"})
	}
	if response == nil {
		return models.S3Reference{}, errors.NewValidationError(
			"response cannot be nil for storing Turn2 response",
			map[string]interface{}{"verification_id": verificationID})
	}

	key := "processing/turn2-processed-response.json"
	stateRef, err := m.stateManager.StoreJSON(m.datePath(verificationID), key, response)
	if err != nil {
		return models.S3Reference{}, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to store Turn2 response", true).
			WithContext("verification_id", verificationID).
			WithContext("category", "processing")
	}

	return m.fromStateReference(stateRef), nil
}

// StoreTurn2RawResponse stores the raw Turn2 Bedrock response
func (m *s3Manager) StoreTurn2RawResponse(ctx context.Context, verificationID string, raw interface{}) (models.S3Reference, error) {
	if verificationID == "" {
		return models.S3Reference{}, errors.NewValidationError(
			"verification ID required for storing Turn2 raw response",
			map[string]interface{}{"operation": "store_turn2_raw"})
	}

	// Convert raw input to map for augmentation
	var rawMap map[string]interface{}
	b, err := json.Marshal(raw)
	if err == nil {
		_ = json.Unmarshal(b, &rawMap)
		if resp, ok := rawMap["response"].(map[string]interface{}); ok {
			if thinking, ok := resp["thinking"]; ok {
				rawMap["thinking"] = thinking
			}
		}
	} else {
		rawMap = map[string]interface{}{"error": err.Error()}
	}

	key := "responses/turn2-raw-response.json"
	stateRef, err := m.stateManager.StoreJSON(m.datePath(verificationID), key, rawMap)
	if err != nil {
		return models.S3Reference{}, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to store Turn2 raw response", true).
			WithContext("verification_id", verificationID).
			WithContext("category", "responses")
	}

	return m.fromStateReference(stateRef), nil
}

// StoreTurn2ProcessedResponse stores the processed Turn2 analysis
func (m *s3Manager) StoreTurn2ProcessedResponse(ctx context.Context, verificationID string, processed *bedrockparser.ParsedTurn2Data) (models.S3Reference, error) {
	if verificationID == "" {
		return models.S3Reference{}, errors.NewValidationError(
			"verification ID required for storing Turn2 processed response",
			map[string]interface{}{"operation": "store_turn2_processed"})
	}
	if processed == nil {
		return models.S3Reference{}, errors.NewValidationError(
			"processed data cannot be nil",
			map[string]interface{}{"verification_id": verificationID})
	}

	key := "processing/turn2-processed-response.json"
	stateRef, err := m.stateManager.StoreJSON(m.datePath(verificationID), key, processed)
	if err != nil {
		return models.S3Reference{}, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to store Turn2 processed response", true).
			WithContext("verification_id", verificationID).
			WithContext("category", "processing")
	}

	return m.fromStateReference(stateRef), nil
}

// StoreTurn2Markdown stores the Markdown version of the Turn2 analysis
func (m *s3Manager) StoreTurn2Markdown(ctx context.Context, verificationID string, markdownContent string) (models.S3Reference, error) {
	if verificationID == "" {
		return models.S3Reference{}, errors.NewValidationError(
			"verification ID required for storing Turn2 markdown",
			map[string]interface{}{"operation": "store_turn2_markdown"})
	}

	key := "responses/turn2-processed-response.md"
	dataBytes := []byte(markdownContent)
	stateRef, err := m.stateManager.StoreWithContentType(m.datePath(verificationID), key, dataBytes, "text/markdown; charset=utf-8")
	if err != nil {
		return models.S3Reference{}, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to store Turn2 markdown", true).
			WithContext("verification_id", verificationID).
			WithContext("category", "responses")
	}

	return m.fromStateReference(stateRef), nil
}

// StoreTurn2Conversation stores full conversation messages for turn2
func (m *s3Manager) StoreTurn2Conversation(ctx context.Context, verificationID string, turn1Messages []schema.BedrockMessage, systemPrompt string, userPrompt string, base64Image string, base64Ref models.S3Reference, assistantResponse string, tokenUsage *schema.TokenUsage, latencyMs int64, bedrockRequestId string, modelId string, bedrockResponseMetadata map[string]interface{}) (models.S3Reference, error) {
	if verificationID == "" {
		return models.S3Reference{}, errors.NewValidationError(
			"verification ID required for storing Turn2 conversation",
			map[string]interface{}{"operation": "store_turn2_conversation"})
	}

	// Build messages array
	messages := []map[string]interface{}{}

	if len(turn1Messages) > 0 && turn1Messages[0].Role == "system" {
		messages = append(messages, bedrockMessageToMap(turn1Messages[0]))
		turn1Messages = turn1Messages[1:]
	} else {
		messages = append(messages, map[string]interface{}{
			"role":    "system",
			"content": []map[string]interface{}{{"text": systemPrompt}},
		})
	}

	for _, msg := range turn1Messages {
		messages = append(messages, bedrockMessageToMap(msg))
	}

	userMessage := map[string]interface{}{
		"role": "user",
		"content": []map[string]interface{}{
			{"text": userPrompt},
			{
				"image": map[string]interface{}{
					"format": "png",
					"source": map[string]interface{}{
						"s3Uri": fmt.Sprintf("s3://%s/%s", base64Ref.Bucket, base64Ref.Key),
					},
				},
			},
		},
	}
	messages = append(messages, userMessage)

	assistantMessage := map[string]interface{}{
		"role":    "assistant",
		"content": buildAssistantContent(assistantResponse),
	}
	messages = append(messages, assistantMessage)

	data := map[string]interface{}{
		"verificationId": verificationID,
		"timestamp":      schema.FormatISO8601(),
		"turnId":         2,
		"analysisStage":  bedrock.AnalysisStageTurn2,
		"messages":       messages,
	}

	if tokenUsage != nil {
		data["tokenUsage"] = map[string]interface{}{
			"input":    tokenUsage.InputTokens,
			"output":   tokenUsage.OutputTokens,
			"thinking": tokenUsage.ThinkingTokens,
			"total":    tokenUsage.TotalTokens,
		}
	}

	if latencyMs > 0 {
		data["latencyMs"] = latencyMs
		data["processingMetadata"] = map[string]interface{}{
			"executionTimeMs": latencyMs,
			"retryAttempts":   0,
		}
	}

	if bedrockRequestId != "" {
		meta := map[string]interface{}{
			"modelId":    modelId,
			"requestId":  bedrockRequestId,
			"stopReason": "end_turn",
		}
		for k, v := range bedrockResponseMetadata {
			meta[k] = v
		}
		data["bedrockMetadata"] = meta
	}

	key := "responses/turn2-conversation.json"
	stateRef, err := m.stateManager.StoreJSON(m.datePath(verificationID), key, data)
	if err != nil {
		return models.S3Reference{}, errors.WrapError(err, errors.ErrorTypeS3,
			"failed to store turn2 conversation", true).
			WithContext("verification_id", verificationID)
	}
	return m.fromStateReference(stateRef), nil
}
