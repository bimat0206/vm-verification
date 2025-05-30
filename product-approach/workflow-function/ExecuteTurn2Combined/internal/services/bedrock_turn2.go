package services

import (
	"context"

	"workflow-function/ExecuteTurn2Combined/internal/bedrock"
	"workflow-function/ExecuteTurn2Combined/internal/config"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
)

// BedrockServiceTurn2 extends BedrockService with Turn2-specific functionality
type BedrockServiceTurn2 interface {
	BedrockService

	// ConverseWithHistory handles Turn2 conversation with history from Turn1
	ConverseWithHistory(ctx context.Context, systemPrompt, turn2Prompt, base64Image, imageFormat string, turn1Response *schema.TurnResponse) (*schema.BedrockResponse, error)
	// MODIFICATION END
}

// bedrockServiceTurn2 implements BedrockServiceTurn2
type bedrockServiceTurn2 struct {
	*bedrockService
	clientTurn2 *bedrock.ClientTurn2
}

// NewBedrockServiceTurn2 creates a new BedrockServiceTurn2 instance
func NewBedrockServiceTurn2(cfg config.Config, log logger.Logger) (BedrockServiceTurn2, error) {
	// Create base service
	baseService, err := NewBedrockService(context.Background(), cfg)
	if err != nil {
		return nil, err
	}

	// Create Turn2 client
	clientTurn2, err := bedrock.NewClientTurn2(cfg, log)
	if err != nil {
		return nil, err
	}

	return &bedrockServiceTurn2{
		bedrockService: baseService.(*bedrockService),
		clientTurn2:    clientTurn2,
	}, nil
}

// ConverseWithHistory handles Turn2 conversation with history from Turn1
func (s *bedrockServiceTurn2) ConverseWithHistory(ctx context.Context, systemPrompt, turn2Prompt, base64Image, imageFormat string, turn1Response *schema.TurnResponse) (*schema.BedrockResponse, error) {
	return s.clientTurn2.ProcessTurn2(ctx, systemPrompt, turn2Prompt, base64Image, imageFormat, turn1Response)
}
