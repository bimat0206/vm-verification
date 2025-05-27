package bedrock

import (
	"context"
	"time"

	"workflow-function/ExecuteTurn2Combined/internal/config"
	sharedBedrock "workflow-function/shared/bedrock"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
)

// AdapterTurn2 provides Turn2-specific Bedrock functionality
type AdapterTurn2 struct {
	cfg    *config.Config
	log    logger.Logger
	client *sharedBedrock.BedrockClient
}

// NewAdapterTurn2 creates a new AdapterTurn2 instance
func NewAdapterTurn2(client *sharedBedrock.BedrockClient, cfg *config.Config, log logger.Logger) *AdapterTurn2 {
	return &AdapterTurn2{
		client: client,
		cfg:    cfg,
		log:    log,
	}
}

// ConverseWithHistory handles Turn2 conversation with history from Turn1
func (a *AdapterTurn2) ConverseWithHistory(ctx context.Context, systemPrompt, turn2Prompt, base64Image string, turn1Response *schema.Turn1ProcessedResponse) (*schema.BedrockResponse, error) {
	startTime := time.Now()

	// Validate inputs
	if systemPrompt == "" {
		return nil, errors.NewValidationError(
			"system prompt cannot be empty",
			map[string]interface{}{"operation": "bedrock_converse_with_history"})
	}

	if turn2Prompt == "" {
		return nil, errors.NewValidationError(
			"turn2 prompt cannot be empty",
			map[string]interface{}{"operation": "bedrock_converse_with_history"})
	}

	if base64Image == "" {
		return nil, errors.NewValidationError(
			"base64 image cannot be empty",
			map[string]interface{}{"operation": "bedrock_converse_with_history"})
	}

	if turn1Response == nil {
		return nil, errors.NewValidationError(
			"turn1 response cannot be nil",
			map[string]interface{}{"operation": "bedrock_converse_with_history"})
	}

	// Build Turn1 message from Turn1 response
	turn1Message := ""
	if turn1Response.InitialConfirmation != "" {
		turn1Message += turn1Response.InitialConfirmation + "\n\n"
	}

	if turn1Response.MachineStructure != "" {
		turn1Message += turn1Response.MachineStructure + "\n\n"
	}

	if turn1Response.ReferenceRowStatus != "" {
		turn1Message += turn1Response.ReferenceRowStatus
	}

	// Build request with conversation history using the correct types
	request := &sharedBedrock.ConverseRequest{
		ModelId: a.cfg.AWS.BedrockModel,
		System:  systemPrompt,
		Messages: []sharedBedrock.MessageWrapper{
			{
				Role: "user",
				Content: []sharedBedrock.ContentBlock{
					{
						Type: "text",
						Text: "[Turn 1] Please analyze this image.",
					},
				},
			},
			{
				Role: "assistant",
				Content: []sharedBedrock.ContentBlock{
					{
						Type: "text",
						Text: turn1Message,
					},
				},
			},
			{
				Role: "user",
				Content: []sharedBedrock.ContentBlock{
					{
						Type: "text",
						Text: "[Turn 2] " + turn2Prompt,
					},
					{
						Type: "image",
						Image: &sharedBedrock.ImageBlock{
							Format: "jpeg", // Assume JPEG for now
							Source: sharedBedrock.ImageSource{
								Type:  "bytes",
								Bytes: base64Image,
							},
						},
					},
				},
			},
		},
		InferenceConfig: sharedBedrock.InferenceConfig{
			MaxTokens:   a.cfg.Processing.MaxTokens,
			Temperature: &[]float64{0.7}[0], // Default temperature
		},
	}

	// Log request details
	a.log.Info("bedrock_turn2_request_prepared", map[string]interface{}{
		"model_id":           a.cfg.AWS.BedrockModel,
		"max_tokens":         a.cfg.Processing.MaxTokens,
		"system_prompt_size": len(systemPrompt),
		"turn2_prompt_size":  len(turn2Prompt),
		"turn1_message_size": len(turn1Message),
		"image_size":         len(base64Image),
		"message_count":      len(request.Messages),
	})

	// Invoke Bedrock (note: returns 3 values)
	response, latencyMs, err := a.client.Converse(ctx, request)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrorTypeBedrock,
			"failed to invoke Bedrock for Turn2", true).
			WithContext("model_id", a.cfg.AWS.BedrockModel).
			WithContext("operation", "bedrock_converse_with_history")
	}

	// Extract text content from response
	textContent := sharedBedrock.ExtractTextFromResponse(response)

	// Translate to schema.BedrockResponse
	schemaResponse := &schema.BedrockResponse{
		Content:          textContent,
		CompletionReason: response.StopReason,
		InputTokens:      response.Usage.InputTokens,
		OutputTokens:     response.Usage.OutputTokens,
		LatencyMs:        latencyMs,
		ModelId:          response.ModelID,
		Timestamp:        time.Now().Format(time.RFC3339),
		Turn:             2,
		ProcessingTimeMs: time.Since(startTime).Milliseconds(),
	}

	// Log response details
	a.log.Info("bedrock_turn2_response_received", map[string]interface{}{
		"model_id":           response.ModelID,
		"completion_reason":  response.StopReason,
		"input_tokens":       response.Usage.InputTokens,
		"output_tokens":      response.Usage.OutputTokens,
		"latency_ms":         latencyMs,
		"content_length":     len(textContent),
		"processing_time_ms": schemaResponse.ProcessingTimeMs,
	})

	return schemaResponse, nil
}

// ProcessTurn2 handles the complete Turn2 processing
func (a *AdapterTurn2) ProcessTurn2(ctx context.Context, systemPrompt, turn2Prompt, base64Image string, turn1Response *schema.Turn1ProcessedResponse) (*schema.BedrockResponse, error) {
	// Apply operational timeout using Processing config
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(a.cfg.Processing.BedrockCallTimeoutSec)*time.Second)
	defer cancel()

	// Invoke Bedrock with conversation history
	response, err := a.ConverseWithHistory(timeoutCtx, systemPrompt, turn2Prompt, base64Image, turn1Response)
	if err != nil {
		return nil, err
	}

	// Enrich response with metadata using available config
	response.ModelConfig = &schema.ModelConfig{
		ModelId:     a.cfg.AWS.BedrockModel,
		Temperature: 0.7, // Default temperature
		TopP:        0.9, // Default TopP
		MaxTokens:   a.cfg.Processing.MaxTokens,
	}

	return response, nil
}
