// internal/services/bedrock.go
package services

import (
    "context"

    "ExecuteTurn1Combined/internal/config"
    "ExecuteTurn1Combined/internal/models"
    shared "workflow-function/shared/bedrock"
    "ExecuteTurn1Combined/internal/utils"
)

// BedrockService defines the Converse API integration.
type BedrockService interface {
    // Converse sends the system prompt, turn prompt, and Base64 image
    // to the Bedrock Converse endpoint and returns both raw and processed responses.
    Converse(ctx context.Context, systemPrompt, turnPrompt, base64Image string) (*models.BedrockResponse, error)
}

type bedrockService struct {
    client *shared.Client
}

// NewBedrockService constructs a BedrockService using the provided configuration.
func NewBedrockService(cfg config.Config) BedrockService {
    client := shared.NewClient(
        cfg.AWSRegion,
        cfg.BEDROCK_MODEL,
        cfg.ANTHROPIC_VERSION,
    )
    return &bedrockService{client: client}
}

// Converse performs the multimodal Converse call to Bedrock.
func (s *bedrockService) Converse(ctx context.Context, systemPrompt, turnPrompt, base64Image string) (*models.BedrockResponse, error) {
    // build the request payload
    req := shared.NewConverseRequest(systemPrompt, turnPrompt, base64Image).
        WithInferenceConfig(shared.InferenceConfig{
            MaxTokens:   cfg.MAX_TOKENS,
            Thinking:    cfg.THINKING_TYPE == "enable",
            BudgetTokens: cfg.BUDGET_TOKENS,
        })

    // invoke Bedrock with built-in retry/backoff
    resp, err := s.client.Invoke(ctx, req)
    if err != nil {
        return nil, errors.WrapRetryable(err, errors.StageBedrockCall, "failed to invoke Bedrock Converse API")
    }

    // map shared response into our domain model
    return &models.BedrockResponse{
        Raw:        resp.Raw,               // raw JSON payload
        Processed:  resp.Processed,         // structured analysis object
        TokenUsage: resp.TokenUsage(),      // counts of input/output/thinking/total
        RequestID:  resp.RequestID,         // Bedrock request identifier
    }, nil
}
