package services

import (
	"context"
	"encoding/json"

	"ExecuteTurn1Combined/internal/config"
	"ExecuteTurn1Combined/internal/errors"
	"ExecuteTurn1Combined/internal/models"
	"workflow-function/shared/bedrock"
)

// BedrockService defines the Converse API integration.
type BedrockService interface {
	Converse(ctx context.Context, systemPrompt, turnPrompt, base64Image string) (*models.BedrockResponse, error)
}

type bedrockService struct {
	client    *bedrock.BedrockClient
	modelID   string
	maxTokens int
}

// NewBedrockService constructs a BedrockService using the provided configuration.
func NewBedrockService(ctx context.Context, cfg config.Config) (BedrockService, error) {
	clientCfg := bedrock.CreateClientConfig(
		cfg.AWS.Region,
		cfg.AWS.AnthropicVersion,
		cfg.Processing.MaxTokens,
		cfg.Processing.ThinkingType,
		cfg.Processing.BudgetTokens,
	)
	c, err := bedrock.NewBedrockClient(ctx, cfg.AWS.BedrockModel, clientCfg)
	if err != nil {
		return nil, err
	}
	return &bedrockService{
		client:    c,
		modelID:   cfg.AWS.BedrockModel,
		maxTokens: cfg.Processing.MaxTokens,
	}, nil
}

// Converse performs the multimodal Converse call to Bedrock.
func (s *bedrockService) Converse(ctx context.Context, systemPrompt, turnPrompt, base64Image string) (*models.BedrockResponse, error) {
	img := bedrock.CreateImageContentFromBytes("jpeg", base64Image)
	userMsg := bedrock.CreateUserMessageWithContent(turnPrompt, []bedrock.ContentBlock{img})
	req := bedrock.CreateConverseRequest(s.modelID, []bedrock.MessageWrapper{userMsg}, systemPrompt, s.maxTokens, nil, nil)

	resp, _, err := s.client.Converse(ctx, req)
	if err != nil {
		return nil, errors.WrapRetryable(err, errors.StageBedrockCall, "failed to invoke Bedrock Converse API")
	}

	raw, _ := json.Marshal(resp)
	usage := models.TokenUsage{
		InputTokens:    resp.Usage.InputTokens,
		OutputTokens:   resp.Usage.OutputTokens,
		ThinkingTokens: 0,
		TotalTokens:    resp.Usage.TotalTokens,
	}

	return &models.BedrockResponse{
		Raw:        raw,
		Processed:  resp,
		TokenUsage: usage,
		RequestID:  resp.RequestID,
	}, nil
}
