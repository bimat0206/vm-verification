package services

import (
	"context"
	"time"

	sharedbedrock "workflow-function/shared/bedrock"
	"workflow-function/shared/errors"

	"workflow-function/ExecuteTurn2Combined/internal/config"
)

// BedrockService provides access to the Bedrock Converse API for Turn 2.
type BedrockService interface {
	ConverseTurn2(ctx context.Context, systemPrompt string, prevTurn *sharedbedrock.Turn1Response, prompt string, imageData string) (*sharedbedrock.Turn2Response, int64, error)
}

// bedrockService implements BedrockService using the shared client.
type bedrockService struct {
	client    *sharedbedrock.BedrockClient
	processor *sharedbedrock.ResponseProcessor
	cfg       config.Config
}

// NewBedrockService initializes a bedrock service.
func NewBedrockService(ctx context.Context, cfg config.Config) (BedrockService, error) {
	clientCfg := &sharedbedrock.ClientConfig{
		Region:           cfg.AWS.Region,
		AnthropicVersion: cfg.AWS.AnthropicVersion,
		MaxTokens:        cfg.Processing.MaxTokens,
		Temperature:      0.7,
	}
	client, err := sharedbedrock.NewBedrockClient(ctx, cfg.AWS.BedrockModel, clientCfg)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrorTypeBedrock, "bedrock client init", false)
	}
	return &bedrockService{
		client:    client,
		processor: sharedbedrock.NewResponseProcessor(),
		cfg:       cfg,
	}, nil
}

func (b *bedrockService) ConverseTurn2(ctx context.Context, systemPrompt string, prevTurn *sharedbedrock.Turn1Response, prompt string, imageData string) (*sharedbedrock.Turn2Response, int64, error) {
	img := sharedbedrock.CreateImageContentFromBytes("jpeg", imageData)
	req := sharedbedrock.CreateConverseRequestForTurn2WithImages(
		b.cfg.AWS.BedrockModel,
		prevTurn,
		prompt,
		[]sharedbedrock.ContentBlock{img},
		systemPrompt,
		b.cfg.Processing.MaxTokens,
		nil,
		nil,
	)
	resp, latency, err := b.client.Converse(ctx, req)
	if err != nil {
		return nil, latency, errors.WrapError(err, errors.ErrorTypeBedrock, "bedrock invoke", true)
	}
	turn2, err := b.processor.ProcessTurn2Response(resp, prompt, latency, time.Now(), prevTurn)
	if err != nil {
		return nil, latency, errors.WrapError(err, errors.ErrorTypeInternal, "response parse", false)
	}
	return turn2, latency, nil
}
