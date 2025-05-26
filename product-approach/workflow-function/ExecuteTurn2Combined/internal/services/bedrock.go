package services

import "context"

// BedrockService abstracts LLM invocation

type BedrockService interface {
    Invoke(ctx context.Context, prompt string, imageData string) (interface{}, error)
}

func NewBedrockService(log interface{}) BedrockService {
    return &noopBedrock{}
}

type noopBedrock struct{}

func (n *noopBedrock) Invoke(ctx context.Context, prompt string, imageData string) (interface{}, error) {
    return map[string]interface{}{"message": "ok"}, nil
}
